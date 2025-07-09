// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"github.com/pkg/errors"
)

var _ BusinessError = (*businessError)(nil)

type BusinessError interface {
	// WithError setting error message
	WithStack(err error) BusinessError

	// BusinessCode get business code
	BusinessCode() string

	// HTTPCode get the HTTP status code
	HTTPCode() int

	// Message get the error description
	Message() string

	// StackError get the error message with stack
	StackError() error

	Error() string
}

var _ BusinessError = (*businessError)(nil)

type businessError struct {
	httpCode     int    // HTTP status code
	businessCode string // business code
	message      string // error description
	stackError   error  // error with stack information
}

func Error(businessCode, message string) BusinessError {
	return &businessError{
		businessCode: businessCode,
		message:      message,
	}
}

func (e *businessError) WithStack(err error) BusinessError {
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

func (e *businessError) Error() string {
	return e.message
}
