// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	sc "github.com/CloudDetail/apo/backend/pkg/model/amconfig/slienceconfig"
	dbdriver "github.com/CloudDetail/apo/backend/pkg/repository/database/driver"

	"github.com/CloudDetail/apo/backend/config"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	"gorm.io/gorm"
	core "github.com/CloudDetail/apo/backend/pkg/core"
)

type AlertInput interface {
	// Manage AlertSource
	CreateAlertSource(ctx_core core.Context, source *alert.AlertSource) error
	GetAlertSource(ctx_core core.Context, sourceId string) (*alert.AlertSource, error)
	UpdateAlertSource(ctx_core core.Context, alertSource *alert.AlertSource) error
	DeleteAlertSource(ctx_core core.Context, alertSource alert.SourceFrom) (*alert.AlertSource, error)
	ListAlertSource(ctx_core core.Context,) ([]alert.AlertSource, error)

	// Manage AlertEnrichRule
	AddAlertEnrichRule(ctx_core core.Context, enrichRule []alert.AlertEnrichRule) error
	AddAlertEnrichConditions(ctx_core core.Context, enrichConditions []alert.AlertEnrichCondition) error
	AddAlertEnrichSchemaTarget(ctx_core core.Context, enrichSchemaTarget []alert.AlertEnrichSchemaTarget) error
	GetAlertEnrichRule(ctx_core core.Context, sourceId string) ([]alert.AlertEnrichRule, error)
	GetAlertEnrichConditions(ctx_core core.Context, sourceId string) ([]alert.AlertEnrichCondition, error)
	GetAlertEnrichSchemaTarget(ctx_core core.Context, sourceId string) ([]alert.AlertEnrichSchemaTarget, error)
	DeleteAlertEnrichRule(ctx_core core.Context, ruleIds []string) error
	DeleteAlertEnrichRuleBySourceId(ctx_core core.Context, sourceId string) error
	DeleteAlertEnrichConditions(ctx_core core.Context, ruleIds []string) error
	DeleteAlertEnrichConditionsBySourceId(ctx_core core.Context, sourceId string) error
	DeleteAlertEnrichSchemaTarget(ctx_core core.Context, ruleIds []string) error
	DeleteAlertEnrichSchemaTargetBySourceId(ctx_core core.Context, sourceId string) error

	// Manage schema
	CreateSchema(ctx_core core.Context, schema string, columns []string) error
	DeleteSchema(ctx_core core.Context, schema string) error
	CheckSchemaIsUsed(ctx_core core.Context, schema string) ([]string, error)
	ListSchema(ctx_core core.Context,) ([]string, error)
	ListSchemaColumns(ctx_core core.Context, schema string) ([]string, error)
	InsertSchemaData(ctx_core core.Context, schema string, columns []string, fullRows [][]string) error
	GetSchemaData(ctx_core core.Context, schema string) ([]string, map[int64][]string, error)
	UpdateSchemaData(ctx_core core.Context, schema string, columns []string, rows map[int][]string) error
	ClearSchemaData(ctx_core core.Context, schema string) error
	SearchSchemaTarget(ctx_core core.Context, schema string, sourceField string, sourceValue string, targets []alert.AlertEnrichSchemaTarget) ([]string, error)

	ListAlertTargetTags(ctx_core core.Context, lang string) ([]alert.TargetTag, error)

	// Load complate alertEnrichRule
	LoadAlertEnrichRule(ctx_core core.Context,) ([]alert.AlertSource, map[alert.SourceFrom][]alert.AlertEnrichRuleVO, error)

	GetAlertSlience(ctx_core core.Context,) ([]sc.AlertSlienceConfig, error)
	AddAlertSlience(ctx_core core.Context, slience *sc.AlertSlienceConfig) error
	UpdateAlertSlience(ctx_core core.Context, slience *sc.AlertSlienceConfig) error
	DeleteAlertSlience(ctx_core core.Context, id int) error
}

type subRepo struct {
	db *gorm.DB
}

func NewAlertInputRepo(db *gorm.DB, cfg *config.Config) (*subRepo, error) {
	repo := &subRepo{db}

	if err := repo.db.AutoMigrate(
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
