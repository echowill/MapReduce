package worker

import (
	"MapReduce/common"
	rpc "MapReduce/common/proto"
	"container/list"
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"io/ioutil"
	"net"
	"strconv"
	"sync"
	"time"
)

/* TODO : 一个worker的启动流程大致分为如下步骤
1. 读取worker.json获取默认配置,里面至少包含master的默认ip
2. 启动workerServer，由于一台机器上可能有多个workerServer，因此port需要提起协商
3. 启动MasterClient，向MasterServer注册，此时注册的port为workerServer的port，
   因为Master使用workerClient向worker同步信息
3.1定时同步Health状态,taskList小于阈值时，标记worker为idle，反之标记为busy，
   master会定期将所以worker状态改为unknown，unknown再次被标记为unknown会被移除worker队列
4. 等待Master派发任务，将任务丢进任务队列，及时消费，推送结果到s3，与master同步结果地址(bucket/file)
*/

type Worker struct {
	UUID          string
	ID            int64
	state         int32
	Type          string `json:"WorkerType"`
	TaskThreshold int    `json:"TaskThreshold"`
	MasterIp      string `json:"MasterIp"`
	WorkerIp      string `json:"WorkerIp"`
	mux           sync.Mutex
	tasks         list.List
	dataCache     []byte
	deliverList   list.List

	rpc.UnimplementedWorkerServer
}

func readWorkerJson(worker *Worker) {
	data, err := ioutil.ReadFile("./worker.json")
	if err != nil {
		fmt.Println("Read json failed,error code is ", err)
		panic(err)
	}
	err = json.Unmarshal(data, worker)
	if err != nil {
		fmt.Println("Parse Json failed,error code is ", err)
		panic(err)
	}
}

// 已解决 TODO : worker向master注册时需要将worker server的ip:port发送过去，而不是发送master client的ip:port

func NewWorker(workerType string) *Worker {
	worker := &Worker{
		UUID:  uuid.New().String(),
		state: int32(common.WORKER_IDLE),
	}
	readWorkerJson(worker)
	return worker
}

func (wr *Worker) WorkerRegister() bool {
	conn, err := grpc.Dial(wr.MasterIp, grpc.WithInsecure())
	if err != nil {
		logrus.Error(err)
		return false
	}
	c := rpc.NewMasterClient(conn)
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	// 已解决 TODO : 要先启动本机的WorkerServer
	if wr.StartWorkerServer() == false {
		fmt.Println("start worker server error")
		return false
	}
	workInfo := &rpc.WorkerInfo{
		Uuid: wr.UUID,
		Ip:   wr.WorkerIp,
	}
	if wr.Type == "map" { // 已确认，master大小写无关,worker json为map(小写) TODO:确认map是全小写还是Map
		workInfo.IsMap = true
	} else {
		workInfo.IsMap = false
	}
	r, err := c.WorkerRegister(ctx, workInfo)

	if err != nil {
		logrus.Error(err)
		return false
	}

	wr.ID = r.Id
	return r.Result
}

func (wr *Worker) StartWorkerServer() bool {
	l, err := net.Listen("tcp", wr.WorkerIp+":0") // port为0时系统会自动分配
	if err != nil {
		fmt.Println("err")
		return false
	}
	wr.WorkerIp += ":" + strconv.Itoa(l.Addr().(*net.TCPAddr).Port)
	baseServer := grpc.NewServer()
	rpc.RegisterWorkerServer(baseServer, wr) // 启动worker server
	return true
}

func (wr *Worker) Health() bool {
	conn, err := grpc.Dial(wr.MasterIp, grpc.WithInsecure())
	if err != nil {
		logrus.Error(err)
		return false
	}
	c := rpc.NewMasterClient(conn)
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	health := &rpc.WorkerState{
		Uuid:  wr.UUID,
		State: wr.state,
	}
	if wr.Type == "map" {
		health.IsMap = true
	} else {
		health.IsMap = false
	}
	r, err := c.Health(ctx, health)
	if r.RpcRes == "WORKER_UNKNOWN" { //如果发现状态被标记成unknown，再次发送健康请求
		return wr.Health()
	}
	return true
}