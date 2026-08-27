// Copyright 2026 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"errors"
	"strings"

	model "github.com/CloudDetail/apo/backend/pkg/model/integration"
)

func normalizeClusterIdentity(cluster *model.Cluster) error {
	cluster.ID = strings.TrimSpace(cluster.ID)
	if cluster.ID == "" {
		return errors.New("cluster_id is required")
	}

	cluster.Name = strings.TrimSpace(cluster.Name)
	if cluster.Name == "" {
		cluster.Name = cluster.ID
	}
	return nil
}
