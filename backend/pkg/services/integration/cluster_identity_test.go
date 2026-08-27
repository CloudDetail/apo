// Copyright 2026 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"testing"

	model "github.com/CloudDetail/apo/backend/pkg/model/integration"
	"github.com/stretchr/testify/require"
)

func TestNormalizeClusterIdentity(t *testing.T) {
	t.Run("requires cluster id", func(t *testing.T) {
		cluster := &model.Cluster{Name: "display-name"}
		require.EqualError(t, normalizeClusterIdentity(cluster), "cluster_id is required")
	})

	t.Run("uses cluster id as an empty display name", func(t *testing.T) {
		cluster := &model.Cluster{ID: " cluster-a ", Name: "  "}
		require.NoError(t, normalizeClusterIdentity(cluster))
		require.Equal(t, "cluster-a", cluster.ID)
		require.Equal(t, "cluster-a", cluster.Name)
	})

	t.Run("keeps an explicit display name", func(t *testing.T) {
		cluster := &model.Cluster{ID: "cluster-a", Name: " Production ", ClusterType: "kubernetes"}
		require.NoError(t, normalizeClusterIdentity(cluster))
		require.Equal(t, "Production", cluster.Name)
		require.Equal(t, model.ClusterTypeK8s, cluster.ClusterType)
	})

	t.Run("normalizes vm aliases", func(t *testing.T) {
		cluster := &model.Cluster{ID: "cluster-vm", ClusterType: "binary"}
		require.NoError(t, normalizeClusterIdentity(cluster))
		require.Equal(t, model.ClusterTypeVM, cluster.ClusterType)
	})

	t.Run("rejects unknown cluster type", func(t *testing.T) {
		cluster := &model.Cluster{ID: "cluster-a", ClusterType: "unknown"}
		require.EqualError(t, normalizeClusterIdentity(cluster), `unsupported cluster type "unknown"; expected k8s or vm`)
	})
}
