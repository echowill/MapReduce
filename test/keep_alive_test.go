package test

import (
	"sync"
	"testing"
	"time"
)

func TestKeepAlive(t *testing.T) {
	var wg sync.WaitGroup
	go func() {
		wg.Add(1)
		startMaster()
		wg.Done()
	}()
	go func() {
		wg.Add(1)
		startReduce()
		wg.Done()
	}()
	go func() {
		wg.Add(1)
		startMap()
		wg.Done()
	}()
	time.Sleep(1 * time.Second)
	wg.Wait()
}
