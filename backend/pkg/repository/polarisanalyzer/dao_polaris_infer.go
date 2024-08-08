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
	params.Add("startTime", strconv.Itoa(int(startTime)))
	params.Add("endTime", strconv.Itoa(int(endTime)))
	params.Add("stepStr", stepStr)
	params.Add("service", service)
	params.Add("endpoint", endpoint)
	fullUrl := fmt.Sprintf("%s%s?%s", polarisAnalyzerAddress, PolarisInferAPI, params.Encode())
	req, err := http.NewRequest("GET", fullUrl, nil)
	if err != nil {
		return &PolarisInferRes{}, err
	}
	// 发送http请求
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return &PolarisInferRes{}, err
	}
	defer res.Body.Close()

	// 从res body中解析json数据
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
