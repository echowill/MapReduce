package worker

import (
	"sync"
	"testing"
)

/*
	TO DO:1. 新建worker 			√
		  2. 启动WorkerServer	√
		  3. 向Master注册自身 	√
		  4. 运行任务处理函数 		√
		  5. 健康监测 			√
*/
func TestWorker_Reduce(t *testing.T) {
	var wg sync.WaitGroup
	wr := NewWorker("reduce")
	wr.WorkerRegister()
	go func() {
		wg.Add(1)
		wr.Health()
		wg.Done()
	}()
	go func() {
		wg.Add(1)
		wr.StartWork()
		wg.Done()
	}()
	wg.Wait()
}
