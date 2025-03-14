// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package dify

import (
	"net/http"
)

const (
	DIFY_WORKFLOWS_RUN   = "/v1/workflows/run"
	DIFY_ADD_USER        = "/console/api/workspaces/apo/members/add"
	DIFY_REMOVE_USER     = "/console/api/workspaces/apo/members/"
	DIFY_PASSWORD_UPDATE = "/console/api/apo/account/password"
	DIFY_RESET_PASSWORD  = "/console/api/apo/account/reset-password"
)

type DifyUser struct {
	Password    string `json:"password"`
	NewPassword string `json:"new_password"`
	Role        string `json:"role"`
	Username    string `json:"username"`
}

type DifyResponse struct {
	Result  string `json:"result"`
	Message string `json:"message"`
}

type DifyRepo interface {
	AddUser(username string, password string, role string) (*DifyResponse, error)
	UpdatePassword(username string, oldPassword string, newPassword string) (*DifyResponse, error)
	RemoveUser(username string) (*DifyResponse, error)
	ResetPassword(username string, newPassword string) (*DifyResponse, error)
}

type difyRepo struct {
	cli *http.Client
}

func New() (DifyRepo, error) {
	client := &http.Client{}
	return &difyRepo{cli: client}, nil
}
