package clickhouse

import "fmt"

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
