package test

import (
	rpc "MapReduce/common/proto"
	. "MapReduce/master"
	wr "MapReduce/worker"
	"fmt"
	"google.golang.org/grpc"
	"net"
	"sync"
	"time"
)

func startMaster() {
	var wg sync.WaitGroup
	listener, err := net.Listen("tcp", "192.168.1.121:12345")
	if err != nil {
		fmt.Errorf("create listener error,code is %s \n", err)
		return
	}
	ms := NewMaster(10, 2)

	baseServer := grpc.NewServer()

	rpc.RegisterMasterServer(baseServer, ms)

	fmt.Println("[master] INFO : register master server successful")

	go baseServer.Serve(listener)

	go func() {
		wg.Add(1)
		ms.RemakeWorkerState(10) // 每隔10秒重写一次worker状态为unknown
		wg.Done()
	}()

	ms.WaitForEnoughWorker()

	go func() {
		wg.Add(1)
		ms.AssignWorks(5000) // 5s  一次
		wg.Done()
	}()
	go func() {
		wg.Add(1)
		for {
			ms.GetMapTaskList()
			time.Sleep(100 * time.Second) //TODO : 间隔100s主动获取一次任务
		}
		wg.Done()
	}()
	go func() {
		wg.Add(1)
		for {
			time.Sleep(20 * time.Second) // 每隔20秒发一次结果
			ms.Mux.Lock()
			for ip, res := range ms.SendToApp {
				ToAppMRResult(ip, res)
			}
			ms.Mux.Unlock()
		}
		wg.Done()
	}()
	wg.Wait()
}

func startMap() {
	var wg sync.WaitGroup
	wr := wr.NewWorker("map")
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

func startReduce() {
	var wg sync.WaitGroup
	wr := wr.NewWorker("reduce")
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
