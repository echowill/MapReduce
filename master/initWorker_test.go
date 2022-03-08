package master

import (
	rpc "MapReduce/common/proto"
	"google.golang.org/grpc"
	"net"
	"testing"
)

func TestMaster(t *testing.T) {
	listener, err := net.Listen("tcp", "10086")
	if err != nil {
		t.Errorf("create listener error,code is %s \n", err)
		return
	}
	ms := NewMaster(10, 2)

	baseServer := grpc.NewServer()
	rpc.RegisterWorkerServer(baseServer, ms)
	go baseServer.Serve(listener)
	ms.(*Master).WaitForEnoughWorker()
}
