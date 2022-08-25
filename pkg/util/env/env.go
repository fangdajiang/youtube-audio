package env

import (
	"fmt"
	"os"
	"strings"
)

func GetEnvVariable(env string) (string, error) {
	recommendation := fmt.Sprintf("Run the following command: 'export %s=\"<%s>\"'", env, env)
	envContent, ok := os.LookupEnv(env)
	if !ok {
		return "", fmt.Errorf("%s doesn't exists. %s", env, recommendation)
	}
	if len(strings.TrimSpace(envContent)) == 0 {
		return "", fmt.Errorf("%s is empty. %s", env, recommendation)
	}
	return envContent, nil
}
