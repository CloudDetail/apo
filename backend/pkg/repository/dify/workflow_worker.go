package dify

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	"go.uber.org/zap"
)

const MAX_CACHE_SIZE = 100

type inputChan struct {
	Ch         chan alert.AlertEvent
	IsShutDown bool
}

func newInputChan() *inputChan {
	return &inputChan{
		Ch:         make(chan alert.AlertEvent, MAX_CACHE_SIZE+10),
		IsShutDown: false,
	}
}

type worker struct {
	logger *zap.Logger
	*AlertCheckConfig

	expiredTS int64
}

func (w *worker) run(c *DifyClient, eventInput <-chan alert.AlertEvent, results chan<- model.WorkflowRecord, wg *sync.WaitGroup) {
	defer wg.Done()
	for event := range eventInput {
		endTime := event.UpdateTime.UnixMicro()
		if w.expiredTS > 0 && endTime < w.expiredTS {
			continue
		}

		startTime := event.UpdateTime.Add(-15 * time.Minute).UnixMicro()
		inputs, _ := json.Marshal(map[string]interface{}{
			"alert":     event.Name,
			"params":    event.TagsInStr(),
			"startTime": startTime,
			"endTime":   endTime,
		})
		resp, err := c.alertCheck(&WorkflowRequest{Inputs: inputs}, w.Authorization, w.User)
		if err != nil {
			w.logger.Error("failed to to alert check", zap.Error(err))
		}

		tw := time.Duration(w.CacheMinutes) * time.Minute
		roundedTime := event.UpdateTime.Truncate(tw).Add(tw)

		var record model.WorkflowRecord
		if resp == nil {
			record = model.WorkflowRecord{
				WorkflowRunID: "",
				WorkflowID:    w.FlowId,
				WorkflowName:  w.FlowName,
				Ref:           event.AlertID,
				Input:         event.ID.String(),
				Output:        "failed: workflow execution failed due to API call failure",
				CreatedAt:     roundedTime.UnixMicro(),
				RoundedTime:   roundedTime.UnixMicro(),
			}
		} else {
			record = model.WorkflowRecord{
				WorkflowRunID: resp.WorkflowRunID(),
				WorkflowID:    w.FlowId,
				WorkflowName:  w.FlowName,
				Ref:           event.AlertID,
				Input:         event.ID.String(),                                  // TODO record input param
				Output:        resp.getOutput("failed: not find expected output"), // 'false' means valid alert
				CreatedAt:     resp.CreatedAt(),
				RoundedTime:   roundedTime.UnixMicro(),
			}
		}
		results <- record
	}
}
