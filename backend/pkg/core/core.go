// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"runtime/debug"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/gin-gonic/gin"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type HandlerFunc func(c Context)

type Mux struct {
	engine *gin.Engine
}

func New(logger *zap.Logger) (*Mux, error) {
	if logger == nil {
		return nil, errors.New("logger required")
	}

	gin.SetMode(gin.ReleaseMode)
	mux := &Mux{
		engine: gin.New(),
	}

	// register swagger
	mux.engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.NewHandler()))

	mux.engine.Use(func(ctx *gin.Context) {
		if ctx.Writer.Status() == http.StatusNotFound {
			return
		}
		ts := time.Now()
		context := newContext(ctx)
		defer releaseContext(context)

		defer func() {
			var (
				httpCode        int
				businessCode    string
				businessCodeMsg string
				abortErr        error
			)

			// Panic occurs
			if err := recover(); err != nil {
				stackInfo := string(debug.Stack())
				logger.Error("got panic", zap.String("panic", fmt.Sprintf("%+v", err)), zap.String("stack", stackInfo))

				// BuisnessError return code is 500
				context.AbortWithError(Error(
					http.StatusInternalServerError,
					code.ServerError,
					code.Text(code.ServerError),
				))
			}

			// Error occurred, return
			if ctx.IsAborted() {
				for i := range ctx.Errors {
					multierr.AppendInto(&abortErr, ctx.Errors[i])
				}
				if err := context.abortError(); err != nil {
					// customer err
					multierr.AppendInto(&abortErr, err.StackError())
					httpCode = err.HTTPCode()
					businessCode = err.BusinessCode()
					businessCodeMsg = err.Message()
				} else {
					// There is no Error
					httpCode = http.StatusInternalServerError
					businessCode = code.ServerError
					businessCodeMsg = code.Text(code.ServerError)
				}
				ctx.JSON(httpCode, &code.Failure{
					Code:    businessCode,
					Message: businessCodeMsg,
				})
			} else {
				if len(ctx.GetHeader("X-Data-Flow")) > 0 {
					// No need to log debug for X-Data-Flow = Meta type data
					// At the same time, the response data is processed and returned inside the handler.
					return
				}

				// region returned correctly
				ctx.JSON(http.StatusOK, context.getPayload())
			}

			success := !ctx.IsAborted() && (ctx.Writer.Status() == http.StatusOK)
			decodedURL, _ := url.QueryUnescape(ctx.Request.URL.RequestURI())
			costSeconds := time.Since(ts).Seconds()
			if !success {
				logger.Error("query-request",
					zap.Any("method", ctx.Request.Method),
					zap.Any("path", decodedURL),
					zap.Any("http_code", ctx.Writer.Status()),
					zap.Any("business_code", businessCode),
					zap.Any("business_message", businessCodeMsg),
					zap.Any("cost_seconds", costSeconds),
					zap.Error(abortErr),
				)
			} else if ce := logger.Check(zapcore.DebugLevel, ""); ce != nil {
				logger.Debug("query-request",
					zap.Any("method", ctx.Request.Method),
					zap.Any("path", decodedURL),
					zap.Any("http_code", ctx.Writer.Status()),
					zap.Any("business_code", businessCode),
					zap.Any("business_message", businessCodeMsg),
					zap.Any("success", success),
					zap.Any("cost_seconds", costSeconds),
					zap.Error(abortErr),
				)
			}
		}()

		ctx.Next()
	})

	return mux, nil
}

func (m *Mux) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	m.engine.ServeHTTP(w, req)
}

func (m *Mux) Group(relativePath string, handlers ...HandlerFunc) *Router {
	return &Router{
		group: m.engine.Group(relativePath, wrapHandlers(handlers...)...),
	}
}

type Router struct {
	group *gin.RouterGroup
}

func (r *Router) GET(relativePath string, handlers ...HandlerFunc) {
	r.group.GET(relativePath, wrapHandlers(handlers...)...)
}

func (r *Router) GET_Gin(relativePath string, handlers []gin.HandlerFunc) {
	r.group.GET(relativePath, handlers...)
}

func (r *Router) POST(relativePath string, handlers ...HandlerFunc) {
	r.group.POST(relativePath, wrapHandlers(handlers...)...)
}

func (r *Router) POST_Gin(relativePath string, handlers []gin.HandlerFunc) {
	r.group.POST(relativePath, handlers...)
}

func (r *Router) DELETE(relativePath string, handlers ...HandlerFunc) {
	r.group.DELETE(relativePath, wrapHandlers(handlers...)...)
}

func (r *Router) PATCH(relativePath string, handlers ...HandlerFunc) {
	r.group.PATCH(relativePath, wrapHandlers(handlers...)...)
}

func (r *Router) PUT(relativePath string, handlers ...HandlerFunc) {
	r.group.PUT(relativePath, wrapHandlers(handlers...)...)
}

func (r *Router) Use(middleware HandlerFunc) *Router {
	r.group.Use(wrapHandler(middleware))
	return r
}

func wrapHandler(handler HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := newContext(c)
		defer releaseContext(ctx)
		handler(ctx)
	}
}

func wrapHandlers(handlers ...HandlerFunc) []gin.HandlerFunc {
	funcs := make([]gin.HandlerFunc, len(handlers))
	for i, handler := range handlers {
		handler := handler
		funcs[i] = func(c *gin.Context) {
			ctx := newContext(c)
			defer releaseContext(ctx)

			handler(ctx)
		}
	}

	return funcs
}
