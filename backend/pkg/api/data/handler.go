package data

import (
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/repository/kubernetes"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
	"github.com/CloudDetail/apo/backend/pkg/services/data"
	"go.uber.org/zap"
)

type Handler interface {
	// GetDatasource Gets all datasource.
	// @Tags API.data
	// @Router /api/data/datasource [get]
	GetDatasource() core.HandlerFunc

	// CreateDataGroup Create a data group.
	// @Tags API.data
	// @Router /api/data/group/create [post]
	CreateDataGroup() core.HandlerFunc

	// DeleteDataGroup Delete the data group.
	// @Tags API.data
	// @Router /api/data/group/delete [post]
	DeleteDataGroup() core.HandlerFunc

	// UpdateDataGroup Updates data group's name.
	// @Tags API.data
	// @Router /api/data/group/update [post]
	UpdateDataGroup() core.HandlerFunc

	// GetDataGroup Get data group.
	// @Tags API.data
	// @Router /api/data/group [post]
	GetDataGroup() core.HandlerFunc

	// GetGroupDatasource Get group's datasource.
	// @Tags API.data
	// @Router /api/data/group/data [get]
	GetGroupDatasource() core.HandlerFunc

	// DataGroupOperation Assign data groups to users or teams, or remove them from data groups.
	// @Tags API.data
	// @Router /api/data/group/operation [post]
	DataGroupOperation() core.HandlerFunc

	// GetSubjectDataGroup Get subject's assigned data group.
	// @Tags API.data
	// @Router /api/data/sub/group [get]
	GetSubjectDataGroup() core.HandlerFunc

	// GroupSubsOperation Manage group's assigned subject.
	// @Tags API.data
	// @Router /api/data/subs/operation [post]
	GroupSubsOperation() core.HandlerFunc

	// GetGroupSubs Get group's assigned subjects.
	// @Tags API.data
	// @Router /api/data/subs [get]
	GetGroupSubs() core.HandlerFunc
}

type handler struct {
	logger      *zap.Logger
	dataService data.Service
}

func New(logger *zap.Logger, dbRepo database.Repo, promRepo prometheus.Repo, k8sRepo kubernetes.Repo) Handler {
	return &handler{
		logger:      logger,
		dataService: data.New(dbRepo, promRepo, k8sRepo),
	}
}
