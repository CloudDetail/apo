// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/amconfig"
	"github.com/CloudDetail/apo/backend/pkg/model/request"

	"github.com/CloudDetail/apo/backend/config"
	"github.com/CloudDetail/apo/backend/pkg/logger"
	"github.com/CloudDetail/apo/backend/pkg/repository/database/driver"
	"github.com/CloudDetail/apo/backend/pkg/repository/database/input/alert"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Define the Database query interface
type Repo interface {
	CreateOrUpdateThreshold(model *Threshold) error
	GetOrCreateThreshold(serviceName string, endPoint string, level string) (Threshold, error)
	DeleteThreshold(serviceName string, endPoint string) error
	OperateLogTableInfo(model *LogTableInfo, op Operator) error
	GetAllLogTable() ([]LogTableInfo, error)
	UpdateLogParseRule(model *LogTableInfo) error
	GetAllOtherLogTable() ([]OtherLogTable, error)
	OperatorOtherLogTable(model *OtherLogTable, op Operator) error
	CreateDingTalkReceiver(dingTalkConfig *amconfig.DingTalkConfig) error
	// GetDingTalkReceiver get the webhook URL secret corresponding to the uuid.
	GetDingTalkReceiver(uuid string) (amconfig.DingTalkConfig, error)
	GetDingTalkReceiverByAlertName(configFile string, alertName string, page, pageSize int) ([]*amconfig.DingTalkConfig, int64, error)
	UpdateDingTalkReceiver(dingTalkConfig *amconfig.DingTalkConfig, oldName string) error
	DeleteDingTalkReceiver(configFile, alertName string) error

	ListQuickAlertRuleMetric() ([]AlertMetricsData, error)

	Login(username, password string) (*User, error)
	CreateUser(ctx context.Context, user *User) error
	UpdateUserPhone(userID int64, phone string) error
	UpdateUserEmail(userID int64, email string) error
	UpdateUserPassword(userID int64, oldPassword, newPassword string) error
	UpdateUserInfo(req *request.UpdateUserInfoRequest) error
	GetUserInfo(userID int64) (User, error)
	GetAnonymousUser() (User, error)
	GetUserList(req *request.GetUserListRequest) ([]User, int64, error)
	RemoveUser(ctx context.Context, userID int64) error
	RestPassword(userID int64, newPassword string) error
	UserExists(userID int64) (bool, error)

	GetUserRole(userID int64) ([]UserRole, error)
	GetUsersRole(userIDs []int64) ([]UserRole, error)
	GetRoles(filter model.RoleFilter) ([]Role, error)
	GetFeature(featureIDs []int) ([]Feature, error)
	GetFeatureByName(name string) (int, error)
	GrantRole(ctx context.Context, userID int64, roleIDs []int) error
	RevokeRole(ctx context.Context, userID int64, roleIDs []int) error
	GetSubjectPermission(subID int64, subType string, typ string) ([]int, error)
	GetSubjectsPermission(subIDs []int64, subType string, typ string) ([]AuthPermission, error)
	RoleExists(roleID int64) (bool, error)
	GrantPermission(ctx context.Context, subID int64, subType string, typ string, permissionIDs []int) error
	RevokePermission(ctx context.Context, subID int64, subType string, typ string, permissionIDs []int) error
	RoleGranted(userID int64, roleID int) (bool, error)
	GetItemRouter(items *[]MenuItem) error
	GetRouterInsertedPage(routers []*Router) error
	GetFeatureTans(features *[]Feature, language string) error
	GetMenuItemTans(menuItems *[]MenuItem, language string) error

	GetFeatureMapping(featureIDs []int, mappedType string) ([]FeatureMapping, error)

	GetMenuItems() ([]MenuItem, error)

	// GetContextDB Gets transaction form ctx.
	GetContextDB(ctx context.Context) *gorm.DB
	// WithTransaction Puts transaction into ctx.
	WithTransaction(ctx context.Context, tx *gorm.DB) context.Context
	// Transaction Starts a transaction and automatically commit and rollback.
	Transaction(ctx context.Context, funcs ...func(txCtx context.Context) error) error

	alert.AlertInput
}

type daoRepo struct {
	db             *gorm.DB
	sqlDB          *sql.DB
	transactionCtx struct{}

	alert.AlertInput
}

// Connect to connect to the database
func New(zapLogger *zap.Logger) (repo Repo, err error) {
	var dbConfig gorm.Dialector
	globalCfg := config.Get()
	databaseCfg := globalCfg.Database
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
	err = migrateTable(database)
	if err != nil {
		return nil, err
	}

	daoRepo := &daoRepo{
		db:    database,
		sqlDB: sqlDb,
	}

	var alertScript string
	if len(databaseCfg.InitScript.QuickAlertRuleMetric) > 0 {
		alertScript = databaseCfg.InitScript.QuickAlertRuleMetric
	} else {
		alertScript = "./sqlscripts/default_quick_alert_rule_metric.sql"
	}
	if err = daoRepo.initSql(AlertMetricsData{}, alertScript); err != nil {
		return nil, err
	}
	if err = daoRepo.initRole(); err != nil {
		return nil, err
	}
	if err = daoRepo.initFeature(); err != nil {
		return nil, err
	}
	if err = daoRepo.initRouterData(); err != nil {
		return nil, err
	}
	if err = daoRepo.initMenuItems(); err != nil {
		return nil, err
	}
	if err = daoRepo.initInsertPages(); err != nil {
		return nil, err
	}
	if err = daoRepo.initRouterPage(); err != nil {
		return nil, err
	}
	if err = daoRepo.initFeatureMenuItems(); err != nil {
		return nil, err
	}
	if err = daoRepo.initPermissions(); err != nil {
		return nil, err
	}
	if err = daoRepo.initI18nTranslation(); err != nil {
		return nil, err
	}
	if err = daoRepo.createAdmin(); err != nil {
		return nil, err
	}
	if err = daoRepo.createAnonymousUser(); err != nil {
		return nil, err
	}

	if daoRepo.AlertInput, err = alert.NewAlertInputRepo(daoRepo.db, globalCfg); err != nil {
		return nil, err
	}

	return daoRepo, nil
}

func migrateTable(db *gorm.DB) error {
	return db.AutoMigrate(
		&amconfig.DingTalkConfig{},
		&Feature{},
		&FeatureMapping{},
		&I18nTranslation{},
		&InsertPage{},
		&LogTableInfo{},
		&RouterInsertPage{},
		&MenuItem{},
		&OtherLogTable{},
		&AuthPermission{},
		&AlertMetricsData{},
		&Role{},
		&UserRole{},
		&Router{},
		&Threshold{},
		&User{},
	)
}
