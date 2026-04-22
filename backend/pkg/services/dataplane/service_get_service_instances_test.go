// Copyright 2026 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package dataplane

import (
	"errors"
	"sort"
	"testing"
	"time"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

func TestGetServiceInstancesReturnsClickHouseFastPath(t *testing.T) {
	chRepo := &fakeServiceInstancesCHRepo{
		instances: []*model.ServiceInstance{
			newServiceInstance("svc-a", "pod-a", 11),
		},
	}
	promRepo := &fakeServiceInstancesPromRepo{}
	dbRepo := &fakeServiceInstancesDBRepo{allowApo: true}
	svc := newTestDataplaneService(chRepo, promRepo, dbRepo, 50*time.Millisecond, time.Second)

	resp := svc.GetServiceInstances(core.EmptyCtx(), newQueryServiceInstancesRequest())

	if resp.Msg != "" {
		t.Fatalf("expected empty message, got %q", resp.Msg)
	}
	assertInstancePods(t, resp.Results, "pod-a")
	if chRepo.calls != 1 {
		t.Fatalf("expected one clickhouse call, got %d", chRepo.calls)
	}
	if promRepo.calls != 0 {
		t.Fatalf("expected no prometheus fallback, got %d calls", promRepo.calls)
	}
	if dbRepo.calls != 0 {
		t.Fatalf("expected no ownership gate check, got %d calls", dbRepo.calls)
	}
}

func TestGetServiceInstancesFallsBackOnEmptyWhenApoFallbackAllowed(t *testing.T) {
	chRepo := &fakeServiceInstancesCHRepo{}
	promRepo := &fakeServiceInstancesPromRepo{
		instanceList: serviceInstancesList(
			newServiceInstance("svc-a", "pod-prom", 21),
		),
	}
	dbRepo := &fakeServiceInstancesDBRepo{allowApo: true}
	svc := newTestDataplaneService(chRepo, promRepo, dbRepo, 50*time.Millisecond, time.Second)

	resp := svc.GetServiceInstances(core.EmptyCtx(), newQueryServiceInstancesRequest())

	if resp.Msg != "" {
		t.Fatalf("expected empty message, got %q", resp.Msg)
	}
	assertInstancePods(t, resp.Results, "pod-prom")
	if promRepo.calls != 1 {
		t.Fatalf("expected one prometheus fallback call, got %d", promRepo.calls)
	}
	if dbRepo.calls != 1 {
		t.Fatalf("expected one ownership gate check, got %d", dbRepo.calls)
	}
}

func TestGetServiceInstancesFallsBackOnErrorWhenApoFallbackAllowed(t *testing.T) {
	chRepo := &fakeServiceInstancesCHRepo{
		err: errors.New("clickhouse failed"),
	}
	promRepo := &fakeServiceInstancesPromRepo{
		instanceList: serviceInstancesList(
			newServiceInstance("svc-a", "pod-prom", 22),
		),
	}
	dbRepo := &fakeServiceInstancesDBRepo{allowApo: true}
	svc := newTestDataplaneService(chRepo, promRepo, dbRepo, 50*time.Millisecond, time.Second)

	resp := svc.GetServiceInstances(core.EmptyCtx(), newQueryServiceInstancesRequest())

	if resp.Msg != "" {
		t.Fatalf("expected empty message, got %q", resp.Msg)
	}
	assertInstancePods(t, resp.Results, "pod-prom")
	if promRepo.calls != 1 {
		t.Fatalf("expected one prometheus fallback call, got %d", promRepo.calls)
	}
	if dbRepo.calls != 1 {
		t.Fatalf("expected one ownership gate check, got %d", dbRepo.calls)
	}
}

func TestGetServiceInstancesFallsBackOnTimeoutWhenApoFallbackAllowed(t *testing.T) {
	chRepo := &fakeServiceInstancesCHRepo{
		waitForContextDone: true,
	}
	promRepo := &fakeServiceInstancesPromRepo{
		instanceList: serviceInstancesList(
			newServiceInstance("svc-a", "pod-prom", 23),
		),
	}
	dbRepo := &fakeServiceInstancesDBRepo{allowApo: true}
	svc := newTestDataplaneService(chRepo, promRepo, dbRepo, 10*time.Millisecond, time.Second)

	resp := svc.GetServiceInstances(core.EmptyCtx(), newQueryServiceInstancesRequest())

	if resp.Msg != "" {
		t.Fatalf("expected empty message, got %q", resp.Msg)
	}
	assertInstancePods(t, resp.Results, "pod-prom")
	if promRepo.calls != 1 {
		t.Fatalf("expected one prometheus fallback call, got %d", promRepo.calls)
	}
	if dbRepo.calls != 1 {
		t.Fatalf("expected one ownership gate check, got %d", dbRepo.calls)
	}
}

func TestGetServiceInstancesSuppressesPrometheusFallbackForNonApoOwnership(t *testing.T) {
	t.Run("empty_result_returns_empty", func(t *testing.T) {
		chRepo := &fakeServiceInstancesCHRepo{}
		promRepo := &fakeServiceInstancesPromRepo{
			instanceList: serviceInstancesList(
				newServiceInstance("svc-a", "pod-prom", 24),
			),
		}
		dbRepo := &fakeServiceInstancesDBRepo{allowApo: false}
		svc := newTestDataplaneService(chRepo, promRepo, dbRepo, 50*time.Millisecond, time.Second)

		resp := svc.GetServiceInstances(core.EmptyCtx(), newQueryServiceInstancesRequest())

		if resp.Msg != "" {
			t.Fatalf("expected empty message, got %q", resp.Msg)
		}
		if len(resp.Results) != 0 {
			t.Fatalf("expected empty results, got %d items", len(resp.Results))
		}
		if promRepo.calls != 0 {
			t.Fatalf("expected no prometheus fallback, got %d calls", promRepo.calls)
		}
	})

	t.Run("timeout_returns_error_message", func(t *testing.T) {
		chRepo := &fakeServiceInstancesCHRepo{
			waitForContextDone: true,
		}
		promRepo := &fakeServiceInstancesPromRepo{}
		dbRepo := &fakeServiceInstancesDBRepo{allowApo: false}
		svc := newTestDataplaneService(chRepo, promRepo, dbRepo, 10*time.Millisecond, time.Second)

		resp := svc.GetServiceInstances(core.EmptyCtx(), newQueryServiceInstancesRequest())

		if resp.Msg == "" {
			t.Fatal("expected timeout error message, got empty message")
		}
		if promRepo.calls != 0 {
			t.Fatalf("expected no prometheus fallback, got %d calls", promRepo.calls)
		}
	})
}

func TestGetServiceInstancesCachesSuccessfulResultsByRequestKey(t *testing.T) {
	chRepo := &fakeServiceInstancesCHRepo{
		instances: []*model.ServiceInstance{
			newServiceInstance("svc-a", "pod-a", 11),
		},
	}
	promRepo := &fakeServiceInstancesPromRepo{}
	dbRepo := &fakeServiceInstancesDBRepo{allowApo: true}
	svc := newTestDataplaneService(chRepo, promRepo, dbRepo, 50*time.Millisecond, time.Second)
	req := newQueryServiceInstancesRequest()

	first := svc.GetServiceInstances(core.EmptyCtx(), req)
	second := svc.GetServiceInstances(core.EmptyCtx(), req)

	assertInstancePods(t, first.Results, "pod-a")
	assertInstancePods(t, second.Results, "pod-a")
	if chRepo.calls != 1 {
		t.Fatalf("expected cached second request to skip clickhouse, got %d calls", chRepo.calls)
	}
	if promRepo.calls != 0 {
		t.Fatalf("expected no prometheus calls, got %d", promRepo.calls)
	}
}

func TestGetServiceInstancesDoesNotCacheTimeoutOrErrorResponses(t *testing.T) {
	chRepo := &fakeServiceInstancesCHRepo{
		err: errors.New("clickhouse failed"),
	}
	promRepo := &fakeServiceInstancesPromRepo{}
	dbRepo := &fakeServiceInstancesDBRepo{allowApo: false}
	svc := newTestDataplaneService(chRepo, promRepo, dbRepo, 50*time.Millisecond, time.Second)
	req := newQueryServiceInstancesRequest()

	first := svc.GetServiceInstances(core.EmptyCtx(), req)
	second := svc.GetServiceInstances(core.EmptyCtx(), req)

	if first.Msg == "" || second.Msg == "" {
		t.Fatal("expected clickhouse error responses")
	}
	if chRepo.calls != 2 {
		t.Fatalf("expected error responses to bypass cache, got %d clickhouse calls", chRepo.calls)
	}
}

func TestGetServiceInstancesCacheKeyIncludesRequestWindowAndCluster(t *testing.T) {
	chRepo := &fakeServiceInstancesCHRepo{
		instances: []*model.ServiceInstance{
			newServiceInstance("svc-a", "pod-a", 11),
		},
	}
	promRepo := &fakeServiceInstancesPromRepo{}
	dbRepo := &fakeServiceInstancesDBRepo{allowApo: true}
	svc := newTestDataplaneService(chRepo, promRepo, dbRepo, 50*time.Millisecond, time.Second)

	reqA := newQueryServiceInstancesRequest()
	reqB := newQueryServiceInstancesRequest()
	reqB.Cluster = "cluster-b"
	reqC := newQueryServiceInstancesRequest()
	reqC.EndTime = reqC.EndTime + 1

	svc.GetServiceInstances(core.EmptyCtx(), reqA)
	svc.GetServiceInstances(core.EmptyCtx(), reqB)
	svc.GetServiceInstances(core.EmptyCtx(), reqC)

	if chRepo.calls != 3 {
		t.Fatalf("expected distinct cache keys to avoid reuse, got %d clickhouse calls", chRepo.calls)
	}
}

func TestServiceInstancesCacheSetSweepsExpiredEntries(t *testing.T) {
	cache := newServiceInstancesCache(time.Second)
	expiredKey := serviceInstancesCacheKey{
		serviceName: "svc-a",
		cluster:     "cluster-a",
		startTime:   100,
		endTime:     200,
	}
	freshKey := serviceInstancesCacheKey{
		serviceName: "svc-a",
		cluster:     "cluster-a",
		startTime:   101,
		endTime:     201,
	}
	cache.entries[expiredKey] = serviceInstancesCacheEntry{
		expiresAt: time.Now().Add(-time.Second),
		instances: []*model.ServiceInstance{
			newServiceInstance("svc-a", "pod-expired", 31),
		},
	}

	cache.set(freshKey, []*model.ServiceInstance{
		newServiceInstance("svc-a", "pod-fresh", 32),
	})

	cache.mutex.RLock()
	_, expiredExists := cache.entries[expiredKey]
	_, freshExists := cache.entries[freshKey]
	entryCount := len(cache.entries)
	cache.mutex.RUnlock()
	if expiredExists {
		t.Fatal("expected set to sweep expired cache entry")
	}
	if !freshExists {
		t.Fatal("expected set to keep newly written cache entry")
	}
	if entryCount != 1 {
		t.Fatalf("expected one cache entry after sweep, got %d", entryCount)
	}
}

func TestServiceInstancesCacheGetExpiredEntryDoesNotMutateCache(t *testing.T) {
	cache := newServiceInstancesCache(time.Second)
	key := serviceInstancesCacheKey{
		serviceName: "svc-a",
		cluster:     "cluster-a",
		startTime:   100,
		endTime:     200,
	}
	cache.entries[key] = serviceInstancesCacheEntry{
		expiresAt: time.Now().Add(-time.Second),
		instances: []*model.ServiceInstance{
			newServiceInstance("svc-a", "pod-expired", 41),
		},
	}

	if _, ok := cache.get(key); ok {
		t.Fatal("expected expired cache entry to miss")
	}

	cache.mutex.RLock()
	_, exists := cache.entries[key]
	cache.mutex.RUnlock()
	if !exists {
		t.Fatal("expected get miss to leave expired entry for write-path sweep")
	}
}

func newTestDataplaneService(
	chRepo *fakeServiceInstancesCHRepo,
	promRepo *fakeServiceInstancesPromRepo,
	dbRepo *fakeServiceInstancesDBRepo,
	lookupBudget time.Duration,
	cacheTTL time.Duration,
) *service {
	return &service{
		instancesLookup: newServiceInstancesLookup(chRepo, promRepo, dbRepo, lookupBudget, cacheTTL, func(string, ...any) {}),
	}
}

func newQueryServiceInstancesRequest() *request.QueryServiceInstancesRequest {
	return &request.QueryServiceInstancesRequest{
		Cluster:     "cluster-a",
		StartTime:   100,
		EndTime:     200,
		ServiceName: "svc-a",
	}
}

func newServiceInstance(serviceName, pod string, pid int64) *model.ServiceInstance {
	return &model.ServiceInstance{
		ServiceName: serviceName,
		PodName:     pod,
		NodeName:    "node-a",
		Namespace:   "default",
		Pid:         pid,
		ClusterID:   "cluster-a",
	}
}

func serviceInstancesList(instances ...*model.ServiceInstance) *model.ServiceInstances {
	list := model.NewServiceInstances()
	list.AddInstances(instances)
	return list
}

func assertInstancePods(t *testing.T, instances []*model.ServiceInstance, want ...string) {
	t.Helper()
	got := make([]string, 0, len(instances))
	for _, instance := range instances {
		got = append(got, instance.PodName)
	}
	sort.Strings(got)
	sort.Strings(want)
	if len(got) != len(want) {
		t.Fatalf("expected pods %v, got %v", want, got)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("expected pods %v, got %v", want, got)
		}
	}
}

type fakeServiceInstancesCHRepo struct {
	instances          []*model.ServiceInstance
	err                error
	waitForContextDone bool
	calls              int
}

func (f *fakeServiceInstancesCHRepo) GetDataplaneServiceInstances(ctx core.Context, startTime int64, endTime int64, cluster string, serviceNames []string) ([]*model.ServiceInstance, error) {
	f.calls++
	if f.waitForContextDone {
		<-ctx.Done()
		return nil, ctx.Err()
	}
	if f.err != nil {
		return nil, f.err
	}
	return f.instances, nil
}

type fakeServiceInstancesPromRepo struct {
	instanceList *model.ServiceInstances
	err          error
	calls        int
}

func (f *fakeServiceInstancesPromRepo) GetActiveInstanceList(ctx core.Context, startTime int64, endTime int64, clusterId string, serviceNames []string) (*model.ServiceInstances, error) {
	f.calls++
	if f.err != nil {
		return nil, f.err
	}
	if f.instanceList == nil {
		return model.NewServiceInstances(), nil
	}
	return f.instanceList, nil
}

type fakeServiceInstancesDBRepo struct {
	allowApo bool
	err      error
	calls    int
}

func (f *fakeServiceInstancesDBRepo) CheckGetApoMetric(ctx core.Context) (bool, error) {
	f.calls++
	if f.err != nil {
		return false, f.err
	}
	return f.allowApo, nil
}
