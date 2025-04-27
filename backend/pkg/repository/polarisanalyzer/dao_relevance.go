// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package polarisanalyzer

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

const (
	RelevanceSortAPI = "/relevance/sort"
)

// SortDescendantByRelevance
func (p *polRepo) SortDescendantByRelevance(
	startTime int64,
	endTime int64,
	stepStr string,
	targetService string,
	targetEndpoint string,
	descendants []ServiceNode,
	sortBy string,
) (sortResp *RelevanceResponse, err error) {
	if len(sortBy) == 0 {
		sortBy = "latency"
	}
	sortRequest := &SortRelevanceRequest{
		StartTime: startTime,
		EndTime:   endTime,
		StepStr:   stepStr,
		Target: Target{
			Service:  targetService,
			Endpoint: targetEndpoint,
		},
		UnsortedDescendant: descendants,
		SortBy:             sortBy,
	}

	body, err := json.Marshal(sortRequest)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s%s", polarisAnalyzerAddress, RelevanceSortAPI), bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to sort relevance, polarisanalyzer response status code: " + resp.Status)
	}

	relevanceRes := &RelevanceResponse{}
	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err = byteUnmarshallingValidator.ValidateAndUnmarshalJSON(respBytes, relevanceRes); err != nil {
		return nil, err
	}

	return relevanceRes, nil
}

type SortRelevanceRequest struct {
	StartTime          int64         `json:"startTime"`
	EndTime            int64         `json:"endTime"`
	StepStr            string        `json:"stepStr"`
	Target             Target        `json:"target"`
	UnsortedDescendant []ServiceNode `json:"unsortedDescendant"`
	SortBy             string        `json:"sortBy"`
}

type Target struct {
	Service  string `json:"service"`
	Endpoint string `json:"endpoint"`
}

type RelevanceResponse struct {
	SortedDescendant   []Relevance `json:"sortedDescendant"`
	UnsortedDescendant []Relevance `json:"unsortedDescendant"`
	DistanceType       string      `json:"distanceType"`
}

type ServiceNode struct {
	Service  string `json:"service"`
	Endpoint string `json:"endpoint"`
	Group    string `json:"group"`
	System   string `json:"system"`
}

type Relevance struct {
	ServiceNode
	Relevance float64 `json:"relevance,omitempty"`
}
