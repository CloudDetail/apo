// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package dataplane

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

func (s *service) DeleteCustomTopology(ctx core.Context, req *request.DeleteCustomTopologyRequest) error {
	return s.dbRepo.DeleteCustomServiceTopology(ctx, req.ID)
}
