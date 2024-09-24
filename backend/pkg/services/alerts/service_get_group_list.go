package alerts

import "github.com/CloudDetail/apo/backend/pkg/model/response"

const (
	appLabelVal       = "应用指标"
	infraLabelVal     = "主机相关"
	netLabelVal       = "网络相关"
	containerLabelVal = "容器相关"
	customLabelVal    = "用户自定义组"

	appLabelKey       = "app"
	infraLabelKey     = "infra"
	netLabelKey       = "network"
	containerLabelKey = "container"
	customLabelKey    = "custom"
)

var groupsLabel = map[string]string{
	appLabelKey:       appLabelVal,
	infraLabelKey:     infraLabelVal,
	netLabelKey:       netLabelVal,
	containerLabelKey: containerLabelVal,
	customLabelKey:    customLabelVal,
}

func (s *service) GetGroupList() response.GetGroupListResponse {
	resp := response.GetGroupListResponse{
		GroupsLabel: groupsLabel,
	}

	return resp
}
