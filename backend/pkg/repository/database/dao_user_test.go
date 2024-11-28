package database

import (
	"github.com/CloudDetail/apo/backend/config"
	"github.com/CloudDetail/apo/backend/pkg/repository/database/driver"
	"github.com/CloudDetail/apo/backend/pkg/util"
	"gorm.io/gorm"
	"os"
	"testing"
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
	conf.AnonymousUser.Role = "user"
	err = createAnonymousUser(database)
	if err != nil {
		t.Fatal(err)
	}
	var count int64
	database.Model(&User{}).Count(&count)
	var user User
	database.Where("username = ?", conf.AnonymousUser.Username).First(&user)

	validator := util.NewValidator(t, "upsert anonymous user")
	validator.CheckInt64Value("count", int64(1), count)
	validator.CheckStringValue("role", conf.AnonymousUser.Role, user.Role)
}
