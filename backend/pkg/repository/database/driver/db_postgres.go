package driver

import (
	"fmt"

	"github.com/CloudDetail/apo/backend/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewPostgresDialector() gorm.Dialector {
	// Build DSN information
	postgresCfg := config.Get().Database.Postgres

	if len(postgresCfg.SSLMode) == 0 {
		postgresCfg.SSLMode = "disable"
	}
	if len(postgresCfg.Timezone) == 0 {
		postgresCfg.Timezone = "Asia/Shanghai"
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
		postgresCfg.Host,
		postgresCfg.UserName,
		postgresCfg.Password,
		postgresCfg.Database,
		postgresCfg.Port,
		postgresCfg.SSLMode,
		postgresCfg.Timezone,
	)
	return postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true,
	})
}
