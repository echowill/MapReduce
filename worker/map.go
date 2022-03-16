package worker

import (
	"MapReduce/SimpleStorageService"
	rpc "MapReduce/common/proto"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"sort"
	"strings"
	"time"
)

/*
	TODO : Map worker应该要做的事情
		1. 从s3://bucket/file拉取log文件
		1.1 文件内容的格式应该为 time-service-sensor.number-value
		2. 只需要将不符合规定的数据保留，按照设备，时间排序，传到s3即可
		3. 传入s3://bucket/file（other bucket）
		4. 地址返回master
		5. reduce获取master收到的地址集
		6. 再次将结果返回master
		6.1 结果集格式，要过滤的内容待定
		7. master将数据上传至app
*/

func (wr *Worker) NewContent(bucket, object string) bool {
	cfg := SimpleStorageService.GetDefaultS3Config()
	reader, err := SimpleStorageService.DownloadObjectToRAM(bucket, object, cfg.GetDefaultS3Session())
	if err != nil {
		return false
	}
	var raw []byte
	(*reader).Read(raw)
	(*reader).Close()
	return wr.doMap(wr.bucketName, object+time.Now().String(), strings.Split(string(raw), "\n"))
}

func (wr *Worker) doMap(bucket, object string, content []string) bool {
	/*
		TODO: 数据格式 time-service-value
			  1.将整个文件中不符合规则的数据全都去除
	*/
	var times, services, values []string
	for _, it := range content {
		time, service, value := contentParse(it)
		times = append(times, time)
		services = append(services, service)
		values = append(values, value)
	}
	sort.Strings(values)
	result := getTimePeriod(times)
	result = result + ":" + getServices(services)
	result = result + ":" + values[0] + "-" + values[len(values)-1]
	//result content is time1-time2:service1-service2:value1-value2
	var tmp []string
	wr.mux.Lock()
	wr.dataCache = append(wr.dataCache, result)
	if len(wr.dataCache) > 1000 {
		tmp = wr.dataCache
		wr.dataCache = []string{}
	}
	wr.mux.Unlock()
	if len(tmp) > 1000 {
		cfg := SimpleStorageService.GetDefaultS3Config() // 已解决TODO : upload to s3
		if SimpleStorageService.UploadObjectFromRAM(bucket, object, tmp, cfg.GetDefaultS3Session()) == false {
			wr.dataCache = append(wr.dataCache, tmp...) // 已解决TODO : 如果此处返回false，即上传s3失败，则需要把结果写回dataCache
			return false
		}
		// 已解决TODO : notice master
		wr.MapNoticeMaster(bucket + object) // TODO:此处没加ip:port
	}
	return true
}

func contentParse(content string) (time, service, value string) {
	split := strings.Split(content, "-")
	time = split[0]
	service = split[1]
	value = split[2]
	return time, service, value
}

func getTimePeriod(content []string) (time string) {
	sort.Strings(content)
	time = content[0] + "-" + content[len(content)-1]
	return time
}

func getServices(content []string) (service string) {
	sort.Strings(content)
	service = content[0] + "-" + content[len(content)-1]
	return service
}

func (wr *Worker) MapNoticeMaster(s3Addr string) bool {
	conn, err := grpc.Dial(wr.MasterIp, grpc.WithInsecure())
	if err != nil {
		logrus.Error(err)
		return false
	}

	c := rpc.NewMasterClient(conn)
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	result := &rpc.Result{}

	res, err := c.Map(ctx, result)
	if err != nil {
		logrus.Error(err)
		return false
	}
	if res.RpcRes != "successful" {
		if res.RpcRes == "retry" {
			return wr.MapNoticeMaster(s3Addr)
		} else {
			logrus.Error(res.RpcRes)
			return false
		}
	}
	return true
}
