// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/CloudDetail/apo/backend/config"
	_ "github.com/CloudDetail/apo/backend/docs" // import Swagger docs package
	"github.com/CloudDetail/apo/backend/pkg/logger"
	"github.com/CloudDetail/apo/backend/pkg/router"
	"github.com/CloudDetail/apo/backend/pkg/util"
	"go.uber.org/zap"
)

// @title swagger interface documentation
// @version 2.0
// @description

// @contact.name
// @contact.url
// @contact.email

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @BasePath /
func main() {
	// Initialize access logger
	logCfg := config.Get().Logger
	accessLogger := logger.NewLogger(
		logger.WithConsole(logCfg.EnableConsole),
		logger.WithLevel(logCfg.Level),
		logger.WithTimeLayout(logger.CSTLayout),
		logger.WithFileRotationP(logCfg.EnableFile, logCfg.FilePath, logCfg.FileNum, logCfg.FileSize),
	)
	defer func() {
		_ = accessLogger.Sync()
	}()

	// Initialize the HTTP service
	s, err := router.NewHTTPServer(accessLogger)
	if err != nil {
		panic(err)
	}

	serverPort := config.Get().Server.Port
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", serverPort),
		Handler: s.Mux,
	}

	go func() {
		accessLogger.Info("Start Http Server", zap.Any("Port", serverPort))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			accessLogger.Fatal("http server startup err", zap.Error(err))
		}
	}()

	// Graceful shutdown
	util.NewShutdownHook().Close(
		// Close http server
		func() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
			defer cancel()

			if err := server.Shutdown(ctx); err != nil {
				accessLogger.Error("server shutdown err", zap.Error(err))
			}
		},
	)
}
