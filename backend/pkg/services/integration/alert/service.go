// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	"log"
	"strings"
	"sync"
	core "github.com/CloudDetail/apo/backend/pkg/core"
	input "github.com/CloudDetail/apo/backend/pkg/model/integration"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/repository/dify"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
)

var _ Service = &service{}

type Service interface {
	CreateAlertSource(ctx_core core.Context, source *alert.AlertSource) (*alert.AlertSource, error)
	GetAlertSource(ctx_core core.Context, source *alert.SourceFrom) (*alert.AlertSource, error)
	UpdateAlertSource(ctx_core core.Context, source *alert.AlertSource) (*alert.AlertSource, error)
	DeleteAlertSource(ctx_core core.Context, source alert.SourceFrom) (*alert.AlertSource, error)
	ListAlertSource(ctx_core core.Context,) ([]alert.AlertSource, error)

	UpdateAlertEnrichRule(ctx_core core.Context, req *alert.AlertEnrichRuleConfigRequest) error
	GetAlertEnrichRule(ctx_core core.Context, sourceID string) ([]alert.AlertEnrichRuleVO, error)

	ProcessAlertEvents(ctx_core core.Context, source alert.SourceFrom, data []byte) error

	GetAlertEnrichRuleTags(ctx core.Context) ([]alert.TargetTag, error)

	CreateSchema(ctx_core core.Context, req *alert.CreateSchemaRequest) error
	DeleteSchema(ctx_core core.Context, schema string) error
	ListSchema(ctx_core core.Context,) ([]string, error)
	ListSchemaColumns(ctx_core core.Context, schema string) ([]string, error)
	UpdateSchemaData(ctx_core core.Context, req *alert.UpdateSchemaDataRequest) error
	CheckSchemaIsUsed(ctx_core core.Context, schema string) ([]string, error)
	GetSchemaData(ctx_core core.Context, schema string) ([]string, map[int64][]string, error)

	CreateCluster(ctx_core core.Context, cluster *input.Cluster) error
	ListCluster(ctx_core core.Context,) ([]input.Cluster, error)
	UpdateCluster(ctx_core core.Context, cluster *input.Cluster) error
	DeleteCluster(ctx_core core.Context, cluster *input.Cluster) error

	GetDefaultAlertEnrichRule(ctx_core core.Context, sourceType string) (string, []alert.AlertEnrichRuleVO)
	ClearDefaultAlertEnrichRule(ctx_core core.Context, sourceType string) (bool, error)
	SetDefaultAlertEnrichRule(ctx_core core.Context, sourceType string, tagEnrichRules []alert.AlertEnrichRuleVO) error
}

type service struct {
	promRepo	prometheus.Repo
	dbRepo		database.Repo
	ckRepo		clickhouse.Repo
	difyRepo	dify.DifyRepo

	dispatcher		Dispatcher
	AddAlertSourceLock	sync.Mutex

	// sourceType -> []alert.AlertEnrichRuleVO
	defaultEnrichRules	sync.Map
}

func New(
	promRepo prometheus.Repo,
	dbRepo database.Repo,
	chRepo clickhouse.Repo,
	difyRepo dify.DifyRepo,
) Service {
	var service = &service{
		promRepo:	promRepo,
		dbRepo:		dbRepo,
		ckRepo:		chRepo,
	}

	_, enrichMaps, err := service.dbRepo.LoadAlertEnrichRule()
	if err != nil {
		log.Printf("failed to init alertinput module,err: %v", err)
		return service
	}

	for source, enricherRules := range enrichMaps {
		if strings.HasPrefix(source.SourceName, defaultSourceName) {
			service.defaultEnrichRules.Store(source.SourceType, enricherRules)
			continue
		}

		enricher, err := service.initExistedAlertSource(source, enricherRules)
		if err != nil {
			log.Printf("failed to init enricherFor AlertSource,err: %v", err)
			continue
		}
		service.dispatcher.AddAlertSource(source, enricher)
	}

	service.difyRepo = difyRepo
	return service
}
