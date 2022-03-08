package SimpleStorageService

import (
	"testing"
)

func TestUploadDir(t *testing.T) {
	cfg := GetDefaultConfig()
	if UploadDir("bucket", "/Users/echo/Desktop/毕设/MapReduce", cfg.GetDefaultS3Session()) == false {
		t.Errorf("upload dir error\n")
	}
}

func TestDownloadObject(t *testing.T) {
	cfg := GetDefaultConfig()
	if DownloadBucket("/Users/echo/Desktop", "bucket", cfg.GetDefaultS3Session()) == false {
		t.Errorf("upload dir error\n")
	}
}

func TestCopyBucketContent(t *testing.T) {
	cfg := GetDefaultConfig()
	if CopyBucketContent(cfg.GetDefaultS3Session(), "bucket", "test") == false {
		t.Errorf("copy bucket content error\n")
	}
}

func TestDeleteBucket(t *testing.T) {
	cfg := GetDefaultConfig()
	if DeleteBucketContent("bucket", cfg.GetDefaultS3Session()) == false {
		t.Errorf("del bucket content error\n")
	}
}
