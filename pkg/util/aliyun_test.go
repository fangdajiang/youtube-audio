package util

import (
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestListBuckets(t *testing.T) {
	ListBuckets()
}

func TestGetAliCloudEnvName(t *testing.T) {
	r := require.New(t)

	accessKeyName, err := GetEnvVariable(EnvAliCloudAccessKeyName)
	r.NoError(err)
	log.Infof("accessKeyName: %s", accessKeyName)

	secretKeyName, err := GetEnvVariable(EnvAliCloudSecretKeyName)
	r.NoError(err)
	log.Infof("secretKeyName: %s", secretKeyName)
}

func TestGetResourceJson(t *testing.T) {
	r := require.New(t)

	fetchBaseJson, err := GetResourceJson(FetchBaseFileName)
	r.NoError(err)
	log.Infof("fetchBaseJson: %s", fetchBaseJson)

	fetchHistoryJson, err := GetResourceJson(FetchHistoryFileName)
	r.NoError(err)
	log.Infof("fetchHistoryJson: %s", fetchHistoryJson)
}

func TestUpdateResourceJson(t *testing.T) {
	UpdateResourceJson("tmp_file.json", "{}")
}
