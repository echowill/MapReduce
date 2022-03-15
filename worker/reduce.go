package worker

import (
	rpc "MapReduce/common/proto"
	"context"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"sort"
	"strconv"
	"strings"
	"time"
)

/*
	TODO :  reduce worker工作流程
		 1. 读取配置文件，启动默认配置，向master注册自身
		 2. 等待Master推送任务
		 2.1定期更新自己的健康状态
		 3. 收到Master推送任务放入任务队列，立即返回接受成功的消息
		 4. 处理完成后向Master推送符合筛选结果的数据集
*/

// 返回一个文件的极值，均值，方差，各值的倒排索引
// time-service-value
func (wr *Worker) doReduce(content []string) {

	/* TODO : 返回一个文件的极值，均值，方差，各值的倒排索引
	1. get time start - end
	2. get max,min,avg
	3. get variance
	4. get inverted index
	5. toMasterReduce
	*/
	// 1. get time start - end
	start, end, values := splitContent(content)
	// 2. get max,min,avg
	min, max := values[0], values[len(values)-1]

	meanValue := func() float64 {
		var sum, cnt float64
		for _, it := range values {
			tmp, err := strconv.Atoi(it)
			if err != nil {
				continue
			}
			sum += float64(tmp)
			cnt++
		}
		return sum / cnt
	}()

	// 3. get variance
	variance := func() float64 {
		var sum, cnt float64
		for _, it := range values {
			tmp, err := strconv.Atoi(it)
			if err != nil {
				continue
			}
			sum += (float64(tmp) - meanValue) * (float64(tmp) - meanValue)
			cnt++
		}
		return sum / cnt
	}()

	// TODO : inverted index
	res := rpc.ReduceResult{
		MaxValue:  max,
		MinValue:  min,
		EndTime:   end,
		StartTime: start,
		Variance:  float32(variance),
	}
	wr.ReduceNoticeMaster(&res)
}

func splitContent(content []string) (string, string, []string) {
	var val1, val3 []string
	for _, it := range content {
		t := strings.Split(it, "-")
		if len(t) == 3 {
			val1 = append(val1, t[0])
			val3 = append(val3, t[3])
		}
	}
	sort.Strings(val1)
	if len(val1) == 0 {
		var err []string
		return "", "", err
	}
	sort.Strings(val3)
	return val1[0], val1[len(val1)-1], val3
}

func (wr *Worker) ReduceNoticeMaster(result *rpc.ReduceResult) bool {
	conn, err := grpc.Dial(wr.MasterIp, grpc.WithInsecure())
	if err != nil {
		logrus.Error(err)
		return false
	}

	c := rpc.NewMasterClient(conn)
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	res, err := c.Reduce(ctx, result)
	if err != nil {
		logrus.Error(err)
		return false
	}
	if res.RpcRes != "successful" {
		if res.RpcRes == "retry" {
			return wr.ReduceNoticeMaster(result)
		} else {
			logrus.Error(res.RpcRes)
			return false
		}
	}
	return true
}
