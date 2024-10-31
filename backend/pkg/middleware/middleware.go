package middleware

import (
	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/gin-gonic/gin"
	"net/http"
)

const UserKey = "username"

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO 检查是否在黑名单中
		rawToken := c.Request.Header.Get("Authorization")
		token := parseRawToken(rawToken)
		if len(token) == 0 {
			err := core.Error(http.StatusBadRequest, code.UnAuth, code.Text(code.UnAuth))
			c.AbortWithStatus(http.StatusBadRequest)
			c.Set("_abort_error_", err)
			return
		}

		claims, err := ParseAccessToken(token)
		if err != nil {
			err := core.Error(http.StatusBadRequest, code.InValidToken, code.Text(code.InValidToken))
			c.AbortWithStatus(http.StatusBadRequest)
			c.Set("_abort_error_", err)
			return
		}

		c.Set(UserKey, claims.Username)
		c.Next()
	}
}
