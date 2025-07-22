// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package dataplane

import (
	"github.com/CloudDetail/apo/backend/pkg/code"
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/integration"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

var _ DPService = (*service)(nil)

type DPService interface {
	ListDataPlaneType(ctx core.Context) (*response.ListDataPlaneTypeResponse, error)

	CreateDataPlane(ctx core.Context, req *request.CreateDataPlaneRequest) error
	ListDataPlane(ctx core.Context) (*response.ListDataPlaneResponse, error)
	UpdateDataPlane(ctx core.Context, req *request.UpdateDataPlaneRequest) error
	DeleteDataPlane(ctx core.Context, req *request.DeleteDataPlaneRequest) error
}

func (s *service) ListDataPlaneType(ctx core.Context) (*response.ListDataPlaneTypeResponse, error) {
	dpTypes, err := s.dbRepo.ListDataPlaneType(ctx)
	return &response.ListDataPlaneTypeResponse{
		DPTypes: dpTypes,
	}, err
}

func (s *service) CreateDataPlane(ctx core.Context, req *request.CreateDataPlaneRequest) error {
	dpType, err := s.dbRepo.GetDataPlaneType(ctx, req.Typ)
	if err != nil {
		return core.Error(code.DataPlaneTypeNotExistError, "data plane type not exist")
	}

	if err := integration.CheckInvalid(&req.DataPlane, dpType); err != nil {
		return core.Error(code.DataPlaneParamInvalidError, err.Error())
	}

	return s.dbRepo.CreateDataPlane(ctx, &req.DataPlane)
}

func (s *service) ListDataPlane(ctx core.Context) (*response.ListDataPlaneResponse, error) {
	res, err := s.dbRepo.ListDataPlane(ctx)
	return &response.ListDataPlaneResponse{
		DataPlanes: res,
	}, err
}

func (s *service) UpdateDataPlane(ctx core.Context, req *request.UpdateDataPlaneRequest) error {
	if find, err := s.dbRepo.CheckDataPlaneExist(ctx, req.DataPlane.ID); err != nil || !find {
		return core.Error(code.DataPlaneNotExistError, "data plane not exist")
	}

	dpType, err := s.dbRepo.GetDataPlaneType(ctx, req.Typ)
	if err != nil {
		return core.Error(code.DataPlaneTypeNotExistError, "data plane type not exist")
	}

	if err := integration.CheckInvalid(&req.DataPlane, dpType); err != nil {
		return core.Error(code.DataPlaneParamInvalidError, err.Error())
	}

	return s.dbRepo.UpdateDataPlane(ctx, &req.DataPlane)
}

func (s *service) DeleteDataPlane(ctx core.Context, req *request.DeleteDataPlaneRequest) error {
	if find, err := s.dbRepo.CheckDataPlaneExist(ctx, req.ID); err != nil || !find {
		return core.Error(code.DataPlaneNotExistError, "data plane not exist")
	}

	return s.dbRepo.DeleteDataPlane(ctx, req.ID)
}
