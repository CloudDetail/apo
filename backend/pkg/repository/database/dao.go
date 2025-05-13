// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/amconfig"
	"github.com/CloudDetail/apo/backend/pkg/model/request"

	"github.com/CloudDetail/apo/backend/config"
	"github.com/CloudDetail/apo/backend/pkg/logger"
	"github.com/CloudDetail/apo/backend/pkg/repository/database/driver"
	"github.com/CloudDetail/apo/backend/pkg/repository/database/integration"
	"go.uber.org/zap"
	"gorm.io/gorm"

	sc "github.com/CloudDetail/apo/backend/pkg/model/amconfig/slienceconfig"
	core "github.com/CloudDetail/apo/backend/pkg/core"
)

// Define the Database query interface
type Repo interface {
	CreateOrUpdateThreshold(ctx_core core.Context, model *Threshold) error
	GetOrCreateThreshold(ctx_core core.Context, serviceName string, endPoint string, level string) (Threshold, error)
	DeleteThreshold(ctx_core core.Context, serviceName string, endPoint string) error
	OperateLogTableInfo(ctx_core core.Context, model *LogTableInfo, op Operator) error
	GetAllLogTable(ctx_core core.Context,) ([]LogTableInfo, error)
	UpdateLogParseRule(ctx_core core.Context, model *LogTableInfo) error
	GetAllOtherLogTable(ctx_core core.Context,) ([]OtherLogTable, error)
	OperatorOtherLogTable(ctx_core core.Context, model *OtherLogTable, op Operator) error
	CreateDingTalkReceiver(ctx_core core.Context, dingTalkConfig *amconfig.DingTalkConfig) error
	// GetDingTalkReceiver get the webhook URL secret corresponding to the uuid.
	GetDingTalkReceiver(ctx_core core.Context, uuid string) (amconfig.DingTalkConfig, error)
	GetDingTalkReceiverByAlertName(ctx_core core.Context, configFile string, alertName string, page, pageSize int) ([]*amconfig.DingTalkConfig, int64, error)
	UpdateDingTalkReceiver(ctx_core core.Context, dingTalkConfig *amconfig.DingTalkConfig, oldName string) error
	DeleteDingTalkReceiver(ctx_core core.Context, configFile, alertName string) error

	ListQuickAlertRuleMetric(ctx_core core.Context, lang string) ([]AlertMetricsData, error)

	Login(ctx_core core.Context, username, password string) (*User, error)
	CreateUser(ctx_core core.Context, ctx context.Context, user *User) error
	UpdateUserPhone(ctx_core core.Context, userID int64, phone string) error
	UpdateUserEmail(ctx_core core.Context, userID int64, email string) error
	UpdateUserPassword(ctx_core core.Context, userID int64, oldPassword, newPassword string) error
	UpdateUserInfo(ctx_core core.Context, ctx context.Context, userID int64, phone string, email string, corporation string) error
	GetUserInfo(ctx_core core.Context, userID int64) (User, error)
	GetAnonymousUser(ctx_core core.Context,) (User, error)
	GetUserList(ctx_core core.Context, req *request.GetUserListRequest) ([]User, int64, error)
	RemoveUser(ctx_core core.Context, ctx context.Context, userID int64) error
	RestPassword(ctx_core core.Context, userID int64, newPassword string) error
	UserExists(ctx_core core.Context, userID ...int64) (bool, error)

	GetUserRole(ctx_core core.Context, userID int64) ([]UserRole, error)
	GetUsersRole(ctx_core core.Context, userIDs []int64) ([]UserRole, error)
	GetRoles(ctx_core core.Context, filter model.RoleFilter) ([]Role, error)
	GetFeature(ctx_core core.Context, featureIDs []int) ([]Feature, error)
	GetFeatureByName(ctx_core core.Context, name string) (int, error)
	GrantRoleWithUser(ctx_core core.Context, ctx context.Context, userID int64, roleIDs []int) error
	GrantRoleWithRole(ctx_core core.Context, ctx context.Context, roleID int, userIDs []int64) error
	RevokeRole(ctx_core core.Context, ctx context.Context, userID int64, roleIDs []int) error
	RevokeRoleWithRole(ctx_core core.Context, ctx context.Context, roleID int) error
	GetSubjectPermission(ctx_core core.Context, subID int64, subType string, typ string) ([]int, error)
	GetSubjectsPermission(ctx_core core.Context, subIDs []int64, subType string, typ string) ([]AuthPermission, error)
	RoleExists(ctx_core core.Context, roleID int) (bool, error)
	GrantPermission(ctx_core core.Context, ctx context.Context, subID int64, subType string, typ string, permissionIDs []int) error
	RevokePermission(ctx_core core.Context, ctx context.Context, subID int64, subType string, typ string, permissionIDs []int) error
	GetAddAndDeletePermissions(ctx_core core.Context, subID int64, subType, typ string, permList []int) (toAdd []int, toDelete []int, err error)
	RoleGrantedToUser(ctx_core core.Context, userID int64, roleID int) (bool, error)
	RoleGranted(ctx_core core.Context, roleID int) (bool, error)
	FillItemRouter(ctx_core core.Context, items *[]MenuItem) error
	GetItemsRouter(ctx_core core.Context, itemIDs []int) ([]Router, error)
	GetRouterByIDs(ctx_core core.Context, routerIDs []int) ([]Router, error)
	GetRouterInsertedPage(ctx_core core.Context, routers []*Router, language string) error
	GetFeatureTans(ctx_core core.Context, features *[]Feature, language string) error
	GetMenuItemTans(ctx_core core.Context, menuItems *[]MenuItem, language string) error

	CreateDataGroup(ctx_core core.Context, ctx context.Context, group *DataGroup) error
	DeleteDataGroup(ctx_core core.Context, ctx context.Context, groupID int64) error
	CreateDatasourceGroup(ctx_core core.Context, ctx context.Context, datasource []model.Datasource, dataGroupID int64) error
	DeleteDSGroup(ctx_core core.Context, ctx context.Context, groupID int64) error
	DataGroupExist(ctx_core core.Context, filter model.DataGroupFilter) (bool, error)
	UpdateDataGroup(ctx_core core.Context, ctx context.Context, groupID int64, groupName string, description string) error
	GetDataGroup(ctx_core core.Context, filter model.DataGroupFilter) ([]DataGroup, int64, error)
	RetrieveDataFromGroup(ctx_core core.Context, ctx context.Context, groupID int64, datasource []string) error
	GetGroupDatasource(ctx_core core.Context, groupID ...int64) ([]DatasourceGroup, error)

	GetFeatureMappingByFeature(ctx_core core.Context, featureIDs []int, mappedType string) ([]FeatureMapping, error)
	GetFeatureMappingByMapped(ctx_core core.Context, mappedID int, mappedType string) (FeatureMapping, error)
	GetMenuItems(ctx_core core.Context,) ([]MenuItem, error)

	GetTeamList(ctx_core core.Context, req *request.GetTeamRequest) ([]Team, int64, error)
	DeleteTeam(ctx_core core.Context, ctx context.Context, teamID int64) error
	CreateTeam(ctx_core core.Context, ctx context.Context, team Team) error
	TeamExist(ctx_core core.Context, filter model.TeamFilter) (bool, error)
	GetTeam(ctx_core core.Context, teamID int64) (Team, error)
	UpdateTeam(ctx_core core.Context, ctx context.Context, team Team) error
	InviteUserToTeam(ctx_core core.Context, ctx context.Context, teamID int64, userIDs []int64) error
	AssignUserToTeam(ctx_core core.Context, ctx context.Context, userID int64, teamIDs []int64) error
	GetUserTeams(ctx_core core.Context, userID int64) ([]int64, error)
	GetTeamUsers(ctx_core core.Context, teamID int64) ([]int64, error)
	GetTeamUserList(ctx_core core.Context, teamID int64) ([]User, error)
	RemoveFromTeamByUser(ctx_core core.Context, ctx context.Context, userID int64, teamIDs []int64) error
	RemoveFromTeamByTeam(ctx_core core.Context, ctx context.Context, teamID int64, userIDs []int64) error
	DeleteAllUserTeam(ctx_core core.Context, ctx context.Context, id int64, by string) error
	GetAssignedTeam(ctx_core core.Context, userID int64) ([]Team, error)

	CreateRole(ctx_core core.Context, ctx context.Context, role *Role) error
	DeleteRole(ctx_core core.Context, ctx context.Context, roleID int) error
	UpdateRole(ctx_core core.Context, ctx context.Context, roleID int, roleName, description string) error

	GetAuthDataGroupBySub(ctx_core core.Context, subjectID int64, subjectType string) ([]AuthDataGroup, error)
	GetGroupAuthDataGroupByGroup(ctx_core core.Context, groupID int64, subjectType string) ([]AuthDataGroup, error)
	AssignDataGroup(ctx_core core.Context, ctx context.Context, authDataGroups []AuthDataGroup) error
	RevokeDataGroupByGroup(ctx_core core.Context, ctx context.Context, dataGroupIDs []int64, subjectID int64) error
	RevokeDataGroupBySub(ctx_core core.Context, ctx context.Context, subjectIDs []int64, groupID int64) error
	GetSubjectDataGroupList(ctx_core core.Context, subjectID int64, subjectType string, category string) ([]DataGroup, error)
	GetModifyAndDeleteDataGroup(ctx_core core.Context, subjectID int64, subjectType string, dgPermissions []request.DataGroupPermission) (toModify []AuthDataGroup, toDelete []int64, err error)
	DeleteAuthDataGroup(ctx_core core.Context, ctx context.Context, subjectID int64, subjectType string) error
	GetDataGroupUsers(ctx_core core.Context, groupID int64) ([]AuthDataGroup, error)
	GetDataGroupTeams(ctx_core core.Context, groupID int64) ([]AuthDataGroup, error)
	CheckGroupPermission(ctx_core core.Context, userID, groupID int64, typ string) (bool, error)

	GetAPIByPath(ctx_core core.Context, path string, method string) (*API, error)

	// GetContextDB Gets transaction form ctx.
	GetContextDB(ctx_core core.Context, ctx context.Context) *gorm.DB
	// WithTransaction Puts transaction into ctx.
	WithTransaction(ctx_core core.Context, ctx context.Context, tx *gorm.DB) context.Context
	// Transaction Starts a transaction and automatically commit and rollback.
	Transaction(ctx_core core.Context, ctx context.Context, funcs ...func(txCtx context.Context) error) error

	GetAMConfigReceiver(ctx_core core.Context, filter *request.AMConfigReceiverFilter, pageParam *request.PageParam) ([]amconfig.Receiver, int, error)
	AddAMConfigReceiver(ctx_core core.Context, receiver amconfig.Receiver) error
	UpdateAMConfigReceiver(ctx_core core.Context, receiver amconfig.Receiver, oldName string) error
	DeleteAMConfigReceiver(ctx_core core.Context, name string) error

	CheckAMReceiverCount(ctx_core core.Context,) int64
	MigrateAMReceiver(ctx_core core.Context, receivers []amconfig.Receiver) ([]amconfig.Receiver, error)

	integration.ObservabilityInputManage
}

type daoRepo struct {
	db		*gorm.DB
	sqlDB		*sql.DB
	transactionCtx	struct{}

	integration.ObservabilityInputManage
}

// Connect to connect to the database
func New(zapLogger *zap.Logger) (repo Repo, err error) {
	var dbConfig gorm.Dialector
	globalCfg := config.Get()
	databaseCfg := globalCfg.Database
	switch databaseCfg.Connection {
	case config.DB_MYSQL:
		dbConfig, err = driver.NewMySqlDialector()
	case config.DB_SQLLITE:
		dbConfig = driver.NewSqlliteDialector()
	case config.DB_POSTGRES:
		dbConfig, err = driver.NewPostgresDialector()
	default:
		return nil, errors.New("database connection not supported")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database, err: %v", err)
	}

	// Connect to the database and set the log mode of GORM
	database, err := gorm.Open(dbConfig, &gorm.Config{
		Logger: logger.NewGormLogger(zapLogger),
	})

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
		db:	database,
		sqlDB:	sqlDb,
	}

	if err = driver.InitSQL(daoRepo.db, &AlertMetricsData{}); err != nil {
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
	if err = daoRepo.initFeatureRouter(); err != nil {
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

	if daoRepo.ObservabilityInputManage, err = integration.NewObservabilityInputManage(daoRepo.db, globalCfg); err != nil {
		return nil, err
	}

	return daoRepo, nil
}

func migrateTable(db *gorm.DB) error {
	err := db.AutoMigrate(&AuthPermission{})
	if err != nil {
		return err
	}
	err = db.AutoMigrate(&RouterInsertPage{})
	if err != nil {
		return err
	}
	migrator := db.Migrator()
	if migrator.HasIndex(&RouterInsertPage{}, "idx_router_insert_page_router_id") {
		err := migrator.DropIndex(&RouterInsertPage{}, "idx_router_insert_page_router_id")
		if err != nil {
			return err
		}
	}

	return db.AutoMigrate(
		&amconfig.Receiver{},
		&sc.AlertSlienceConfig{},
		&amconfig.DingTalkConfig{},
		&Feature{},
		&FeatureMapping{},
		&I18nTranslation{},
		&InsertPage{},
		&LogTableInfo{},
		&MenuItem{},
		&OtherLogTable{},
		&AlertMetricsData{},
		&Role{},
		&UserRole{},
		&Router{},
		&Threshold{},
		&User{},
		&API{},
		&AuthDataGroup{},
		&DataGroup{},
		&DatasourceGroup{},
		&Team{},
		&UserTeam{},
	)
}
