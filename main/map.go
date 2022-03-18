package main

import (
	. "MapReduce/worker"
	"sync"
	"time"
)

// TODO : start map worker
func main() {
	var wg sync.WaitGroup
	wr := NewWorker("map")
	wr.WorkerRegister()
	go func() {
		wg.Add(1)
		for {
			wr.Health()
			time.Sleep(10 * time.Second)
		}
		wg.Done()
	}()
	go func() {
		wg.Add(1)
		wr.StartWork()
		wg.Done()
	}()
	time.Sleep(1 * time.Second)
	wg.Wait()
}
