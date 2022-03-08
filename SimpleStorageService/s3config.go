package SimpleStorageService

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"io/ioutil"
)

type Config struct {
	Endpoint         string `json:"Endpoint"`
	AccessKey        string `json:"AccessKey"`
	SecretKey        string `json:"SecretKey"`
	DisableSSL       bool   `json:"DisableSSL"`
	S3ForcePathStyle bool   `json:"S3ForcePathStyle"`
}

func GetDefaultConfig() (cfg Config) {
	data, err := ioutil.ReadFile("./config.json")
	if err != nil {
		fmt.Println("Read json failed,error code is ", err)
		panic(err)
	}
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		fmt.Println("Parse Json failed,error code is ", err)
		panic(err)
	}
	return cfg
}

func (cfg *Config) GetDefaultS3Session() *session.Session {
	s3Config := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(cfg.AccessKey, cfg.SecretKey, ""),
		Endpoint:         aws.String(cfg.Endpoint),
		Region:           aws.String("us-east-1"),
		DisableSSL:       aws.Bool(cfg.DisableSSL),
		S3ForcePathStyle: aws.Bool(cfg.S3ForcePathStyle),
	}
	return session.New(s3Config)
}

func GetNewS3Session(accessKey, secretKey, EndPoint, Region string, DisableSSL, S3Style bool) *session.Session {
	s3Config := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(accessKey, secretKey, ""),
		Endpoint:         aws.String(EndPoint),
		Region:           aws.String(Region),
		DisableSSL:       aws.Bool(DisableSSL),
		S3ForcePathStyle: aws.Bool(S3Style),
	}
	return session.New(s3Config)
}
