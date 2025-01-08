package response

import (
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
)

type GetDatasourceResponse []model.Datasource

type GetDataGroupResponse []database.DataGroup

type GetGroupDatasourceResponse database.DataGroup

type GetSubjectDataGroupResponse []database.DataGroup
