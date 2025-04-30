// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package clickhouse

import (
	"fmt"
	"strings"

	"github.com/ClickHouse/clickhouse-go/v2"
)

var (
	ALWAYS_TRUE  = &whereSQL{Wheres: "TRUE"}
	ALWAYS_FALSE = &whereSQL{Wheres: "FALSE"}
)

// whereSQL support concatenation of complex conditions in SQL statements
type whereSQL struct {
	Wheres string
	Values []any
}

func equals(key string, value any) *whereSQL {
	return valueCmp(key, value, "=")
}

func notEquals(key string, value any) *whereSQL {
	return valueCmp(key, value, "!=")
}

// When the equalsIfNotEmpty value length is 0, always true is returned
func equalsIfNotEmpty(key string, value string) *whereSQL {
	if len(key) <= 0 {
		return ALWAYS_TRUE
	}
	if len(value) > 0 {
		return &whereSQL{
			Wheres: fmt.Sprintf("%s = ?", key),
			Values: []any{value},
		}
	}
	return ALWAYS_TRUE
}

func like(key string, value any) *whereSQL {
	return valueCmp(key, value, "LIKE")
}

func notLike(key string, value any) *whereSQL {
	return valueCmp(key, value, "NOT LIKE")
}

func in(key string, values clickhouse.ArraySet) *whereSQL {
	if len(key) <= 0 {
		return ALWAYS_TRUE
	}
	if len(values) <= 0 {
		return ALWAYS_FALSE
	}
	if len(values) > 0 {
		return &whereSQL{
			Wheres: fmt.Sprintf("%s IN ?", key),
			Values: []any{values},
		}
	}
	return ALWAYS_FALSE
}

func notIn(key string, values clickhouse.ArraySet) *whereSQL {
	if len(key) <= 0 {
		return ALWAYS_TRUE
	}
	if len(values) <= 0 {
		return ALWAYS_FALSE
	}
	if len(values) > 0 {
		return &whereSQL{
			Wheres: fmt.Sprintf("%s NOT IN ?", key),
			Values: []any{values},
		}
	}
	return ALWAYS_FALSE
}

// ValueInGroups is used to pass in multiple sets of InGroups parameters in the OrInGroups and make OR connections.
// Each ValueInGroups generates the following SQL, where x is each value in the EqualIfNotEmpty
// (keys) IN (ValueGroups)
type ValueInGroups struct {
	Keys        []string
	ValueGroups []clickhouse.GroupSet
}

func inGroup(vgs ValueInGroups) *whereSQL {
	if len(vgs.Keys) <= 0 {
		return ALWAYS_TRUE
	}
	if len(vgs.ValueGroups) <= 0 {
		return ALWAYS_FALSE
	}
	return &whereSQL{
		Wheres: fmt.Sprintf("(%s) IN ?", strings.Join(vgs.Keys, ",")),
		Values: []any{clickhouse.GroupSet{
			Value: []any{vgs.ValueGroups},
		}},
	}
}

func contains(key string, value any) *whereSQL {
	if len(key) <= 0 {
		return ALWAYS_TRUE
	}
	return &whereSQL{
		Wheres: fmt.Sprintf("POSITION(%s, ?) > 0", key),
		Values: []any{value},
	}
}

func notContains(key string, value any) *whereSQL {
	if len(key) <= 0 {
		return ALWAYS_TRUE
	}
	return &whereSQL{
		Wheres: fmt.Sprintf("POSITION(%s, ?) = 0", key),
		Values: []any{value},
	}
}

func lessThan(key string, value any) *whereSQL {
	if len(key) <= 0 {
		return ALWAYS_TRUE
	}
	return &whereSQL{
		Wheres: fmt.Sprintf("%s < ?", key),
		Values: []any{value},
	}
}

func greaterThan(key string, value any) *whereSQL {
	if len(key) <= 0 {
		return ALWAYS_TRUE
	}
	return &whereSQL{
		Wheres: fmt.Sprintf("%s > ?", key),
		Values: []any{value},
	}
}

func exists(key string) *whereSQL {
	return &whereSQL{
		Wheres: fmt.Sprintf("%s EXISTS", key),
	}
}

func notExists(key string) *whereSQL {
	return &whereSQL{
		Wheres: fmt.Sprintf("%s NOT EXISTS", key),
	}
}

type mergeSep string

const (
	AndSep mergeSep = " AND "
	OrSep  mergeSep = " OR "
)

func getMergeSep(sep string) mergeSep {
	if strings.Contains(strings.ToLower(sep), "or") {
		return OrSep
	}
	return AndSep
}

// mergeWheres merge multiple conditions
func mergeWheres(sep mergeSep, whereSQLs ...*whereSQL) *whereSQL {
	var wheres []string
	var values []any

	var allFalse = true
	var allTrue = true

	if len(whereSQLs) <= 0 {
		// No conditions added
		return ALWAYS_TRUE
	}

	for _, where := range whereSQLs {
		if where == nil || where == ALWAYS_FALSE {
			if sep == AndSep {
				return ALWAYS_FALSE
			} else {
				allTrue = false
				continue
			}
		} else if where == ALWAYS_TRUE {
			if sep == AndSep {
				allFalse = false
				continue
			} else {
				return ALWAYS_TRUE
			}
		}

		allFalse = false
		allTrue = false

		wheres = append(wheres, where.Wheres)
		values = append(values, where.Values...)
	}

	if allTrue {
		return ALWAYS_TRUE
	} else if allFalse {
		return ALWAYS_FALSE
	} else if len(wheres) <= 0 {
		// No conditions added
		return ALWAYS_TRUE
	}

	return &whereSQL{
		Wheres: fmt.Sprintf("(%s)", strings.Join(wheres, string(sep))),
		Values: values,
	}
}

func valueCmp(key string, value any, cmp string) *whereSQL {
	if len(key) <= 0 {
		return ALWAYS_TRUE
	}
	return &whereSQL{
		Wheres: fmt.Sprintf("%s %s ?", key, cmp),
		Values: []any{value},
	}
}
