// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package k8s

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/kubernetes"
)

type Service interface {
	GetNamespaceList() (*response.GetNamespaceListResponse, error)
	GetNamespaceInfo(req *request.GetNamespaceInfoRequest) (*response.GetNamespaceInfoResponse, error)
	GetPodList(req *request.GetPodListRequest) (*response.GetPodListResponse, error)
	GetPodInfo(req *request.GetPodInfoRequest) (*response.GetPodInfoResponse, error)
}

type service struct {
	k8sRepo kubernetes.Repo
}

func New(k8sRepo kubernetes.Repo) Service {
	return &service{
		k8sRepo: k8sRepo,
	}
}
