package test

import (
	"sync"
	"testing"
	"time"
)

func TestDataStream(t *testing.T) {
	var wg sync.WaitGroup
	go func() {
		wg.Add(1)
		startMaster()
		wg.Done()
	}()
	go func() {
		wg.Add(1)
		startWorker("map")
		wg.Done()
	}()
	go func() {
		wg.Add(1)
		startWorker("reduce")
		wg.Done()
	}()
	time.Sleep(1 * time.Second)
	wg.Wait()
}
