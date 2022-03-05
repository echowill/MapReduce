package master

import (
	"MapReduce/common"
	rpc "MapReduce/common/proto"
	"fmt"
	"time"
)

/*
	TaskDispatch TODO : Map任务列表非空且map worker列表存在空闲worker时，会触发map任务派发，reduce同理
				 1. 需要定期检测task列表以及worker列表		ok
  				 2. 向worker派发任务 						ok
				 3. 暂定一个worker一次派发一个任务			ok
				 4. 任务发出去后需要及时从任务列队中清除		ok
				 5. 设置worker权重,采用轮询派发的模式
*/
func (ms *Master) DispatchTasks(monitorFrequencyMs int64) {
	for {
		ms.mux.Lock()
		if ms.MapTasks.Len() != 0 {
			for _, it := range ms.MapWorker {
				if it.WorkerState == common.WORKER_IDLE {
					// 给该worker派发任务，同时将任务从task list中移除
					task := ms.MapTasks.Front().Value.(rpc.TaskInfo)
					res, _ := toWorkerMap(it.IP, &task)
					if &res == nil {
						fmt.Println("Error occurrence")
					}
					ms.MapTasks.Remove(ms.MapTasks.Front())
					break
				}
			}
		}

		if ms.ReduceTasks.Len() != 0 {
			for _, it := range ms.ReduceWorker {
				if it.WorkerState == common.WORKER_IDLE {
					// 给该worker派发任务，同时将任务从task list中移除
					task := ms.ReduceTasks.Front().Value.(rpc.TaskInfo)
					res, _ := toWorkerReduce(it.IP, &task)
					if &res == nil {
						fmt.Println("Error occurrence")
					}
					ms.ReduceTasks.Remove(ms.ReduceTasks.Front())
					break
				}
			}
		}
		ms.mux.Unlock()
		time.Sleep(time.Duration(monitorFrequencyMs * 1000))
	}
}
