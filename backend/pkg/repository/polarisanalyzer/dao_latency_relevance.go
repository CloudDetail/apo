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

// SortDescendantByLatencyRelevance
func (p *polRepo) SortDescendantByLatencyRelevance(
	startTime int64,
	endTime int64,
	stepStr string,
	targetService string,
	targetEndpoint string,
	descendants []LatencyRelevance,
) (sorted []LatencyRelevance, unsorted []LatencyRelevance, err error) {
	sortRequest := &SortRelevanceRequest{
		StartTime: startTime,
		EndTime:   endTime,
		StepStr:   stepStr,
		Target: Target{
			Service:  targetService,
			Endpoint: targetEndpoint,
		},
		UnsortedDescendant: descendants,
	}

	body, err := json.Marshal(sortRequest)
	if err != nil {
		return nil, descendants, err
	}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s%s", polarisAnalyzerAddress, RelevanceSortAPI), bytes.NewBuffer(body))
	if err != nil {
		return nil, descendants, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, descendants, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, descendants, errors.New("failed to sort relevance, polarisanalyzer response status code: " + resp.Status)
	}

	relevanceRes := &LatencyRelevanceResponse{}
	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, descendants, err
	}

	err = json.Unmarshal(respBytes, relevanceRes)
	if err != nil {
		return nil, descendants, err
	}

	return relevanceRes.SortedDescendant, relevanceRes.UnsortedDescendant, nil
}

type SortRelevanceRequest struct {
	StartTime          int64              `json:"startTime"`
	EndTime            int64              `json:"endTime"`
	StepStr            string             `json:"stepStr"`
	Target             Target             `json:"target"`
	UnsortedDescendant []LatencyRelevance `json:"unsortedDescendant"`
}

type Target struct {
	Service  string `json:"service"`
	Endpoint string `json:"endpoint"`
}

type LatencyRelevanceResponse struct {
	SortedDescendant   []LatencyRelevance `json:"sortedDescendant"`
	UnsortedDescendant []LatencyRelevance `json:"unsortedDescendant"`
}

type LatencyRelevance struct {
	Service   string  `json:"service"`
	Endpoint  string  `json:"endpoint"`
	Relevance float64 `json:"relevance"`
}
