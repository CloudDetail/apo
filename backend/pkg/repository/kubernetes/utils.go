// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package kubernetes

const (
	AppLabelVal        = "应用指标"
	InfraLabelVal      = "主机相关"
	NetLabelVal        = "网络相关"
	ContainerLabelVal  = "容器相关"
	MiddlewareLabelVal = "数据库和中间件相关"
	CustomLabelVal     = "用户自定义组"

	MutationAppLabelVal        = "异常检测-应用指标"
	MutationInfraLabelVal      = "异常检测-主机相关"
	MutationNetLabelVal        = "异常检测-网络相关"
	MutationContainerLabelVal  = "异常检测-容器相关"
	MutationMiddlewareLabelVal = "异常检测-数据库和中间件相关"
	MutationCustomLabelVal     = "异常检测-用户自定义组"

	AppLabelKey        = "app"
	InfraLabelKey      = "infra"
	NetLabelKey        = "network"
	ContainerLabelKey  = "container"
	MiddlewareLabelKey = "middleware"
	CustomLabelKey     = "custom"

	MutationAppLabelKey        = "mutation-app"
	MutationInfraLabelKey      = "mutation-infra"
	MutationNetLabelKey        = "mutation-network"
	MutationContainerLabelKey  = "mutation-container"
	MutationMiddlewareLabelKey = "mutation-middleware"
	MutationCustomLabelKey     = "mutation-custom"
)

var GroupsLabel = map[string]string{
	AppLabelKey:        AppLabelVal,
	InfraLabelKey:      InfraLabelVal,
	NetLabelKey:        NetLabelVal,
	ContainerLabelKey:  ContainerLabelVal,
	MiddlewareLabelKey: MiddlewareLabelVal,
	CustomLabelKey:     CustomLabelVal,

	MutationAppLabelKey:        MutationAppLabelVal,
	MutationInfraLabelKey:      MutationInfraLabelVal,
	MutationNetLabelKey:        MutationNetLabelVal,
	MutationContainerLabelKey:  MutationContainerLabelVal,
	MutationMiddlewareLabelKey: MutationMiddlewareLabelVal,
	MutationCustomLabelKey:     MutationCustomLabelVal,
}

var reversedGroupsLabel = map[string]string{
	AppLabelVal:        AppLabelKey,
	InfraLabelVal:      InfraLabelKey,
	NetLabelVal:        NetLabelKey,
	ContainerLabelVal:  ContainerLabelKey,
	MiddlewareLabelVal: MiddlewareLabelKey,
	CustomLabelVal:     CustomLabelKey,

	MutationAppLabelVal:        MutationAppLabelKey,
	MutationInfraLabelVal:      MutationInfraLabelKey,
	MutationNetLabelVal:        MutationNetLabelKey,
	MutationContainerLabelVal:  MutationContainerLabelKey,
	MutationMiddlewareLabelVal: MutationMiddlewareLabelKey,
	MutationCustomLabelVal:     MutationCustomLabelKey,
}

func GetLabel(group string) (string, bool) {
	label, ok := reversedGroupsLabel[group]
	return label, ok
}
