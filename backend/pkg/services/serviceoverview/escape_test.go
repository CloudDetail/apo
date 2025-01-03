// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package serviceoverview

import (
	"fmt"
	"github.com/CloudDetail/apo/backend/config"
	"github.com/CloudDetail/apo/backend/pkg/logger"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
	"github.com/magiconair/properties/assert"
	"os"
	"testing"
	"time"
)

func init() {
	os.Setenv("APO_CONFIG", "../../../config/apo.yml")
}

func getRepo() (prometheus.Repo, error) {
	cfg := config.Get().Promethues
	zapLog := logger.NewLogger(logger.WithLevel("debug"))
	return prometheus.New(zapLog, cfg.Address, cfg.Storage)
}

func TestMultiNamespaceEscape(t *testing.T) {
	repo, err := getRepo()
	if err != nil {
		t.Fatalf(err.Error())
	}
	filter := EndpointsFilter{
		MultiService: []string{"foo", "foo|bar"},
	}

	filterStr := filter.ExtractFilterStr()
	query := fmt.Sprintf("last_over_time(test_metric{%s\"%s\"}[24h])", filterStr[0], filterStr[1])
	data, err := repo.QueryData(time.Now(), query)
	if err != nil {
		t.Fatalf(err.Error())
	}

	assert.Equal(t, len(data), 2)
}

func TestContainsSvcEscape(t *testing.T) {
	repo, err := getRepo()
	if err != nil {
		t.Fatalf(err.Error())
	}
	filter := EndpointsFilter{
		ContainsSvcName: "foo|bar",
	}

	filterStr := filter.ExtractFilterStr()
	query := fmt.Sprintf("last_over_time(test_metric{%s\"%s\"}[24h])", filterStr[0], filterStr[1])
	data, err := repo.QueryData(time.Now(), query)
	if err != nil {
		t.Fatalf(err.Error())
	}

	assert.Equal(t, len(data), 1)
}

func TestNamespaceEscape(t *testing.T) {
	repo, err := getRepo()
	if err != nil {
		t.Fatalf(err.Error())
	}
	filter := EndpointsFilter{
		Namespace: "foo|bar",
	}

	filterStr := filter.ExtractFilterStr()
	query := fmt.Sprintf("last_over_time(test_metric{%s\"%s\"}[24h])", filterStr[0], filterStr[1])
	data, err := repo.QueryData(time.Now(), query)
	if err != nil {
		t.Fatalf(err.Error())
	}

	assert.Equal(t, len(data), 1)
}

func TestServiceEscape(t *testing.T) {
	repo, err := getRepo()
	if err != nil {
		t.Fatalf(err.Error())
	}
	filter := EndpointsFilter{
		ServiceName: "foo|bar",
	}

	filterStr := filter.ExtractFilterStr()
	query := fmt.Sprintf("last_over_time(test_metric{%s\"%s\"}[24h])", filterStr[0], filterStr[1])
	data, err := repo.QueryData(time.Now(), query)
	if err != nil {
		t.Fatalf(err.Error())
	}

	assert.Equal(t, len(data), 1)
}

func TestMultiServiceEscape(t *testing.T) {
	repo, err := getRepo()
	if err != nil {
		t.Fatalf(err.Error())
	}
	filter := EndpointsFilter{
		MultiService: []string{"foo|bar", "foo"},
	}

	filterStr := filter.ExtractFilterStr()
	query := fmt.Sprintf("last_over_time(test_metric{%s\"%s\"}[24h])", filterStr[0], filterStr[1])
	data, err := repo.QueryData(time.Now(), query)
	if err != nil {
		t.Fatalf(err.Error())
	}

	assert.Equal(t, len(data), 2)
}

func TestEndpointEscape(t *testing.T) {
	repo, err := getRepo()
	if err != nil {
		t.Fatalf(err.Error())
	}
	filter := EndpointsFilter{
		ContainsEndpointName: "foo|bar",
	}

	filterStr := filter.ExtractFilterStr()
	query := fmt.Sprintf("last_over_time(test_metric{%s\"%s\"}[24h])", filterStr[0], filterStr[1])
	data, err := repo.QueryData(time.Now(), query)
	if err != nil {
		t.Fatalf(err.Error())
	}

	assert.Equal(t, len(data), 1)
}

func TestMultiEndpointEscape(t *testing.T) {
	repo, err := getRepo()
	if err != nil {
		t.Fatalf(err.Error())
	}
	filter := EndpointsFilter{
		MultiEndpoint: []string{"foo|bar", "foo"},
	}

	filterStr := filter.ExtractFilterStr()
	query := fmt.Sprintf("last_over_time(test_metric{%s\"%s\"}[24h])", filterStr[0], filterStr[1])
	data, err := repo.QueryData(time.Now(), query)
	if err != nil {
		t.Fatalf(err.Error())
	}

	assert.Equal(t, len(data), 2)
}
