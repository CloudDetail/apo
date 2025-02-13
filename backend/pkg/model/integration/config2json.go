// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type JSONField[T any] struct {
	Obj T
}

func (j JSONField[T]) Value() (driver.Value, error) {
	// 将结构体序列化为 JSON 字符串
	val, err := json.Marshal(j.Obj)
	if err != nil {
		return nil, err
	}
	return string(val), nil
}

func (j *JSONField[T]) Scan(value interface{}) error {
	// 如果值为 nil，则设置为零值
	if value == nil {
		j.Obj = *new(T)
		return nil
	}
	// 将字符串反序列化为结构体
	val, ok := value.(string)
	if !ok {
		return fmt.Errorf("failed to scan JSONField, expected string, got %T", value)
	}
	return json.Unmarshal([]byte(val), &j.Obj)
}

func (j *JSONField[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(j.Obj)
}

func (j *JSONField[T]) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &j.Obj)
}
