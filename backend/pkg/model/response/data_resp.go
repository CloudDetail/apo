package response

import (
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
)

type GetDatasourceResponse struct {
	NamespaceList []model.Datasource `json:"namespaceList"`
	ServiceList   []model.Datasource `json:"serviceList"`
}

type GetDataGroupResponse struct {
	DataGroupList    []database.DataGroup `json:"dataGroupList"`
	model.Pagination `json:",inline"`
}

type GetGroupDatasourceResponse struct {
	NamespaceMap map[string][]string `json:"namespaceMap"` // namespace: services
	ServiceMap   map[string][]string `json:"serviceMap"`   // service: endpoints
}

type GetSubjectDataGroupResponse []database.DataGroup

type GetGroupSubsResponse []database.AuthDataGroup
