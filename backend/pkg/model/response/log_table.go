// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package response

type LogTableResponse struct {
	Sqls []string `json:"sqls"`
	Err  string   `json:"error"`
}
