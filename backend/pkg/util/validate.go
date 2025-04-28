// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"encoding/json"
	"io"
	"testing"
)

type Validator struct {
	test      *testing.T
	Operation string
}

func NewValidator(t *testing.T, operation string) *Validator {
	return &Validator{
		test:      t,
		Operation: operation,
	}
}

func (v *Validator) CheckIntValue(key string, expect int, got int) *Validator {
	if expect != got {
		v.test.Errorf("[Check %s-%s] want=%d, got=%d", v.Operation, key, expect, got)
	}
	return v
}

func (v *Validator) CheckInt64Value(key string, expect int64, got int64) *Validator {
	if expect != got {
		v.test.Errorf("[Check %s-%s] want=%d, got=%d", v.Operation, key, expect, got)
	}
	return v
}

func (v *Validator) CheckStringValue(key string, expect string, got string) *Validator {
	if expect != got {
		v.test.Errorf("[Check %s-%s] want=%s, got=%s", v.Operation, key, expect, got)
	}
	return v
}

func (v *Validator) CheckBoolValue(key string, expect bool, got bool) *Validator {
	if expect != got {
		v.test.Errorf("[Check %s-%s] want=%t, got=%t", v.Operation, key, expect, got)
	}
	return v
}

func IsValidStatusCode(statusCode int) bool {
	// 定义合法的状态码范围
	return statusCode >= 100 && statusCode <= 599
}

func ValidateResponse(body io.ReadCloser) (io.ReadCloser, bool) {
	return body, true
}
func ValidateResponseBytes(body []byte) ([]byte, bool) {
	if len(body) > 2*1024*1024 { // 限制最大2MB
		return nil, false
	}
	if !json.Valid(body) {
		return nil, false
	}
	return body, true
}
