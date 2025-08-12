// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package dataplane

import (
	"net/http"
	"time"

	"github.com/CloudDetail/apo/backend/config"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
)

type DataplaneRepo interface {
	CheckDataSource(request *model.CheckDataSourceRequest) (*model.CheckDataSourceResponse, error)
	QueryServices(providerId int, startTime int64, endTime int64) ([]*model.Service, error)
	QueryApmServiceInstances(providerId int, startTime int64, endTime int64, service *model.Service) ([]*model.ApmServiceInstance, error)
	QueryServiceRedCharts(providerId int, startTime int64, endTime int64, service *model.Service, step int64) (map[string]*model.RedCharts, error)
	QueryServiceRedValue(providerId int, startTime int64, endTime int64, service *model.Service) (*model.RedMetricValue, error)
	QueryServiceToplogy(providerId int, startTime int64, endTime int64, clusterId string, datasource string) ([]*model.ServiceToplogy, error)
}

type dataplaneRepo struct {
	address string
	client  *http.Client
	ch      clickhouse.Repo
	db      database.Repo
}

func New(ch clickhouse.Repo, db database.Repo) (DataplaneRepo, error) {
	dataplaneConf := config.Get().Dataplane
	return &dataplaneRepo{
		address: dataplaneConf.Address,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		ch: ch,
		db: db,
	}, nil
}
