package database

import (
	"os"
	"testing"

	"github.com/CloudDetail/apo/backend/config"
	"github.com/CloudDetail/apo/backend/pkg/repository/database/driver"
	"github.com/CloudDetail/apo/backend/pkg/util"
	"gorm.io/gorm"
)

func TestUpsertAnonymous(t *testing.T) {
	os.Setenv("APO_CONFIG", "../../../config/apo.yml")
	database, err := gorm.Open(driver.NewSqlliteDialector(), &gorm.Config{})
	database.AutoMigrate(&User{})
	if err != nil {
		t.Fatal(err)
	}
	err = createAnonymousUser(database)
	if err != nil {
		t.Fatal(err)
	}
	conf := config.Get()
	conf.User.AnonymousUser.Role = "user"
	err = createAnonymousUser(database)
	if err != nil {
		t.Fatal(err)
	}
	var count int64
	database.Model(&User{}).Count(&count)
	var user User
	database.Where("username = ?", conf.User.AnonymousUser.Username).First(&user)

	validator := util.NewValidator(t, "upsert anonymous user")
	validator.CheckInt64Value("count", int64(1), count)
	validator.CheckStringValue("role", conf.User.AnonymousUser.Role, user.Role)
}
