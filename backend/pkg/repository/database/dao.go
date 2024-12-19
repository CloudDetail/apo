package database

import (
	"context"
	"database/sql"
	"errors"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/amconfig"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"time"

	"github.com/CloudDetail/apo/backend/config"
	"github.com/CloudDetail/apo/backend/pkg/logger"
	"github.com/CloudDetail/apo/backend/pkg/repository/database/driver"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// 定义Database查询接口
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
	// GetDingTalkReceiver 获取uuid对应的webhook url secret
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
	GrantRole(ctx context.Context, userID int64, roleIDs []int) error
	RevokeRole(ctx context.Context, userID int64, roleIDs []int) error
	GetSubjectPermission(subID int64, subType string, typ string) ([]int, error)
	GetSubjectsPermission(subIDs []int64, subType string, typ string) ([]AuthPermission, error)
	RoleExists(roleID int64) (bool, error)
	GrantPermission(ctx context.Context, subID int64, subType string, typ string, permissionIDs []int) error
	RevokePermission(ctx context.Context, subID int64, subType string, typ string, permissionIDs []int) error
	RoleGranted(userID int64, roleID int) (bool, error)
	GetItemRouter(items *[]MenuItem) error
	GetItemInsertPage(items *[]MenuItem) error

	GetMappedMenuItem(featureIDs []int) ([]FeatureMenuItem, error)

	GetMenuItems() ([]MenuItem, error)

	// GetContextDB Gets transaction form ctx.
	GetContextDB(ctx context.Context) *gorm.DB
	// WithTransaction Puts transaction into ctx.
	WithTransaction(ctx context.Context, tx *gorm.DB) context.Context
	// Transaction Starts a transaction and automatically commit and rollback.
	Transaction(ctx context.Context, funcs ...func(txCtx context.Context) error) error

	initSql(model interface{}, sqlScript string) error
}

type daoRepo struct {
	db             *gorm.DB
	sqlDB          *sql.DB
	transactionCtx struct{}
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
	// 创建阈值表
	err = database.AutoMigrate(&Threshold{})
	if err != nil {
		return nil, err
	}
	err = database.AutoMigrate(&LogTableInfo{})
	if err != nil {
		return nil, err
	}
	err = database.AutoMigrate(&OtherLogTable{})
	if err != nil {
		return nil, err
	}
	err = database.AutoMigrate(&amconfig.DingTalkConfig{})
	if err != nil {
		return nil, err
	}
	err = database.AutoMigrate(&User{})
	if err != nil {
		return nil, err
	}
	err = database.AutoMigrate(&Role{})
	if err != nil {
		return nil, err
	}
	err = database.AutoMigrate(&UserRole{})
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
	if err = daoRepo.initSql(Role{}, "./sqlscripts/default_role.sql"); err != nil {
		return nil, err
	}
	if err = daoRepo.initSql(Feature{}, "./sqlscripts/default_feature.sql"); err != nil {
		return nil, err
	}
	if err = daoRepo.initSql(FeatureMenuItem{}, "./sqlscripts/default_feature_mapping.sql"); err != nil {
		return nil, err
	}
	if err = daoRepo.initSql(AuthPermission{}, "./sqlscripts/default_role_permission.sql"); err != nil {
		return nil, err
	}
	if err = daoRepo.initSql(InsertPage{}, "./sqlscripts/default_inserted_page.sql"); err != nil {
		return nil, err
	}
	if err = daoRepo.initSql(Router{}, "./sqlscripts/default_router.sql"); err != nil {
		return nil, err
	}
	if err = daoRepo.initSql(MenuItem{}, "./sqlscripts/menu_item.sql"); err != nil {
		return nil, err
	}
	if err = daoRepo.createAdmin(); err != nil {
		return nil, err
	}
	if err = daoRepo.createAnonymousUser(); err != nil {
		return nil, err
	}
	return daoRepo, nil
}
