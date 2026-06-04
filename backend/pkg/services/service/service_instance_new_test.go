// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"testing"

	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func TestAppendInstanceDataByNamePrefersRicherDuplicate(t *testing.T) {
	empty := response.InstanceData{
		Name:     "guoxin_app@FIN-guoxin-web-6@1136429",
		NodeName: "FIN-guoxin-web-6",
		Latency: response.TempChartObject{
			ChartData: map[int64]float64{1779998052000000: 0},
		},
		Tps: response.TempChartObject{
			ChartData: map[int64]float64{1779998052000000: 0},
		},
	}
	latencyValue := 4915360.0
	tpsValue := 0.034722222222222224
	richer := response.InstanceData{
		Name:     empty.Name,
		NodeName: empty.NodeName,
		Latency: response.TempChartObject{
			Value:     &latencyValue,
			ChartData: map[int64]float64{1780012452000000: 3834000},
		},
		Tps: response.TempChartObject{
			Value:     &tpsValue,
			ChartData: map[int64]float64{1780012452000000: 0.11428571428571428},
		},
	}

	var list []response.InstanceData
	indexByName := make(map[string]int)
	list = appendInstanceDataByName(list, indexByName, empty)
	list = appendInstanceDataByName(list, indexByName, richer)

	if len(list) != 1 {
		t.Fatalf("expected one deduplicated instance, got %d", len(list))
	}
	if list[0].Latency.Value == nil || *list[0].Latency.Value != latencyValue {
		t.Fatalf("expected richer duplicate to replace empty instance, got %#v", list[0].Latency.Value)
	}
}

func TestAppendInstanceDataByNameKeepsDifferentNames(t *testing.T) {
	list := []response.InstanceData{}
	indexByName := make(map[string]int)

	list = appendInstanceDataByName(list, indexByName, response.InstanceData{Name: "svc@node-a@1"})
	list = appendInstanceDataByName(list, indexByName, response.InstanceData{Name: "svc@node-b@2"})

	if len(list) != 2 {
		t.Fatalf("expected different instances to be preserved, got %d", len(list))
	}
}
