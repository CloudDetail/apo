// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package response

type CreateResponse struct {
	Id uint `json:"id"` // primary key ID
}

type ListResponse struct {
	List       []*ListData `json:"list"`
	Pagination *Pagination `json:"pagination"`
}

type ListData struct {
	Id   int    `json:"id"`   // ID
	Name string `json:"name"` // username
}

type Pagination struct {
	Total        int64 `json:"total"`          // total number of records
	CurrentPage  int   `json:"current_page"`   // current page number
	PerPageCount int   `json:"per_page_count"` // number of pieces per page
}

type DetailResponse struct {
	Id   uint   `json:"id"`
	Name string `json:"name"`
}

type DeleteResponse struct {
	Id uint `json:"id"` // primary key ID
}
