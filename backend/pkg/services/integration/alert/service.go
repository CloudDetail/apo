// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	"log"
	"strings"
	"sync"

	"github.com/CloudDetail/apo/backend/pkg/core"
	input "github.com/CloudDetail/apo/backend/pkg/model/integration"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
	"github.com/CloudDetail/apo/backend/pkg/services/integration/workflow"
)

var _ Service = &service{}

type Service interface {
	CreateAlertSource(source *alert.AlertSource) (*alert.AlertSource, error)
	GetAlertSource(source *alert.SourceFrom) (*alert.AlertSource, error)
	UpdateAlertSource(source *alert.AlertSource) (*alert.AlertSource, error)
	DeleteAlertSource(source alert.SourceFrom) (*alert.AlertSource, error)
	ListAlertSource() ([]alert.AlertSource, error)

	UpdateAlertEnrichRule(*alert.AlertEnrichRuleConfigRequest) error
	GetAlertEnrichRule(sourceID string) ([]alert.AlertEnrichRuleVO, error)

	ProcessAlertEvents(source alert.SourceFrom, data []byte) error

	GetAlertEnrichRuleTags(ctx core.Context) ([]alert.TargetTag, error)

	CreateSchema(req *alert.CreateSchemaRequest) error
	DeleteSchema(schema string) error
	ListSchema() ([]string, error)
	ListSchemaColumns(schema string) ([]string, error)
	UpdateSchemaData(req *alert.UpdateSchemaDataRequest) error
	CheckSchemaIsUsed(schema string) ([]string, error)
	GetSchemaData(schema string) ([]string, map[int64][]string, error)

	CreateCluster(cluster *input.Cluster) error
	ListCluster() ([]input.Cluster, error)
	UpdateCluster(cluster *input.Cluster) error
	DeleteCluster(cluster *input.Cluster) error

	GetDefaultAlertEnrichRule(sourceType string) (string, []alert.AlertEnrichRuleVO)
	ClearDefaultAlertEnrichRule(sourceType string) (bool, error)
	SetDefaultAlertEnrichRule(sourceType string, tagEnrichRules []alert.AlertEnrichRuleVO) error
}

type service struct {
	promRepo prometheus.Repo
	dbRepo   database.Repo
	ckRepo   clickhouse.Repo

	dispatcher         Dispatcher
	AddAlertSourceLock sync.Mutex

	// sourceType -> []alert.AlertEnrichRuleVO
	defaultEnrichRules sync.Map

	alertSubmit *workflow.AlertWorkflow
}

func New(
	promRepo prometheus.Repo,
	dbRepo database.Repo,
	chRepo clickhouse.Repo,
	alertWorkflow *workflow.AlertWorkflow,
) Service {
	var service = &service{
		promRepo: promRepo,
		dbRepo:   dbRepo,
		ckRepo:   chRepo,
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

	service.alertSubmit = alertWorkflow
	return service
}
