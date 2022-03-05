package master

import (
	rpc "MapReduce/common/proto"
	"context"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"time"
)

func toWorkerMap(workerIp string, task *rpc.TaskInfo) (*rpc.WEmpty, error) {
	conn, err := grpc.Dial(workerIp, grpc.WithInsecure())
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	c := rpc.NewWorkerClient(conn)
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	r, err := c.Map(ctx, task)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	return r, nil
}

func toWorkerReduce(workerIp string, task *rpc.TaskInfo) (*rpc.WEmpty, error) {
	conn, err := grpc.Dial(workerIp, grpc.WithInsecure())
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	c := rpc.NewWorkerClient(conn)
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	r, err := c.Map(ctx, task)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	return r, nil
}
