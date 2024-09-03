package clickhouse

import (
	"fmt"
	"strings"

	"github.com/ClickHouse/clickhouse-go/v2"
)

type FieldBuilder struct {
	fields []string
}

type QueryBuilder struct {
	where  []string
	values []interface{}
}

type ByLimitBuilder struct {
	groupBy []string
	order   []string
	limit   int
	offset  int
}

func NewFieldBuilder() *FieldBuilder {
	return &FieldBuilder{
		fields: make([]string, 0),
	}
}

func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{
		where:  make([]string, 0),
		values: make([]interface{}, 0),
	}
}

func NewByLimitBuilder() *ByLimitBuilder {
	return &ByLimitBuilder{
		order:   make([]string, 0),
		groupBy: make([]string, 0),
		limit:   0,
		offset:  0,
	}
}

func (builder *FieldBuilder) Alias(key string, alias string) *FieldBuilder {
	builder.fields = append(builder.fields, fmt.Sprintf("%s as %s", key, alias))
	return builder
}

func (builder *FieldBuilder) Fields(keys ...string) *FieldBuilder {
	builder.fields = append(builder.fields, keys...)
	return builder
}

// 返回检索字段
func (builder *FieldBuilder) String() string {
	labels := ""
	for i, field := range builder.fields {
		if i > 0 {
			labels += ", "
		}
		labels += field
	}
	return labels
}

func (builder *QueryBuilder) Between(key string, from int64, to int64) *QueryBuilder {
	builder.where = append(builder.where, fmt.Sprintf("%s BETWEEN ? AND ?", key))
	builder.values = append(builder.values, from, to)
	return builder
}

func (builder *QueryBuilder) Equals(key string, value interface{}) *QueryBuilder {
	builder.where = append(builder.where, fmt.Sprintf("%s = ?", key))
	builder.values = append(builder.values, value)
	return builder
}

func (builder *QueryBuilder) GreaterThan(key string, value any) *QueryBuilder {
	builder.where = append(builder.where, fmt.Sprintf("%s > ?", key))
	builder.values = append(builder.values, value)
	return builder
}

// 组合生成SQL中的 key in (values) 语句, values内部为值数组
func (builder *QueryBuilder) In(key string, values clickhouse.GroupSet) *QueryBuilder {
	builder.where = append(builder.where, fmt.Sprintf("%s in ?", key))
	builder.values = append(builder.values, values)
	return builder
}

func (builder *QueryBuilder) InStrings(key string, values []string) *QueryBuilder {
	builder.where = append(builder.where, fmt.Sprintf("%s in ?", key))
	builder.values = append(builder.values, values)
	return builder
}

// ValueInGroups 用于在OrInGroups中传入多组InGroups参数并做OR连接
// 每个ValueInGroups生成如下SQL,x是EqualIfNotEmpty中的每个值
// (keys) IN (ValueGroups)
type ValueInGroups struct {
	Keys        []string
	ValueGroups []clickhouse.GroupSet
}

// whereSQL SQL片段
// !!! nil有特殊的含义,后续And或OR合并时, nil等价于ALWAYS_TRUE
type whereSQL struct {
	Wheres string
	Values []any
}

var (
	ALWAYS_TRUE = &whereSQL{
		Wheres: "TRUE",
	}

	ALWAYS_FALSE = &whereSQL{
		Wheres: "FALSE",
	}
)

func In(key string, values clickhouse.ArraySet) *whereSQL {
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

func InGroup(vgs ValueInGroups) *whereSQL {
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

// EqualsIfNotEmpty value长度为0时,返回always true
func EqualsIfNotEmpty(key string, value string) *whereSQL {
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

func Equals(key string, value string) *whereSQL {
	if len(key) <= 0 {
		return ALWAYS_TRUE
	}
	return &whereSQL{
		Wheres: fmt.Sprintf("%s = ?", key),
		Values: []any{value},
	}
}

type MergeSep string

const (
	AndSep MergeSep = " AND "
	OrSep  MergeSep = " OR "
)

// MergeWheres 合并多个条件
func MergeWheres(sep MergeSep, whereSQLs ...*whereSQL) *whereSQL {
	var wheres []string
	var values []any
	for _, where := range whereSQLs {
		if where == nil || where == ALWAYS_FALSE {
			if sep == AndSep {
				return ALWAYS_FALSE
			} else {
				continue
			}
		} else if where == ALWAYS_TRUE {
			if sep == AndSep {
				continue
			} else {
				return ALWAYS_TRUE
			}
		}

		wheres = append(wheres, where.Wheres)
		values = append(values, where.Values...)
	}

	if len(wheres) <= 0 {
		return ALWAYS_TRUE
	}

	return &whereSQL{
		Wheres: fmt.Sprintf("(%s)", strings.Join(wheres, string(sep))),
		Values: values,
	}
}

// And 将一系列条件whereSQL以And方式加入到QueryBuilder
func (builder *QueryBuilder) And(where *whereSQL) *QueryBuilder {
	if where == nil || where == ALWAYS_FALSE {
		builder.where = append(builder.where, "FALSE")
		return builder
	} else if where == ALWAYS_TRUE {
		return builder
	}
	builder.where = append(builder.where, where.Wheres)
	builder.values = append(builder.values, where.Values...)
	return builder
}

func (builder *QueryBuilder) EqualsNotEmpty(key string, value string) *QueryBuilder {
	if value != "" {
		builder.where = append(builder.where, fmt.Sprintf("%s = ?", key))
		builder.values = append(builder.values, value)
	}
	return builder
}

func (builder *QueryBuilder) Statement(where string) *QueryBuilder {
	builder.where = append(builder.where, where)
	return builder
}

// 返回查询条件
func (builder *QueryBuilder) String() string {
	whereSql := ""
	for i, where := range builder.where {
		if i == 0 {
			whereSql += "WHERE "
		} else {
			whereSql += " AND "
		}
		whereSql += where
	}
	return whereSql
}

func (builder *ByLimitBuilder) GroupBy(keys ...string) *ByLimitBuilder {
	builder.groupBy = append(builder.groupBy, keys...)
	return builder
}

func (builder *ByLimitBuilder) OrderBy(key string, asc bool) *ByLimitBuilder {
	order := "DESC"
	if asc {
		order = "ASC"
	}
	builder.order = append(builder.order, fmt.Sprintf("%s %s", key, order))
	return builder
}

func (builder *ByLimitBuilder) Limit(limit int) *ByLimitBuilder {
	builder.limit = limit
	return builder
}

func (builder *ByLimitBuilder) Offset(offset int) *ByLimitBuilder {
	builder.offset = offset
	return builder
}

// 返回GroupBy、OrderBy和Limit
func (builder *ByLimitBuilder) String() string {
	sql := ""
	for i, key := range builder.groupBy {
		if i == 0 {
			sql += " GROUP BY "
		} else {
			sql += ", "
		}
		sql += key
	}
	for i, order := range builder.order {
		if i == 0 {
			sql += " ORDER BY "
		} else {
			sql += ", "
		}
		sql += order
	}
	if builder.limit > 0 {
		sql += fmt.Sprintf(" LIMIT %d", builder.limit)
	}
	if builder.offset > 0 {
		sql += fmt.Sprintf(" OFFSET %d", builder.offset)
	}
	return sql
}
