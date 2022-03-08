package master

import (
	"MapReduce/common"
	rpc "MapReduce/common/proto"
	"container/list"
	"fmt"
	"github.com/google/uuid"
	"sync"
	"time"
)

type Master struct {
	MapWorker    map[string]common.WorkerInfo
	ReduceWorker map[string]common.WorkerInfo

	WorkQueue map[string]common.WorkerInfo

	MapTasks     list.List
	ReduceTasks  list.List
	WorkerNums   common.NumControl
	enoughWorker chan bool
	crashChan    chan bool
	mux          sync.Mutex
	appIp        string
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

func (ms *Master) AssignWork(idleSleepTimeMs int) {
	for {
		// 如果工人全都空闲或者任务队列全为空，着睡眠指定时间
		if len(ms.WorkQueue) == 0 || (ms.ReduceTasks.Len() == 0 && ms.MapTasks.Len() == 0) {
			time.Sleep(time.Duration(idleSleepTimeMs * 1000))
		}
		var wg = sync.WaitGroup{}
		for workerType, workerInfo := range ms.WorkQueue {
			var task rpc.TaskInfo
			if workerType[0] == 'R' { // 如果是reduce类型的worker

				ms.mux.Lock()
				if ms.ReduceTasks.Len() == 0 { //如果此时taskInfo队列空闲
					ms.mux.Unlock() //先解锁，在跳出本次循环
					continue
				}
				task = ms.ReduceTasks.Front().Value.(rpc.TaskInfo)
				ms.ReduceTasks.Remove(ms.ReduceTasks.Front())
				ms.mux.Unlock()

				go func() {
					wg.Add(1)
					rpcRes, err := toWorkerReduce(workerInfo.IP, &task)
					if err != nil {
						fmt.Println(rpcRes, err) //TODO : 暂时先打印
					}
					wg.Done()
				}()

			} else {
				var task rpc.TaskInfo
				ms.mux.Lock()
				if ms.MapTasks.Len() == 0 {
					ms.mux.Unlock()
					continue
				}
				task = ms.MapTasks.Front().Value.(rpc.TaskInfo)
				ms.MapTasks.Remove(ms.MapTasks.Front())
				ms.mux.Unlock()

				go func() {
					wg.Add(1)
					rpcRes, err := toWorkerMap(workerInfo.IP, &task)
					if err != nil {
						fmt.Println(rpcRes, err)
					}
					wg.Done()
				}()

			}

		}
		wg.Wait()
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

// TODO : SimpleStorageService get object list, IP = ip:port/bucket
func getTaskList(ip string) (res []string, err error) {

	return res, err
}

func checkWorkerIsIdle(info common.WorkerInfo) bool {
	if info.WorkerState == common.WORKER_IDLE {
		return true
	}
	return false
}

func resultToTaskInfo(result *rpc.Result) rpc.TaskInfo {
	return rpc.TaskInfo{
		Address: result.Address,
		Uuid:    uuid.New().String(), //每个任务都有自己的uuid
		//TODO : Task 任务类型暂未跟新
	}
}
