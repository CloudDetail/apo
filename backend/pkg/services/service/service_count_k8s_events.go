// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

// CountK8sEvents get K8s events
func (s *service) CountK8sEvents(req *request.GetK8sEventsRequest) (*response.GetK8sEventsResponse, error) {
	startTime := req.StartTime
	endTime := req.EndTime
	// Get all the instance information of the service first
	instanceList, err := s.promRepo.GetInstanceList(startTime, endTime, req.ServiceName, "")
	if err != nil {
		return nil, err
	}
	// If no Pod instance exists, null is returned.
	podInstances := instanceList.GetPodInstances()
	resp := &response.GetK8sEventsResponse{
		Status:  "normal",
		Reasons: []string{},
		Data:    make(map[string]*response.K8sEventStatistics),
	}
	if len(podInstances) == 0 {
		return resp, nil
	}
	// Use all pod instances as filter criteria to return a list of events related to the time period
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
				DisplayName: count.Reason,
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
	// Container creation
	CreatedContainer = "Created"
	// Container start
	StartedContainer = "Started"
	// Failed to start container
	FailedToCreateContainer = "Failed"
	FailedToStartContainer  = "Failed"
	// Delete container
	KillingContainer = "Killing"
	PreemptContainer = "Preempting"
	// Failed to start container and retry
	BackOffStartContainer = "BackOff"
	// container timeout
	ExceededGracePeriod = "ExceededGracePeriod"
)

// Pod event reason list
const (
	// Failed to delete pod
	FailedToKillPod = "FailedKillPod"
	// Failed to create pod
	FailedToCreatePodContainer     = "FailedCreatePodContainer"
	FailedToMakePodDataDirectories = "Failed"
	// Network not ready
	NetworkNotReady = "NetworkNotReady"
)

// Image event reason list
const (
	// Getting mirror
	PullingImage = "Pulling"
	// Image obtained successfully
	PulledImage = "Pulled"
	// Failed to get mirror
	FailedToPullImage       = "Failed"
	FailedToInspectImage    = "InspectFailed"
	ErrImageNeverPullPolicy = "ErrImageNeverPull"
	// Retry pulling image
	BackOffPullImage = "BackOff"
)

// kubelet event reason list
const (
	// Host node is ready
	NodeReady = "NodeReady"
	// Host node is not ready
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
	// Container health check failed
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
