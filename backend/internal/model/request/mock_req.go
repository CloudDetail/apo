// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package request

type CreateRequest struct {
	Name string `form:"name" binding:"required"` // 名称
}

type ListRequest struct {
	PageNum  int    `form:"page_num" binding:"required"` // 第几页
	PageSize int    `form:"page_size"`                   // 每页显示条数
	Name     string `form:"name"`                        // 用户名
}

type DetailRequest struct {
	Id uint `uri:"id" binding:"required"` // ID
}

type DeleteRequest struct {
	Id uint `uri:"id" binding:"required"` // ID
}
