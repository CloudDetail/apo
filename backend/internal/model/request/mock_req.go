// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package request

type CreateRequest struct {
	Name string `form:"name" binding:"required"` // name
}

type ListRequest struct {
	PageNum  int    `form:"page_num" binding:"required"` // page
	PageSize int    `form:"page_size"`                   // display number per page
	Name     string `form:"name"`                        // username
}

type DetailRequest struct {
	Id uint `uri:"id" binding:"required"` // ID
}

type DeleteRequest struct {
	Id uint `uri:"id" binding:"required"` // ID
}
