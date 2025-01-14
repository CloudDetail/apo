package response

import (
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
)

type GetDatasourceResponse []model.Datasource

type GetDataGroupResponse struct {
	DataGroupList    []database.DataGroup `json:"dataGroupList"`
	model.Pagination `json:",inline"`
}

type GetGroupDatasourceResponse database.DataGroup

type GetSubjectDataGroupResponse []database.DataGroup

type GetUserDatasourceResponse []model.Datasource
