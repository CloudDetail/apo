// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func WrapHandlerFunctions(rawFuncs ...http.HandlerFunc) []gin.HandlerFunc {
	funcs := make([]gin.HandlerFunc, len(rawFuncs))
	for i, f := range rawFuncs {
		funcs[i] = func(c *gin.Context) {
			f(c.Writer, c.Request)
		}
	}

	return funcs
}
