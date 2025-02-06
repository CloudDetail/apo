package request

type CreateTeamRequest struct {
	TeamName             string                `json:"teamName" form:"teamName" binding:"required"`
	Description          string                `json:"description" form:"description"`
	FeatureList          []int                 `json:"featureList" form:"featureList"`
	DataGroupPermissions []DataGroupPermission `json:"dataGroupPermission" form:"dataGroupPermission"`
	UserList             []int64               `json:"userList" form:"userList"`
}

type UpdateTeamRequest struct {
	TeamID               int64                 `json:"teamId" form:"teamId" binding:"required"`
	TeamName             string                `json:"teamName" form:"teamName" binding:"required"`
	Description          string                `json:"description" form:"description"`
	FeatureList          []int                 `json:"featureList" form:"featureList"`
	DataGroupPermissions []DataGroupPermission `json:"dataGroupPermission" form:"dataGroupPermission"`
	UserList             []int64               `json:"userList" form:"userList"`
}

type GetTeamRequest struct {
	TeamName      string `form:"teamName"`
	FeatureList   []int  `form:"featureList"`
	DataGroupList []int  `form:"datasourceList"`
	*PageParam
}

type DeleteTeamRequest struct {
	TeamID int64 `form:"teamId" binding:"required"`
}

type TeamOperationRequest struct {
	UserID   int64   `form:"userId" binding:"required"`
	TeamList []int64 `form:"teamList"`
}

type GetUserTeamRequest struct {
	UserID int64 `form:"userId" binding:"required"`
}

type AssignToTeamRequest struct {
	TeamID   int64   `form:"teamId" binding:"required"`
	UserList []int64 `form:"userList"`
}

type GetTeamUserRequest struct {
	TeamID int64 `form:"teamId" binding:"required"`
}

type DataGroupOperationRequest struct {
	SubjectID           int64                 `json:"subjectId" binding:"required"`
	SubjectType         string                `json:"subjectType" binding:"required"`
	DataGroupPermission []DataGroupPermission `json:"dataGroupPermission"`
}

type DataGroupPermission struct {
	DataGroupID    int64  `json:"groupId" binding:"required"`
	PermissionType string `json:"type" binding:"required"`
}
