// Copyright 2026 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package dify

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoginByAPOSession(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != DIFY_APO_LOGIN {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		var body struct {
			Email      string `json:"email"`
			RememberMe bool   `json:"remember_me"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body.Email != "admin@apo.com" || !body.RememberMe {
			t.Fatalf("unexpected login body: %+v", body)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"result":"success","data":{"access_token":"access","refresh_token":"refresh"}}`))
	}))
	defer server.Close()

	repo := &difyRepo{
		cli: &DifyClient{Client: server.Client(), BaseURL: server.URL},
		url: server.URL,
	}
	result, err := repo.LoginByAPOSession("admin")
	if err != nil {
		t.Fatal(err)
	}
	if result.Data.AccessToken != "access" || result.Data.RefreshToken != "refresh" {
		t.Fatalf("unexpected token pair: %+v", result.Data)
	}
}

func TestLoginByAPOSessionRejectsIncompleteCredentials(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"result":"success","data":{"access_token":"access"}}`))
	}))
	defer server.Close()

	repo := &difyRepo{
		cli: &DifyClient{Client: server.Client(), BaseURL: server.URL},
		url: server.URL,
	}
	if _, err := repo.LoginByAPOSession("admin"); err == nil {
		t.Fatal("expected incomplete credentials to fail")
	}
}
