package worker

import (
	"MapReduce/SimpleStorageService"
	"MapReduce/common"
	rpc "MapReduce/common/proto"
	"container/list"
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"io"
	"io/ioutil"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

/*
TODO : 一个worker的启动流程大致分为如下步骤
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
	Port          string `json:"Port"`
	mux           sync.Mutex
	tasks         list.List
	dataCache     []string
	deliverList   list.List
	bucketName    string `json:"Bucket"`
	rpc.UnimplementedWorkerServer
}

func (wr *Worker) Map(ctx context.Context, in *rpc.TaskInfo) (*rpc.WResult, error) {
	task := *in
	res := &rpc.WResult{}
	wr.mux.Lock()
	old := wr.tasks.Len()
	wr.tasks.PushBack(task)
	now := wr.tasks.Len()
	if now-old == 1 {
		res.RpcRes = "successful"
	} else {
		res.RpcRes = "failed"
	}
	wr.mux.Unlock()
	return res, nil
}

func (wr *Worker) Reduce(ctx context.Context, in *rpc.TaskInfo) (*rpc.WResult, error) {
	task := *in
	res := &rpc.WResult{}
	wr.mux.Lock()
	old := wr.tasks.Len()
	wr.tasks.PushBack(task)
	now := wr.tasks.Len()
	if now-old == 1 {
		res.RpcRes = "successful"
	} else {
		res.RpcRes = "failed"
	}
	wr.mux.Unlock()
	return res, nil
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
	// 已解决 TO DO : 要先启动本机的WorkerServer
	if wr.StartWorkerServer() == false {
		fmt.Println("start worker server error")
		return false
	}
	workInfo := &rpc.WorkerInfo{
		Uuid: wr.UUID,
		Port: wr.Port,
	}
	if wr.Type == "map" { // 已确认，master大小写无关,worker json为map(小写) TODO:确认map是全小写还是Map
		workInfo.IsMap = true
	} else {
		workInfo.IsMap = false
	}

	fmt.Printf("[%s worker] INFO: worker server start successful \n", wr.Type)

	r, err := c.WorkerRegister(ctx, workInfo)

	if err != nil {
		logrus.Error(err)
		return false
	} else {
		fmt.Printf("[%s worker] INFO: worker register successful \n", wr.Type)
	}

	wr.ID = r.Id
	return r.Result
}

func (wr *Worker) StartWorkerServer() bool {
	l, err := net.Listen("tcp", ":0") // port为0时系统会自动分配
	if err != nil {
		fmt.Println("err")
		return false
	}

	wr.Port = strconv.Itoa(l.Addr().(*net.TCPAddr).Port) // 获取到worker服务器到port，ip由master client发送

	fmt.Println("[worker] INFO : local worker server Port is ", wr.Port)

	baseServer := grpc.NewServer()
	rpc.RegisterWorkerServer(baseServer, wr) // 启动worker server

	go baseServer.Serve(l)

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
	if err != nil {
		return false
	}
	if r.RpcRes == "WORKER_UNKNOWN" { //如果发现状态被标记成unknown，再次发送健康请求
		fmt.Println("[worker] Rpc.Res is", r.RpcRes)
		return wr.Health()
	}
	fmt.Println("[worker] Rpc.Res is", r.RpcRes)
	return true
}

func (wr *Worker) StartWork() {
	if wr.Type == "map" {
		for {
			if wr.tasks.Len() == 0 {
				time.Sleep(10 * time.Second)
				continue
			}
			// TODO : 取出task，doMap
			task := wr.tasks.Front().Value.(rpc.TaskInfo)
			// Map http://192.168.1.152/bucket/object
			_, bucket, object := parseS3Url(task.Address)
			cfg := SimpleStorageService.GetDefaultS3Config()
			reader, err := SimpleStorageService.DownloadObjectToRAM(bucket, object, cfg.GetDefaultS3Session())
			if err != nil {
				// TODO : 下载内容失败
			}
			content := readerToStrings(reader)
			if wr.doMap(task.Bucket, task.Object, *content) == false {
				fmt.Println("[worker] do map error,time is : ", time.Now().String())
			}
		}
	} else {
		for {
			if wr.tasks.Len() == 0 {
				time.Sleep(10 * time.Second)
				continue
			}
			task := wr.tasks.Front().Value.(rpc.TaskInfo)
			_, bucket, object := parseS3Url(task.Address)
			cfg := SimpleStorageService.GetDefaultS3Config()
			reader, err := SimpleStorageService.DownloadObjectToRAM(bucket, object, cfg.GetDefaultS3Session())
			if err != nil {
				// TODO : 下载内容失败
			}
			content := readerToStrings(reader)
			wr.doReduce(*content)
		}

	}

}

// TODO : parse url http://192.168.1.152/bucket/object --> http://192.168.1.152 bucket object
func parseS3Url(url string) (ip, bucket, object string) {
	strs := strings.Split(url, "/")
	if len(strs) < 2 {
		return "", "", ""
	}
	bucket = strs[len(strs)-2]
	object = strs[len(strs)-1]
	for i, it := range url {
		if it == '/' && i < len(url)-1 && url[i+1] != '/' {
			break
		}
		ip += string(byte(it))
	}
	return ip, bucket, object
}

func readerToStrings(in *io.ReadCloser) *[]string {
	var buffer []byte
	(*in).Read(buffer)
	(*in).Close()
	str := strings.Split(string(buffer), "\n")
	return &str
}
