package kubernetes

const (
	AppLabelVal       = "应用指标"
	InfraLabelVal     = "主机相关"
	NetLabelVal       = "网络相关"
	ContainerLabelVal = "容器相关"
	CustomLabelVal    = "用户自定义组"

	AppLabelKey       = "app"
	InfraLabelKey     = "infra"
	NetLabelKey       = "network"
	ContainerLabelKey = "container"
	CustomLabelKey    = "custom"
)

var GroupsLabel = map[string]string{
	AppLabelKey:       AppLabelVal,
	InfraLabelKey:     InfraLabelVal,
	NetLabelKey:       NetLabelVal,
	ContainerLabelKey: ContainerLabelVal,
	CustomLabelKey:    CustomLabelVal,
}

var reversedGroupsLabel = map[string]string{
	AppLabelVal:       AppLabelKey,
	InfraLabelVal:     InfraLabelKey,
	NetLabelVal:       NetLabelKey,
	ContainerLabelVal: ContainerLabelKey,
	CustomLabelVal:    CustomLabelKey,
}

func GetLabel(group string) (string, bool) {
	label, ok := reversedGroupsLabel[group]
	return label, ok
}
