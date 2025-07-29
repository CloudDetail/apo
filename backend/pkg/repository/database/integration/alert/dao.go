// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	sc "github.com/CloudDetail/apo/backend/pkg/model/amconfig/slienceconfig"
	"github.com/CloudDetail/apo/backend/pkg/repository/database/driver"
	dbdriver "github.com/CloudDetail/apo/backend/pkg/repository/database/driver"

	"github.com/CloudDetail/apo/backend/config"
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	"gorm.io/gorm"
)

type AlertInput interface {
	// Manage AlertSource
	CreateAlertSource(ctx core.Context, source *alert.AlertSource) error
	GetAlertSource(ctx core.Context, sourceId string) (*alert.AlertSource, error)
	UpdateAlertSource(ctx core.Context, alertSource *alert.AlertSource) error
	DeleteAlertSource(ctx core.Context, alertSource alert.SourceFrom) (*alert.AlertSource, error)
	ListAlertSource(ctx core.Context) ([]alert.AlertSource, error)

	// Manage AlertEnrichRule
	AddAlertEnrichRule(ctx core.Context, enrichRule []alert.AlertEnrichRule) error
	AddAlertEnrichConditions(ctx core.Context, enrichConditions []alert.AlertEnrichCondition) error
	AddAlertEnrichSchemaTarget(ctx core.Context, enrichSchemaTarget []alert.AlertEnrichSchemaTarget) error
	GetAlertEnrichRule(ctx core.Context, sourceId string) ([]alert.AlertEnrichRule, error)
	GetAlertEnrichConditions(ctx core.Context, sourceId string) ([]alert.AlertEnrichCondition, error)
	GetAlertEnrichSchemaTarget(ctx core.Context, sourceId string) ([]alert.AlertEnrichSchemaTarget, error)
	DeleteAlertEnrichRule(ctx core.Context, ruleIds []string) error
	DeleteAlertEnrichRuleBySourceId(ctx core.Context, sourceId string) error
	DeleteAlertEnrichConditions(ctx core.Context, ruleIds []string) error
	DeleteAlertEnrichConditionsBySourceId(ctx core.Context, sourceId string) error
	DeleteAlertEnrichSchemaTarget(ctx core.Context, ruleIds []string) error
	DeleteAlertEnrichSchemaTargetBySourceId(ctx core.Context, sourceId string) error

	// Manage schema
	CreateSchema(ctx core.Context, schema string, columns []string) error
	DeleteSchema(ctx core.Context, schema string) error
	CheckSchemaIsUsed(ctx core.Context, schema string) ([]string, error)
	ListSchema(ctx core.Context) ([]string, error)
	ListSchemaColumns(ctx core.Context, schema string) ([]string, error)
	InsertSchemaData(ctx core.Context, schema string, columns []string, fullRows [][]string) error
	GetSchemaData(ctx core.Context, schema string) ([]string, map[int64][]string, error)
	UpdateSchemaData(ctx core.Context, schema string, columns []string, rows map[int][]string) error
	ClearSchemaData(ctx core.Context, schema string) error
	SearchSchemaTarget(ctx core.Context, schema string, sourceField string, sourceValue string, targets []alert.AlertEnrichSchemaTarget) ([]string, error)

	ListAlertTargetTags(ctx core.Context) ([]alert.TargetTag, error)

	// Load complate alertEnrichRule
	LoadAlertEnrichRule(ctx core.Context) ([]alert.AlertSource, map[alert.SourceFrom][]alert.AlertEnrichRuleVO, error)

	GetAlertSlience(ctx core.Context) ([]sc.AlertSlienceConfig, error)
	AddAlertSlience(ctx core.Context, slience *sc.AlertSlienceConfig) error
	UpdateAlertSlience(ctx core.Context, slience *sc.AlertSlienceConfig) error
	DeleteAlertSlience(ctx core.Context, id int) error

	LoadFiringIncidents(ctx core.Context) ([]alert.Incident, error)
	LoadIncidentTemplates(ctx core.Context) ([]alert.IncidentKeyTemp, error)
	CreateIncident(ctx core.Context, incident *alert.Incident) error
	UpdateIncident(ctx core.Context, incident *alert.Incident) error

	CreateIncidentTemplates(ctx core.Context, temp []*alert.IncidentKeyTemp) error
	DeleteIncidentTemplates(ctx core.Context, tempIDs []string) error
	UpdateIncidentTemplates(ctx core.Context, temps []*alert.IncidentKeyTemp) error
	GetIncidentTemplatesBySourceId(ctx core.Context, sourceId string) ([]*alert.IncidentKeyTemp, error)
}

type subRepo struct {
	*driver.DB
}

func NewAlertInputRepo(db *gorm.DB, cfg *config.Config) (*subRepo, error) {
	repo := &subRepo{
		DB: &driver.DB{DB: db},
	}

	if err := repo.GetContextDB(nil).AutoMigrate(
		&alert.AlertSource{},
		&alert.TargetTag{},
		&alert.AlertEnrichRule{},
		&alert.AlertEnrichCondition{},
		&alert.AlertEnrichSchemaTarget{},
		&alert.AlertSource2Cluster{},
	); err != nil {
		return nil, err
	}

	err := dbdriver.InitSQL(db, &alert.TargetTag{})
	if err != nil {
		return nil, err
	}

	return repo, nil
}
