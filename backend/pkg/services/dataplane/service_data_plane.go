// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package dataplane

import (
	"encoding/json"
	"os"
	"sort"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/code"
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/integration"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

const showAPODataPlaneTypeENV = "SHOW_APO_DATA_PLANE_TYPE"

var _ DPService = (*service)(nil)

type DPService interface {
	ListDataPlaneType(ctx core.Context) (*response.ListDataPlaneTypeResponse, error)

	CreateDataPlane(ctx core.Context, req *request.CreateDataPlaneRequest) error
	ListDataPlane(ctx core.Context) (*response.ListDataPlaneResponse, error)
	UpdateDataPlane(ctx core.Context, req *request.UpdateDataPlaneRequest) error
	DeleteDataPlane(ctx core.Context, req *request.DeleteDataPlaneRequest) error

	CheckMinimalDataPlaneExist(ctx core.Context) (*response.CheckMinimalDataPlaneResponse, error)
}

var specialOrder = map[string]int{
	"datadog":           0,
	"jaeger":            1,
	"prometheusmetric":  2,
	"tempo":             3,
	"loki":              4,
	"grafanaPrometheus": 5,
	"signoz":            6,
	"skywalking":        7,
	"arms":              8,
	"apo":               9,
}

func (s *service) ListDataPlaneType(ctx core.Context) (*response.ListDataPlaneTypeResponse, error) {
	dpTypes, err := s.dbRepo.ListDataPlaneType(ctx)
	for i := 0; i < len(dpTypes); i++ {
		dpTypes[i].UseLang(ctx.LANG())
	}

	if os.Getenv(showAPODataPlaneTypeENV) != "true" {
		for i := 0; i < len(dpTypes); i++ {
			if dpTypes[i].TypeName == "apo" {
				dpTypes = append(dpTypes[:i], dpTypes[i+1:]...)
				break
			}
		}
	}

	sort.SliceStable(dpTypes, func(i, j int) bool {
		orderI, findI := specialOrder[dpTypes[i].TypeName]
		if !findI {
			orderI = len(specialOrder)
		}
		orderJ, findJ := specialOrder[dpTypes[j].TypeName]
		if !findJ {
			orderJ = len(specialOrder)
		}
		if orderI == orderJ {
			return dpTypes[i].TypeName < dpTypes[j].TypeName
		}
		return orderI < orderJ
	})

	return &response.ListDataPlaneTypeResponse{
		DPTypes: dpTypes,
	}, err
}

func (s *service) CreateDataPlane(ctx core.Context, req *request.CreateDataPlaneRequest) error {
	dpType, err := s.dbRepo.GetDataPlaneType(ctx, req.Typ)
	if err != nil {
		return core.Error(code.DataPlaneTypeNotExistError, "data plane type not exist")
	}

	if len(req.Capability.Obj) == 0 {
		// Full capability
		req.Capability.Obj = dpType.CapabilitySpec.Obj
	}

	if err := integration.CheckInvalid(&req.DataPlane, dpType); err != nil {
		return core.Error(code.DataPlaneParamInvalidError, err.Error())
	}

	// HACK prometheus Prometheus 数据平面特殊处理
	if dpType.TypeName == "prometheus" {
		var paramsJson map[string]any
		err := json.Unmarshal([]byte(req.DataPlane.Params), &paramsJson)
		if err != nil {
			return core.Error(code.DataPlaneParamInvalidError, err.Error())
		}

		paramsJson["promStorage"] = "prom"
		updatedParams, _ := json.Marshal(paramsJson)
		req.DataPlane.Params = string(updatedParams)
	}

	req.DataPlane.IsDeleted = false

	if len(req.DataPlane.ClusterID) > 0 {
		var createDataPlane = func(ctx core.Context) error {
			return s.dbRepo.CreateDataPlane(ctx, &req.DataPlane)
		}

		var createCluster2DataPlane = func(ctx core.Context) error {
			return s.dbRepo.CreateCluster2DataPlane(ctx, &integration.Cluster2DataPlane{
				ClusterID:   req.DataPlane.ClusterID,
				DataPlaneID: req.DataPlane.ID,
			})
		}

		return s.dbRepo.Transaction(ctx, createDataPlane, createCluster2DataPlane)
	}

	return s.dbRepo.CreateDataPlane(ctx, &req.DataPlane)
}

func (s *service) ListDataPlane(ctx core.Context) (*response.ListDataPlaneResponse, error) {
	res, err := s.dbRepo.ListDataPlane(ctx)

	for i := 0; i < len(res); i++ {
		res[i].UseLang(ctx.LANG())
	}

	// 过滤掉 APO 数据平面
	// if os.Getenv(showAPODataPlaneTypeENV) != "true" {
	// 	for i := 0; i < len(res); i++ {
	// 		if res[i].Typ == "apo" {
	// 			res = append(res[:i], res[i+1:]...)
	// 			i--
	// 		}
	// 	}
	// }

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
	// if find, err := s.dbRepo.CheckDataPlaneExist(ctx, req.ID); err != nil || !find {
	// 	return core.Error(code.DataPlaneNotExistError, "data plane not exist")
	// }
	dp, err := s.dbRepo.GetDataplane(ctx, req.ID)
	if err != nil {
		return core.Error(code.DataPlaneNotExistError, "data plane not exist")
	}

	if dp.Typ == "apo" {
		return core.Error(code.DeleteDataPlaneError, "forbidden to delete apo data plane")
	}

	var deleteCluster2DataPlane = func(ctx core.Context) error {
		return s.dbRepo.DeleteCluster2DataPlane(ctx, req.ID)
	}
	var deleteDataPlane = func(ctx core.Context) error {
		return s.dbRepo.DeleteDataPlane(ctx, req.ID)
	}
	return s.dbRepo.Transaction(ctx, deleteCluster2DataPlane, deleteDataPlane)
}

func (s *service) CheckMinimalDataPlaneExist(ctx core.Context) (*response.CheckMinimalDataPlaneResponse, error) {
	now := time.Now()
	clusterIDs, err := s.promRepo.GetMinimalAgentExisted(ctx, now, 7*time.Hour*24)
	if err != nil {
		return nil, err
	}

	dpList, err := s.dbRepo.ListDataPlane(ctx)
	if err != nil {
		return nil, err
	}

	var dataSources []string
	for _, dp := range dpList {
		dataSources = append(dataSources, dp.Name)
	}

	return &response.CheckMinimalDataPlaneResponse{
		AgentReady:     len(clusterIDs) > 0,
		Clusters:       clusterIDs,
		DataPlaneExist: len(dpList) > 0,
		DataSource:     dataSources,
	}, nil

}
