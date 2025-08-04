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

	assert.NotNil(t, clonedCtx, "克隆后的上下文不应为 nil")
	assert.NotSame(t, originalCtx, clonedCtx, "克隆后的上下文应该是一个新对象")

	originalUserID, _ := originalCtx.Get("user_id")
	originalIsAdmin, _ := originalCtx.Get("is_admin")

	clonedUserID, _ := clonedCtx.Get("user_id")
	clonedIsAdmin, _ := clonedCtx.Get("is_admin")

	assert.Equal(t, originalUserID, clonedUserID, "克隆对象的 user_id 应该与原始对象相同")
	assert.Equal(t, originalIsAdmin, clonedIsAdmin, "克隆对象的 is_admin 应该与原始对象相同")

	clonedCtx.Set("user_id", 456)
	clonedCtx.Set("new_key", "test")

	modifiedOriginalUserID, _ := originalCtx.Get("user_id")
	assert.Equal(t, originalUserID, modifiedOriginalUserID, "修改克隆对象不应影响原始对象")

	_, exists := originalCtx.Get("new_key")
	assert.False(t, exists, "原始对象中不应该有新添加的键")
}
