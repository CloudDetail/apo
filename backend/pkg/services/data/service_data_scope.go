package data

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/datagroup"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) ListDataScopeByGroupID(ctx core.Context, req *request.DGScopeListRequest) (*response.ListDataScopesResponse, error) {
	options, err := s.dbRepo.GetScopesOptionByGroupID(ctx, req.GroupID)
	if err != nil {
		return nil, err
	}
	selected, err := s.dbRepo.GetScopesSelectedByGroupID(ctx, req.GroupID)
	if err != nil {
		return nil, err
	}

	var scopes *datagroup.DataScopeTreeNode
	if req.SkipNotChecked {
		scopes = s.DataGroupStore.CloneScopeWithPermission(options, selected)
	} else {
		scopes = s.DataGroupStore.CloneScopeWithPermission(selected, nil)
	}

	return &response.ListDataScopesResponse{
		Scopes:      scopes,
		DataSources: selected,
	}, nil
}
