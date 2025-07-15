// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package polarisanalyzer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/util"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
	prom "github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
)

const (
	PolarisInferAPI = "/analyze"
)

// QueryPolarisInfer implements Repo.
func (p *polRepo) QueryPolarisInfer(req *request.GetPolarisInferRequest) (*PolarisInferRes, error) {
	if req.Step < 60e6 {
		req.Step = 60e6 // interval must be large than 1m
	}

	payload := PolarisInferReq{
		StartTime:  req.StartTime,
		EndTime:    req.EndTime,
		StepStr:    prom.VecFromDuration(time.Duration(req.Step) * time.Microsecond),
		Service:    req.Service,
		Endpoint:   req.Endpoint,
		Language:   req.Language,
		Timezone:   req.Timezone,
		ClusterIDs: req.ClusterIDs,
	}

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	fullUrl := fmt.Sprintf("%s%s", polarisAnalyzerAddress, PolarisInferAPI)
	request, err := http.NewRequest("POST", fullUrl, bytes.NewReader(jsonBytes))
	if err != nil {
		return &PolarisInferRes{}, err
	}
	// Send http request

	request.Header.Add("Content-Type", "application/json")
	// request.Header.Add("X-Timezone", req.Timezone)

	res, err := http.DefaultClient.Do(request)
	if err != nil {
		return &PolarisInferRes{}, err
	}
	defer res.Body.Close()

	// parse json data from res body
	inferRes := &PolarisInferRes{}
	respBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	validateBody, ok := util.ValidateResponseBytes(respBytes)
	if !ok {
		return nil, fmt.Errorf("reponse body is invalid")
	}
	if err = json.Unmarshal(validateBody, inferRes); err != nil {
		return nil, err
	}
	return inferRes, nil
}

type PolarisInferReq struct {
	StartTime int64  `json:"startTime"`
	EndTime   int64  `json:"endTime"`
	StepStr   string `json:"stepStr"`
	Service   string `json:"service"`
	Endpoint  string `json:"endpoint"`

	Language string `json:"language"`
	Timezone string `json:"timezone"`

	ClusterIDs []string `json:"clusterIds"`
}

type PolarisInferRes struct {
	InferMetricsPng string `json:"inferMetricsPng"`
	InferCause      string `json:"inferCause"`
}
