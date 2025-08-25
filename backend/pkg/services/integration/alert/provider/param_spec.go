// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package provider

import "fmt"

type JSONType string

const (
	JSONTypeObject  JSONType = "object"
	JSONTypeArray   JSONType = "array"
	JSONTypeString  JSONType = "string"
	JSONTypeNumber  JSONType = "number"
	JSONTypeBoolean JSONType = "boolean"
	JSONTypeNull    JSONType = "null"
)

type ParamSpec struct {
	Name     string      `json:"name"`
	Type     JSONType    `json:"type"`
	Optional bool        `json:"optional,omitempty"`
	Children []ParamSpec `json:"children,omitempty"`
	Desc     string      `json:"desc,omitempty"`
	DescEN   string      `json:"desc_en,omitempty"`
}

func ValidateJSON(value any, schema ParamSpec) error {
	switch schema.Type {
	case JSONTypeObject:
		obj, ok := value.(map[string]any)
		if !ok {
			return fmt.Errorf("field '%s' expected object", schema.Name)
		}
		for _, child := range schema.Children {
			val, exists := obj[child.Name]
			if !exists {
				if child.Optional {
					continue
				}
				return fmt.Errorf("missing required field: %s", child.Name)
			}
			if err := ValidateJSON(val, child); err != nil {
				return err
			}
		}
	case JSONTypeArray:
		arr, ok := value.([]any)
		if !ok {
			return fmt.Errorf("field '%s' expected array", schema.Name)
		}
		for _, item := range arr {
			// 数组元素类型统一使用 schema.Children[0]
			if len(schema.Children) == 0 {
				return fmt.Errorf("field '%s' array has no element type defined", schema.Name)
			}
			if err := ValidateJSON(item, schema.Children[0]); err != nil {
				return err
			}
		}
	case JSONTypeString:
		if _, ok := value.(string); !ok {
			return fmt.Errorf("field '%s' expected string", schema.Name)
		}
	case JSONTypeNumber:
		switch value.(type) {
		case float64, float32, int, int64, int32:
			// ok
		default:
			return fmt.Errorf("field '%s' expected number", schema.Name)
		}
	case JSONTypeBoolean:
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("field '%s' expected boolean", schema.Name)
		}
	case JSONTypeNull:
		if value != nil {
			return fmt.Errorf("field '%s' expected null", schema.Name)
		}
	default:
		return fmt.Errorf("unknown type for field '%s'", schema.Name)
	}
	return nil
}
