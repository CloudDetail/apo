// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package dify

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

var _ asyncAlertCheck = &checkWorkers{}
var _ asyncAlertCheck = &sampleWithFirstRecord{}
var _ asyncAlertCheck = &sampleWithLastRecord{}

type asyncAlertCheck interface {
	Run(ctx context.Context, client *DifyClient) (<-chan *model.WorkflowRecord, error)
	AddEvents(events []alert.AlertEvent)
}

type AlertCheckConfig struct {
	FlowId        string
	FlowName      string
	APIKey        string
	Authorization string
	AnalyzeAPIKey string
	AnalyzeAuth   string
	AnalyzeFlowId string

	User string

	Sampling       string
	CacheMinutes   int
	MaxConcurrency int
	QueueSize      int
	EnqueueTimeout time.Duration

	Prom prometheus.Repo
	CH   clickhouse.Repo
}

const (
	defaultAlertCheckQueueSize      = 1000
	defaultAlertCheckEnqueueTimeout = 3 * time.Second
)

func (cfg *AlertCheckConfig) queueSize() int {
	if cfg != nil && cfg.QueueSize > 0 {
		return cfg.QueueSize
	}
	return defaultAlertCheckQueueSize
}

func (cfg *AlertCheckConfig) enqueueTimeout() time.Duration {
	if cfg != nil && cfg.EnqueueTimeout > 0 {
		return cfg.EnqueueTimeout
	}
	return defaultAlertCheckEnqueueTimeout
}

func (cfg *AlertCheckConfig) newRecordChan() chan *model.WorkflowRecord {
	return make(chan *model.WorkflowRecord, cfg.queueSize())
}

func (cfg *AlertCheckConfig) enqueueAlertEvent(logger *zap.Logger, eventInput *inputChan, event *alert.AlertEvent) bool {
	if eventInput.IsShutDown {
		return false
	}

	timer := time.NewTimer(cfg.enqueueTimeout())
	defer timer.Stop()

	select {
	case eventInput.Ch <- event:
		return true
	case <-timer.C:
		logger.Info("alert check queue is full, drop after waiting",
			zap.String("alertId", event.AlertID),
			zap.Int("queueSize", cfg.queueSize()),
			zap.Int("remainCount", len(eventInput.Ch)),
		)
		return false
	}
}

func newAsyncAlertCheck(cfg *AlertCheckConfig, logger *zap.Logger) asyncAlertCheck {
	switch strings.ToLower(cfg.Sampling) {
	case "no", "false", "disabled":
		return &checkWorkers{
			AlertCheckConfig: cfg,
			logger:           logger,
			eventInput:       newInputChan(cfg.queueSize()),
		}
	case "last":
		return &sampleWithLastRecord{
			logger:           logger,
			AlertCheckConfig: cfg,
			prepareToRun:     make(map[string]alert.AlertEvent),
			eventInput:       newInputChan(cfg.queueSize()),
		}
	default:
		return &sampleWithFirstRecord{
			logger:           logger,
			AlertCheckConfig: cfg,
			checkedAlert:     make(map[string]struct{}),
			eventInput:       newInputChan(cfg.queueSize()),
		}
	}
}

type checkWorkers struct {
	*AlertCheckConfig
	logger *zap.Logger

	eventInput *inputChan
}

func (c *checkWorkers) Run(ctx context.Context, client *DifyClient) (<-chan *model.WorkflowRecord, error) {
	rChan := c.newRecordChan()
	var wg sync.WaitGroup
	for i := 0; i < c.MaxConcurrency; i++ {
		wg.Add(1)
		worker := worker{
			logger:           c.logger,
			expiredTS:        -1,
			AlertCheckConfig: c.AlertCheckConfig,
		}
		go worker.run(client, c.eventInput.Ch, rChan, &wg)
	}

	go waitForShutDown(ctx, c.eventInput, rChan, &wg, c.logger)
	return rChan, nil
}

func (c *checkWorkers) AddEvents(events []alert.AlertEvent) {
	if c.eventInput.IsShutDown {
		return
	}
	for i := 0; i < len(events); i++ {
		c.enqueueAlertEvent(c.logger, c.eventInput, &events[i])
	}
}

type sampleWithFirstRecord struct {
	logger *zap.Logger
	*AlertCheckConfig

	eventInput *inputChan

	checkedAlert map[string]struct{}
	mutex        sync.Mutex

	workers []*worker
}

func (s *sampleWithFirstRecord) Run(
	ctx context.Context,
	client *DifyClient,
) (<-chan *model.WorkflowRecord, error) {
	rChan := s.newRecordChan()

	var wg sync.WaitGroup
	now := time.Now()
	expiredTS := now.Truncate(time.Duration(s.CacheMinutes) * time.Minute).UnixMicro()
	for i := 0; i < s.MaxConcurrency; i++ {
		wg.Add(1)
		worker := worker{
			logger:           s.logger,
			expiredTS:        expiredTS,
			AlertCheckConfig: s.AlertCheckConfig,
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
	go waitForShutDown(ctx, s.eventInput, rChan, &wg, s.logger)
	return rChan, nil
}

func (s *sampleWithFirstRecord) cleanCache() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.checkedAlert = make(map[string]struct{})
	now := time.Now()
	expiredTS := now.Truncate(time.Duration(s.CacheMinutes)*time.Minute).UnixMicro() - int64(s.CacheMinutes)*60*1e6 // Keep 1~2 cycles

	for _, worker := range s.workers {
		worker.expiredTS = expiredTS
	}
}

func (s *sampleWithFirstRecord) AddEvents(events []alert.AlertEvent) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for i := 0; i < len(events); i++ {
		if events[i].Status == alert.StatusResolved {
			continue
		}

		if _, find := s.checkedAlert[events[i].AlertID]; find {
			continue
		}
		if s.enqueueAlertEvent(s.logger, s.eventInput, &events[i]) {
			s.checkedAlert[events[i].AlertID] = struct{}{}
		}
	}
}

type sampleWithLastRecord struct {
	logger *zap.Logger
	*AlertCheckConfig

	eventInput *inputChan

	prepareToRun map[string]alert.AlertEvent
	mutex        sync.Mutex

	workers []*worker
}

func (s *sampleWithLastRecord) Run(
	ctx context.Context,
	client *DifyClient,
) (<-chan *model.WorkflowRecord, error) {
	rChan := s.newRecordChan()

	var wg sync.WaitGroup
	for i := 0; i < s.MaxConcurrency; i++ {
		wg.Add(1)
		worker := worker{
			logger:           s.logger,
			expiredTS:        -1,
			AlertCheckConfig: s.AlertCheckConfig,
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
	go waitForShutDown(ctx, s.eventInput, rChan, &wg, s.logger)
	return rChan, nil
}

func (s *sampleWithLastRecord) AddEvents(events []alert.AlertEvent) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for _, event := range events {
		if event.Status == alert.StatusResolved {
			continue
		}

		if _, find := s.prepareToRun[event.AlertID]; !find {
			if len(s.prepareToRun) >= s.queueSize() {
				s.logger.Info("too many alerts waiting for check, skip", zap.String("alertId", event.AlertID))
				continue
			}
		}
		s.prepareToRun[event.AlertID] = event
	}
}

func (s *sampleWithLastRecord) submit() {
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
				return
			case s.eventInput.Ch <- &cachedEvents[i]:
			}
		}
	}()
}

var runner atomic.Int32

func waitForShutDown(
	ctx context.Context,
	eventInput *inputChan,
	rChan chan *model.WorkflowRecord,
	wg *sync.WaitGroup,
	logger *zap.Logger,
) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			eventInput.IsShutDown = true
			close(eventInput.Ch)
			wg.Wait()
			close(rChan)
		case <-ticker.C:
			logger.Debug("alert waiting for check in dify workflow", zap.Int32("runningWorker", runner.Load()), zap.Int("remainCount", len(eventInput.Ch)))
		}
	}
}
