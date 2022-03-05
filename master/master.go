package master

import (
	"MapReduce/common"
	rpc "MapReduce/common/proto"
	"container/list"
	"sync"
	"time"
)

type Master struct {
	MapWorker    map[string]common.WorkerInfo
	ReduceWorker map[string]common.WorkerInfo
	MapTasks     list.List
	ReduceTasks  list.List
	WorkerNums   common.NumControl
	enoughWorker chan bool
	crashChan    chan bool
	mux          sync.Mutex
	rpc.UnimplementedMasterServer
}

func (ms *Master) WaitForEnoughWorker() {
	for {
		ms.mux.Lock()
		if ms.WorkerNums.TotalMapNums > ms.WorkerNums.MapMinNums &&
			ms.WorkerNums.TotalReduceNums > ms.WorkerNums.ReduceMinNums {
			break
		}
		//fmt.Printf("[master] time is %s,map num is %s,reduce num is %s", time.Now().String(), ms.WorkerNums.TotalMapNums, ms.WorkerNums.ReduceMinNums)
		time.Sleep(1000)
		ms.mux.Unlock()
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

// TODO : s3 get object list, IP = ip:port/bucket
func getTaskList(ip string) (res []string, err error) {

	return res, err
}
