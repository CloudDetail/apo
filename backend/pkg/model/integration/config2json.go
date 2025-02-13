// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"reflect"
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

func (j *JSONField[T]) ReplaceSecret() {
	replaceSecrets(&j.Obj)
}

func (j *JSONField[T]) AcceptExistedSecret(oldV T) {
	va := reflect.ValueOf(j.Obj).Elem()
	vb := reflect.ValueOf(oldV).Elem()

	if va.Type() != vb.Type() {
		fmt.Println("Error: a and b must be of the same type")
		return
	}
	acceptSecrets(va, vb)
}

const secretFieldValue = "<secret>"

func acceptSecrets(va, vb reflect.Value) {
	for i := 0; i < va.NumField(); i++ {
		fieldA := va.Field(i)
		fieldB := vb.Field(i)

		if fieldA.Kind() == reflect.Ptr {
			if !fieldA.IsNil() && !fieldB.IsNil() {
				acceptSecrets(fieldA.Elem(), fieldB.Elem())
			}
		} else if fieldA.Kind() == reflect.Struct {
			acceptSecrets(fieldA, fieldB)
		} else if fieldA.Kind() == reflect.String {
			if fieldA.String() == secretFieldValue {
				if fieldB.Kind() == reflect.String {
					fieldA.SetString(fieldB.String())
				}
			}
		}
	}
}

// replaceSecrets 遍历结构体字段，将带有 "secret" 标签的字段值替换为 "<secret>"
func replaceSecrets(v interface{}) {
	val := reflect.ValueOf(v).Elem() // 获取结构体的值
	typ := reflect.TypeOf(v).Elem()  // 获取结构体的类型

	// 遍历结构体的每个字段
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)     // 获取字段值
		fieldType := typ.Field(i) // 获取字段类型信息

		switch field.Kind() {
		case reflect.String:
			// 判断字段是否标记了 "secret" 标签
			if fieldType.Tag.Get("secret") == "true" {
				if len(field.String()) > 0 {
					field.SetString(secretFieldValue)
				}
			}
		case reflect.Struct:
			replaceSecrets(field.Addr().Interface()) // 传递指针类型
		case reflect.Ptr:
			if !field.IsNil() {
				replaceSecrets(field.Interface()) // 递归处理指针指向的内容
			}
		}

	}
}
