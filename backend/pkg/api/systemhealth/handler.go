// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package systemhealth

import (
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/repository/dify"
	"github.com/CloudDetail/apo/backend/pkg/repository/kubernetes"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
)

type Handler interface {
	GetOverview() core.HandlerFunc
	GetNewAbnormalMetrics() core.HandlerFunc
	GetSystemDetail() core.HandlerFunc
	AnalyzePendingIssue() core.HandlerFunc
	GetReport() core.HandlerFunc
	ListReportRecords() core.HandlerFunc
	GetReportRecord() core.HandlerFunc
}

type handler struct {
	chRepo                    clickhouse.Repo
	dbRepo                    database.Repo
	difyRepo                  dify.DifyRepo
	k8sRepo                   kubernetes.Repo
	promRepo                  prometheus.Repo
	reportCache               *systemHealthReportCache
	pendingIssueAnalysisCache *systemHealthPendingIssueAnalysisCache
	alertManagerURL           string
	vmAlertURL                string
}

func New(chRepo clickhouse.Repo, dbRepo database.Repo, difyRepo dify.DifyRepo, k8sRepo kubernetes.Repo, promRepo prometheus.Repo, alertManagerURL string, vmAlertURL string) Handler {
	return &handler{
		chRepo:                    chRepo,
		dbRepo:                    dbRepo,
		difyRepo:                  difyRepo,
		k8sRepo:                   k8sRepo,
		promRepo:                  promRepo,
		reportCache:               newSystemHealthReportCache(),
		pendingIssueAnalysisCache: newSystemHealthPendingIssueAnalysisCache(0, 512),
		alertManagerURL:           alertManagerURL,
		vmAlertURL:                vmAlertURL,
	}
}
