// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package polarisanalyzer

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

const (
	PolarisInferAPI = "/analyze"
)

// QueryPolarisInfer implements Repo.
func (p *polRepo) QueryPolarisInfer(
	startTime, endTime int64, stepStr string,
	service, endpoint string,
) (*PolarisInferRes, error) {

	params := url.Values{}
	params.Add("startTime", strconv.FormatInt(startTime, 10))
	params.Add("endTime", strconv.FormatInt(endTime, 10))
	params.Add("stepStr", stepStr)
	params.Add("service", service)
	params.Add("endpoint", endpoint)
	fullUrl := fmt.Sprintf("%s%s?%s", polarisAnalyzerAddress, PolarisInferAPI, params.Encode())
	req, err := http.NewRequest("GET", fullUrl, nil)
	if err != nil {
		return &PolarisInferRes{}, err
	}
	// Send http request
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return &PolarisInferRes{}, err
	}
	defer res.Body.Close()

	// parse json data from res body
	var inferRes PolarisInferRes
	err = json.NewDecoder(res.Body).Decode(&inferRes)
	if err != nil {
		return &PolarisInferRes{}, err
	}
	return &inferRes, nil
}

type PolarisInferRes struct {
	InferMetricsPng string `json:"inferMetricsPng"`
	InferCause      string `json:"inferCause"`
}
