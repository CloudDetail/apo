package clickhouse

import (
	"log"
	"testing"
	"time"

	"github.com/spf13/viper"

	"github.com/CloudDetail/apo/backend/pkg/logger"
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
