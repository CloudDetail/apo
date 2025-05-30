// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package k8s

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/kubernetes"
)

type Service interface {
	GetNamespaceList(ctx core.Context) (*response.GetNamespaceListResponse, error)
	GetNamespaceInfo(ctx core.Context, req *request.GetNamespaceInfoRequest) (*response.GetNamespaceInfoResponse, error)
	GetPodList(ctx core.Context, req *request.GetPodListRequest) (*response.GetPodListResponse, error)
	GetPodInfo(ctx core.Context, req *request.GetPodInfoRequest) (*response.GetPodInfoResponse, error)
}

type service struct {
	k8sRepo kubernetes.Repo
}

func New(k8sRepo kubernetes.Repo) Service {
	return &service{
		k8sRepo: k8sRepo,
	}
}
