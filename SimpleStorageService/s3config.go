package SimpleStorageService

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"os"
)

func S3test() {
	bucket := aws.String("bucket") //bucket名称
	key := aws.String("keyring")   //object name
	access_key := "echo"
	secret_key := "echo"
	end_point := "192.168.1.152:7480" //endpoint设置，不要动
	s3Config := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(access_key, secret_key, ""),
		Endpoint:         aws.String(end_point),
		Region:           aws.String("us-east-1"),
		DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(true), // 此处设置为true格式为http://host:port/bucket
	}
	newSession := session.New(s3Config)
	s3Client := s3.New(newSession)
	cparams := &s3.HeadBucketInput{
		Bucket: bucket, // Required
	}
	_, err := s3Client.HeadBucket(cparams)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	uploader := s3manager.NewUploader(newSession)
	filename := "/Users/echo/Desktop/Google-MapReduce中文版_1.0.pdf" //上传文件路径
	f, err := os.Open(filename)
	if err != nil {
		fmt.Errorf("failed to open file %q, %v", filename, err)
		return
	}
	// Upload the file to S3.
	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: bucket,
		Key:    key,
		Body:   f,
	}, func(u *s3manager.Uploader) {
		u.PartSize = 10 * 1024 * 1024 // 分块大小,当文件体积超过10M开始进行分块上传
		u.LeavePartsOnError = true
		u.Concurrency = 3
	}) //并发数
	if err != nil {
		fmt.Printf("Failed to upload data to %s/%s, %s\n", *bucket, *key, err.Error())
		return
	}
	fmt.Printf("file uploaded to, %s\n", result.Location)
	down_file := "/Users/echo/Desktop/down_file.pdf" //下载路径
	file, err := os.Create(down_file)
	if err != nil {
		fmt.Println("Failed to create file", err)
		return
	}
	defer file.Close()
	downloader := s3manager.NewDownloader(newSession)
	numBytes, err := downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: bucket,
			Key:    key,
		})
	if err != nil {
		fmt.Println("Failed to download file", err)
		return
	}
	fmt.Println("Downloaded file", file.Name(), numBytes, "bytes")
}
