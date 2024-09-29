package kubernetes

import (
	"testing"

	"github.com/CloudDetail/apo/backend/config"
	"github.com/CloudDetail/apo/backend/pkg/logger"
	"github.com/spf13/viper"
)

func NewTestRepo(t *testing.T) Repo {
	viper.SetConfigFile("testdata/config.yml")
	viper.ReadInConfig()

	authType := viper.GetString("kubernetes.authType")
	authFilePath := viper.GetString("kubernetes.authFilePath")

	zapLog := logger.NewLogger(logger.WithLevel("debug"))
	repo, err := New(zapLog, authType, authFilePath, config.MetadataSettings{})
	if err != nil {
		t.Fatalf("Error to connect clickhouse: %v", err)
	}
	return repo
}
