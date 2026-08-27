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
		cluster := &model.Cluster{ID: "cluster-a", Name: " Production "}
		require.NoError(t, normalizeClusterIdentity(cluster))
		require.Equal(t, "Production", cluster.Name)
	})
}
