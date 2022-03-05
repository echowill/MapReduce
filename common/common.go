package common

const (
	WORKER_IDLE int = iota
	WORKER_BUSY
	WORKER_UNKNOWN
)

type WorkerInfo struct {
	UUID        string // 由worker生成
	IP          string // 由worker生成
	WorkerState int    // 对应 WORKER_IDLE WORKER_BUSY WORKER_UNKNOWN
}

type NumControl struct {
	MapMinNums      int
	ReduceMinNums   int
	TotalMapNums    int
	TotalReduceNums int
}
