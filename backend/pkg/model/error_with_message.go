// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package model

type ErrWithMessage struct {
	Err  error
	Code string
}

func (e ErrWithMessage) Error() string {
	return e.Err.Error()
}

func NewErrWithMessage(err error, code string) ErrWithMessage {
	return ErrWithMessage{
		Err:  err,
		Code: code,
	}
}
