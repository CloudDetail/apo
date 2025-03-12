// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alerts

import (
	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/kubernetes"
)

func (s *service) GetGroupList(ctx core.Context) response.GetGroupListResponse {

	if ctx.LANG() == code.LANG_ZH {
		return response.GetGroupListResponse{
			GroupsLabel: kubernetes.GroupsCNLabel,
		}
	} else {
		return response.GetGroupListResponse{
			GroupsLabel: kubernetes.GroupsENLabel,
		}
	}
}
