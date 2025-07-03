package data

import (
	"errors"
	"time"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/datagroup"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
)

type DataGroupStore struct {
	*datagroup.DataGroupTreeNode
	*datagroup.DataScopeTree

	// ScopesID map[datagroup.ScopeLabels]string // TODO using another scopeID construct
	ExistedScope map[datagroup.DataScope]struct{}
	stopCh       chan struct{}
}

func NewDatasourceStoreMap(prom prometheus.Repo, ch clickhouse.Repo, db database.Repo) *DataGroupStore {
	dgStore := &DataGroupStore{
		ExistedScope: make(map[datagroup.DataScope]struct{}),
		stopCh:       make(chan struct{}),
	}

	dgTree, err := db.LoadDataGroupTree(core.EmptyCtx())
	if err != nil {
		panic(err)
	}

	dgStore.DataGroupTreeNode = dgTree

	scopeTree, err := db.LoadScopes(core.EmptyCtx())
	if err != nil {
		panic(err)
	}
	dgStore.DataScopeTree = scopeTree
	return dgStore
}

func (m *DataGroupStore) KeepWatchScope(
	ctx core.Context,
	promRepo prometheus.Repo,
	chRepo clickhouse.Repo,
	dbRepo database.Repo,
	interval time.Duration,
) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.scanAndSave(ctx, promRepo, chRepo, dbRepo, interval)
		case <-m.stopCh:
			return
		}
	}
}

func (m *DataGroupStore) scanAndSave(
	ctx core.Context,
	promRepo prometheus.Repo,
	chRepo clickhouse.Repo,
	dbRepo database.Repo,
	interval time.Duration,
) error {
	now := time.Now()
	start := now.Add(-2 * interval)

	var errs []error

	scopes, err := m.scanInProm(ctx, promRepo, start.UnixMicro(), now.UnixMicro())
	if err != nil {
		errs = append(errs, err)
	}
	promScopes := generateParent(scopes)

	scopes, err = m.scanInCH(ctx, chRepo, start.UnixMicro(), now.UnixMicro())
	if err != nil {
		errs = append(errs, err)
	}
	chScopes := generateParent(scopes)

	scopes = append(promScopes, chScopes...)
	if err := dbRepo.SaveScopes(ctx, scopes); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

func (m *DataGroupStore) scanInProm(ctx core.Context, prom prometheus.Repo, startTime, endTime int64) ([]datagroup.DataScope, error) {
	var newScope []datagroup.DataScope

	metricRes, err := prom.QueryMetricsWithPQLFilter(ctx,
		prometheus.PQLMetricSeries(prometheus.SPAN_TRACE_COUNT),
		startTime, endTime, "cluster_id,namespace,svc_name", nil)
	if err != nil {
		return nil, err
	}

	for _, metric := range metricRes {
		scopeLabels := datagroup.ScopeLabels{
			ClusterID: metric.Metric.ClusterID,
			Namespace: metric.Metric.Namespace,
			Service:   metric.Metric.SvcName,
		}
		fillEmptyLabel(&scopeLabels, datagroup.DATASOURCE_TYP_SERVICE)
		ds := datagroup.DataScope{
			ScopeID:     scopeLabels.ScopeID(),
			Name:        scopeLabels.Service,
			Type:        datagroup.DATASOURCE_TYP_SERVICE,
			Category:    datagroup.DATASOURCE_CATEGORY_APM,
			ScopeLabels: scopeLabels,
		}
		if _, find := m.ExistedScope[ds]; find {
			continue
		}
		newScope = append(newScope, ds)
	}

	metricRes, err = prom.QueryMetricsWithPQLFilter(ctx,
		prometheus.PQLMetricSeries(
			prometheus.LOG_LEVEL_COUNT,
			prometheus.LOG_EXCEPTION_COUNT,
		),
		startTime, endTime, "cluster_id,namespace", nil,
	)
	if err != nil {
		return nil, err
	}
	for _, metric := range metricRes {
		scopeLabels := datagroup.ScopeLabels{
			ClusterID: metric.Metric.ClusterID,
			Namespace: metric.Metric.Namespace,
		}

		fillEmptyLabel(&scopeLabels, datagroup.DATASOURCE_TYP_NAMESPACE)
		ds := datagroup.DataScope{
			ScopeID:     scopeLabels.ScopeID(),
			Name:        scopeLabels.Namespace,
			Type:        datagroup.DATASOURCE_TYP_NAMESPACE,
			Category:    datagroup.DATASOURCE_CATEGORY_LOG,
			ScopeLabels: scopeLabels,
		}

		if _, find := m.ExistedScope[ds]; find {
			continue
		}
		newScope = append(newScope, ds)
	}

	return newScope, nil
}

func (m *DataGroupStore) scanInCH(ctx core.Context, ch clickhouse.Repo, startTime, endTime int64) ([]datagroup.DataScope, error) {
	var newScope []datagroup.DataScope
	scopes, err := ch.GetAlertDataScope(
		ctx,
		time.UnixMicro(startTime),
		time.UnixMicro(endTime),
	)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(scopes); i++ {
		fillEmptyLabel(&scopes[i].ScopeLabels, scopes[i].Type)
		scopes[i].ScopeID = scopes[i].ScopeLabels.ScopeID()
		if _, find := m.ExistedScope[scopes[i]]; find {
			continue
		}
		newScope = append(newScope, scopes[i])
	}
	return newScope, nil
}

func generateParent(scopes []datagroup.DataScope) []datagroup.DataScope {
	parentScopeTmp := make(map[datagroup.ScopeLabels]struct{})
	extraScopes := make([]datagroup.DataScope, 0)

	for _, scope := range scopes {
		switch scope.Type {
		case datagroup.DATASOURCE_TYP_SERVICE:
			nsLabels := scope.ScopeLabels
			nsLabels.Service = ""
			if addIfNotExists(nsLabels, parentScopeTmp) {
				extraScopes = append(extraScopes, createNamespaceScope(scope, nsLabels))
			}

			clusterLabels := datagroup.ScopeLabels{ClusterID: scope.ClusterID}
			if addIfNotExists(clusterLabels, parentScopeTmp) {
				extraScopes = append(extraScopes, createClusterScope(scope, clusterLabels))
			}

		case datagroup.DATASOURCE_TYP_NAMESPACE:
			clusterLabels := datagroup.ScopeLabels{ClusterID: scope.ClusterID}
			if addIfNotExists(clusterLabels, parentScopeTmp) {
				extraScopes = append(extraScopes, createClusterScope(scope, clusterLabels))
			}
		}
	}

	return append(scopes, extraScopes...)
}

func addIfNotExists(labels datagroup.ScopeLabels, seen map[datagroup.ScopeLabels]struct{}) bool {
	if _, exists := seen[labels]; exists {
		return false
	}
	seen[labels] = struct{}{}
	return true
}

func createNamespaceScope(serviceScope datagroup.DataScope, labels datagroup.ScopeLabels) datagroup.DataScope {
	return datagroup.DataScope{
		ScopeID:     labels.ScopeID(),
		Category:    serviceScope.Category,
		Name:        serviceScope.Namespace,
		Type:        datagroup.DATASOURCE_TYP_NAMESPACE,
		ScopeLabels: labels,
	}
}

func createClusterScope(baseScope datagroup.DataScope, labels datagroup.ScopeLabels) datagroup.DataScope {
	return datagroup.DataScope{
		ScopeID:     labels.ScopeID(),
		Category:    baseScope.Category,
		Name:        baseScope.ClusterID,
		Type:        datagroup.DATASOURCE_TYP_CLUSTER,
		ScopeLabels: labels,
	}
}

func fillEmptyLabel(s *datagroup.ScopeLabels, typ string) {
	switch typ {
	case datagroup.DATASOURCE_TYP_SERVICE:
		if s.Service == "" {
			s.Service = "unknown"
		}
		if s.Namespace == "" {
			s.Namespace = "unknown"
		}
		if s.ClusterID == "" {
			s.ClusterID = "unknown"
		}
	case datagroup.DATASOURCE_TYP_NAMESPACE:
		if s.Namespace == "" {
			s.Namespace = "unknown"
		}
		if s.ClusterID == "" {
			s.ClusterID = "unknown"
		}
	case datagroup.DATASOURCE_TYP_CLUSTER:
		if s.ClusterID == "" {
			s.ClusterID = "unknown"
		}
	}
}
