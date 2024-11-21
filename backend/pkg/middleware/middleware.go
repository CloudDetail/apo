package middleware

import (
	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/repository/cache"
	"github.com/CloudDetail/apo/backend/pkg/util"
	"github.com/gin-gonic/gin"
	"net/http"
)

const UserKey = "username"

func Auth(tokenCache cache.Repo) gin.HandlerFunc {
	return func(c *gin.Context) {
		rawToken := c.Request.Header.Get("Authorization")
		token := util.ParseRawToken(rawToken)
		if len(token) == 0 {
			err := core.Error(http.StatusBadRequest, code.UnAuth, code.Text(code.UnAuth))
			c.AbortWithStatus(http.StatusBadRequest)
			c.Set("_abort_error_", err)
			return
		}

		// TODO handle error when switch to redis
		if ok, _ := tokenCache.IsInBlackList(token); ok {
			err := core.Error(http.StatusBadRequest, code.InValidToken, code.Text(code.InValidToken))
			c.AbortWithStatus(http.StatusBadRequest)
			c.Set("_abort_error_", err)
			return
		}
		claims, err := util.ParseAccessToken(token)
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
