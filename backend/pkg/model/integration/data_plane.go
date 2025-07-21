// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"encoding/json"
	"fmt"
	"slices"
	"time"
)

type DataPlane struct {
	ID   int    `gorm:"id;primaryKey;autoIncrement" json:"id"`
	Name string `gorm:"name" json:"name"`

	Typ        string              `gorm:"typ" json:"typ"`
	Params     string              `gorm:"params" json:"params"`         // JSONStr
	Capability JSONField[[]string] `gorm:"capability" json:"capability"` // Str Slice

	UpdatedAt time.Time `gorm:"update_at;autoUpdateTime:second" json:"-"`
	IsDelete  bool      `gorm:"is_delete" json:"-"`
}

func (DataPlane) TableName() string {
	return "data_plane"
}

type DataPlaneWithType struct {
	DataPlane
	DataPlaneType
}

type DataPlaneWithClusterIDs struct {
	DataPlaneWithType
	ClusterIDs []string `json:"clusterIds"`
}

func CheckInvalid(d *DataPlane, dpt *DataPlaneType) error {
	if len(dpt.CapabilitySpec.Obj) == 0 {
		return fmt.Errorf("unexpected empty capability in data plane spec: %s", dpt.TypeName)
	}

	for _, capability := range d.Capability.Obj {
		if !slices.Contains(dpt.CapabilitySpec.Obj, capability) {
			return fmt.Errorf("unsupported capability %s in data plane: %s", capability, d.Name)
		}
	}

	var obj interface{}
	if err := json.Unmarshal([]byte(d.Params), &obj); err != nil {
		return err
	}
	return ValidateJSON(obj, dpt.ParamSpec.Obj)
}

func (d *DataPlaneWithType) CheckInvalid() error {
	return CheckInvalid(&d.DataPlane, &d.DataPlaneType)
}

type Cluster2DataPlane struct {
	ClusterID   string `gorm:"cluster_id;primary_key" json:"clusterId"`
	DataPlaneID int    `gorm:"data_plane_id;primary_key" json:"dataPlaneId"`
}
