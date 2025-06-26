// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package prometheus

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPQLFilter(t *testing.T) {
	A := EqualFilter("key1", "val1")
	assert.Equal(t,
		`test{key1="val1"}`,
		vector("test", A).String(),
	)

	B := RegexMatchFilter("key2", "val2")
	assert.Equal(t,
		`test{key2=~"val2"}`,
		vector("test", B).String(),
	)
	C := And(A, B)
	assert.Equal(t,
		`test{key1="val1",key2=~"val2"}`,
		vector("test", C).String(),
	)

	A2 := EqualFilter("key3", "val3")
	B2 := RegexMatchFilter("key4", "val4")

	D := Or(A2, B2)
	assert.Equal(t,
		`test{key3="val3" or key4=~"val4"}`,
		vector("test", D).String(),
	)

	E := And(C, D)

	assert.Equal(t,
		`test{key1="val1",key2=~"val2",key3="val3" or key1="val1",key2=~"val2",key4=~"val4"}`,
		vector("test", E).String(),
	)

	D2 := Or(A, B)
	F := Or(D, D2)

	assert.Equal(t,
		`test{key3="val3" or key4=~"val4" or key1="val1" or key2=~"val2"}`,
		vector("test", F).String(),
	)
}

func TestStructPQLFilter(t *testing.T) {
	EnableStrictPQL()

	A := EqualFilter("key1", "val1")
	assert.Equal(t,
		`test{key1="val1"}`,
		vector("test", A).String(),
	)

	B := RegexMatchFilter("key2", "val2")
	assert.Equal(t,
		`test{key2=~"val2"}`,
		vector("test", B).String(),
	)
	C := And(A, B)
	assert.Equal(t,
		`test{key1="val1",key2=~"val2"}`,
		vector("test", C).String(),
	)

	A2 := EqualFilter("key3", "val3")
	B2 := RegexMatchFilter("key4", "val4")

	D := Or(A2, B2)
	assert.Equal(t,
		`test{key3="val3"} or test{key4=~"val4"}`,
		vector("test", D).String(),
	)

	E := And(C, D)

	assert.Equal(t,
		`test{key1="val1",key2=~"val2",key3="val3"} or test{key1="val1",key2=~"val2",key4=~"val4"}`,
		vector("test", E).String(),
	)

	D2 := Or(A, B)
	F := Or(D, D2)
	assert.Equal(t,
		`test{key3="val3"} or test{key4=~"val4"} or test{key1="val1"} or test{key2=~"val2"}`,
		vector("test", F).String(),
	)
}
