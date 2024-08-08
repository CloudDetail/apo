package clickhouse

import (
	"log"
	"testing"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/logger"
)

func NewTestRepo(t *testing.T) Repo {
	address := "192.168.1.6:30189"
	username := "admin"
	password := "27ff0399-0d3a-4bd8-919d-17c2181e6fb9"
	database := "apo"
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

func TestChRepo_NetworkAlert(t *testing.T) {
	repo := NewTestRepo(t)
	currentTime := time.Now()
	// 获取1小时前的时间
	oneHourAgo := currentTime.Add(-24 * time.Hour)
	pods := []string{"train-ticket-mysql-master-0", "ts-station-service-5bc59b4494-drm7w"}
	nodeNames := []string{"worker-23"}
	pids := []string{"123", "22147", "22187"}
	NetworkAlert, err := repo.NetworkAlert(oneHourAgo, currentTime, pods, nodeNames, pids)
	if err != nil {
		t.Fatalf("Error to get NetworkAlert: %v", err)
	}
	log.Printf("NetworkAlert is: %v", NetworkAlert)
}

func TestChRepo_InfrastructureAlert(t *testing.T) {
	repo := NewTestRepo(t)
	currentTime := time.Now()
	// 获取1小时前的时间
	oneHourAgo := currentTime.Add(-1 * time.Hour)
	instances := []string{"ts-travel2-service-7d9bb6cfb9-pxz5r"}
	InfrastructureAlert, err := repo.InfrastructureAlert(oneHourAgo, currentTime, instances)
	if err != nil {
		t.Fatalf("Error to get InfrastructureAlert: %v", err)
	}
	log.Printf("k8sAlert is: %v", InfrastructureAlert)
}

func TestZeroTime(t *testing.T) {
	zeroTime := time.Unix(0, 0)
	t.Logf("zeroTime: %v", zeroTime)
	t.Logf("zeroTime: %v", zeroTime.UnixMicro())
	t.Logf("zeroTime is Zero: %v", zeroTime.IsZero())
}
