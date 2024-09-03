package clickhouse

import (
	"log"
	"testing"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/logger"
	"github.com/spf13/viper"
)

func NewTestRepo(t *testing.T) Repo {
	viper.SetConfigFile("testdata/config.yml")
	viper.ReadInConfig()

	address := viper.GetString("clickhouse.address")
	database := viper.GetString("clickhouse.database")
	username := viper.GetString("clickhouse.username")
	password := viper.GetString("clickhouse.password")

	zapLog := logger.NewLogger(logger.WithLevel("debug"))
	repo, err := New(zapLog, []string{address}, database, username, password)
	if err != nil {
		t.Fatalf("Error to connect clickhouse: %v", err)
	}
	return repo
}
func TestChRepo_RebootTime(t *testing.T) {
	repo := NewTestRepo(t)
	var endTime = time.Now().UnixMicro()
	instances := []string{"stuck-demo-69b56fd6b6-8bkqr"}
	lastUpdateTime, err := repo.RebootTime(endTime, instances)
	if err != nil {
		t.Fatalf("Error to get update time: %v", err)
	}
	log.Printf("lastUpdateTime: %v", lastUpdateTime)
}

func TestChRepo_K8sAlert(t *testing.T) {
	repo := NewTestRepo(t)
	currentTime := time.Now()
	// 获取1小时前的时间
	oneHourAgo := currentTime.Add(-1 * time.Hour)
	instances := []string{"ts-travel2-service-7d9bb6cfb9-pxz5r"}
	k8sAlert, err := repo.K8sAlert(oneHourAgo, currentTime, instances)
	if err != nil {
		t.Fatalf("Error to get k8sAlert: %v", err)
	}
	log.Printf("k8sAlert is: %v", k8sAlert)
}
