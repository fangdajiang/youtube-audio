package oss

import (
	"github.com/stretchr/testify/require"
	"testing"
	"youtube-audio/pkg/util/env"
	"youtube-audio/pkg/util/log"
)

func TestListBuckets(t *testing.T) {
	ListBuckets()
}

func TestGetAliCloudEnvName(t *testing.T) {
	r := require.New(t)

	accessKeyName, err := env.GetEnvVariable(EnvAliCloudAccessKeyName)
	r.NoError(err)
	log.Debugf("accessKeyName: %s", accessKeyName)

	secretKeyName, err := env.GetEnvVariable(EnvAliCloudSecretKeyName)
	r.NoError(err)
	log.Debugf("secretKeyName: %s", secretKeyName)
}

func TestGetResourceJson(t *testing.T) {
	r := require.New(t)

	fetchBaseJson, err := GetResourceJson(FetchBaseFileName)
	r.NoError(err)
	log.Debugf("fetchBaseJson: %s", fetchBaseJson)

	fetchHistoryJson, err := GetResourceJson(FetchHistoryFileName)
	r.NoError(err)
	log.Debugf("fetchHistoryJson: %s", fetchHistoryJson)
}

func TestUpdateResourceJson(t *testing.T) {
	UpdateResourceJson("tmp_file.json", "{}")
}

func TestUpload2Oss(t *testing.T) {
	r := require.New(t)

	Upload2Oss(log.LoggingFilePath)

	r.NoFileExists(log.LoggingFilePath)

}
