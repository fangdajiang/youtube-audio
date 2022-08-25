package oss

import (
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"io"
	"os"
	"strings"
	"youtube-audio/pkg/util/env"
	"youtube-audio/pkg/util/log"
	"youtube-audio/pkg/util/myio"
)

const (
	OssYouTubeAudioBucketName string = "youtube-audio"
	OssEndpoint               string = "oss-cn-hongkong.aliyuncs.com"
	EnvAliCloudAccessKeyName  string = "ALICLOUD_ACCESS_KEY"
	EnvAliCloudSecretKeyName  string = "ALICLOUD_SECRET_KEY"

	FetchBaseFileName    string = "fetch_base.json"
	FetchHistoryFileName string = "fetch_history.json"
)

func GetBucket(name string) (*oss.Bucket, error) {
	accessKeyName, err := env.GetEnvVariable(EnvAliCloudAccessKeyName)
	if err != nil {
		return nil, fmt.Errorf("get accessKeyName %s error:%v", accessKeyName, err)
	}
	secretKeyName, err := env.GetEnvVariable(EnvAliCloudSecretKeyName)
	if err != nil {
		return nil, fmt.Errorf("get secretKeyName %s error:%v", secretKeyName, err)
	}

	ossClient, err := oss.New(OssEndpoint, accessKeyName, secretKeyName)
	if err != nil {
		return nil, fmt.Errorf("oss new error:%v", err)
	}
	return ossClient.Bucket(name)
}

func GetResourceJson(ossFileName string) (string, error) {
	bucket, err := GetBucket(OssYouTubeAudioBucketName)
	if err != nil {
		return "", fmt.Errorf("get bucket name error: %s, %s", OssYouTubeAudioBucketName, err)
	}

	body, err := bucket.GetObject(ossFileName)
	if err != nil {
		return "", fmt.Errorf("get object %s error:%s", ossFileName, err)
	}

	data, err := io.ReadAll(body)
	defer func(body io.ReadCloser) {
		err := body.Close()
		if err != nil {
			log.Errorf("io read body error, ossFileName: %s, error: %s", ossFileName, err)
		}
	}(body)
	return string(data), nil
}

func Upload2Oss(filePath string) {
	log.Debugf("log file path:%s", filePath)
	content, err := os.ReadFile(filePath)
	if err != nil {
		log.Errorf("read file error, filePath: %s, error: %s", filePath, err)
		return
	}
	UpdateResourceJson(filePath, string(content))
	myio.Cleanup(filePath)
}

func UpdateResourceJson(ossFileName string, ossFileBody string) {
	bucket, err := GetBucket(OssYouTubeAudioBucketName)
	if err != nil {
		log.Errorf("get bucket error, bucket name: %s, error: %s", OssYouTubeAudioBucketName, err)
		return
	}
	err = bucket.PutObject(ossFileName, strings.NewReader(ossFileBody))
	if err != nil {
		log.Errorf("io put object error, %s", err)
	}
}

func ListBuckets() {
	ossClient, err := oss.New(OssEndpoint, EnvAliCloudAccessKeyName, EnvAliCloudSecretKeyName)
	if err != nil {
		log.Errorf("oss new error:%v", err)
	}
	lsRes, err := ossClient.ListBuckets()
	if err != nil {
		// HandleError(err)
	}

	for _, bucket := range lsRes.Buckets {
		fmt.Println("Buckets:", bucket.Name)
	}
}
