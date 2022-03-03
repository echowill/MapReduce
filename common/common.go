package common

import (
	rpc "MapReduce/common/proto"
	"sync"
)

const (
	WORKER_IDLE int = iota
	WORKER_BUSY
	WORKER_UNKNOWN
)

type WorkerInfo struct {
	UUID        string     // 由worker生成
	IP          string     // 由worker生成
	WorkerState int        // 对应 WORKER_IDLE WORKER_BUSY WORKER_UNKNOWN
	mux         sync.Mutex // 并发控制
}

type TaskInfo struct {
	IsMap  bool // 默认false，false为MapInfo
	Map    rpc.MapInfo
	Reduce rpc.ReduceInfo
}

type NumControl struct {
	MapMinNums      int
	ReduceMinNums   int
	TotalMapNums    int
	TotalReduceNums int
}
