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
	val, err := json.Marshal(j.Obj)
	return string(val), err
}

func (j *JSONField[T]) Scan(value interface{}) error {
	if value == nil {
		j.Obj = *new(T)
		return nil
	}

	var val []byte
	switch s := value.(type) {
	case string:
		val = []byte(s)
	case []byte:
		val = s
	default:
		return fmt.Errorf("failed to scan JSONField, expected string, got %T", value)
	}
	return json.Unmarshal(val, &j.Obj)
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
	va := reflect.ValueOf(&j.Obj).Elem()
	vb := reflect.ValueOf(&oldV).Elem()

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

// replace field value with "<secret>" if it has "secret" tag
func replaceSecrets(v interface{}) {
	val := reflect.ValueOf(v).Elem()
	typ := reflect.TypeOf(v).Elem()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		switch field.Kind() {
		case reflect.String:
			if fieldType.Tag.Get("secret") == "true" {
				if len(field.String()) > 0 {
					field.SetString(secretFieldValue)
				}
			}
		case reflect.Struct:
			replaceSecrets(field.Addr().Interface())
		case reflect.Ptr:
			if !field.IsNil() {
				replaceSecrets(field.Interface())
			}
		}
	}
}

func (c APOCollector) Value() (driver.Value, error) {
	val, err := json.Marshal(c)
	return string(val), err
}

func (c *APOCollector) Scan(value interface{}) error {
	if value == nil {
		*c = APOCollector{}
		return nil
	}
	var val []byte
	switch s := value.(type) {
	case string:
		val = []byte(s)
	case []byte:
		val = s
	default:
		return fmt.Errorf("failed to scan JSONField, expected string, got %T", value)
	}
	return json.Unmarshal(val, c)
}
