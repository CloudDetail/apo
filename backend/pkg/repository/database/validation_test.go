package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidMenuItem(t *testing.T) {
	tests := []struct {
		name     string
		menuItem string
		want     bool
	}{
		{"Valid menu item: service", "service", true},
		{"Valid menu item: workflows", "workflows", true},
		{"Invalid menu item", "invalid", false},
		{"Empty menu item", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidMenuItem(tt.menuItem)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsValidRoleName(t *testing.T) {
	tests := []struct {
		name     string
		roleName string
		want     bool
	}{
		{"Valid role: admin", "admin", true},
		{"Valid role: viewer", "viewer", true},
		{"Invalid role", "manager", false},
		{"Empty role", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidRoleName(tt.roleName)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsValidFeature(t *testing.T) {
	tests := []struct {
		name        string
		featureName string
		want        bool
	}{
		{"Valid feature: 服务概览", "服务概览", true},
		{"Valid feature: 工作流", "工作流", true},
		{"Invalid feature", "feature3", false},
		{"Empty feature", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidFeature(tt.featureName)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsValidMethod(t *testing.T) {
	tests := []struct {
		name   string
		method string
		want   bool
	}{
		{"Valid method: GET", "GET", true},
		{"Valid method: post", "post", true},
		{"Valid method: PUT", "PUT", true},
		{"Valid method: DELETE", "DELETE", true},
		{"Valid method: PATCH", "PATCH", true},
		{"Invalid method", "OPTIONS", false},
		{"Empty method", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidMethod(tt.method)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsValidPath(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		{"Valid path: /api", "/api", true},
		{"Valid path: /users/123", "/users/123", true},
		{"Invalid path: no leading slash", "api", false},
		{"Invalid path: contains ;", "/api;test", false},
		{"Invalid path: contains &", "/api&test", false},
		{"Invalid path: contains '", "/api'test", false},
		{"Invalid path: contains =", "/api=test", false},
		{"Empty path", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidPath(tt.path)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsValidRouter(t *testing.T) {
	tests := []struct {
		name     string
		routerTo string
		want     bool
	}{
		{"Valid router: /service", "/service", true},
		{"Valid router: /logs/full", "/logs/full", true},
		{"Invalid router", "/api/v3", false},
		{"Empty router", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidRouter(tt.routerTo)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsValidPageURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{"Valid page: grafana/d/b0102ebf-9e5e-4f21-80aa-9c2565cd3dcb/originx-polaris-metrics-service-level", "grafana/d/b0102ebf-9e5e-4f21-80aa-9c2565cd3dcb/originx-polaris-metrics-service-level", true},
		{"Invalid page", "/login", false},
		{"Empty page", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidPageURL(tt.url)
			assert.Equal(t, tt.want, got)
		})
	}
}
