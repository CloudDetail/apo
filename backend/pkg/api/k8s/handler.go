// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package k8s

import (
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/repository/kubernetes"
	"github.com/CloudDetail/apo/backend/pkg/services/k8s"
)

type Handler interface {
	// GetNamespaceList 获取所有namespace信息
	// @Tags API.k8s
	// @Router /api/k8s/namespaces [get]
	GetNamespaceList() core.HandlerFunc
	// GetNamespaceInfo 获取namespace信息
	// @Tags API.k8s
	// @Router /api/k8s/namespace/info [get]
	GetNamespaceInfo() core.HandlerFunc
	// GetPodList 获取namespace下所有pod信息
	// @Tags API.k8s
	// @Router /api/k8s/pods [get]
	GetPodList() core.HandlerFunc
	// GetPodInfo 获取pod信息
	// @Tags API.k8s
	// @Router /api/k8s/pod/info [get]
	GetPodInfo() core.HandlerFunc
}

type handler struct {
	k8sService k8s.Service
}

func New(k8sRepo kubernetes.Repo) Handler {
	return &handler{
		k8sService: k8s.New(k8sRepo),
	}
}
