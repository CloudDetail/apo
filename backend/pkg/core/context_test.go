// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestContext_Clone(t *testing.T) {
	originalCtx := &context{
		ctx: &gin.Context{},
	}
	originalCtx.Set("user_id", 123)
	originalCtx.Set("is_admin", true)

	clonedCtx := originalCtx.Clone()

	assert.NotNil(t, clonedCtx, "cloned context should not be nil")
	assert.NotSame(t, originalCtx, clonedCtx, "cloned context should be a new struct")

	originalUserID, _ := originalCtx.Get("user_id")
	originalIsAdmin, _ := originalCtx.Get("is_admin")

	clonedUserID, _ := clonedCtx.Get("user_id")
	clonedIsAdmin, _ := clonedCtx.Get("is_admin")

	assert.Equal(t, originalUserID, clonedUserID, "cloned context should have the same user_id")
	assert.Equal(t, originalIsAdmin, clonedIsAdmin, "cloned context should have the same is_admin")

	clonedCtx.Set("user_id", 456)
	clonedCtx.Set("new_key", "test")

	modifiedOriginalUserID, _ := originalCtx.Get("user_id")
	assert.Equal(t, originalUserID, modifiedOriginalUserID, "")

	_, exists := originalCtx.Get("new_key")
	assert.False(t, exists, "There should be no newly added keys in the original object")
}
