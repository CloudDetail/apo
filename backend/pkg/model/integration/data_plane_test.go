// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDataPlaneValidate(t *testing.T) {
	tests := []struct {
		name      string
		paramSpec string
		param     string
	}{
		{
			name:      "apo",
			paramSpec: `{"name":"root","type":"object","children":[{"name":"promAddress","type":"string"},{"name":"promStorage","type":"string"}]}`,
			param:     `{"promAddress": "http://0.0.0.0:8428", "promStorage": "vm"}`,
		},
		{
			name:      "datadog",
			paramSpec: `{"name":"root","type":"object","children":[{"name":"site","type":"string"},{"name":"apiKey","type":"string"},{"name":"appKey","type":"string"},{"name":"env","type":"string"}]}`,
			param:     `{"site": "datadoghq.com", "apiKey": "", "appKey": "", "env": "dev"}`,
		},
		{
			name:      "nginxLog",
			paramSpec: `{"name":"root","type":"object","children":[{"name":"address","type":"string"},{"name":"userName","type":"string"},{"name":"password","type":"string"},{"name":"database","type":"string"},{"name":"logTable","type":"string"}]}`,
			param:     `{"address": "0.0.0.0:9000", "userName": "uname", "password": "pwdtest", "database": "xxx", "logTable": "logs_nginx_access_log"}`,
		},
		{
			name:      "arms",
			paramSpec: `{"name":"root","type":"object","children":[{"name":"address","type":"string"},{"name":"accessKey","type":"string"},{"name":"accessSecret","type":"string"}]}`,
			param:     `{"address": "arms.cn-xxxx.aliyuncs.com", "accessKey": "exampleaccessKey", "accessSecret": "examplesecret"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var obj interface{}
			if err := json.Unmarshal([]byte(tt.param), &obj); err != nil {
				t.Fatal(err)
			}

			var paramSpec ParamSpec
			if err := json.Unmarshal([]byte(tt.paramSpec), &paramSpec); err != nil {
				t.Fatal(err)
			}
			err := ValidateJSON(obj, paramSpec)
			assert.NoError(t, err)
		})
	}
}

func TestUnsatisfiedParam(t *testing.T) {
	tests := []struct {
		name      string
		paramSpec string
		param     string
		want      error
	}{
		{
			name:      "missing field",
			paramSpec: `{"name":"root","type":"object","children":[{"name":"promAddress","type":"string"},{"name":"promStorage","type":"string"}]}`,
			param:     `{"promStorage": "vm"}`,
			want:      fmt.Errorf("missing required field: promAddress"),
		},
		{
			name:      "type mismatch",
			paramSpec: `{"name":"root","type":"object","children":[{"name":"site","type":"string"},{"name":"apiKey","type":"string"},{"name":"appKey","type":"string"},{"name":"env","type":"string"}]}`,
			param:     `{"site": 123, "apiKey": "", "appKey": "", "env": "dev"}`,
			want:      fmt.Errorf("field 'site' expected string"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var obj interface{}
			if err := json.Unmarshal([]byte(tt.param), &obj); err != nil {
				t.Fatal(err)
			}

			var paramSpec ParamSpec
			if err := json.Unmarshal([]byte(tt.paramSpec), &paramSpec); err != nil {
				t.Fatal(err)
			}
			err := ValidateJSON(obj, paramSpec)
			assert.Equal(t, tt.want, err)
		})
	}
}
