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
	NamespaceMap map[string][]string `json:"namespaceMap"`
	ServiceList  []string            `json:"serviceList"`
}

type GetSubjectDataGroupResponse []database.DataGroup

type GetUserDatasourceResponse struct {
}

type GetGroupSubsResponse []database.AuthDataGroup
