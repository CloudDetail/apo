package service

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

// CountK8sEvents 获取K8s事件
func (s *service) CountK8sEvents(req *request.GetK8sEventsRequest) (*response.GetK8sEventsResponse, error) {
	startTime := req.StartTime
	endTime := req.EndTime
	// 首先获取所有该服务下的实例信息
	instanceList, err := s.promRepo.GetInstanceList(startTime, endTime, req.ServiceName, "")
	if err != nil {
		return nil, err
	}
	// 如果不存在Pod实例，则直接返回空
	podInstances := instanceList.GetPodInstances()
	resp := &response.GetK8sEventsResponse{
		Status:  "normal",
		Reasons: []string{},
		Data:    make(map[string]*response.K8sEventStatistics),
	}
	if len(podInstances) == 0 {
		return resp, nil
	}
	// 将所有Pod实例作为筛选条件，返回时间段相关的事件列表
	counts, err := s.chRepo.CountK8sEvents(startTime, endTime, podInstances)
	if err != nil {
		return resp, err
	}
	eventCountMap := make(map[string]*response.K8sEventStatistics)
	warningReasons := make(map[string]bool)
	for _, count := range counts {
		if count.TimeRange == "current" && count.Count > 0 && count.Severity == "Warning" {
			resp.Status = "critical"
			warningReasons[count.Reason] = true
		}
		eventCount, ok := eventCountMap[count.Reason]
		if !ok {
			eventCountMap[count.Reason] = &response.K8sEventStatistics{
				EventName:   count.Reason,
				DisplayName: convertReason(count.Reason),
				Severity:    count.Severity,
				Counts:      response.K8sEventCountValues{},
			}
			eventCountMap[count.Reason].Counts.AddCount(count)
		} else {
			eventCount.Counts.AddCount(count)
		}
	}

	resp.Data = eventCountMap
	for warningReason := range warningReasons {
		resp.Reasons = append(resp.Reasons, warningReason)
	}
	return resp, nil
}

func convertReason(reason string) string {
	switch reason {
	case CreatedContainer:
		return "容器创建成功"
	case StartedContainer:
		return "容器启动成功"
	case FailedToCreateContainer:
		return "容器创建失败"
	case BackOffStartContainer:
		return "容器启动失败重试"
	case ContainerUnhealthy:
		return "容器健康检查失败"
	case NodeReady:
		return "节点已就绪"
	case NodeNotReady:
		return "节点未就绪"
	default:
		return reason
	}
}

// Copied from kubernetes/pkg/kubelet/events/event.go
// Container event reason list
const (
	// 容器创建
	CreatedContainer = "Created"
	// 容器启动
	StartedContainer = "Started"
	// 启动容器失败
	FailedToCreateContainer = "Failed"
	FailedToStartContainer  = "Failed"
	// 删除容器
	KillingContainer = "Killing"
	PreemptContainer = "Preempting"
	// 启动容器失败重试
	BackOffStartContainer = "BackOff"
	// 容器超时
	ExceededGracePeriod = "ExceededGracePeriod"
)

// Pod event reason list
const (
	// 删除 Pod 失败
	FailedToKillPod = "FailedKillPod"
	// 创建 Pod 失败
	FailedToCreatePodContainer     = "FailedCreatePodContainer"
	FailedToMakePodDataDirectories = "Failed"
	// 网络未就绪
	NetworkNotReady = "NetworkNotReady"
)

// Image event reason list
const (
	// 正在获取镜像
	PullingImage = "Pulling"
	// 获取镜像成功
	PulledImage = "Pulled"
	// 获取镜像失败
	FailedToPullImage       = "Failed"
	FailedToInspectImage    = "InspectFailed"
	ErrImageNeverPullPolicy = "ErrImageNeverPull"
	// 重试拉取镜像
	BackOffPullImage = "BackOff"
)

// kubelet event reason list
const (
	// 主机节点就绪
	NodeReady = "NodeReady"
	// 主机节点未就绪
	NodeNotReady                         = "NodeNotReady"
	NodeSchedulable                      = "NodeSchedulable"
	NodeNotSchedulable                   = "NodeNotSchedulable"
	StartingKubelet                      = "Starting"
	KubeletSetupFailed                   = "KubeletSetupFailed"
	FailedAttachVolume                   = "FailedAttachVolume"
	FailedMountVolume                    = "FailedMount"
	VolumeResizeFailed                   = "VolumeResizeFailed"
	VolumeResizeSuccess                  = "VolumeResizeSuccessful"
	FileSystemResizeFailed               = "FileSystemResizeFailed"
	FileSystemResizeSuccess              = "FileSystemResizeSuccessful"
	FailedMapVolume                      = "FailedMapVolume"
	WarnAlreadyMountedVolume             = "AlreadyMountedVolume"
	SuccessfulAttachVolume               = "SuccessfulAttachVolume"
	SuccessfulMountVolume                = "SuccessfulMountVolume"
	NodeRebooted                         = "Rebooted"
	NodeShutdown                         = "Shutdown"
	ContainerGCFailed                    = "ContainerGCFailed"
	ImageGCFailed                        = "ImageGCFailed"
	FailedNodeAllocatableEnforcement     = "FailedNodeAllocatableEnforcement"
	SuccessfulNodeAllocatableEnforcement = "NodeAllocatableEnforced"
	SandboxChanged                       = "SandboxChanged"
	FailedCreatePodSandBox               = "FailedCreatePodSandBox"
	FailedStatusPodSandBox               = "FailedPodSandBoxStatus"
	FailedMountOnFilesystemMismatch      = "FailedMountOnFilesystemMismatch"
	FailedPrepareDynamicResources        = "FailedPrepareDynamicResources"
	PossibleMemoryBackedVolumesOnDisk    = "PossibleMemoryBackedVolumesOnDisk"
	CgroupV1                             = "CgroupV1"
)

// Image manager event reason list
const (
	InvalidDiskCapacity = "InvalidDiskCapacity"
	FreeDiskSpaceFailed = "FreeDiskSpaceFailed"
)

// Probe event reason list
const (
	// 容器健康检查失败
	ContainerUnhealthy    = "Unhealthy"
	ContainerProbeWarning = "ProbeWarning"
)

// Pod worker event reason list
const (
	FailedSync = "FailedSync"
)

// Config event reason list
const (
	FailedValidation = "FailedValidation"
)

// Lifecycle hooks
const (
	FailedPostStartHook = "FailedPostStartHook"
	FailedPreStopHook   = "FailedPreStopHook"
)
