// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	sc "github.com/CloudDetail/apo/backend/pkg/model/amconfig/slienceconfig"
	dbdriver "github.com/CloudDetail/apo/backend/pkg/repository/database/driver"

	"github.com/CloudDetail/apo/backend/config"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	"gorm.io/gorm"
)

type AlertInput interface {
	// Manage AlertSource
	CreateAlertSource(source *alert.AlertSource) error
	GetAlertSource(sourceId string) (*alert.AlertSource, error)
	UpdateAlertSource(alertSource *alert.AlertSource) error
	DeleteAlertSource(alertSource alert.SourceFrom) (*alert.AlertSource, error)
	ListAlertSource() ([]alert.AlertSource, error)

	// Manage AlertEnrichRule
	AddAlertEnrichRule(enrichRule []alert.AlertEnrichRule) error
	AddAlertEnrichConditions(enrichConditions []alert.AlertEnrichCondition) error
	AddAlertEnrichSchemaTarget(enrichSchemaTarget []alert.AlertEnrichSchemaTarget) error
	GetAlertEnrichRule(sourceId string) ([]alert.AlertEnrichRule, error)
	GetAlertEnrichConditions(sourceId string) ([]alert.AlertEnrichCondition, error)
	GetAlertEnrichSchemaTarget(sourceId string) ([]alert.AlertEnrichSchemaTarget, error)
	DeleteAlertEnrichRule(ruleIds []string) error
	DeleteAlertEnrichRuleBySourceId(sourceId string) error
	DeleteAlertEnrichConditions(ruleIds []string) error
	DeleteAlertEnrichConditionsBySourceId(sourceId string) error
	DeleteAlertEnrichSchemaTarget(ruleIds []string) error
	DeleteAlertEnrichSchemaTargetBySourceId(sourceId string) error

	// Manage schema
	CreateSchema(schema string, columns []string) error
	DeleteSchema(schema string) error
	CheckSchemaIsUsed(schema string) ([]string, error)
	ListSchema() ([]string, error)
	ListSchemaColumns(schema string) ([]string, error)
	InsertSchemaData(schema string, columns []string, fullRows [][]string) error
	GetSchemaData(schema string) ([]string, map[int64][]string, error)
	UpdateSchemaData(schema string, columns []string, rows map[int][]string) error
	ClearSchemaData(schema string) error
	SearchSchemaTarget(schema string, sourceField string, sourceValue string, targets []alert.AlertEnrichSchemaTarget) ([]string, error)

	ListAlertTargetTags(lang string) ([]alert.TargetTag, error)

	// Load complate alertEnrichRule
	LoadAlertEnrichRule() ([]alert.AlertSource, map[alert.SourceFrom][]alert.AlertEnrichRuleVO, error)

	GetAlertSlience() ([]sc.AlertSlienceConfig, error)
	AddAlertSlience(slience *sc.AlertSlienceConfig) error
	UpdateAlertSlience(slience *sc.AlertSlienceConfig) error
	DeleteAlertSlience(id int) error
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
