// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	"log"
	"strings"
	"sync"
	"time"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	input "github.com/CloudDetail/apo/backend/pkg/model/integration"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/repository/dify"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
	"github.com/CloudDetail/apo/backend/pkg/services/integration/alert/provider"
)

var _ Service = &service{}

type Service interface {
	GetAlertProviderParamsSpec(sourceType string) *response.GetAlertProviderParamsSpecResponse

	CreateAlertSource(ctx core.Context, source *alert.AlertSource) (*alert.AlertSource, error)
	GetAlertSource(ctx core.Context, source *alert.SourceFrom) (*alert.AlertSource, error)
	UpdateAlertSource(ctx core.Context, source *alert.AlertSource) (*alert.AlertSource, error)
	DeleteAlertSource(ctx core.Context, source alert.SourceFrom) (*alert.AlertSource, error)
	ListAlertSource(ctx core.Context) ([]alert.AlertSource, error)

	UpdateAlertEnrichRule(ctx core.Context, req *alert.AlertEnrichRuleConfigRequest) error
	GetAlertEnrichRule(ctx core.Context, sourceID string) ([]alert.AlertEnrichRuleVO, error)

	ProcessAlertEvents(ctx core.Context, source alert.SourceFrom, data []byte) error

	GetAlertEnrichRuleTags(ctx core.Context) ([]alert.TargetTag, error)

	CreateSchema(ctx core.Context, req *alert.CreateSchemaRequest) error
	DeleteSchema(ctx core.Context, schema string) error
	ListSchema(ctx core.Context) ([]string, error)
	ListSchemaColumns(ctx core.Context, schema string) ([]string, error)
	UpdateSchemaData(ctx core.Context, req *alert.UpdateSchemaDataRequest) error
	CheckSchemaIsUsed(ctx core.Context, schema string) ([]string, error)
	GetSchemaData(ctx core.Context, schema string) ([]string, map[int64][]string, error)

	CreateCluster(ctx core.Context, cluster *input.Cluster) error
	ListCluster(ctx core.Context) ([]input.Cluster, error)
	UpdateCluster(ctx core.Context, cluster *input.Cluster) error
	DeleteCluster(ctx core.Context, cluster *input.Cluster) error

	GetDefaultAlertEnrichRule(ctx core.Context, sourceType string) (string, []alert.AlertEnrichRuleVO)
	ClearDefaultAlertEnrichRule(ctx core.Context, sourceType string) (bool, error)
	SetDefaultAlertEnrichRule(ctx core.Context, sourceType string, tagEnrichRules []alert.AlertEnrichRuleVO) error
}

type service struct {
	promRepo prometheus.Repo
	dbRepo   database.Repo
	ckRepo   clickhouse.Repo
	difyRepo dify.DifyRepo

	dispatcher         Dispatcher
	AddAlertSourceLock sync.Mutex

	// sourceType -> []alert.AlertEnrichRuleVO
	defaultEnrichRules sync.Map
}

func New(
	promRepo prometheus.Repo,
	dbRepo database.Repo,
	chRepo clickhouse.Repo,
	difyRepo dify.DifyRepo,
) Service {
	var service = &service{
		promRepo: promRepo,
		dbRepo:   dbRepo,
		ckRepo:   chRepo,
	}

	alertSources, enrichMaps, err := service.dbRepo.LoadAlertEnrichRule(nil)
	if err != nil {
		log.Printf("failed to init alertinput module,err: %v", err)
		return service
	}

	targetTags, err := dbRepo.ListAlertTargetTags(core.EmptyCtx())
	if err != nil {
		log.Printf("failed to init alertinput module,err: %v", err)
		return service
	}

	for source, enricherRules := range enrichMaps {
		if strings.HasPrefix(source.SourceName, defaultSourceName) {
			service.defaultEnrichRules.Store(source.SourceType, enricherRules)
			continue
		}

		enricher, err := service.initExistedAlertSource(source, enricherRules, targetTags)
		if err != nil {
			log.Printf("failed to init enricherFor AlertSource,err: %v", err)
			continue
		}
		service.dispatcher.AddAlertSource(source, enricher)
	}

	for _, source := range alertSources {
		if !source.EnabledPull {
			continue
		}

		pType, find := provider.ProviderRegistry[source.SourceType]
		if !find {
			log.Printf("failed to init provider of AlertSource,err: %v", err)
			continue
		}

		if err = provider.ValidateJSON(source.Params.Obj, pType.ParamSpec); err != nil {
			log.Printf("failed to init provider of AlertSource,err: %v", err)
			continue
		}

		provider := pType.New(source.SourceFrom, source.Params.Obj)
		if err != nil {
			log.Printf("failed to init provider of AlertSource,err: %v", err)
			continue
		}
		go service.KeepPullAlert(core.EmptyCtx(), source, time.Minute, provider)
	}

	service.difyRepo = difyRepo
	return service
}
