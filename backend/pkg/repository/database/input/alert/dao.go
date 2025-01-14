// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	"github.com/CloudDetail/apo/backend/config"
	input "github.com/CloudDetail/apo/backend/pkg/model/input/alert"
	"gorm.io/gorm"
)

type AlertInput interface {
	// Manage Cluster
	CreateCluster(cluster *input.Cluster) error
	UpdateCluster(cluster *input.Cluster) error
	DeleteCluster(cluster *input.Cluster) error
	ListCluster() ([]input.Cluster, error)

	// Manage AlertSource
	CreateAlertSource(*input.AlertSource) error
	GetAlertSource(sourceId string) (*input.AlertSource, error)
	UpdateAlertSource(alertSource *input.AlertSource) error
	DeleteAlertSource(alertSource input.SourceFrom) (*input.AlertSource, error)
	ListAlertSource() ([]input.AlertSource, error)

	// Manage AlertEnrichRule
	AddAlertEnrichRule(enrichRule []input.AlertEnrichRule) error
	AddAlertEnrichConditions(enrichConditions []input.AlertEnrichCondition) error
	AddAlertEnrichSchemaTarget(enrichSchemaTarget []input.AlertEnrichSchemaTarget) error
	GetAlertEnrichRule(sourceId string) ([]input.AlertEnrichRule, error)
	GetAlertEnrichConditions(sourceId string) ([]input.AlertEnrichCondition, error)
	GetAlertEnrichSchemaTarget(sourceId string) ([]input.AlertEnrichSchemaTarget, error)
	DeleteAlertEnrichRule(ruleIds []string) error
	DeleteAlertEnrichRuleBySourceId(sourceId string) error
	DeleteAlertEnrichConditions(ruleIds []string) error
	DeleteAlertEnrichConditionsBySourceId(sourceId string) error
	DeleteAlertEnrichSchemaTarget(ruleIds []string) error
	DeleteAlertEnrichSchemaTargetBySourceId(sourceId string) error

	// Manage schema
	CreateSchema(schema string, columns []string) error
	DeleteSchema(string) error
	CheckSchemaIsUsed(schema string) ([]string, error)
	ListSchema() ([]string, error)
	ListSchemaColumns(schema string) ([]string, error)
	InsertSchemaData(schema string, columns []string, fullRows [][]string) error
	GetSchemaData(schema string) ([]string, map[int64][]string, error)
	UpdateSchemaData(schema string, columns []string, rows map[int][]string) error
	ClearSchemaData(schema string) error
	SearchSchemaTarget(schema string, sourceField string, sourceValue string, targets []input.AlertEnrichSchemaTarget) ([]string, error)

	ListAlertTargetTags() ([]input.TargetTag, error)

	// Load complate alertEnrichRule
	LoadAlertEnrichRule() ([]input.AlertSource, map[input.SourceFrom][]input.AlertEnrichRuleVO, error)
}

type subRepo struct {
	db *gorm.DB
}

func NewAlertInputRepo(db *gorm.DB, cfg *config.Config) (*subRepo, error) {
	repo := &subRepo{db}
	err := repo.initDefaultAlertTagMapping(
		cfg.Database.InitScript.DefaultAlertTagMapping)

	return repo, err
}
