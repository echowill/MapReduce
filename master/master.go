package master

import (
	"MapReduce/common"
	rpc "MapReduce/common/proto"
	"context"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

type Master struct {
	MapWorker    []common.WorkerInfo
	ReduceWorker []common.WorkerInfo
	MapTasks     []common.TaskInfo
	ReduceTasks  []common.TaskInfo
	WorkerNums   common.NumControl
	enoughWorker chan bool
	crashChan    chan bool
	mux          sync.Mutex
	rpc.UnimplementedWorkerServer
}

// NewMaster : init a master node
func NewMaster(nMap int, nReduce int) rpc.WorkerServer {
	return &Master{
		WorkerNums: common.NumControl{
			ReduceMinNums: nReduce,
			MapMinNums:    nMap,
		},
		enoughWorker: make(chan bool, 1),
		crashChan:    make(chan bool, 100),
	}
}

// Normal function

func (ms *Master) WaitForEnoughWorker() {
	for {
		ms.mux.Lock()
		if ms.WorkerNums.TotalMapNums > ms.WorkerNums.MapMinNums &&
			ms.WorkerNums.TotalReduceNums > ms.WorkerNums.ReduceMinNums {
			break
		}
		logrus.Info("[master] time is %s,map num is %s,reduce num is %s", time.Now().String(), ms.WorkerNums.TotalMapNums, ms.WorkerNums.ReduceMinNums)
		time.Sleep(100)
		ms.mux.Unlock()
	}
}

// newTaskInfo : append task to task queue
func newTaskInfo(IsMap bool, data0, data1 []string) common.TaskInfo {
	if IsMap == false {
		return common.TaskInfo{
			IsMap: IsMap,
			Map: rpc.MapInfo{
				Uuid:      uuid.New().String(),
				DataName:  data0,
				DataRange: data1,
			},
		}
	}
	return common.TaskInfo{
		IsMap: IsMap,
		Reduce: rpc.ReduceInfo{
			Uuid:          uuid.New().String(),
			S3FileAddress: data0,
			ToMapAddress:  data1,
		},
	}
}

// NewWorker : new a worker
func NewWorker(uuid string, ip string) common.WorkerInfo {
	return common.WorkerInfo{
		UUID:        uuid,
		IP:          ip,
		WorkerState: common.WORKER_IDLE,
	}
}

// WorkerRegister : register worker is map worker
func (ms *Master) WorkerRegister(ctx context.Context, in *rpc.WorkerInfo) (*rpc.RegisterResult, error) {
	var num int
	ms.mux.Lock()
	if in.IsMap == false { // false is map worker
		ms.MapWorker = append(ms.MapWorker, NewWorker(in.Uuid, in.Ip))
		ms.WorkerNums.TotalMapNums++
		num = ms.WorkerNums.TotalMapNums
	} else {
		ms.ReduceWorker = append(ms.ReduceWorker, NewWorker(in.Uuid, in.Ip))
		ms.WorkerNums.TotalReduceNums++
		num = ms.WorkerNums.TotalReduceNums + 0x80000000
	}
	ms.mux.Unlock()
	logrus.Info("[Master] worker register success")
	return &rpc.RegisterResult{
		Result: true,
		Id:     int64(num - 1),
	}, nil
}
