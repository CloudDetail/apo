package trace

import (
	"github.com/CloudDetail/apo/backend/pkg/util"
	"testing"
)

func TestMapEqual(t *testing.T) {
	m1 := map[string]string{"key": "value"}
	m2 := map[string]string{"key": "value"}
	got := checkLabelsEqual(m1, m2)
	validator := util.NewValidator(t, "Test map equal")
	validator.CheckBoolValue("isEqual", true, got)
}

func TestLenMismatch(t *testing.T) {
	m1 := map[string]string{"key": "value"}
	m2 := map[string]string{"key": "value", "key2": "value2"}
	got := checkLabelsEqual(m1, m2)
	validator := util.NewValidator(t, "Test len mismatch")
	validator.CheckBoolValue("isEqual", false, got)
}

func TestKeyMismatch(t *testing.T) {
	m1 := map[string]string{"key": "value"}
	m2 := map[string]string{"key1": "value"}
	got := checkLabelsEqual(m1, m2)
	validator := util.NewValidator(t, "Test key mismatch")
	validator.CheckBoolValue("isEqual", false, got)
}

func TestValueMismatch(t *testing.T) {
	m1 := map[string]string{"key": "value"}
	m2 := map[string]string{"key": "values"}
	got := checkLabelsEqual(m1, m2)
	validator := util.NewValidator(t, "Test value mismatch")
	validator.CheckBoolValue("isEqual", false, got)
}
