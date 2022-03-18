package master

import (
	"MapReduce/common"
	"fmt"
	"time"
)

func (ms *Master) RemakeWorkerState(defaultTimeS int64) {
	fmt.Println("[master] remake start,", time.Now().String())
	for {

		ms.Mux.Lock()
		for i, it := range ms.ReduceWorker {
			it.WorkerState = common.WORKER_UNKNOWN
			ms.ReduceWorker[i] = it
		}

		for i, it := range ms.MapWorker {
			it.WorkerState = common.WORKER_UNKNOWN
			ms.MapWorker[i] = it
		}
		ms.Mux.Unlock()
		fmt.Println("[master] remake worker state,", time.Now().String())
		time.Sleep(time.Duration(defaultTimeS * 1000 * 1000 * 1000))
	}
}
