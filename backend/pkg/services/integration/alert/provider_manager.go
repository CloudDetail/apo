package alert

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	"github.com/CloudDetail/apo/backend/pkg/services/integration/alert/provider"
)

type ProviderManager struct {
	mux          sync.RWMutex
	providersMap map[string]*ProviderInstance // sourceID -> provider
}

type ProviderInstance struct {
	provider.Provider

	cancel func()
}

func NewProviderManager() *ProviderManager {
	return &ProviderManager{
		providersMap: make(map[string]*ProviderInstance),
	}
}

func (s *service) DeleteAlertProvider(ctx core.Context, source alert.SourceFrom) {
	pm := s.providerManager
	pm.mux.Lock()
	defer pm.mux.Unlock()
	instance, find := pm.providersMap[source.SourceID]
	if !find {
		return
	}
	instance.cancel()
	delete(pm.providersMap, source.SourceID)
}

// can use for init or update
func (s *service) SetupAlertProvider(ctx core.Context, source alert.AlertSource, pullInterval time.Duration) error {
	pm := s.providerManager
	pm.mux.Lock()
	defer pm.mux.Unlock()

	instance, find := pm.providersMap[source.SourceID]
	if find {
		instance.cancel()
		instance.Provider.UpdateAlertSource(source)
	} else {
		if !source.EnabledPull {
			return nil
		}

		pType, find := provider.ProviderRegistry[source.SourceType]
		if !find || !pType.SupportPull {
			return fmt.Errorf("provider %s not support pull", source.SourceType)
		}

		if err := provider.ValidateJSON(source.Params.Obj, pType.ParamSpec); err != nil {
			return fmt.Errorf("failed to validate provider params, err: %v", err)
		}

		provider := pType.New(source.SourceFrom, source.Params.Obj)
		instance = &ProviderInstance{
			Provider: provider,
		}
		pm.providersMap[source.SourceID] = instance
	}

	cCtx, cancel := context.WithCancel(ctx)
	instance.cancel = cancel

	go s.keepPullAlert(cCtx, source, pullInterval, instance.Provider)
	return nil
}

func (s *service) keepPullAlert(ctx context.Context, source alert.AlertSource, interval time.Duration, p provider.Provider) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	now := time.Now()

	// usually come from database
	lastPullTime := time.UnixMilli(source.LastPullMillTS)
	if lastPullTime.Add(15 * 24 * time.Hour).Before(now) {
		lastPullTime = now.Add(-15 * 24 * time.Hour)
	}

	// first pull immediately
	s.pullAndProcessAlertEvent(source, p, &lastPullTime, now)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			now := time.Now()
			if now.Sub(lastPullTime) < interval {
				continue
			}
			s.pullAndProcessAlertEvent(source, p, &lastPullTime, now)
		}
	}
}

func (s *service) pullAndProcessAlertEvent(source alert.AlertSource, p provider.Provider, lastPullTime *time.Time, to time.Time) {
	events, err := p.PullAlerts(provider.GetAlertParams{
		From: *lastPullTime,
		To:   to,
	})
	if err != nil {
		log.Printf("failed to pull alerts, err: %v", err)
		return
	}

	if err := s.dispatcher.DispatchDecodedEvents(&source.SourceFrom, events); err != nil {
		log.Printf("failed to dispatch events, err: %v", err)
	}

	s.difyRepo.SubmitAlertEvents(events)

	// TODO pass ctx to control database storage
	if err := s.ckRepo.InsertAlertEvent(core.EmptyCtx(), events, source.SourceFrom); err != nil {
		log.Printf("failed to insert alert event, err: %v", err)
		return
	}

	// TODO pass ctx to control database storage
	if err := s.dbRepo.UpdateAlertSourceLastPullTime(core.EmptyCtx(), source.SourceID, to); err != nil {
		log.Printf("failed to update alert source last pull time, err: %v", err)
	}

	// update lastPullTime after successful pull
	*lastPullTime = to
}
