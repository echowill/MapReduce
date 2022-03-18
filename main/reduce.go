package main

import (
	. "MapReduce/worker"
	"sync"
)

func main() {
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
