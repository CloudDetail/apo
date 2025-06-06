// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package core

import "github.com/gin-gonic/gin"

func EmptyCtx() Context {
	ctx := &context{ctx: &gin.Context{}}
	// ctx.Set(UserIDKey, 0)
	return ctx
}
