// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package model

type Pagination struct {
	Total       int64 `json:"total"`       // 总记录数
	CurrentPage int   `json:"currentPage"` // 当前页码
	PageSize    int   `json:"pageSize"`    // 每页条数
}

type I18nLanguage struct {
	Language string `json:"language" form:"language"` // I18n language
}
