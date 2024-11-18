package model

import (
	"fmt"
	"sort"
	"strconv"
)

// ServiceInstances 包含了 Pod、Container、VM三种场景的所有映射情况，已剔除未设置Pod的数据
type ServiceInstances struct {
	InstanceMap map[string]*ServiceInstance
}

func NewServiceInstances() *ServiceInstances {
	return &ServiceInstances{
		InstanceMap: make(map[string]*ServiceInstance),
	}
}

func (instances *ServiceInstances) AddInstances(list []*ServiceInstance) {
	for _, instance := range list {
		if instance.PodName != "" {
			instances.InstanceMap[instance.getPodInstanceId()] = instance
			instances.InstanceMap[instance.getContainerInstanceId()] = instance
			if instance.Pid != 1 {
				instances.InstanceMap[instance.getVMInstanceId()] = instance
			}
		} else {
			instanceId := ""
			if instance.ContainerId != "" {
				instanceId = instance.getContainerInstanceId()
			} else {
				if instance.Pid == 1 {
					continue
				}
				instanceId = instance.getVMInstanceId()
			}
			if _, exist := instances.InstanceMap[instanceId]; !exist {
				// 如果已存在Pod则不覆盖
				instances.InstanceMap[instanceId] = instance
			}
		}
	}
}

func (instances *ServiceInstances) GetPodInstances() []string {
	pods := make([]string, 0)
	for instanceId, instance := range instances.InstanceMap {
		// 去重
		if instance.PodName != "" && instanceId == instance.PodName {
			pods = append(pods, instance.PodName)
		}
	}
	return pods
}

func (instances *ServiceInstances) GetInstances() []*ServiceInstance {
	instanceList := make([]*ServiceInstance, 0)
	if len(instances.InstanceMap) == 0 {
		return instanceList
	}

	for _, instance := range instances.GetInstanceIdMap() {
		instanceList = append(instanceList, instance)
	}
	return instanceList
}

func (instances *ServiceInstances) GetInstanceIds() []string {
	instanceIdList := make([]string, 0)
	if len(instances.InstanceMap) == 0 {
		return instanceIdList
	}

	for instanceId := range instances.GetInstanceIdMap() {
		instanceIdList = append(instanceIdList, instanceId)
	}
	// 基于名称排序
	sort.Strings(instanceIdList)
	return instanceIdList
}

func (instances *ServiceInstances) GetInstanceIdMap() map[string]*ServiceInstance {
	// 使用Map去重
	instanceMap := make(map[string]*ServiceInstance)
	for _, instance := range instances.InstanceMap {
		if instance.PodName != "" {
			instanceMap[instance.getPodInstanceId()] = instance
		} else if instance.ContainerId != "" {
			instanceMap[instance.getContainerInstanceId()] = instance
		} else {
			instanceMap[instance.getVMInstanceId()] = instance
		}
	}
	return instanceMap
}

type ServiceInstance struct {
	ServiceName string `json:"service"`     // 服务名
	ContainerId string `json:"containerId"` // 容器ID
	PodName     string `json:"podName"`     // Pod名
	Namespace   string `json:"-"`
	NodeName    string `json:"nodeName"` // 主机名
	Pid         int64  `json:"pid"`      // 进程号
	NodeIP      string `json:"nodeIp"`
}

func (i *ServiceInstance) MatchSvcTags(group string, tags map[string]string) bool {
	switch group {
	case "app":
		if len(i.ServiceName) > 0 {
			return i.ServiceName == tags["svc_name"]
		}
	case "container":
		if len(i.PodName) > 0 {
			pod, find := tags["pod"]
			if !find {
				return false
			}
			namespace, find := tags["namespace"]
			if !find {
				return false
			}
			return i.PodName == pod && i.Namespace == namespace
		}
	case "network":
		if len(i.PodName) > 0 {
			pod, find := tags["src_pod"]
			if !find {
				return false
			}
			namespace, find := tags["src_namespace"]
			if !find {
				return false
			}
			return i.PodName == pod && i.Namespace == namespace
		} else if i.Pid > 0 {
			pid, find := tags["pid"]
			if !find {
				return false
			}
			node, find := tags["src_node"]
			if !find {
				return false
			}

			return strconv.Itoa(int(i.Pid)) == pid && i.NodeName == node
		}
	case "infra":
		if len(i.NodeName) > 0 {
			node, find := tags["instance_name"]
			if !find {
				return false
			}
			return node == i.NodeName
		}
	}
	return false
}

func (instance *ServiceInstance) GetInstanceId() string {
	if instance.PodName != "" {
		return instance.getPodInstanceId()
	}
	if instance.ContainerId != "" {
		return instance.getContainerInstanceId()
	}
	return instance.getVMInstanceId()
}

func (instance *ServiceInstance) getPodInstanceId() string {
	return instance.PodName
}

func (instance *ServiceInstance) getContainerInstanceId() string {
	return fmt.Sprintf("%s@%s@%s", instance.ServiceName, instance.NodeName, instance.ContainerId)
}

func (instance *ServiceInstance) getVMInstanceId() string {
	return fmt.Sprintf("%s@%s@%d", instance.ServiceName, instance.NodeName, instance.Pid)
}
