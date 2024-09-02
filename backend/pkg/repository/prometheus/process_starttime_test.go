package prometheus

import (
	"testing"
	"time"

	"github.com/CloudDetail/apo/backend/config"
	"github.com/CloudDetail/apo/backend/pkg/logger"
	"github.com/CloudDetail/apo/backend/pkg/model"
)

func TestQueryProcessStartTime(t *testing.T) {
	cfg := config.Get().Promethues
	zapLog := logger.NewLogger(logger.WithLevel("debug"))
	repo, err := New(zapLog, cfg.Address, cfg.Storage)
	if err != nil {
		t.Fatalf("Error to connect prometheus: %v", err)
	}

	endTime := time.Now()
	startTime := endTime.Add(-time.Minute * 30)
	ret, err := repo.QueryProcessStartTime(startTime, endTime, time.Minute, []string{}, []string{"1e77a6527b4b"})
	if err != nil {
		t.Fatalf("QueryProcessStartTime failed, err: %v", err)
	}
	for k, v := range ret {
		t.Logf("QueryProcessStartTime result key: %v, %v", k, v)
	}
	queryInstance := model.ServiceInstance{
		ContainerId: "1e77a6527b4b",
		NodeName:    "worker-23",
	}
	t.Logf("worker-23 pid 1 start time: %v", ret[queryInstance])
}
