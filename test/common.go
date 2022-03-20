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
	ms := NewMaster(1, 1)

	baseServer := grpc.NewServer()

	rpc.RegisterMasterServer(baseServer, ms)

	fmt.Println("[master] INFO : register master server successful")

	go baseServer.Serve(listener)

	// 此处由于没有运维程序，补上 bucket 获取地址
	task := rpc.MREmpty{}
	task.RpcRes = append(task.RpcRes, "map")
	//ms.RawBuckets.PushBack()
	ms.RawBuckets.PushBack(task)
	go func() {
		wg.Add(1)
		ms.RemakeWorkerState(50) // 每隔 10 秒重写一次 worker 状态为 unknown
		wg.Done()
	}()

	ms.WaitForEnoughWorker()

	fmt.Println("[master] INFO :  there have plenty worker")

	//go func() {
	//	wg.Add(1)
	go ms.AssignWorks(5000) // 5s  一次
	//	wg.Done()
	//}()
	//go func() {
	//	wg.Add(1)
	// for { // FIXME : 暂时只获取一次
	fmt.Println("[master] INFO : start get map task list")
	go ms.GetMapTaskList()
	// time.Sleep(100 * time.Second) //TO DO : 间隔100s主动获取一次任务
	// }
	//	wg.Done()
	//}()
	go func() {
		wg.Add(1)
		for {
			time.Sleep(20 * time.Second) // 每隔20秒发一次结果
			ms.Mux.Lock()
			for ip, res := range ms.SendToApp {
				// ToAppMRResult(ip, res) 暂时print
				fmt.Printf("ip is %s,res is %s \n", ip, res)
			}
			ms.SendToApp = make(map[string][]string) // 情况整个map
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

func startWorker(workerType string) {
	if workerType == "map" {
		startMap()
	} else {
		startReduce()
	}
}
