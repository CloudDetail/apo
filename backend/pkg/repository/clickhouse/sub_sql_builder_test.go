package clickhouse

import (
	"testing"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/stretchr/testify/assert"
)

func TestWhereSQL(t *testing.T) {
	subQuery1 := mergeWheres(
		getMergeSep("OR"),
		equals("key1", "value1"),
		equalsIfNotEmpty("key2", "value2"),
		like("key3", "value3"),
		in("key4", clickhouse.ArraySet{"value4_1", "value4_2"}),
		contains("key5", "value5"),
		inGroup(ValueInGroups{
			Keys: []string{"k6_1", "k6_2"},
			ValueGroups: []clickhouse.GroupSet{
				{
					Value: []any{"v6_1_1", "v6_1_2"},
				},
				{
					Value: []any{"v6_2_1", "v6_2_2"},
				},
			},
		}),
	)

	subQuery2 := mergeWheres(
		getMergeSep("or"),
		notEquals("key7", "value7"),
		notLike("key8", "value8"),
		notIn("key9", clickhouse.ArraySet{"v9_1", "v9_2"}),
		notContains("key10", "value10"),
	)

	subSQL := mergeWheres(AndSep, subQuery1, subQuery2)

	want := `((key1 = ? OR key2 = ? OR key3 LIKE ? OR key4 IN ? OR POSITION(key5, ?) > 0 OR (k6_1,k6_2) IN ?) AND (key7 != ? OR key8 NOT LIKE ? OR key9 NOT IN ? OR POSITION(key10, ?) = 0))`
	assert.Equal(t, want, subSQL.Wheres)

	valueWant := []any{
		"value1",
		"value2",
		"value3",
		clickhouse.ArraySet{"value4_1", "value4_2"},
		"value5",
		clickhouse.GroupSet{
			Value: []any{
				[]clickhouse.GroupSet{
					{
						Value: []any{"v6_1_1", "v6_1_2"},
					},
					{
						Value: []any{"v6_2_1", "v6_2_2"},
					},
				},
			},
		},
		"value7",
		"value8",
		clickhouse.ArraySet{"v9_1", "v9_2"},
		"value10",
	}

	assert.Equal(t, valueWant, subSQL.Values)

}
