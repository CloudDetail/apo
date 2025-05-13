// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	core "github.com/CloudDetail/apo/backend/pkg/core"
)

func (s *service) GetServiceEntryEndpoints(ctx_core core.Context, req *request.GetServiceEntryEndpointsRequest) ([]clickhouse.EntryNode, error) {
	return s.chRepo.ListEntryEndpoints(req)
}
