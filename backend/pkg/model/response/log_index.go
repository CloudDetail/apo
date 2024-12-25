// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package response

type LogIndexResponse struct {
	Indexs []IndexItem `json:"indexs"`
}

type IndexItem struct {
	IndexName string  `json:"indexName"`
	Count     uint64  `json:"count"`
	Percent   float64 `json:"percent"`
}
