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

// Define the Database query interface
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

// Connect to connect to the database
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

	// Connect to the database and set the log mode of GORM
	database, err := gorm.Open(dbConfig, &gorm.Config{
		Logger: logger.NewGormLogger(zapLogger),
	})
	// Handling errors
	if err != nil {
		return nil, err
	}

	// Get the underlying sqlDB
	sqlDb, err := database.DB()
	if err != nil {
		return nil, err
	}

	// Set the maximum number of connections
	sqlDb.SetMaxOpenConns(databaseCfg.MaxOpen)
	// Set the maximum number of idle connections
	sqlDb.SetMaxIdleConns(databaseCfg.MaxIdle)
	// Set the expiration time for each connection
	sqlDb.SetConnMaxLifetime(time.Duration(databaseCfg.MaxLife) * time.Second)

	// Automatically create table mock
	err = database.AutoMigrate(&Mock{})
	if err != nil {
		return nil, err
	}
	return &daoRepo{
		db:    database,
		sqlDB: sqlDb,
	}, nil
}
