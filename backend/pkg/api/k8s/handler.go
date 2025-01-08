// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package k8s

import (
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/repository/kubernetes"
	"github.com/CloudDetail/apo/backend/pkg/services/k8s"
)

type Handler interface {
	// GetNamespaceList get all namespace information
	// @Tags API.k8s
	// @Router /api/k8s/namespaces [get]
	GetNamespaceList() core.HandlerFunc
	// GetNamespaceInfo get namespace information
	// @Tags API.k8s
	// @Router /api/k8s/namespace/info [get]
	GetNamespaceInfo() core.HandlerFunc
	// GetPodList get information about all pods in the namespace
	// @Tags API.k8s
	// @Router /api/k8s/pods [get]
	GetPodList() core.HandlerFunc
	// GetPodInfo get pod information
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
