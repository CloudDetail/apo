package request

import "github.com/CloudDetail/apo/backend/pkg/model"

type CreateDataGroupRequest struct {
	GroupName        string             `json:"groupName" binding:"required"`
	Description      string             `json:"description"`
	DatasourceList   []model.Datasource `json:"datasourceList"`
	AssignedSubjects []AuthDataGroup    `json:"assignedSubjects"`
}

type AuthDataGroup struct {
	SubjectID   int64  `json:"subjectId"`
	SubjectType string `json:"subjectType"`
	Type        string `json:"type"`
}

type DeleteDataGroupRequest struct {
	GroupID int64 `form:"groupId" binding:"required"`
}

type UpdateDataGroupNameRequest struct {
	GroupID        int64              `json:"groupId" form:"groupId" binding:"required"`
	GroupName      string             `json:"groupName" form:"groupName" binding:"required"`
	Description    string             `json:"description"`
	DatasourceList []model.Datasource `json:"datasourceList"`
}

type GetDataGroupRequest struct {
	GroupName      string             `json:"groupName" form:"groupName"`
	DataSourceList []model.Datasource `json:"datasourceList"`
	*PageParam
}

type GetGroupDatasourceRequest struct {
	GroupID int64 `form:"groupId" binding:"required"`
}

type GetSubjectDataGroupRequest struct {
	SubjectID   int64  `form:"subjectId" binding:"required"`
	SubjectType string `form:"subjectType" binding:"required"`
	Category    string `form:"category"`
}
