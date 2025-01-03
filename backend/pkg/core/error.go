// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package core

import "github.com/pkg/errors"

var _ BusinessError = (*businessError)(nil)

type BusinessError interface {
	// WithError 设置错误信息
	WithError(err error) BusinessError

	// BusinessCode 获取业务码
	BusinessCode() string

	// HTTPCode 获取 HTTP 状态码
	HTTPCode() int

	// Message 获取错误描述
	Message() string

	// StackError 获取带堆栈的错误信息
	StackError() error
}

type businessError struct {
	httpCode     int    // HTTP 状态码
	businessCode string // 业务码
	message      string // 错误描述
	stackError   error  // 含有堆栈信息的错误
}

func Error(httpCode int, businessCode, message string) BusinessError {
	return &businessError{
		httpCode:     httpCode,
		businessCode: businessCode,
		message:      message,
	}
}

func (e *businessError) WithError(err error) BusinessError {
	e.stackError = errors.WithStack(err)
	return e
}

func (e *businessError) HTTPCode() int {
	return e.httpCode
}

func (e *businessError) BusinessCode() string {
	return e.businessCode
}

func (e *businessError) Message() string {
	return e.message
}

func (e *businessError) StackError() error {
	return e.stackError
}
