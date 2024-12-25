// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package polarisanalyzer

import (
	"net/http"
	"os"
	"time"
)

var polarisAnalyzerAddress = "http://localhost:5000"

type Repo interface {
	// SortDescendantByRelevance 查询依赖节点延时关联度
	SortDescendantByRelevance(
		startTime, endTime int64, stepStr string,
		targetService, targetEndpoint string,
		unsortedDescendant []ServiceNode, sortBy string,
	) (sortResp *RelevanceResponse, err error)

	QueryPolarisInfer(
		startTime, endTime int64, stepStr string,
		service, endpoint string,
	) (*PolarisInferRes, error)
}

func New() (Repo, error) {
	if value, find := os.LookupEnv("POLARIS_ANALYZER_ADDRESS"); find {
		polarisAnalyzerAddress = value
	}

	return &polRepo{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

type polRepo struct {
	client *http.Client
}
