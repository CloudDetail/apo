// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alertbus

import (
	"fmt"
	"time"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	"go.uber.org/zap"
)

type EventsHandler func(ctx core.Context, events []alert.AlertEvent) error

type ExtraEventHandler struct {
	handlers []EventsHandler
	logger   *zap.Logger

	timeout time.Duration
}

var ExtraHandlers *ExtraEventHandler

func InitExtraEventHandler(logger *zap.Logger, timeoutSeconds int) *ExtraEventHandler {
	if timeoutSeconds < 300 {
		timeoutSeconds = 300
	}

	e := &ExtraEventHandler{
		handlers: make([]EventsHandler, 0),
		timeout:  time.Duration(timeoutSeconds) * time.Second,
	}
	e.logger = logger
	ExtraHandlers = e
	return e
}

func (e *ExtraEventHandler) RegisterHandler(handler EventsHandler) {
	e.handlers = append(e.handlers, handler)
}

func (e *ExtraEventHandler) HandleEvents(ctx core.Context, events []alert.AlertEvent) {
	if e == nil {
		return
	}
	for _, handler := range e.handlers {
		go e.dispatchToHandler(ctx, handler, events)
	}
}

func (e *ExtraEventHandler) dispatchToHandler(rc core.Context, handler EventsHandler, events []alert.AlertEvent) {
	if err := handler(rc.Clone(), events); err != nil {
		e.logger.Error("extra event handler failed",
			zap.String("handler", fmt.Sprintf("%T", handler)),
			zap.Error(err),
		)
	}
}
