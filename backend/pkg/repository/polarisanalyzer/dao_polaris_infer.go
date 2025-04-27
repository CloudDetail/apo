// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package polarisanalyzer

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/util"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
	prom "github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
)

const (
	PolarisInferAPI = "/analyze"
)

var byteUnmarshallingValidator = util.NewByteValidator(1024*1024, []string{}, []string{"$func", "$eval", "constructor", "prototype"}, 10)

// QueryPolarisInfer implements Repo.
func (p *polRepo) QueryPolarisInfer(req *request.GetPolarisInferRequest) (*PolarisInferRes, error) {

	params := url.Values{}
	params.Add("startTime", strconv.FormatInt(req.StartTime, 10))
	params.Add("endTime", strconv.FormatInt(req.EndTime, 10))
	params.Add("stepStr", prom.VecFromDuration(time.Duration(req.Step)*time.Microsecond))
	params.Add("service", req.Service)
	params.Add("endpoint", req.Endpoint)

	params.Add("language", req.Language)
	params.Add("timezone", req.Timezone)

	fullUrl := fmt.Sprintf("%s%s?%s", polarisAnalyzerAddress, PolarisInferAPI, params.Encode())
	request, err := http.NewRequest("GET", fullUrl, nil)
	if err != nil {
		return &PolarisInferRes{}, err
	}
	// Send http request

	// request.Header.Add("Accept-Language", req.Lanaguage)
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
	if err = byteUnmarshallingValidator.ValidateAndUnmarshalJSON(respBytes, inferRes); err != nil {
		return nil, err
	}
	return inferRes, nil
}

type PolarisInferRes struct {
	InferMetricsPng string `json:"inferMetricsPng"`
	InferCause      string `json:"inferCause"`
}
