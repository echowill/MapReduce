package SimpleStorageService

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"io/ioutil"
	"os"
	"strings"
)

func GetObjectList(sess *session.Session, bucketName string) (res []string) {
	svc := s3.New(sess)
	input := &s3.ListObjectsInput{
		Bucket: aws.String(bucketName),
	}
	result, err := svc.ListObjects(input)
	if err != nil {
		fmt.Println(err)
		return res
	}
	for _, it := range result.Contents {
		res = append(res, *it.Key)
	}
	return res
}

func GetBucketList(sess *session.Session) (res []string) {
	svc := s3.New(sess)
	result, err := svc.ListBuckets(&s3.ListBucketsInput{})
	if err != nil {
		fmt.Println(err)
		return res
	}
	for _, it := range result.Buckets {
		res = append(res, *it.Name)
	}
	return res
}

func CopyBucketContent(sess *session.Session, src, des string) (res bool) {
	res = true
	svc := s3.New(sess)
	objList := GetObjectList(sess, src)
	for _, it := range objList {
		input := &s3.CopyObjectInput{
			Bucket:     aws.String(des),
			CopySource: aws.String("/" + src + "/" + it),
			Key:        aws.String(it),
		}
		_, err := svc.CopyObject(input)
		if err != nil {
			res = false
		}
	}
	return res
}

func UploadObject(bucket, object, filePath string, sess *session.Session) bool {
	f, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("open filename is %s error,code is %s", filePath, err)
		return false
	}
	uploader := s3manager.NewUploader(sess)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
		Body:   f,
	}, func(uploader *s3manager.Uploader) {
		uploader.PartSize = 10 * 1024 * 1024 // 大于10M
		uploader.LeavePartsOnError = true
		uploader.Concurrency = 8 // 8并发
	})
	if err != nil {
		fmt.Printf("upload filename is %s error,code is %s", filePath, err)
		return false
	}
	return true
}

func DownloadObject(bucket, object, filePath string, sess *session.Session) bool {
	downloader := s3manager.NewDownloader(sess)
	f, err := os.Create(filePath)
	if err != nil {
		fmt.Printf("crate %s error,code is %s", f.Name(), err)
		return false
	}
	defer f.Close()
	_, err = downloader.Download(f, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	if err != nil {
		fmt.Printf("download object %s error ,code is %s", object, err)
		return false
	}
	return true
}

func DeleteObject(bucket, object string, sess *session.Session) bool {
	svc := s3.New(sess)
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	}
	_, err := svc.DeleteObject(input)
	if err != nil {
		fmt.Printf("delete %s failed,code is %s", object, err)
		return false
	}
	return true
}

func UploadDir(bucket, dir string, sess *session.Session) bool {
	files, err := getAllFileName(dir)
	if err != nil {
		fmt.Printf("get file error,code is %s", err)
		return false
	}
	for _, it := range files {
		if UploadObject(bucket, parseFileName(it), it, sess) == false {
			return false
		}
	}
	return true
}

func DownloadBucket(dir, bucket string, sess *session.Session) bool {
	err := os.Mkdir(dir+"/"+bucket, 0755)
	if err != nil {
		fmt.Printf("create dir error ,code is %s", err)
		return false
	}
	objs := GetObjectList(sess, bucket)
	for _, it := range objs {
		if DownloadObject(bucket, it, dir+"/"+bucket+"/"+it, sess) == false {
			fmt.Printf("download object %s error ,code is %s", it, err)
			return false
		}
	}
	return true
}

func DeleteBucketContent(bucket string, sess *session.Session) bool {
	cfg := GetDefaultConfig()
	objs := GetObjectList(cfg.GetDefaultS3Session(), bucket)
	for _, it := range objs {
		if DeleteObject(bucket, it, sess) == false {
			return false
		}
	}
	return true
}

func DeleteBucket(bucket string, sess *session.Session) bool {
	svc := s3.New(sess)
	_, err := svc.DeleteBucket(&s3.DeleteBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		fmt.Printf("delete bucket error,code is %s", err)
		return false
	}
	return true
}

func CleanUserContent(sess *session.Session) bool {
	svc := s3.New(sess)
	result, err := svc.ListBuckets(&s3.ListBucketsInput{})
	if err != nil {
		fmt.Printf("get bucket list error ,code is %s", err)
		return false
	}
	for _, it := range result.Buckets {
		if DeleteBucketContent(*it.Name, sess) == false {
			fmt.Printf("delete bucket error ,code is %s", *it.Name)
			return false
		}
	}
	return true
}

func readFileToMemory(filePath string) []byte {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return data
}

func writeFileToDisk(filePath string, data []byte) bool {
	err := ioutil.WriteFile(filePath, data, 0644)
	if err != nil {
		return false
	}
	return true
}

func getAllFileName(dir string) (result []string, err error) {

	fis, err := ioutil.ReadDir(dir)
	if err != nil {
		fmt.Printf("读取文件目录失败，pathname=%v, err=%v \n", dir, err)
		return result, err
	}
	// 所有文件/文件夹
	for _, fi := range fis {
		fullName := dir + "/" + fi.Name()
		// 是文件夹则递归进入获取;是文件，则压入数组
		if fi.IsDir() {
			temp, err := getAllFileName(fullName)
			if err != nil {
				fmt.Printf("读取文件目录失败,fullName=%v, err=%v", fullName, err)
				return result, err
			}
			result = append(result, temp...)
		} else {
			result = append(result, fullName)
		}
	}
	return result, err
}

func parseFileName(path string) string {
	res := strings.Split(path, "/")
	return res[len(res)-1]
}
