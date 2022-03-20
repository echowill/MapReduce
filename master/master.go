package master

import (
	"MapReduce/SimpleStorageService"
	"MapReduce/common"
	rpc "MapReduce/common/proto"
	"container/list"
	"fmt"
	"github.com/google/uuid"
	"sync"
	"time"
)

type Master struct {
	MapWorker    map[string]common.WorkerInfo
	ReduceWorker map[string]common.WorkerInfo

	WorkQueue map[string]common.WorkerInfo // Deprecated

	MapTasks    list.List
	ReduceTasks list.List

	RawBuckets list.List //  bucket地址结构体

	WorkerNums   common.NumControl
	enoughWorker chan bool // Deprecated
	crashChan    chan bool // Deprecated
	Mux          sync.Mutex

	AppIp     string
	SendToApp map[string][]string // ip,res

	rpc.UnimplementedMasterServer
}

func (ms *Master) WaitForEnoughWorker() {
	for {
		ms.Mux.Lock()
		if ms.WorkerNums.TotalMapNums >= ms.WorkerNums.MapMinNums &&
			ms.WorkerNums.TotalReduceNums >= ms.WorkerNums.ReduceMinNums {
			ms.Mux.Unlock()
			break
		}
		//fmt.Printf("[master] time is %s,map num is %s,reduce num is %s", time.Now().String(), ms.WorkerNums.TotalMapNums, ms.WorkerNums.ReduceMinNums)
		ms.Mux.Unlock()
		fmt.Println("[master] INFO : wait worker ...")
		fmt.Printf("[master] INFO : now map worker num is %d,need %d\n", ms.WorkerNums.TotalMapNums, ms.WorkerNums.MapMinNums)
		fmt.Printf("[master] INFO : now red worker num is %d,need %d\n", ms.WorkerNums.TotalReduceNums, ms.WorkerNums.ReduceMinNums)
		time.Sleep(10 * time.Second)
	}
}

// AssignWorks FIXME : 此处的WorkQueue更新来源不明
func (ms *Master) AssignWorks(idleSleepTimeMs int) {
	fmt.Println("[master] INFO : assign work start")
	for {
		// 如果工人全都空闲或者任务队列全为空，则睡眠指定时间
		if ms.ReduceTasks.Len() == 0 && ms.MapTasks.Len() == 0 {
			time.Sleep(time.Duration(idleSleepTimeMs * 1000))
		}
		// TODO : 如何task失败记得把任务加回到tasks尾部
		for _, workerInfo := range ms.MapWorker {
			if workerInfo.WorkerState == common.WORKER_IDLE {
				ms.Mux.Lock()
				task := ms.MapTasks.Front().Value.(rpc.TaskInfo)
				ms.MapTasks.Remove(ms.MapTasks.Front())
				ms.Mux.Unlock()
				go func() {
					rpcRes, err := toWorkerMap(workerInfo.IP, &task)
					if err != nil {
						fmt.Println(rpcRes, err) //TODO : 暂时先打印
						ms.Mux.Lock()
						ms.MapTasks.PushBack(task) // 如果任务失败，这里会把任务丢回任务队列
						ms.Mux.Unlock()
					} else {
						fmt.Println("[master] INFO : map task result is ", rpcRes.RpcRes)
					}
				}()
			}
		}

		for _, workerInfo := range ms.ReduceWorker {
			if workerInfo.WorkerState == common.WORKER_IDLE {
				ms.Mux.Lock()
				task := ms.ReduceTasks.Front().Value.(rpc.TaskInfo)
				ms.ReduceTasks.Remove(ms.ReduceTasks.Front())
				ms.Mux.Unlock()
				go func() {
					rpcRes, err := toWorkerReduce(workerInfo.IP, &task)
					if err != nil {
						fmt.Println(rpcRes, err) //TODO : 暂时先打印
						ms.Mux.Lock()
						ms.ReduceTasks.PushBack(task) // 如果任务失败，这里会把任务丢回任务队列
						ms.Mux.Unlock()
					} else {
						fmt.Println("[master] INFO : map task result is ", rpcRes.RpcRes)
					}
				}()
			}
		}
		//for workerType, workerInfo := range ms.WorkQueue {
		//	var task rpc.TaskInfo
		//	if workerType[0] == 'R' { // 如果是reduce类型的worker
		//
		//		ms.Mux.Lock()
		//		if ms.ReduceTasks.Len() == 0 { //如果此时taskInfo队列空闲
		//			ms.Mux.Unlock() //先解锁，在跳出本次循环
		//			continue
		//		}
		//		task = ms.ReduceTasks.Front().Value.(rpc.TaskInfo)
		//		ms.ReduceTasks.Remove(ms.ReduceTasks.Front())
		//		delete(ms.WorkQueue, workerType) // 及时从workQueue中移除已经获取到task的worker
		//		ms.Mux.Unlock()
		//
		//		go func() {
		//			wg.Add(1)
		//			rpcRes, err := toWorkerReduce(workerInfo.IP, &task)
		//			if err != nil {
		//				fmt.Println(rpcRes, err) //TO DO : 暂时先打印
		//				ms.SendToApp[rpcRes.Ip] = append(ms.SendToApp[rpcRes.Ip], rpcRes.RpcRes)
		//				//ToAppMRResult("","",rpcRes.RpcRes) //TO DO : 发送到app的逻辑
		//			}
		//			wg.Done()
		//		}()
		//
		//	} else {
		//		var task rpc.TaskInfo
		//		ms.Mux.Lock()
		//		if ms.MapTasks.Len() == 0 {
		//			ms.Mux.Unlock()
		//			continue
		//		}
		//		task = ms.MapTasks.Front().Value.(rpc.TaskInfo)
		//		ms.MapTasks.Remove(ms.MapTasks.Front())
		//		delete(ms.WorkQueue, workerType)
		//		ms.Mux.Unlock()
		//
		//		go func() {
		//			wg.Add(1)
		//			rpcRes, err := toWorkerMap(workerInfo.IP, &task)
		//			if err != nil {
		//				fmt.Println(rpcRes, err)
		//			}
		//			wg.Done()
		//		}()
		//
		//	}
		//
		//}
	}
}

// GetMapTaskList TODO :  需要有一个指定的bucket或者bucket range
func (ms *Master) GetMapTaskList() {
	if ms.RawBuckets.Len() != 0 {
		task := ms.RawBuckets.Front().Value.(rpc.MREmpty)
		for _, it := range task.RpcRes {

			fmt.Println("[master] INFO : raw bucket name is", it)

			cfg := SimpleStorageService.GetDefaultS3Config()
			objs := SimpleStorageService.GetObjectList(cfg.GetDefaultS3Session(), it)

			fmt.Println("[master] DATA : map task nums ", len(objs))
			for i, it := range objs {
				if i > 3 {
					break
				}
				fmt.Printf("[master] DATA : map task %d is %s\n", i, it)
			}

			for _, obj := range objs {
				mTask := rpc.TaskInfo{
					Uuid:    uuid.New().String(),
					Address: it + "/" + obj, // TODO : 暂定bucket / file
				}
				ms.Mux.Lock()
				ms.MapTasks.PushBack(mTask)
				ms.Mux.Unlock()
			}
			// TODO : 何时删除桶内容？等master向app更新最后一个此bucket内的内容时删除
			fmt.Printf("[master] INFO : read bucket %s end", it)
		}
		ms.RawBuckets.Remove(ms.RawBuckets.Front())
	} else {
		fmt.Println("[master] INFO : raw buckets is empty !")
	}
}

func NewMaster(mapNums, reduceNums int) *Master {

	return &Master{
		WorkerNums: common.NumControl{
			MapMinNums:    mapNums,
			ReduceMinNums: reduceNums,
		},
		MapWorker:    make(map[string]common.WorkerInfo, 100),
		ReduceWorker: make(map[string]common.WorkerInfo, 100),
	}
}

// NewWorker : new a worker
func NewWorker(uuid string, ip string) common.WorkerInfo {
	return common.WorkerInfo{
		UUID:        uuid,
		IP:          ip,
		WorkerState: common.WORKER_IDLE,
	}
}

// TODO : SimpleStorageService get object list, IP = ip:port/bucket
func getTaskList(ip string) (res []string, err error) {

	return res, err
}

func checkWorkerIsIdle(info common.WorkerInfo) bool {
	if info.WorkerState == common.WORKER_IDLE {
		return true
	}
	return false
}

func resultToTaskInfo(result *rpc.Result) rpc.TaskInfo {
	return rpc.TaskInfo{
		Address: result.Address,
		Uuid:    uuid.New().String(), //每个任务都有自己的uuid
		//TODO : Task 任务类型暂未跟新
	}
}
