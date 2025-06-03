// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/amconfig"
	"github.com/CloudDetail/apo/backend/pkg/model/profile"
	"github.com/CloudDetail/apo/backend/pkg/model/request"

	"github.com/CloudDetail/apo/backend/config"
	"github.com/CloudDetail/apo/backend/pkg/logger"
	"github.com/CloudDetail/apo/backend/pkg/repository/database/driver"
	"github.com/CloudDetail/apo/backend/pkg/repository/database/integration"
	"go.uber.org/zap"
	"gorm.io/gorm"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	sc "github.com/CloudDetail/apo/backend/pkg/model/amconfig/slienceconfig"
)

// Define the Database query interface
type Repo interface {
	CreateOrUpdateThreshold(ctx core.Context, model *Threshold) error
	GetOrCreateThreshold(ctx core.Context, serviceName string, endPoint string, level string) (Threshold, error)
	DeleteThreshold(ctx core.Context, serviceName string, endPoint string) error
	OperateLogTableInfo(ctx core.Context, model *LogTableInfo, op Operator) error
	GetAllLogTable(ctx core.Context) ([]LogTableInfo, error)
	UpdateLogParseRule(ctx core.Context, model *LogTableInfo) error
	GetAllOtherLogTable(ctx core.Context) ([]OtherLogTable, error)
	OperatorOtherLogTable(ctx core.Context, model *OtherLogTable, op Operator) error
	CreateDingTalkReceiver(ctx core.Context, dingTalkConfig *amconfig.DingTalkConfig) error
	// GetDingTalkReceiver get the webhook URL secret corresponding to the uuid.
	GetDingTalkReceiver(ctx core.Context, uuid string) (amconfig.DingTalkConfig, error)
	GetDingTalkReceiverByAlertName(ctx core.Context, configFile string, alertName string, page, pageSize int) ([]*amconfig.DingTalkConfig, int64, error)
	UpdateDingTalkReceiver(ctx core.Context, dingTalkConfig *amconfig.DingTalkConfig, oldName string) error
	DeleteDingTalkReceiver(ctx core.Context, configFile, alertName string) error

	ListQuickAlertRuleMetric(ctx core.Context) ([]AlertMetricsData, error)

	Login(ctx core.Context, username, password string) (*profile.User, error)
	CreateUser(ctx core.Context, user *profile.User) error
	UpdateUserPhone(ctx core.Context, userID int64, phone string) error
	UpdateUserEmail(ctx core.Context, userID int64, email string) error
	UpdateUserPassword(ctx core.Context, userID int64, oldPassword, newPassword string) error
	UpdateUserInfo(ctx core.Context, userID int64, phone string, email string, corporation string) error
	GetUserInfo(ctx core.Context, userID int64) (profile.User, error)
	GetAnonymousUser(ctx core.Context) (profile.User, error)
	GetUserList(ctx core.Context, req *request.GetUserListRequest) ([]profile.User, int64, error)
	RemoveUser(ctx core.Context, userID int64) error
	RestPassword(ctx core.Context, userID int64, newPassword string) error
	UserExists(ctx core.Context, userID ...int64) (bool, error)

	GetUserRole(ctx core.Context, userID int64) ([]profile.UserRole, error)
	GetUsersRole(ctx core.Context, userIDs []int64) ([]profile.UserRole, error)
	GetRoles(ctx core.Context, filter model.RoleFilter) ([]profile.Role, error)
	GetFeature(ctx core.Context, featureIDs []int) ([]profile.Feature, error)
	GetFeatureByName(ctx core.Context, name string) (int, error)
	GrantRoleWithUser(ctx core.Context, userID int64, roleIDs []int) error
	GrantRoleWithRole(ctx core.Context, roleID int, userIDs []int64) error
	RevokeRole(ctx core.Context, userID int64, roleIDs []int) error
	RevokeRoleWithRole(ctx core.Context, roleID int) error
	GetSubjectPermission(ctx core.Context, subID int64, subType string, typ string) ([]int, error)
	GetSubjectsPermission(ctx core.Context, subIDs []int64, subType string, typ string) ([]AuthPermission, error)
	RoleExists(ctx core.Context, roleID int) (bool, error)
	GrantPermission(ctx core.Context, subID int64, subType string, typ string, permissionIDs []int) error
	RevokePermission(ctx core.Context, subID int64, subType string, typ string, permissionIDs []int) error
	GetAddAndDeletePermissions(ctx core.Context, subID int64, subType, typ string, permList []int) (toAdd []int, toDelete []int, err error)
	RoleGrantedToUser(ctx core.Context, userID int64, roleID int) (bool, error)
	RoleGranted(ctx core.Context, roleID int) (bool, error)
	FillItemRouter(ctx core.Context, items *[]MenuItem) error
	GetItemsRouter(ctx core.Context, itemIDs []int) ([]Router, error)
	GetRouterByIDs(ctx core.Context, routerIDs []int) ([]Router, error)
	GetRouterInsertedPage(ctx core.Context, routers []*Router, language string) error
	GetFeatureTans(ctx core.Context, features *[]profile.Feature, language string) error
	GetMenuItemTans(ctx core.Context, menuItems *[]MenuItem, language string) error

	CreateDataGroup(ctx core.Context, group *DataGroup) error
	DeleteDataGroup(ctx core.Context, groupID int64) error
	CreateDatasourceGroup(ctx core.Context, datasource []model.Datasource, dataGroupID int64) error
	DeleteDSGroup(ctx core.Context, groupID int64) error
	DataGroupExist(ctx core.Context, filter model.DataGroupFilter) (bool, error)
	UpdateDataGroup(ctx core.Context, groupID int64, groupName string, description string) error
	GetDataGroup(ctx core.Context, filter model.DataGroupFilter) ([]DataGroup, int64, error)
	RetrieveDataFromGroup(ctx core.Context, groupID int64, datasource []string) error
	GetGroupDatasource(ctx core.Context, groupID ...int64) ([]DatasourceGroup, error)

	GetFeatureMappingByFeature(ctx core.Context, featureIDs []int, mappedType string) ([]FeatureMapping, error)
	GetFeatureMappingByMapped(ctx core.Context, mappedID int, mappedType string) (FeatureMapping, error)
	GetMenuItems(ctx core.Context) ([]MenuItem, error)

	GetTeamList(ctx core.Context, req *request.GetTeamRequest) ([]profile.Team, int64, error)
	DeleteTeam(ctx core.Context, teamID int64) error
	CreateTeam(ctx core.Context, team profile.Team) error
	TeamExist(ctx core.Context, filter model.TeamFilter) (bool, error)
	GetTeam(ctx core.Context, teamID int64) (profile.Team, error)
	UpdateTeam(ctx core.Context, team profile.Team) error
	InviteUserToTeam(ctx core.Context, teamID int64, userIDs []int64) error
	AssignUserToTeam(ctx core.Context, userID int64, teamIDs []int64) error
	GetUserTeams(ctx core.Context, userID int64) ([]int64, error)
	GetTeamUsers(ctx core.Context, teamID int64) ([]int64, error)
	GetTeamUserList(ctx core.Context, teamID int64) ([]profile.User, error)
	RemoveFromTeamByUser(ctx core.Context, userID int64, teamIDs []int64) error
	RemoveFromTeamByTeam(ctx core.Context, teamID int64, userIDs []int64) error
	DeleteAllUserTeam(ctx core.Context, id int64, by string) error
	GetAssignedTeam(ctx core.Context, userID int64) ([]profile.Team, error)

	CreateRole(ctx core.Context, role *profile.Role) error
	DeleteRole(ctx core.Context, roleID int) error
	UpdateRole(ctx core.Context, roleID int, roleName, description string) error

	GetAuthDataGroupBySub(ctx core.Context, subjectID int64, subjectType string) ([]AuthDataGroup, error)
	GetGroupAuthDataGroupByGroup(ctx core.Context, groupID int64, subjectType string) ([]AuthDataGroup, error)
	AssignDataGroup(ctx core.Context, authDataGroups []AuthDataGroup) error
	RevokeDataGroupByGroup(ctx core.Context, dataGroupIDs []int64, subjectID int64) error
	RevokeDataGroupBySub(ctx core.Context, subjectIDs []int64, groupID int64) error
	GetSubjectDataGroupList(ctx core.Context, subjectID int64, subjectType string, category string) ([]DataGroup, error)
	GetModifyAndDeleteDataGroup(ctx core.Context, subjectID int64, subjectType string, dgPermissions []request.DataGroupPermission) (toModify []AuthDataGroup, toDelete []int64, err error)
	DeleteAuthDataGroup(ctx core.Context, subjectID int64, subjectType string) error
	GetDataGroupUsers(ctx core.Context, groupID int64) ([]AuthDataGroup, error)
	GetDataGroupTeams(ctx core.Context, groupID int64) ([]AuthDataGroup, error)
	CheckGroupPermission(ctx core.Context, userID, groupID int64, typ string) (bool, error)

	GetAPIByPath(ctx core.Context, path string, method string) (*API, error)

	// GetContextDB Gets transaction form ctx.
	GetContextDB(ctx core.Context) *gorm.DB
	// WithTransaction Puts transaction into ctx.
	WithTransaction(ctx core.Context, tx *gorm.DB) core.Context
	// Transaction Starts a transaction and automatically commit and rollback.
	Transaction(ctx core.Context, funcs ...func(txCtx core.Context) error) error

	GetAMConfigReceiver(ctx core.Context, filter *request.AMConfigReceiverFilter, pageParam *request.PageParam) ([]amconfig.Receiver, int, error)
	AddAMConfigReceiver(ctx core.Context, receiver amconfig.Receiver) error
	UpdateAMConfigReceiver(ctx core.Context, receiver amconfig.Receiver, oldName string) error
	DeleteAMConfigReceiver(ctx core.Context, name string) error

	CheckAMReceiverCount(ctx core.Context) int64
	MigrateAMReceiver(ctx core.Context, receivers []amconfig.Receiver) ([]amconfig.Receiver, error)

	integration.ObservabilityInputManage
}

type daoRepo struct {
	*driver.DB
	sqlDB *sql.DB

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
		DB:    &driver.DB{DB: database},
		sqlDB: sqlDb,
	}

	if err = driver.InitSQL(database, &AlertMetricsData{}); err != nil {
		return nil, err
	}
	if err = daoRepo.initRole(nil); err != nil {
		return nil, err
	}
	if err = daoRepo.initFeature(nil); err != nil {
		return nil, err
	}
	if err = daoRepo.initRouterData(nil); err != nil {
		return nil, err
	}
	if err = daoRepo.initMenuItems(nil); err != nil {
		return nil, err
	}
	if err = daoRepo.initInsertPages(nil); err != nil {
		return nil, err
	}
	if err = daoRepo.initRouterPage(nil); err != nil {
		return nil, err
	}
	if err = daoRepo.initFeatureMenuItems(nil); err != nil {
		return nil, err
	}
	if err = daoRepo.initFeatureRouter(nil); err != nil {
		return nil, err
	}
	if err = daoRepo.initPermissions(nil); err != nil {
		return nil, err
	}
	if err = daoRepo.initI18nTranslation(nil); err != nil {
		return nil, err
	}
	// TODO core.Context
	if err = daoRepo.createAdmin(nil); err != nil {
		return nil, err
	}
	if err = daoRepo.createAnonymousUser(nil); err != nil {
		return nil, err
	}

	if daoRepo.ObservabilityInputManage, err = integration.NewObservabilityInputManage(database, globalCfg); err != nil {
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
		&profile.Feature{},
		&FeatureMapping{},
		&I18nTranslation{},
		&InsertPage{},
		&LogTableInfo{},
		&MenuItem{},
		&OtherLogTable{},
		&AlertMetricsData{},
		&profile.Role{},
		&profile.UserRole{},
		&Router{},
		&Threshold{},
		&profile.User{},
		&API{},
		&AuthDataGroup{},
		&DataGroup{},
		&DatasourceGroup{},
		&profile.Team{},
		&profile.UserTeam{},
	)
}
