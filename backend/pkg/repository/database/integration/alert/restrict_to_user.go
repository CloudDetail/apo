package alert

import (
	"strconv"

	"github.com/CloudDetail/apo/backend/pkg/core"
	"gorm.io/gorm"
)

func RestrictToUser(userID int64) func(*gorm.DB) *gorm.DB {
	return func(d *gorm.DB) *gorm.DB {
		return d.Where("access_info = ? or access_info = ''", strconv.Itoa(int(userID)))
	}
}

func (repo *subRepo) UserByContext(ctx core.Context) *gorm.DB {
	return repo.db.Scopes(RestrictToUser(ctx.UserID()))
}

func (repo *subRepo) User(userID int64) *gorm.DB {
	return repo.db.Scopes(RestrictToUser(userID))
}

func (repo *subRepo) Admin() *gorm.DB {
	return repo.db
}
