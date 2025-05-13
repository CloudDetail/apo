// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	core "github.com/CloudDetail/apo/backend/pkg/core"
)

var _ Service = (*service)(nil)

type Service interface {
	SetTTL(ctx_core core.Context, req *request.SetTTLRequest) error
	SetSingleTableTTL(ctx_core core.Context, req *request.SetSingleTTLRequest) error
	GetTTL(ctx_core core.Context,) (*response.GetTTLResponse, error)
}

type service struct {
	chRepo clickhouse.Repo
}

func New(chRepo clickhouse.Repo) Service {
	return &service{
		chRepo: chRepo,
	}
}
