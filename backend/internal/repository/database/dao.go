// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"database/sql"
	"errors"
	"time"

	"github.com/CloudDetail/apo/backend/config"
	"github.com/CloudDetail/apo/backend/internal/model/request"
	"github.com/CloudDetail/apo/backend/internal/model/response"
	"github.com/CloudDetail/apo/backend/pkg/logger"
	"github.com/CloudDetail/apo/backend/pkg/repository/database/driver"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// 定义Database查询接口
type Repo interface {
	CreateMock(model *Mock) (id uint, err error)
	GetMockById(id uint) (model *Mock, err error)
	ListMocksByCondition(req *request.ListRequest) (r []*response.ListData, count int64, err error)
	UpdateMockById(id uint, m map[string]interface{}) error
	DeleteMockById(id uint) error
}

type daoRepo struct {
	db    *gorm.DB
	sqlDB *sql.DB
}

// Connect 连接数据库
func New(zapLogger *zap.Logger) (repo Repo, err error) {
	var dbConfig gorm.Dialector
	databaseCfg := config.Get().Database
	switch databaseCfg.Connection {
	case config.DB_MYSQL:
		dbConfig = driver.NewMySqlDialector()
	case config.DB_SQLLITE:
		dbConfig = driver.NewSqlliteDialector()
	default:
		return nil, errors.New("database connection not supported")
	}

	// 连接数据库，并设置 GORM 的日志模式
	database, err := gorm.Open(dbConfig, &gorm.Config{
		Logger: logger.NewGormLogger(zapLogger),
	})
	// 处理错误
	if err != nil {
		return nil, err
	}

	// 获取底层的 sqlDB
	sqlDb, err := database.DB()
	if err != nil {
		return nil, err
	}

	// 设置最大连接数
	sqlDb.SetMaxOpenConns(databaseCfg.MaxOpen)
	// 设置最大空闲连接数
	sqlDb.SetMaxIdleConns(databaseCfg.MaxIdle)
	// 设置每个连接的过期时间
	sqlDb.SetConnMaxLifetime(time.Duration(databaseCfg.MaxLife) * time.Second)

	// 自动创建表 mock
	err = database.AutoMigrate(&Mock{})
	if err != nil {
		return nil, err
	}
	return &daoRepo{
		db:    database,
		sqlDB: sqlDb,
	}, nil
}
