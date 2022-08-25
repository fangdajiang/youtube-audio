package util

import (
	"github.com/stretchr/testify/require"
	"testing"
	"youtube-audio/pkg/util/log"
)

func TestListBuckets(t *testing.T) {
	ListBuckets()
}

func TestGetAliCloudEnvName(t *testing.T) {
	r := require.New(t)

	accessKeyName, err := GetEnvVariable(EnvAliCloudAccessKeyName)
	r.NoError(err)
	log.Debugf("accessKeyName: %s", accessKeyName)

	secretKeyName, err := GetEnvVariable(EnvAliCloudSecretKeyName)
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
