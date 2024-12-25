// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"os"
	"os/signal"
	"syscall"
)

var _ ShutdownHook = (*hook)(nil)

// Hook a graceful shutdown hook, default with signals of SIGINT and SIGTERM
type ShutdownHook interface {
	// WithSignals add more signals into hook
	WithSignals(signals ...syscall.Signal) ShutdownHook

	// Close register shutdown handles
	Close(funcs ...func())
}

type hook struct {
	ctx chan os.Signal
}

// NewHook create a Hook instance
func NewShutdownHook() ShutdownHook {
	hook := &hook{
		ctx: make(chan os.Signal, 1),
	}

	return hook.WithSignals(syscall.SIGINT, syscall.SIGTERM)
}

func (h *hook) WithSignals(signals ...syscall.Signal) ShutdownHook {
	for _, s := range signals {
		signal.Notify(h.ctx, s)
	}

	return h
}

func (h *hook) Close(funcs ...func()) {
	select {
	case <-h.ctx:
	}
	signal.Stop(h.ctx)

	for _, f := range funcs {
		f()
	}
}
