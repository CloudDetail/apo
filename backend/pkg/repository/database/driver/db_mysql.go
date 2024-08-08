package driver

import (
	"fmt"

	"github.com/CloudDetail/apo/backend/config"
	"gorm.io/gorm"

	"gorm.io/driver/mysql"
)

func NewMySqlDialector() gorm.Dialector {
	// 构建 DSN 信息
	mysqlCfg := config.Get().Database.MySql
	dsn := fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?charset=%v&parseTime=True&multiStatements=true&loc=Local",
		mysqlCfg.UserName,
		mysqlCfg.Password,
		mysqlCfg.Host,
		mysqlCfg.Port,
		mysqlCfg.Database,
		mysqlCfg.Charset,
	)
	return mysql.New(mysql.Config{
		DSN: dsn,
	})
}
