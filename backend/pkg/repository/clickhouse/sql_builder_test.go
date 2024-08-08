package clickhouse

import (
	"testing"

	"github.com/CloudDetail/apo/backend/pkg/util"
)

func TestFieldBuilder(t *testing.T) {
	got := NewFieldBuilder().
		Alias("timestamp", "ts").
		Alias("Id", "id").
		Fields("name", "type").
		String()
	want := "timestamp as ts, Id as id, name, type"

	util.NewValidator(t, "FieldBuilder").
		CheckStringValue("fields", want, got)
}

func TestQueryBuilder(t *testing.T) {
	var (
		timeFrom   int64  = 0
		timeTo     int64  = 100
		keyValue   string = "test"
		errorValue bool   = true
		nameValue  string = ""
		groupValue string = "g1"
	)

	builder := NewQueryBuilder().
		Between("timestamp", timeFrom, timeTo).
		Equals("key", keyValue).
		Equals("error", errorValue).
		EqualsNotEmpty("name", nameValue).
		EqualsNotEmpty("group", groupValue).
		Statement("name is not null")
	want := "WHERE timestamp BETWEEN ? AND ?" +
		" AND key = ?" +
		" AND error = ?" +
		" AND group = ?" +
		" AND name is not null"

	util.NewValidator(t, "FieldBuilder").
		CheckStringValue("whereSql", want, builder.String()).
		CheckInt64Value("timeFrom", timeFrom, builder.values[0].(int64)).
		CheckInt64Value("timeTo", timeTo, builder.values[1].(int64)).
		CheckStringValue("keyValue", keyValue, builder.values[2].(string)).
		CheckBoolValue("errorValue", errorValue, builder.values[3].(bool)).
		CheckStringValue("groupValue", groupValue, builder.values[4].(string))
}

func TestByLimitBuilder(t *testing.T) {
	builder := NewByLimitBuilder().
		GroupBy("name", "key").
		OrderBy("timestamp", true).
		OrderBy("id", false).
		Limit(10).
		Offset(100)

	want := " GROUP BY name, key" +
		" ORDER BY timestamp ASC, id DESC" +
		" LIMIT 10" +
		" OFFSET 100"
	util.NewValidator(t, "ByLimitBuilder").
		CheckStringValue("byLimitSql", want, builder.String())
}
