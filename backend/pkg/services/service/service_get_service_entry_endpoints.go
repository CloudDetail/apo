// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package service

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"github.com/CloudDetail/apo/backend/pkg/services/common"
)

func (s *service) GetServiceEntryEndpoints(ctx core.Context, req *request.GetServiceEntryEndpointsRequest) ([]clickhouse.EntryNode, error) {
	return s.chRepo.ListEntryEndpoints(ctx, req)
}

func (s *service) GetServiceEntryEndpointsInGroup(ctx core.Context, req *request.GetServiceEntryEndpointsRequest) ([]clickhouse.EntryNode, error) {
	selected, err := s.dbRepo.GetScopeIDsSelectedByGroupID(ctx, req.GroupID)
	if err != nil {
		return nil, err
	}

	permSvcList := common.DataGroupStorage.GetFullPermissionSvcList(selected)

	var permSvcMap = make(map[string]struct{})
	for _, svc := range permSvcList {
		permSvcMap[svc] = struct{}{}
	}

	parents, err := s.chRepo.ListAncestorEndpoints(ctx, req)
	if err != nil {
		return nil, err
	}

	var entryInGroup []clickhouse.EntryNode
	var minDepth uint64 = 999
	for i := 0; i < len(parents); i++ {
		if parents[i].Depth > minDepth {
			continue
		}
		if _, ok := permSvcMap[parents[i].Service]; !ok {
			continue
		}
		if parents[i].Depth < minDepth {
			minDepth = parents[i].Depth
			entryInGroup = make([]clickhouse.EntryNode, 0)
		}
		entryInGroup = append(entryInGroup, clickhouse.EntryNode{
			Service:  parents[i].Service,
			Endpoint: parents[i].Endpoint,
		})
	}
	return entryInGroup, nil
}
