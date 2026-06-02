// Copyright 2026 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package customdashboard

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	cdmodel "github.com/CloudDetail/apo/backend/pkg/model/customdashboard"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	customdashboardservice "github.com/CloudDetail/apo/backend/pkg/services/customdashboard"
	"go.uber.org/zap"
)

func TestVariableCandidatesErrorCodeDistinguishesNoPermission(t *testing.T) {
	if got := variableCandidatesErrorCode(customdashboardservice.ErrNoDataGroupPermission); got != code.UserNoPermissionError {
		t.Fatalf("expected no-permission error code, got %s", got)
	}

	if got := variableCandidatesErrorCode(errors.New("other failure")); got != code.GetCustomDashboardError {
		t.Fatalf("expected generic custom dashboard error code, got %s", got)
	}
}

type dashboardPermissionErrorServiceStub struct {
	customdashboardservice.Service
}

func (s *dashboardPermissionErrorServiceStub) Get(ctx core.Context, id string) (*cdmodel.Dashboard, error) {
	return nil, core.Error(code.UserNoPermissionError, "")
}

func (s *dashboardPermissionErrorServiceStub) Update(ctx core.Context, req *request.UpdateDashboardRequest) (*cdmodel.Dashboard, error) {
	return nil, core.Error(code.UserNoPermissionError, "")
}

func TestDashboardHandlersPreserveNoPermissionErrorCode(t *testing.T) {
	mux, err := core.New(zap.NewNop())
	if err != nil {
		t.Fatalf("create mux: %v", err)
	}
	handler := &handler{service: &dashboardPermissionErrorServiceStub{}}
	group := mux.Group("/api/custom-dashboard")
	group.GET("/detail", handler.Get())
	group.POST("/update", handler.Update())

	cases := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{name: "detail", method: http.MethodGet, path: "/api/custom-dashboard/detail?id=dashboard-1"},
		{name: "update", method: http.MethodPost, path: "/api/custom-dashboard/update", body: `{"id":"dashboard-1"}`},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			if tt.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}
			recorder := httptest.NewRecorder()

			mux.ServeHTTP(recorder, req)

			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("expected bad request, got %d with body %s", recorder.Code, recorder.Body.String())
			}
			var failure code.Failure
			if err := json.NewDecoder(recorder.Body).Decode(&failure); err != nil {
				t.Fatalf("decode failure: %v", err)
			}
			if failure.Code != code.UserNoPermissionError {
				t.Fatalf("expected no-permission code, got %s", failure.Code)
			}
			if failure.Message != code.Text(code.LANG_ZH, code.UserNoPermissionError) {
				t.Fatalf("expected localized no-permission message, got %q", failure.Message)
			}
		})
	}
}
