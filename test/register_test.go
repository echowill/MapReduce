package test

import (
	"MapReduce/SimpleStorageService"
	"testing"
)

/*
	TODO : 待重做
*/
func TestRegister(t *testing.T) {
	//var wg sync.WaitGroup
	//go func() {
	//	wg.Add(1)
	//	listener, err := net.Listen("tcp", "127.0.0.1:7899")
	//	if err != nil {
	//		t.Errorf("create listener error,code is %s \n", err)
	//		return
	//	}
	//	ms := master.NewMaster(10, 2)
	//	baseServer := grpc.NewServer()
	//	rpc.RegisterWorkerServer(baseServer, ms)
	//	go baseServer.Serve(listener)
	//	ms.(*master.Master).WaitForEnoughWorker()
	//	wg.Done()
	//}()
	//go func() {
	//	for i := 0; i < 10; i++ {
	//		wg.Add(1)
	//		w := rpc.WorkerInfo{
	//			Uuid:  uuid.New().String(),
	//			Ip:    "127.0.0.1",
	//			IsMap: false,
	//		}
	//		conn, wErr := grpc.Dial("127.0.0.1:7899", grpc.WithInsecure())
	//		c := rpc.NewWorkerClient(conn)
	//		if wErr != nil {
	//			t.Logf("register worker error ,code is %s", wErr)
	//		}
	//		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	//		defer cancel()
	//		t.Logf("[worker] register start")
	//		r, rErr := c.WorkerRegister(ctx, &w)
	//		if rErr != nil {
	//			t.Logf("register error code is %s", rErr)
	//		}
	//		if r.Result == false {
	//			t.Logf("r.Result is false")
	//		} else {
	//			t.Logf("id is %d \n", r.Id)
	//		}
	//		t.Logf("[worker] register end")
	//		wg.Done()
	//	}
	//}()
	//time.Sleep(1000)
	//wg.Wait()
}

func TestS3Client(t *testing.T) {
	SimpleStorageService.S3test()
}
