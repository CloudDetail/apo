package data

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/repository/kubernetes"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
)

type Service interface {
	GetDataSource() (response.GetDatasourceResponse, error)
	CreateDataGroup(req *request.CreateDataGroupRequest) error
	DeleteDataGroup(req *request.DeleteDataGroupRequest) error
	GetDataGroup(req *request.GetDataGroupRequest) (response.GetDataGroupResponse, error)
	UpdateDataGroupName(req *request.UpdateDataGroupNameRequest) error
	GetGroupDatasource(req *request.GetGroupDatasourceRequest) (response.GetGroupDatasourceResponse, error)
	DataGroupOperation(req *request.DataGroupOperationRequest) error
	GetSubjectDataGroup(req *request.GetSubjectDataGroupRequest) (response.GetSubjectDataGroupResponse, error)
	// CheckDatasourcePermission Filtering data sources that users are not authorised to view. Expected *string or *[]string.
	CheckDatasourcePermission(userID int64, namespaces, services interface{}) (err error)
}

type service struct {
	dbRepo   database.Repo
	promRepo prometheus.Repo
	k8sRepo  kubernetes.Repo
}

func New(dbRepo database.Repo, promRepo prometheus.Repo, k8sRepo kubernetes.Repo) Service {
	return &service{
		dbRepo:   dbRepo,
		promRepo: promRepo,
		k8sRepo:  k8sRepo,
	}
}
