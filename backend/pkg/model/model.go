// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package model

type Pagination struct {
	Total       int64 `json:"total"`       // total number of records
	CurrentPage int   `json:"currentPage"` // current page number
	PageSize    int   `json:"pageSize"`    // number of entries per page
}

type I18nLanguage struct {
	Language string `json:"language" form:"language"` // I18n language
}
