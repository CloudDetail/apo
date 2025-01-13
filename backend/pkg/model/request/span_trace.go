// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package request

import (
	"fmt"
	"reflect"
)

type GetTraceFiltersRequest struct {
	StartTime  int64 `form:"statTime" json:"startTime" binding:"min=0"`                   // query start time
	EndTime    int64 `form:"endTime" json:"endTime" binding:"required,gtfield=StartTime"` // query end time
	NeedUpdate bool  `form:"needUpdate" json:"needUpdate"`                                // whether it needs to be updated immediately
}

type GetTraceFilterValueRequest struct {
	StartTime  int64           `json:"startTime" binding:"min=0"`                    // query start time
	EndTime    int64           `json:"endTime" binding:"required,gtfield=StartTime"` // query end time
	SearchText string          `json:"searchText"`                                   // query keyword
	Filter     SpanTraceFilter `json:"filter"`
}

type Operation string

const (
	OpEqual       Operation = "EQUAL"
	OpNotEqual    Operation = "NOT_EQUAL"
	OpIn          Operation = "IN"
	OpNotIn       Operation = "NOT_IN"
	OpLike        Operation = "LIKE"
	OpNotLike     Operation = "NOT_LIKE"
	OpExists      Operation = "EXISTS"
	OpNotExists   Operation = "NOT_EXISTS"
	OpContains    Operation = "CONTAINS"
	OpNotContains Operation = "NOT_CONTAINS"

	OpGreaterThan Operation = "GREATER_THAN"
	OpLessThan    Operation = "LESS_THAN"
)

type DataType string

const (
	I64Column    DataType = "int64"
	U32Column    DataType = "uint32"
	U64Column    DataType = "uint64"
	StringColumn DataType = "string"
	BoolColumn   DataType = "bool"
)

func (f *DataType) Scan(src interface{}) error {
	v, ok := src.(string)
	if !ok {
		return fmt.Errorf("can not covert %v to ParentField", reflect.TypeOf(src))
	}
	*f = DataType(v)
	return nil
}

type ParentField string

func (f *ParentField) Scan(src interface{}) error {
	v, ok := src.(string)
	if !ok {
		return fmt.Errorf("can not covert %v to ParentField", reflect.TypeOf(src))
	}
	*f = ParentField(v)
	return nil
}

const (
	PF_Labels ParentField = "labels"
	PF_Flags  ParentField = "flags"
)

type SpanTraceFilter struct {
	Key         string      `ch:"key" json:"key"`
	ParentField ParentField `ch:"parent_field" json:"parentField"`
	DataType    DataType    `ch:"data_type" json:"dataType"`
	Operation   Operation   `json:"operation,omitempty"`
	Value       []string    `json:"value,omitempty"`
}
