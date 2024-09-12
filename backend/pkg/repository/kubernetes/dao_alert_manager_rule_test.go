package kubernetes

import (
	"log"
	"testing"

	"github.com/CloudDetail/apo/backend/pkg/logger"
	"github.com/spf13/viper"
)

func NewTestRepo(t *testing.T) Repo {
	viper.SetConfigFile("testdata/config.yml")
	viper.ReadInConfig()

	authType := viper.GetString("kubernetes.authType")
	authFilePath := viper.GetString("kubernetes.authFilePath")

	zapLog := logger.NewLogger(logger.WithLevel("debug"))
	repo, err := New(zapLog, authType, authFilePath)
	if err != nil {
		t.Fatalf("Error to connect clickhouse: %v", err)
	}
	return repo
}

func Test_k8sApi_GetAlertManagerRule(t *testing.T) {
	repo := NewTestRepo(t)

	cm, err := repo.GetAlertManagerRule()
	if err != nil {
		t.Errorf("Error to get alert manager rule: %v", err)
	}

	log.Printf("cm: %+v", cm)
}
