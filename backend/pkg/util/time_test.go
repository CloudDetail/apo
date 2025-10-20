// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"testing"
)

func TestRoundDownToMinute(t *testing.T) {
	tests := []struct {
		name      string
		timestamp int64
		expected  int64
	}{
		{
			name:      "Round down to minute",
			timestamp: 1760059323561000,
			expected:  1760059320000000,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RoundDownToMinute(tt.timestamp); got != tt.expected {
				t.Errorf("RoundDownToMinute() = %v, want %v", got, tt.expected)
			}
		})
	}
}
