package data

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/datagroup"
)

func (s *service) ListDataScopeByGroupID(ctx core.Context, groupID int64) (*datagroup.DataScopeTreeNode, error) {
	options, err := s.dbRepo.GetScopesOptionByGroupID(ctx, groupID)
	if err != nil {
		return nil, err
	}
	selected, err := s.dbRepo.GetScopesSelectedByGroupID(ctx, groupID)
	if err != nil {
		return nil, err
	}
	scopes := s.DataGroupStore.CloneScopeWithPermission(options, selected)
	return scopes, nil
}
