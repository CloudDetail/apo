package log

import (
	"errors"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func log2item(logs []map[string]any) ([]response.LogItem, error) {
	var timestamp int64
	logitems := make([]response.LogItem, len(logs))
	for i, log := range logs {
		content := log["content"]
		delete(log, "content")

		for k, v := range log {
			if k == "timestamp" {
				ts, ok := v.(time.Time)
				if ok {
					timestamp = ts.UnixMicro()
				} else {
					return nil, errors.New("timestamp type error")
				}
				delete(log, k)
			}
			vMap, ok := v.(map[string]string)
			if ok {
				for k2, v2 := range vMap {
					log[k+"."+k2] = v2
				}
				delete(log, k)
			}
		}

		logitems[i] = response.LogItem{
			Content: content,
			Tags:    log,
			Time:    timestamp,
		}
	}
	return logitems, nil
}

func (s *service) QueryLogContext(req *request.LogQueryContextRequest) (*response.LogQueryContextResponse, error) {
	res := &response.LogQueryContextResponse{}
	front, end, _ := s.chRepo.QueryLogContext(req)

	frontItem, err := log2item(front)
	if err != nil {
		return nil, err
	}
	endItem, err := log2item(end)
	if err != nil {
		return nil, err
	}

	res.Front = frontItem
	res.Back = endItem
	return res, nil
}
