// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"errors"
	"net/http"
	"strings"
	"sync"

	"github.com/CloudDetail/apo/backend/pkg/code"

	go_context "context"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

const (
	_PayloadName    = "_payload_"
	_AbortErrorName = "_abort_error_"
)

var contextPool = &sync.Pool{
	New: func() interface{} {
		return new(context)
	},
}

func newContext(ctx *gin.Context) Context {
	context := contextPool.Get().(*context)
	context.ctx = ctx
	return context
}

func releaseContext(ctx Context) {
	c := ctx.(*context)
	c.ctx = nil
	contextPool.Put(c)
}

var _ Context = (*context)(nil)

type Context interface {
	// ShouldBindQuery deserialization querystring
	// tag: `form:"xxx"`(note: do not write query)
	ShouldBindQuery(obj interface{}) error

	// ShouldBindPostForm deserialization postform (querystring will be ignored)
	// tag: `form:"xxx"`
	ShouldBindPostForm(obj interface{}) error

	// ShouldBindJSON deserialization postjson
	// tag: `json:"xxx"`
	ShouldBindJSON(obj interface{}) error

	// ShouldBindURI the deserialization path parameter (if the routing path is/user/:name)
	// tag: `uri:"xxx"`
	ShouldBindURI(obj interface{}) error

	Param(key string) string

	// Payload returned correctly
	GetMethodPath() (method string, path string)

	// Payload 正确返回
	Payload(payload interface{})
	getPayload() interface{}

	// AbortWithError error return
	AbortWithPermissionError(err error, expectCode string, emptyResp any)
	// AbortWithError 错误返回

	// c.AbortWithError(
	// 			http.StatusBadRequest,
	// 			code.ParamBindError,
	// 			err,
	// 		)
	// AbortWithError(err BusinessError)
	// If err is BusinessError, param businessCode will be ignored
	AbortWithError(statusCode int, businessCode string, err error)

	abortError() BusinessError

	// Header Gets the Header object
	Header() http.Header
	// Get the header GetHeader
	GetHeader(key string) string
	// Set the header SetHeader
	SetHeader(key, value string)

	GetContext() go_context.Context
	// Set Sets key-value pairs in context
	Set(key string, value interface{})
	// Get Gets the value of the key in context.
	Get(key string) (any, bool)

	Next()
	Request() *http.Request

	ClientIP() string

	LANG() string
	UserID() int64

	ErrMessage(errCode string) string
}

type context struct {
	ctx *gin.Context
}

// ShouldBindQuery deserialization querystring
// tag: `form:"xxx"`(note: do not write query)
func (c *context) ShouldBindQuery(obj interface{}) error {
	return c.ctx.ShouldBindWith(obj, binding.Query)
}

// ShouldBindPostForm deserialization postform (querystring will be ignored)
// tag: `form:"xxx"`
func (c *context) ShouldBindPostForm(obj interface{}) error {
	return c.ctx.ShouldBindWith(obj, binding.FormPost)
}

// ShouldBindJSON deserialization postjson
// tag: `json:"xxx"`
func (c *context) ShouldBindJSON(obj interface{}) error {
	return c.ctx.ShouldBindWith(obj, binding.JSON)
}

// ShouldBindURI the deserialization path parameter (if the routing path is/user/:name)
// tag: `uri:"xxx"`
func (c *context) ShouldBindURI(obj interface{}) error {
	return c.ctx.ShouldBindUri(obj)
}

func (c *context) Param(key string) string {
	return c.ctx.Param(key)
}

func (c *context) getPayload() interface{} {
	if payload, ok := c.ctx.Get(_PayloadName); ok {
		return payload
	}
	return nil
}

func (c *context) Payload(payload interface{}) {
	c.ctx.Set(_PayloadName, payload)
}

func (c *context) Header() http.Header {
	header := c.ctx.Request.Header

	clone := make(http.Header, len(header))
	for k, v := range header {
		value := make([]string, len(v))
		copy(value, v)

		clone[k] = value
	}
	return clone
}

func (c *context) GetMethodPath() (method string, path string) {
	m := c.ctx.Request.Method
	path = c.ctx.FullPath()
	return m, path
}

func (c *context) GetHeader(key string) string {
	return c.ctx.GetHeader(key)
}

func (c *context) SetHeader(key, value string) {
	c.ctx.Header(key, value)
}

func (c *context) AbortWithPermissionError(err error, expectCode string, emptyResp any) {
	if emptyResp == nil {
		c.AbortWithError(http.StatusBadRequest, expectCode, err)
		return
	}

	var vErr *businessError
	if !errors.As(err, &vErr) {
		c.AbortWithError(http.StatusBadRequest, expectCode, err)
		return
	}
	if vErr.businessCode != code.GroupNoDataError {
		c.AbortWithError(http.StatusBadRequest, expectCode, err)
		return
	}
	c.Payload(emptyResp)
}

func (c *context) AbortWithError(statusCode int, commonCode string, err error) {
	if err == nil {
		return
	}

	if statusCode == 0 {
		statusCode = http.StatusInternalServerError
	}
	c.ctx.AbortWithStatus(statusCode)

	var vErr *businessError
	if errors.As(err, &vErr) {
		vErr.httpCode = statusCode
		if len(vErr.message) == 0 {
			vErr.message = c.ErrMessage(vErr.BusinessCode())
		}
		c.ctx.Set(_AbortErrorName, vErr.WithStack(err))
	} else {
		c.ctx.Set(_AbortErrorName, Error(commonCode, c.ErrMessage(commonCode)).WithStack(err))
	}
}

func (c *context) abortError() BusinessError {
	err, _ := c.ctx.Get(_AbortErrorName)
	return err.(BusinessError)
}

func (c *context) GetContext() go_context.Context {
	return c.ctx
}

func (c *context) Set(key string, value interface{}) { c.ctx.Set(key, value) }

func (c *context) Get(key string) (any, bool) {
	return c.ctx.Get(key)
}

func (c *context) Next() { c.ctx.Next() }

func (c *context) Abort() { c.ctx.Abort() }

func (c *context) Request() *http.Request {
	return c.ctx.Request
}

func (c *context) ClientIP() string {
	return c.ctx.ClientIP()
}

func (c *context) LANG() string {
	if c == nil {
		return code.LANG_ZH
	}
	lang := code.LANG_EN

	if c.ctx.Request != nil {
		lang = c.GetHeader("APO-Language")
	}

	if strings.HasPrefix(strings.ToLower(lang), code.LANG_EN) {
		return code.LANG_EN
	}
	return code.LANG_ZH
}

func (c *context) ErrMessage(errCode string) string {
	return code.Text(c.LANG(), errCode)
}

func (c *context) UserID() int64 {
	userID, ok := c.Get(UserIDKey)
	if !ok {
		return 0
	}
	id, ok := userID.(int64)
	if !ok {
		return 0
	}
	return id
}
