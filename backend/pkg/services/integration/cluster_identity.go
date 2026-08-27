// Copyright 2026 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"errors"
	"fmt"
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
	switch strings.ToLower(strings.TrimSpace(cluster.ClusterType)) {
	case model.ClusterTypeK8s, "kubernetes", "helm", "":
		cluster.ClusterType = model.ClusterTypeK8s
	case model.ClusterTypeVM, "binary":
		cluster.ClusterType = model.ClusterTypeVM
	default:
		return fmt.Errorf("unsupported cluster type %q; expected k8s or vm", cluster.ClusterType)
	}
	return nil
}
