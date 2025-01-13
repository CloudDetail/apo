// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package polarisanalyzer

import (
	"net/http"
	"os"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

var polarisAnalyzerAddress = "http://localhost:5000"

type Repo interface {
	// SortDescendantByRelevance query dependency node latency correlation
	SortDescendantByRelevance(
		startTime, endTime int64, stepStr string,
		targetService, targetEndpoint string,
		unsortedDescendant []ServiceNode, sortBy string,
	) (sortResp *RelevanceResponse, err error)

	QueryPolarisInfer(req *request.GetPolarisInferRequest) (*PolarisInferRes, error)
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
