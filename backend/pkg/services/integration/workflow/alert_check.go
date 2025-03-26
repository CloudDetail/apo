// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

const MAX_CACHE_SIZE = 100

type AlertCheckCfg struct {
	FlowId        string
	FlowName      string
	APIKey        string
	Authorization string
	User          string

	Sampling       string
	CacheMinutes   int
	MaxConcurrency int
}

type alertCheck interface {
	Run(ctx context.Context, client *DifyClient) (<-chan model.WorkflowRecord, error)
	AddEvents(events []alert.AlertEvent)
}

func newAlertCheck(cfg *AlertCheckCfg, logger *zap.Logger) alertCheck {
	switch strings.ToLower(cfg.Sampling) {
	case "no", "false", "disabled":
		return &checkWorkers{
			AlertCheckCfg: cfg,
			logger:        logger,
			eventInput:    newInputChan(),
		}
	case "last":
		return &sampleWithLatestRecord{
			logger:        logger,
			AlertCheckCfg: cfg,
			prepareToRun:  make(map[string]alert.AlertEvent),
			eventInput:    newInputChan(),
		}
	default:
		return &sampleWithFirstRecord{
			logger:        logger,
			AlertCheckCfg: cfg,
			checkedAlert:  make(map[string]struct{}),
			eventInput:    newInputChan(),
		}
	}
}

type checkWorkers struct {
	*AlertCheckCfg
	logger *zap.Logger

	eventInput *inputChan
}

func (c *checkWorkers) Run(ctx context.Context, client *DifyClient) (<-chan model.WorkflowRecord, error) {
	rChan := make(chan model.WorkflowRecord, MAX_CACHE_SIZE+10)
	var wg sync.WaitGroup
	for i := 0; i < c.MaxConcurrency; i++ {
		wg.Add(1)
		worker := worker{
			logger:        c.logger,
			expiredTS:     -1,
			AlertCheckCfg: c.AlertCheckCfg,
		}
		go worker.run(client, c.eventInput.Ch, rChan, &wg)
	}

	go waitForShutDown(ctx, c.eventInput, rChan, &wg)
	return rChan, nil
}

func (c *checkWorkers) AddEvents(events []alert.AlertEvent) {
	if c.eventInput.IsShutDown {
		return
	}
	remainSize := len(c.eventInput.Ch)
	for i := 0; i < len(events); i++ {
		if remainSize > MAX_CACHE_SIZE {
			c.logger.Info("too many alerts waiting for check, skip", zap.String("alertId", events[i].AlertID))
			continue
		}
		remainSize++
		c.eventInput.Ch <- events[i]
	}
}

type sampleWithFirstRecord struct {
	logger *zap.Logger
	*AlertCheckCfg

	eventInput *inputChan

	checkedAlert map[string]struct{}
	mutex        sync.Mutex

	workers   []*worker
	dropCount int // ignore data races.
}

func (s *sampleWithFirstRecord) Run(
	ctx context.Context,
	client *DifyClient,
) (<-chan model.WorkflowRecord, error) {
	rChan := make(chan model.WorkflowRecord, MAX_CACHE_SIZE+10)

	var wg sync.WaitGroup
	now := time.Now()
	expiredTS := now.Truncate(time.Duration(s.CacheMinutes) * time.Minute).UnixMicro()
	for i := 0; i < s.MaxConcurrency; i++ {
		wg.Add(1)
		worker := worker{
			logger:        s.logger,
			expiredTS:     expiredTS,
			AlertCheckCfg: s.AlertCheckCfg,
		}
		s.workers = append(s.workers, &worker)
		go worker.run(client, s.eventInput.Ch, rChan, &wg)
	}
	cronTab := cron.New()
	_, err := cronTab.AddFunc(fmt.Sprintf("*/%d * * * *", s.CacheMinutes), func() {
		s.cleanCache()
	})
	if err != nil {
		return nil, err
	}
	cronTab.Start()
	go waitForShutDown(ctx, s.eventInput, rChan, &wg)
	return rChan, nil
}

func (s *sampleWithFirstRecord) cleanCache() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.checkedAlert = make(map[string]struct{})
	now := time.Now()
	expiredTS := now.Truncate(time.Duration(s.CacheMinutes) * time.Minute).UnixMicro()

	dropCount := 0
	for _, worker := range s.workers {
		worker.expiredTS = expiredTS
		dropCount += worker.dropCount
		worker.dropCount = 0
	}

	if dropCount+s.dropCount > 0 {
		s.logger.Info("check alert failed count", zap.Int("cacheMinutes", s.CacheMinutes), zap.Int("drop count", dropCount+s.dropCount))
	}
	s.dropCount++
}

func (s *sampleWithFirstRecord) AddEvents(events []alert.AlertEvent) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	remainSize := len(s.eventInput.Ch)
	for i := 0; i < len(events); i++ {
		if _, find := s.checkedAlert[events[i].AlertID]; find {
			continue
		}
		s.checkedAlert[events[i].AlertID] = struct{}{}
		if remainSize > MAX_CACHE_SIZE {
			s.dropCount++
			s.logger.Info("too many alerts waiting for check, skip", zap.String("alertId", events[i].AlertID))
			continue
		}
		remainSize++
		s.eventInput.Ch <- events[i]
	}
}

type sampleWithLatestRecord struct {
	logger *zap.Logger
	*AlertCheckCfg

	eventInput *inputChan

	prepareToRun map[string]alert.AlertEvent
	mutex        sync.Mutex

	workers   []*worker
	dropCount int // ignore data races.
}

func (s *sampleWithLatestRecord) Run(
	ctx context.Context,
	client *DifyClient,
) (<-chan model.WorkflowRecord, error) {
	rChan := make(chan model.WorkflowRecord, MAX_CACHE_SIZE+10)

	var wg sync.WaitGroup
	for i := 0; i < s.MaxConcurrency; i++ {
		wg.Add(1)
		worker := worker{
			logger:        s.logger,
			expiredTS:     -1,
			AlertCheckCfg: s.AlertCheckCfg,
		}
		s.workers = append(s.workers, &worker)
		go worker.run(client, s.eventInput.Ch, rChan, &wg)
	}
	cronTab := cron.New()
	_, err := cronTab.AddFunc(fmt.Sprintf("*/%d * * * *", s.CacheMinutes), func() {
		s.submit()
	})
	if err != nil {
		return nil, err
	}
	cronTab.Start()
	go waitForShutDown(ctx, s.eventInput, rChan, &wg)
	return rChan, nil
}

func (s *sampleWithLatestRecord) AddEvents(events []alert.AlertEvent) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for _, event := range events {
		if event.Status == alert.StatusResolved {
			continue
		}

		if _, find := s.prepareToRun[event.AlertID]; !find {
			if len(s.prepareToRun) > MAX_CACHE_SIZE {
				s.dropCount++
				continue
			}
		}
		s.prepareToRun[event.AlertID] = event
	}
}

func (s *sampleWithLatestRecord) submit() {
	dropCount := 0
	for _, worker := range s.workers {
		dropCount += worker.dropCount
		worker.dropCount = 0
	}
	if dropCount+s.dropCount > 0 {
		s.logger.Info("check alert failed count", zap.Int("cacheMinutes", s.CacheMinutes), zap.Int("drop count", dropCount+s.dropCount))
	}
	s.dropCount = 0

	s.mutex.Lock()
	var cachedEvents []alert.AlertEvent
	for _, event := range s.prepareToRun {
		cachedEvents = append(cachedEvents, event)
	}
	s.prepareToRun = make(map[string]alert.AlertEvent)
	s.mutex.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*time.Duration(s.CacheMinutes-1))
	defer cancel()
	go func() {
		for i := 0; i < len(cachedEvents); i++ {
			select {
			case <-ctx.Done():
				s.dropCount += len(cachedEvents) - i
				return
			case s.eventInput.Ch <- cachedEvents[i]:
			}
		}
	}()
}

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
	*AlertCheckCfg

	expiredTS int64
	dropCount int // ignore data races.
}

func (w *worker) run(c *DifyClient, eventInput <-chan alert.AlertEvent, results chan<- model.WorkflowRecord, wg *sync.WaitGroup) {
	defer wg.Done()
	for event := range eventInput {
		endTime := event.ReceivedTime.UnixMicro()
		if w.expiredTS > 0 && endTime < w.expiredTS {
			w.dropCount++
			continue
		}

		startTime := event.ReceivedTime.Add(-15 * time.Minute).UnixMicro()
		inputs, _ := json.Marshal(map[string]interface{}{
			"alert":     event.Name,
			"params":    event.TagsInStr(),
			"startTime": startTime,
			"endTime":   endTime,
		})
		resp, err := c.alertCheck(&DifyRequest{Inputs: inputs}, w.Authorization, w.User)
		if err != nil {
			w.dropCount++
			w.logger.Error("failed to to alert check", zap.Error(err))
			continue
		}

		tw := time.Duration(w.CacheMinutes) * time.Minute
		roundedTime := event.ReceivedTime.Truncate(tw).Add(tw)

		record := model.WorkflowRecord{
			WorkflowRunID: resp.WorkflowRunID(),
			WorkflowID:    w.FlowId,
			WorkflowName:  w.FlowName,
			Ref:           event.AlertID,
			Input:         "",                             // TODO record input param
			Output:        resp.IsValidOrDefault("false"), // 'false' means valid alert
			CreatedAt:     resp.CreatedAt(),
			RoundedTime:   roundedTime.UnixMicro(),
		}

		results <- record
	}
}

func waitForShutDown(
	ctx context.Context,
	eventInput *inputChan,
	rChan chan model.WorkflowRecord,
	wg *sync.WaitGroup,
) {
	<-ctx.Done()
	eventInput.IsShutDown = true
	close(eventInput.Ch)
	wg.Wait()
	close(rChan)
}

func (c *AlertCheckCfg) HasValidAPIKey() bool {
	if c == nil {
		return false
	}
	// TODO check Destination
	return len(c.APIKey) > 0
}
