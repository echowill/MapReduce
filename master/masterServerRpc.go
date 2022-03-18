package master

import (
	"MapReduce/common"
	rpc "MapReduce/common/proto"
	"context"
	"fmt"
	"github.com/google/uuid"
	"strconv"
)

// WorkerRegister : register worker is map worker
func (ms *Master) WorkerRegister(ctx context.Context, in *rpc.WorkerInfo) (*rpc.RegisterResult, error) {
	fmt.Println("[Master] worker register start")
	var num int
	ms.Mux.Lock()
	if in.IsMap == true { // true is map worker
		fmt.Printf("[master] INFO : worker uuid is %s,ip is %s \n", in.Uuid, in.Ip)

		ms.MapWorker[in.Uuid] = NewWorker(in.Uuid, in.Ip)
		ms.WorkerNums.TotalMapNums++
		num = ms.WorkerNums.TotalMapNums
		fmt.Printf("[master] INFO : worker %s add successful\n", in.Uuid)
	} else {
		ms.ReduceWorker[in.Uuid] = NewWorker(in.Uuid, in.Ip)
		ms.WorkerNums.TotalReduceNums++
		num = ms.WorkerNums.TotalReduceNums + 0x80000000
	}
	ms.Mux.Unlock()
	fmt.Println("[Master] worker register success")
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
		fmt.Println("get object list error,code is ", tErr)
		empty.RpcRes = "get object list error"
		return empty, tErr
	}

	ms.Mux.Lock()
	for _, it := range tasks {
		task := rpc.TaskInfo{
			Uuid:    uuid.New().String(),
			Task:    in.DataRange,
			Address: in.Address + "/" + it,
		}
		ms.MapTasks.PushBack(task)
	}
	ms.Mux.Unlock()

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

	task := resultToTaskInfo(in)

	ms.Mux.Lock()

	ms.ReduceTasks.PushBack(task) // pushback的一定不是指针
	for _, it := range ms.ReduceWorker {
		if checkWorkerIsIdle(it) == true { // if reduce worker is idle
			it.WorkerState = common.WORKER_BUSY // 领到新到任务，状态转为busy
			ms.WorkQueue["R"+it.UUID] = it      // 由reduce 队列转到worker队列
		}
	}

	ms.Mux.Unlock()

	return empty, nil
}

/*
	TODO : 需要补充reduce节点调用reduce函数的内容
*/
func (ms *Master) Reduce(ctx context.Context, in *rpc.ReduceResult) (*rpc.Empty, error) {
	empty := &rpc.Empty{
		RpcRes: string("successful"),
	}
	var out []string
	out = append(out, "min : "+in.MinValue)
	out = append(out, "max : "+in.MaxValue)
	out = append(out, "start : "+in.StartTime)
	out = append(out, "end : "+in.EndTime)
	out = append(out, "variance : "+strconv.FormatFloat(float64(in.Variance), 'f', 6, 64))
	ToAppMRResult(ms.AppIp, out) // 更新到app,此时reduce会等着master将结果写入app才会返回
	// 暂时不更新此 TODO : 结果更新到s3 指定bucket,
	return empty, nil
}

/*
	TODO : 待测试
*/
func (ms *Master) Health(ctx context.Context, in *rpc.WorkerState) (*rpc.Empty, error) {
	empty := &rpc.Empty{
		RpcRes: string("successful"),
	}

	ms.Mux.Lock()
	if in.IsMap == true {
		value, ok := ms.MapWorker[in.Uuid]
		if ok {
			if value.WorkerState == common.WORKER_UNKNOWN {
				empty.RpcRes = "WORKER_UNKNOWN"
			}
			value.WorkerState = int(in.State)
			ms.MapWorker[in.Uuid] = value
		}
	} else {
		value, ok := ms.ReduceWorker[in.Uuid]
		if ok {
			if value.WorkerState == common.WORKER_UNKNOWN {
				empty.RpcRes = "WORKER_UNKNOWN"
			}
			value.WorkerState = int(in.State)
			ms.ReduceWorker[in.Uuid] = value
		}
	}
	ms.Mux.Unlock()
	fmt.Println("[master] Rpc.Res is ", empty.RpcRes)
	return empty, nil
}
