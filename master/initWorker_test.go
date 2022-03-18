package master

import (
	rpc "MapReduce/common/proto"
	"google.golang.org/grpc"
	"net"
	"sync"
	"testing"
	"time"
)

func TestMaster(t *testing.T) {
	var wg sync.WaitGroup
	listener, err := net.Listen("tcp", "127.0.0.1:10086")
	if err != nil {
		t.Errorf("create listener error,code is %s \n", err)
		return
	}
	ms := NewMaster(10, 2)

	baseServer := grpc.NewServer()
	rpc.RegisterMasterServer(baseServer, ms)
	go func() {
		wg.Add(1)
		baseServer.Serve(listener)
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
	go func() {
		wg.Add(1)
		ms.RemakeWorkerState(100) // 每隔100秒重写一次worker状态为unknown
		wg.Done()
	}()
	wg.Wait()
}
