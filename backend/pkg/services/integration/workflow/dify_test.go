package workflow

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestDifyClient_alertCheck(t *testing.T) {
	c := NewDifyClient(http.DefaultClient, "192.168.1.10:5001/v1/workflows/run")

	startTime := time.Now().Unix()

	params := `{"pod": "ts-station-service-8588f5f7bd-q8lwp","group": "container","namespace": "train-ticket"}`
	inputs, _ := json.Marshal(map[string]interface{}{
		"alert":     "容器CPU使用率超过80%",
		"params":    params,
		"startTime": (startTime - 15*60) * 1e6,
		"endTime":   startTime * 1e6,
	})

	resp, err := c.alertCheck(
		&DifyRequest{
			Inputs: inputs,
		},
		"Bearer app-mnoDwcsODcgOO0oUYe5XvCze",
		"apo-backend",
	)
	if err != nil {
		t.Error(err)
	}

	fmt.Printf("%+v", resp)
}
