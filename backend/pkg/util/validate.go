package util

import "testing"

type Validator struct {
	test      *testing.T
	Operation string
}

func NewValidator(t *testing.T, operation string) *Validator {
	return &Validator{
		test:      t,
		Operation: operation,
	}
}

func (v *Validator) CheckIntValue(key string, expect int, got int) *Validator {
	if expect != got {
		v.test.Errorf("[Check %s-%s] want=%d, got=%d", v.Operation, key, expect, got)
	}
	return v
}

func (v *Validator) CheckInt64Value(key string, expect int64, got int64) *Validator {
	if expect != got {
		v.test.Errorf("[Check %s-%s] want=%d, got=%d", v.Operation, key, expect, got)
	}
	return v
}

func (v *Validator) CheckStringValue(key string, expect string, got string) *Validator {
	if expect != got {
		v.test.Errorf("[Check %s-%s] want=%s, got=%s", v.Operation, key, expect, got)
	}
	return v
}

func (v *Validator) CheckBoolValue(key string, expect bool, got bool) *Validator {
	if expect != got {
		v.test.Errorf("[Check %s-%s] want=%t, got=%t", v.Operation, key, expect, got)
	}
	return v
}
