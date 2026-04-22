// Copyright 2026 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package dataplane

import (
	go_context "context"
	"fmt"
	"sync"
	"time"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

const (
	defaultServiceInstancesLookupBudget = 2 * time.Second
	defaultServiceInstancesCacheTTL     = 10 * time.Second
)

type serviceInstancesClickHouseRepo interface {
	GetDataplaneServiceInstances(ctx core.Context, startTime int64, endTime int64, cluster string, serviceNames []string) ([]*model.ServiceInstance, error)
}

type serviceInstancesPromRepo interface {
	GetActiveInstanceList(ctx core.Context, startTime int64, endTime int64, clusterId string, serviceNames []string) (*model.ServiceInstances, error)
}

type serviceInstancesDBRepo interface {
	CheckGetApoMetric(ctx core.Context) (bool, error)
}

type serviceInstancesLookup struct {
	chRepo       serviceInstancesClickHouseRepo
	promRepo     serviceInstancesPromRepo
	dbRepo       serviceInstancesDBRepo
	lookupBudget time.Duration
	cache        *serviceInstancesCache
	logf         func(format string, args ...any)
}

func newServiceInstancesLookup(
	chRepo serviceInstancesClickHouseRepo,
	promRepo serviceInstancesPromRepo,
	dbRepo serviceInstancesDBRepo,
	lookupBudget time.Duration,
	cacheTTL time.Duration,
	logf func(format string, args ...any),
) *serviceInstancesLookup {
	if lookupBudget <= 0 {
		lookupBudget = defaultServiceInstancesLookupBudget
	}
	if cacheTTL <= 0 {
		cacheTTL = defaultServiceInstancesCacheTTL
	}
	if logf == nil {
		logf = func(string, ...any) {}
	}
	return &serviceInstancesLookup{
		chRepo:       chRepo,
		promRepo:     promRepo,
		dbRepo:       dbRepo,
		lookupBudget: lookupBudget,
		cache:        newServiceInstancesCache(cacheTTL),
		logf:         logf,
	}
}

func (l *serviceInstancesLookup) query(ctx core.Context, req *request.QueryServiceInstancesRequest) *response.QueryServiceInstancesResponse {
	cacheKey := newServiceInstancesCacheKey(req)
	if instances, ok := l.cache.get(cacheKey); ok {
		return &response.QueryServiceInstancesResponse{
			Results: instances,
		}
	}

	chCtx, cancel := withLookupBudget(ctx, l.lookupBudget)
	defer cancel()

	instances, err := l.chRepo.GetDataplaneServiceInstances(chCtx, req.StartTime, req.EndTime, req.Cluster, []string{req.ServiceName})
	if err == nil && len(instances) > 0 {
		l.cache.set(cacheKey, instances)
		return &response.QueryServiceInstancesResponse{
			Results: cloneServiceInstances(instances),
		}
	}

	apo, gateErr := l.dbRepo.CheckGetApoMetric(ctx)
	if gateErr != nil {
		return &response.QueryServiceInstancesResponse{
			Msg: "query service instances failed: " + gateErr.Error(),
		}
	}

	if !apo {
		if err != nil {
			l.logFallbackSuppressed(req, err)
			return &response.QueryServiceInstancesResponse{
				Msg: "query service instances failed: " + err.Error(),
			}
		}
		l.cache.set(cacheKey, instances)
		return &response.QueryServiceInstancesResponse{
			Results: cloneServiceInstances(instances),
		}
	}

	l.logFallback(req, err, len(instances) == 0)
	return l.queryServiceInstancesByApo(ctx, req, cacheKey)
}

func (l *serviceInstancesLookup) queryServiceInstancesByApo(ctx core.Context, req *request.QueryServiceInstancesRequest, cacheKey serviceInstancesCacheKey) *response.QueryServiceInstancesResponse {
	instanceList, err := l.promRepo.GetActiveInstanceList(ctx, req.StartTime, req.EndTime, req.Cluster, []string{req.ServiceName})
	if err != nil {
		return &response.QueryServiceInstancesResponse{
			Msg: "query service instances failed: " + err.Error(),
		}
	}
	if instanceList == nil {
		instanceList = model.NewServiceInstances()
	}

	instances := make([]*model.ServiceInstance, 0)
	for _, instance := range instanceList.GetInstanceIdMap() {
		instances = append(instances, instance)
	}
	l.cache.set(cacheKey, instances)
	return &response.QueryServiceInstancesResponse{
		Results: cloneServiceInstances(instances),
	}
}

func (l *serviceInstancesLookup) logFallback(req *request.QueryServiceInstancesRequest, err error, empty bool) {
	if err != nil {
		l.logf("dataplane instances clickhouse lookup degraded, falling back to prometheus: service=%s cluster=%s startTime=%d endTime=%d err=%v", req.ServiceName, req.Cluster, req.StartTime, req.EndTime, err)
		return
	}
	if empty {
		l.logf("dataplane instances clickhouse lookup returned empty result, falling back to prometheus: service=%s cluster=%s startTime=%d endTime=%d", req.ServiceName, req.Cluster, req.StartTime, req.EndTime)
	}
}

func (l *serviceInstancesLookup) logFallbackSuppressed(req *request.QueryServiceInstancesRequest, err error) {
	l.logf("dataplane instances prometheus fallback suppressed by non-apo ownership: service=%s cluster=%s startTime=%d endTime=%d err=%v", req.ServiceName, req.Cluster, req.StartTime, req.EndTime, err)
}

type lookupBudgetContext struct {
	core.Context
	ctx go_context.Context
}

func withLookupBudget(ctx core.Context, budget time.Duration) (core.Context, go_context.CancelFunc) {
	if budget <= 0 {
		budget = defaultServiceInstancesLookupBudget
	}
	baseCtx := ctx.GetContext()
	boundedCtx, cancel := go_context.WithTimeout(baseCtx, budget)
	return &lookupBudgetContext{
		Context: ctx,
		ctx:     boundedCtx,
	}, cancel
}

func (c *lookupBudgetContext) GetContext() go_context.Context {
	return c.ctx
}

func (c *lookupBudgetContext) Done() <-chan struct{} {
	return c.ctx.Done()
}

func (c *lookupBudgetContext) Deadline() (deadline time.Time, ok bool) {
	return c.ctx.Deadline()
}

func (c *lookupBudgetContext) Err() error {
	return c.ctx.Err()
}

func (c *lookupBudgetContext) Value(key any) any {
	return c.ctx.Value(key)
}

type serviceInstancesCacheKey struct {
	serviceName string
	cluster     string
	startTime   int64
	endTime     int64
}

func newServiceInstancesCacheKey(req *request.QueryServiceInstancesRequest) serviceInstancesCacheKey {
	return serviceInstancesCacheKey{
		serviceName: req.ServiceName,
		cluster:     req.Cluster,
		startTime:   req.StartTime,
		endTime:     req.EndTime,
	}
}

type serviceInstancesCache struct {
	ttl     time.Duration
	mutex   sync.RWMutex
	entries map[serviceInstancesCacheKey]serviceInstancesCacheEntry
}

type serviceInstancesCacheEntry struct {
	expiresAt time.Time
	instances []*model.ServiceInstance
}

func newServiceInstancesCache(ttl time.Duration) *serviceInstancesCache {
	return &serviceInstancesCache{
		ttl:     ttl,
		entries: make(map[serviceInstancesCacheKey]serviceInstancesCacheEntry),
	}
}

func (c *serviceInstancesCache) get(key serviceInstancesCacheKey) ([]*model.ServiceInstance, bool) {
	if c == nil {
		return nil, false
	}
	c.mutex.RLock()
	entry, ok := c.entries[key]
	c.mutex.RUnlock()
	if !ok {
		return nil, false
	}
	if time.Now().After(entry.expiresAt) {
		return nil, false
	}
	return cloneServiceInstances(entry.instances), true
}

func (c *serviceInstancesCache) set(key serviceInstancesCacheKey, instances []*model.ServiceInstance) {
	if c == nil {
		return
	}
	now := time.Now()
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.deleteExpiredLocked(now)
	c.entries[key] = serviceInstancesCacheEntry{
		expiresAt: now.Add(c.ttl),
		instances: cloneServiceInstances(instances),
	}
}

func (c *serviceInstancesCache) deleteExpiredLocked(now time.Time) {
	for key, entry := range c.entries {
		if now.After(entry.expiresAt) {
			delete(c.entries, key)
		}
	}
}

func cloneServiceInstances(instances []*model.ServiceInstance) []*model.ServiceInstance {
	if instances == nil {
		return []*model.ServiceInstance{}
	}
	cloned := make([]*model.ServiceInstance, 0, len(instances))
	for _, instance := range instances {
		if instance == nil {
			continue
		}
		item := *instance
		cloned = append(cloned, &item)
	}
	return cloned
}

func durationFromSeconds(seconds int, fallback time.Duration) time.Duration {
	if seconds <= 0 {
		return fallback
	}
	return time.Duration(seconds) * time.Second
}

func formatServiceInstancesLookupConfig(lookupBudget time.Duration, cacheTTL time.Duration) string {
	return fmt.Sprintf("lookupBudget=%s cacheTTL=%s", lookupBudget, cacheTTL)
}
