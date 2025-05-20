// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package dify

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddUser(t *testing.T) {
	difyRepo, err := New()
	if err != nil {
		t.Fatalf("Failed to create difyRepo: %v", err)
	}
	resp, err := difyRepo.AddUser("test", "APO2024@admin", "admin")
	if err != nil {
		t.Fatalf("Failed to add user: %v", err)
	}
	assert.Equal(t, "success", resp.Result)
}

func TestUpdatePassword(t *testing.T) {
	difyRepo, err := New()
	if err != nil {
		t.Fatalf("Failed to create difyRepo: %v", err)
	}
	resp, err := difyRepo.UpdatePassword("test", "APO2024@admin", "test123456")
	if err != nil {
		t.Fatalf("Failed to add user: %v", err)
	}
	assert.Equal(t, "success", resp.Result)
}

func TestRemoveUser(t *testing.T) {
	difyRepo, err := New()
	if err != nil {
		t.Fatalf("Failed to create difyRepo: %v", err)
	}
	resp, err := difyRepo.RemoveUser("test")
	if err != nil {
		t.Fatalf("Failed to add user: %v", err)
	}
	fmt.Println(resp.Message)
	assert.Equal(t, "success", resp.Result)
}

func TestRunWorkflows(t *testing.T) {
	difyRepo, err := New()
	if err != nil {
		t.Fatalf("Failed to create difyRepo: %v", err)
	}
	req := &WorkflowRequest{
		Inputs:       json.RawMessage(`{"input": "test"}`),
		ResponseMode: "stream",
		User:         "test",
	}
	_, err = difyRepo.WorkflowsRun(req, "test")
	if err != nil {
		t.Fatalf("Failed to run workflow: %v", err)
	}

}
