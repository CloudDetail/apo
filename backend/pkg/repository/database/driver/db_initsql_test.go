// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"reflect"
	"testing"
)

func Test_sortUpgradeFiles(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		files []string
		want  []string
	}{
		{
			name: "sort with semantic",
			files: []string{
				"1.11.0-upgrade.sql",
				"1.9.2-upgrade.sql",
				"1.13.0-upgrade.sql",
				"1.11.0-upgrade-mysql.sql",
				"1.13.0-upgrade-mysql.sql",
			},
			want: []string{
				"1.9.2-upgrade.sql",
				"1.11.0-upgrade.sql",
				"1.11.0-upgrade-mysql.sql",
				"1.13.0-upgrade.sql",
				"1.13.0-upgrade-mysql.sql",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sortUpgradeFiles(tt.files)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("sortUpgradeFiles() = %v, want %v", got, tt.want)
			}
		})
	}
}
