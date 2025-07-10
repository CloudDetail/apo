package dataplane

import (
	"log"
	"net/http"
	"slices"
	"time"

	"github.com/CloudDetail/apo/backend/config"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
)

type DataplaneRepo interface {
	Start()
	QueryServices(startTime int64, endTime int64) ([]*model.Service, error)
	QueryApmServiceInstances(startTime int64, endTime int64, service *model.Service) ([]*model.ApmServiceInstance, error)
	QueryServiceRedCharts(startTime int64, endTime int64, service *model.Service, step int64) (*model.ServiceRedCharts, error)
	QueryServiceRedValue(startTime int64, endTime int64, service *model.Service) (*model.RedMetricValue, error)
	QueryServiceToplogy(startTime int64, endTime int64, clusterId string, datasource string) ([]*model.ServiceToplogy, error)
}

type dataplaneRepo struct {
	address               string
	client                *http.Client
	ch                    clickhouse.Repo
	db                    database.Repo
	cacheRedInterval      time.Duration
	cacheTopologyInterval time.Duration
}

func New(ch clickhouse.Repo, db database.Repo) (DataplaneRepo, error) {
	dataplaneConf := config.Get().Dataplane
	return &dataplaneRepo{
		address: dataplaneConf.Address,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		ch:                    ch,
		db:                    db,
		cacheRedInterval:      dataplaneConf.QueryRedInterval,
		cacheTopologyInterval: dataplaneConf.QueryTopologyInterval,
	}, nil
}

func (repo *dataplaneRepo) Start() {
	if repo.cacheRedInterval.Seconds() > 0 {
		go repo.loopCacheServiceReds()
	}
	if repo.cacheTopologyInterval.Seconds() > 0 {
		go repo.loopCacheTopologies()
	}
}

func (repo *dataplaneRepo) loopCacheServiceReds() {
	ticker := time.NewTicker(repo.cacheRedInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			now := time.Now().UTC()
			endTime := now.Add(-2 * time.Minute).Truncate(time.Minute)
			startTime := endTime.Add(-30 * time.Minute)
			services, err := repo.QueryServices(startTime.UnixMicro(), endTime.UnixMicro())
			if err != nil {
				log.Printf("failed to query services: %v", err)
			} else {
				if err := repo.cacheServiceReds(startTime.UnixMicro(), endTime.UnixMicro(), services); err != nil {
					log.Printf("failed to cache red metrics: %v", err)
				}

				toResolveApps, err := repo.GetToResolveApps()
				if err != nil {
					log.Printf("failed to get toResolve apps: %v", err)
				} else {
					if len(toResolveApps) > 0 {
						if err := repo.RelateInstances(startTime.UnixMicro(), endTime.UnixMicro(), services, toResolveApps); err != nil {
							log.Printf("failed to relate instance: %v", err)
						}
						if err := repo.RelateServices(toResolveApps); err != nil {
							log.Printf("failed to relate service: %v", err)
						}
					}
				}
			}
		}
	}
}

func (repo *dataplaneRepo) loopCacheTopologies() {
	ticker := time.NewTicker(repo.cacheTopologyInterval)
	defer ticker.Stop()

	repo.CacheServiceTopologies()
	for {
		select {
		case <-ticker.C:
			repo.CacheServiceTopologies()
		}
	}
}

func (repo *dataplaneRepo) cacheServiceReds(startTime int64, endTime int64, services []*model.Service) error {
	for _, service := range services {
		redCharts, err := repo.QueryServiceRedCharts(startTime, endTime, service, 60_000_000)
		if err != nil {
			return err
		}
		if err := repo.ch.WriteServiceRedMetrics(core.EmptyCtx(), redCharts); err != nil {
			return err
		}
	}
	return nil
}

func (repo *dataplaneRepo) CacheServiceTopologies() error {
	now := time.Now().UTC()
	endTime := now.Truncate(time.Minute)
	startTime := endTime.Add(-60 * time.Minute)

	services, err := repo.QueryServices(startTime.UnixMicro(), endTime.UnixMicro())
	if err != nil {
		return err
	}
	dataSources := make(map[string][]string)
	for _, service := range services {
		if clusterIds, ok := dataSources[service.Source]; ok {
			if !slices.Contains(clusterIds, service.ClusterId) {
				clusterIds = append(clusterIds, service.ClusterId)
				dataSources[service.Source] = clusterIds
			}
		} else {
			clusterIds = []string{service.ClusterId}
			dataSources[service.Source] = clusterIds
		}
	}

	for dataSource, clusterIds := range dataSources {
		for _, clusterId := range clusterIds {
			topologies, err := repo.QueryServiceToplogy(startTime.UnixMicro(), endTime.UnixMicro(), clusterId, dataSource)
			if err != nil {
				return err
			}
			if len(topologies) > 0 {
				count, err := repo.ch.QueryServiceTopologyCount(core.EmptyCtx(), endTime.UnixMicro(), clusterId, dataSource)
				if err != nil {
					return err
				}
				if count == 0 {
					if err = repo.ch.WriteServiceTopology(core.EmptyCtx(), endTime.UnixMicro(), clusterId, dataSource, topologies); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func (repo *dataplaneRepo) GetToResolveApps() ([]*model.AppInfo, error) {
	return repo.ch.GetToResolveApps(core.EmptyCtx())
}

func (repo *dataplaneRepo) RelateInstances(startTime int64, endTime int64, services []*model.Service, toResolveApps []*model.AppInfo) error {
	if len(services) == 0 {
		return nil
	}

	for _, service := range services {
		instances, err := repo.QueryApmServiceInstances(startTime, endTime, service)
		if err != nil {
			return err
		}
		for _, instance := range instances {
			if matchedApp := instance.MatchApp(toResolveApps); matchedApp != nil {
				matchedApp.Labels["source"] = service.Source
				matchedApp.Labels["service_id"] = service.Id
				matchedApp.Labels["service_name"] = service.Name
				log.Printf("[Match Instance] ServiceName: %s, HostName: %s, Pid: %d", instance.ServiceName, matchedApp.Labels["host_name"], matchedApp.HostPid)
				if err := repo.ch.WriteRelateApp(core.EmptyCtx(), matchedApp); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (repo *dataplaneRepo) RelateServices(toResolveApps []*model.AppInfo) error {
	rules, err := repo.db.ListAllServiceNameRule(core.EmptyCtx())
	if err != nil {
		return err
	}
	if len(rules) == 0 {
		// No Rule is set.
		return nil
	}
	conditions, err := repo.db.ListAllServiceNameRuleCondition(core.EmptyCtx())
	if err != nil {
		return err
	}
	if len(conditions) == 0 {
		// No Condition is set.
		return nil
	}
	ruleMap := map[int]*ServiceRuleConditions{}
	for _, rule := range rules {
		ruleMap[rule.ID] = newServiceRuleConditions(rule.ClusterId, rule.Service)
	}
	for _, condition := range conditions {
		if serviceRuleConditions, ok := ruleMap[condition.RuleID]; ok {
			serviceRuleConditions.addCondition(&condition)
		}
	}

	for _, app := range toResolveApps {
		if app.Labels["service_name"] != "" {
			continue
		}
		if serviceName := getMatchedServiceByRules(app, ruleMap); serviceName != "" {
			app.Labels["source"] = ""
			app.Labels["service_name"] = serviceName
			app.Labels["service_id"] = serviceName
			log.Printf("[Match Service By Rule] ServiceName: %s, HostName: %s, Pid: %d", serviceName, app.Labels["host_name"], app.HostPid)
			if err := repo.ch.WriteRelateApp(core.EmptyCtx(), app); err != nil {
				return err
			}
		}
	}
	return nil
}

func getMatchedServiceByRules(app *model.AppInfo, ruleMap map[int]*ServiceRuleConditions) string {
	for _, serviceRuleConditions := range ruleMap {
		if serviceRuleConditions.ClusterId != app.Labels["cluster_id"] {
			continue
		}
		if serviceName := serviceRuleConditions.GetMatchedServiceName(app); serviceName != "" {
			return serviceName
		}
	}
	return ""
}

type ServiceRuleConditions struct {
	ClusterId   string
	ServiceName string
	Conditions  []*database.ServiceNameRuleCondition
}

func newServiceRuleConditions(clusterId string, serviceName string) *ServiceRuleConditions {
	return &ServiceRuleConditions{
		ClusterId:   clusterId,
		ServiceName: serviceName,
		Conditions:  make([]*database.ServiceNameRuleCondition, 0),
	}
}

func (condtions *ServiceRuleConditions) addCondition(condition *database.ServiceNameRuleCondition) {
	condtions.Conditions = append(condtions.Conditions, condition)
}

func (conditions *ServiceRuleConditions) GetMatchedServiceName(app *model.AppInfo) string {
	for _, condition := range conditions.Conditions {
		if !condition.Match(app.Labels) {
			return ""
		}
	}
	return conditions.ServiceName
}
