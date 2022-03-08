package master

import (
	rpc "MapReduce/common/proto"
	"context"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// WorkerRegister : register worker is map worker
func (ms *Master) WorkerRegister(ctx context.Context, in *rpc.WorkerInfo) (*rpc.RegisterResult, error) {
	var num int
	ms.mux.Lock()
	if in.IsMap == true { // true is map worker
		ms.MapWorker[in.Uuid] = NewWorker(in.Uuid, in.Ip)
		ms.WorkerNums.TotalMapNums++
		num = ms.WorkerNums.TotalMapNums
	} else {
		ms.ReduceWorker[in.Uuid] = NewWorker(in.Uuid, in.Ip)
		ms.WorkerNums.TotalReduceNums++
		num = ms.WorkerNums.TotalReduceNums + 0x80000000
	}
	ms.mux.Unlock()
	logrus.Info("[Master] worker register success")
	return &rpc.RegisterResult{
		Result: true,
		Id:     int64(num - 1),
	}, nil
}

/*
	TODO : 待测试
*/
func (ms *Master) APP(ctx context.Context, in *rpc.DataAddress) (*rpc.Empty, error) {
	empty := &rpc.Empty{
		RpcRes: string("successful"),
	}
	tasks, tErr := getTaskList(in.Address)
	if tErr != nil {
		logrus.Error("get object list error,code is %s", tErr)
		empty.RpcRes = "get object list error"
		return empty, tErr
	}

	ms.mux.Lock()
	for _, it := range tasks {
		task := rpc.TaskInfo{
			Uuid:    uuid.New().String(),
			Task:    in.DataRange,
			Address: in.Address + "/" + it,
		}
		ms.MapTasks.PushBack(task)
	}
	ms.mux.Unlock()

	return empty, nil
}

/*
	TODO : 需要补充map节点调用map函数的内容
 		   1. 此处应写入Reduce专用bucket/file,等reduce调用读取该file
*/
func (ms *Master) Map(ctx context.Context, in *rpc.Result) (*rpc.Empty, error) {
	empty := &rpc.Empty{
		RpcRes: string("successful"),
	}
	// TODO : 文件传入ceph , 记录目的地址 ip:port/bucket/fileName,写入ReduceTask
	return empty, nil
}

/*
	TODO : 需要补充reduce节点调用reduce函数的内容
*/
func (ms *Master) Reduce(ctx context.Context, in *rpc.Result) (*rpc.Empty, error) {
	empty := &rpc.Empty{
		RpcRes: string("successful"),
	}

	toAppMRResult(ms.appIp, in.Address, in.Result) // 更新到app

	// 暂时不更新此 TODO : 更新到s3 指定bucket,

	return empty, nil
}

/*
	TODO : 待测试
*/
func (ms *Master) Health(ctx context.Context, in *rpc.WorkerState) (*rpc.Empty, error) {
	empty := &rpc.Empty{
		RpcRes: string("successful"),
	}
	ms.mux.Lock()
	if in.IsMap == true {
		value, ok := ms.MapWorker[in.Uuid]
		if ok {
			value.WorkerState = int(in.State)
			ms.MapWorker[in.Uuid] = value
		}
	} else {
		value, ok := ms.ReduceWorker[in.Uuid]
		if ok {
			value.WorkerState = int(in.State)
			ms.ReduceWorker[in.Uuid] = value
		}
	}
	ms.mux.Unlock()
	return empty, nil
}