package master

import (
	"MapReduce/common"
	"time"
)

/*
	TODO : 待测试
*/
func (ms *Master) remakeWorkerState(defaultTimeS int64) {
	for {
		ms.mux.Lock()
		for _, it := range ms.ReduceWorker {
			it.WorkerState = common.WORKER_UNKNOWN
		}
		for _, it := range ms.MapWorker {
			it.WorkerState = common.WORKER_UNKNOWN
		}
		ms.mux.Unlock()
		time.Sleep(time.Duration(defaultTimeS * 1000 * 1000))
	}
}
