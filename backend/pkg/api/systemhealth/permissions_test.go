// Copyright 2026 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package systemhealth

import (
	"reflect"
	"testing"

	"github.com/CloudDetail/apo/backend/pkg/model/datagroup"
	"github.com/CloudDetail/apo/backend/pkg/model/profile"
)

func TestSystemHealthReportGroupIDsOnlyIncludesVisibleGroups(t *testing.T) {
	groups := []*datagroup.DataGroupTreeNode{
		{DataGroup: datagroup.DataGroup{GroupID: 102}},
		nil,
		{DataGroup: datagroup.DataGroup{GroupID: 101}},
	}

	if got := systemHealthReportGroupIDs(groups, false); !reflect.DeepEqual(got, []int64{102, 101}) {
		t.Fatalf("unexpected visible report groups: %v", got)
	}
	if got := systemHealthReportGroupIDs(groups, true); !reflect.DeepEqual(got, []int64{0, 102, 101}) {
		t.Fatalf("expected root report group for root permission, got %v", got)
	}
	if systemHealthReportGroupAllowed([]int64{101}, 0) {
		t.Fatal("root report group must require explicit root permission")
	}
}

func TestIsSystemHealthAdminUser(t *testing.T) {
	viewer := profile.User{RoleList: []profile.Role{{RoleName: "viewer"}}}
	admin := profile.User{RoleList: []profile.Role{{RoleName: "viewer"}, {RoleName: "admin"}}}
	if isSystemHealthAdminUser(viewer) {
		t.Fatal("viewer must not manage global system health config")
	}
	if !isSystemHealthAdminUser(admin) {
		t.Fatal("admin must be allowed to manage global system health config")
	}
}
