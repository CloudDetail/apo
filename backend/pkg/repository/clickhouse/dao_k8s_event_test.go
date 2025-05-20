// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package clickhouse

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
)

func TestChRepo_K8sAlert(t *testing.T) {
	repo := NewTestRepo(t)
	currentTime := time.Now()
	// Get the time 1 hour ago
	oneHourAgo := currentTime.Add(-1 * time.Hour)
	instances := []*model.ServiceInstance{
		{
			PodName:  "apisix-etcd-0",
			NodeName: "worker-23",
		},
	}
	k8sAlert, err := repo.GetK8sAlertEventsSample(core.EmptyCtx(), oneHourAgo, currentTime, instances)
	if err != nil {
		t.Fatalf("Error to get k8sAlert: %v", err)
	}
	for _, event := range k8sAlert {
		info := fmt.Sprintf("%s: %s %s:%s", event.Timestamp.Format("15:04:05"), event.GetObjName(), event.GetReason(), event.Body)
		log.Println(info)
	}
}
