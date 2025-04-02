// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package code

const (
	// Simplified Chinese
	LANG_ZH = "zh"
	// English
	LANG_EN = "en"
)

const (
	ServerError           = "A0001"
	ParamBindError        = "A0002"
	DbConnectError        = "A0003"
	UnAuth                = "A0004"
	InValidToken          = "A0005"
	UserNoPermissionError = "A0006"
	GroupNoDataError      = "A0007"

	MockCreateError = "B0101"
	MockListError   = "B0102"
	MockDetailError = "B0103"
	MockDeleteError = "B0104"

	GetServiceUrlRelationError          = "B0301"
	GetServiceUrlTopologyError          = "B0301" // TODO to be deleted
	GetDescendantMetricsError           = "B0302"
	GetDescendantRelevanceError         = "B0303"
	GetPolarisInferError                = "B0304"
	GetErrorInstanceError               = "B0306"
	GetErrorInstanceLogsError           = "B0307"
	GetLogMetricsError                  = "B0308"
	GetLogLogsError                     = "B0309"
	GetTraceMetricsError                = "B0310"
	GetTraceLogsError                   = "B0311"
	GetServiceListError                 = "B0312"
	GetServiceInstanceListError         = "B0313" // TODO to be deleted.
	GetServiceInstanceOptionsError      = "B0313"
	GetK8sEventError                    = "B0314"
	GetOverviewServiceInstanceListError = "B0315"
	GetServiceMoreUrlListError          = "B0316"
	GetThresholdError                   = "B0317"
	GetTop3UrlListError                 = "B0318"
	SetThresholdError                   = "B0319"
	GetServiceEndPointListError         = "B0320"
	GetServicesAlertError               = "B0321"
	GetSQLMetricError                   = "B0322"
	GetServiceEntryEndpointsError       = "B0323"
	GetServiceRYGLightError             = "B0324"
	GetNamespaceListError               = "B0325"
	GetServiceREDChartsError            = "B0326"

	GetFaultLogPageListError = "B0401"
	GetFaultLogContentError  = "B0402"
	QueryLogContextError     = "B0405"
	QueryLogError            = "B0406"
	GetLogChartError         = "B0407"
	GetLogIndexError         = "B0408"

	GetLogTableInfoError = "B0409"

	GetLogParseRuleError    = "B0410"
	UpdateLogParseRuleError = "B0411"
	AddLogParseRuleError    = "B0412"
	DeleteLogParseRuleError = "B0413"

	GetAllOtherLogTableError = "B0414"
	GetOtherLogTableError    = "B0415"
	AddOtherLogTableError    = "B0416"
	DeleteOtherLogTableError = "B0417"

	GetServiceRouteError = "B0418"

	GetTracePageListError    = "B0501"
	GetTraceFiltersError     = "B0502"
	GetTraceFilterValueError = "B0503"
	GetOnOffCPUError         = "B0504"
	GetSingleTraceError      = "B0505"
	GetFlameGraphError       = "B0506"

	SetTTLError            = "B0601"
	GetTTLError            = "B0602"
	SetSingleTableTTLError = "B0603"

	GetAlertEventsError = "B0701"

	// Alert Rule configure
	GetAlertEventsSampleError = "B0702"
	GetAlertRuleError         = "B0703"
	AddAlertRuleError         = "B0710"
	UpdateAlertRuleError      = "B0704"
	DeleteAlertRuleError      = "B0705"

	// AlertManager Receiver configure
	GetAMConfigReceiverError     = "B0706"
	AddAMConfigReceiverError     = "B0720"
	UpdateAMConfigReceiverError  = "B0707"
	DeleteConfigReceiverError    = "B0708"
	UpdateAlertRuleValidateError = "B0709"

	// Alert Rule Check
	AlertGroupAndLabelMismatchError = "B0711"
	AlertKeepFiringForIllegalError  = "B0712"
	AlertForIllegalError            = "B0713"
	AlertOldGroupNotExistError      = "B0714"
	AlertAlertNotExistError         = "B0715"
	AlertAlertAlreadyExistError     = "B0716"
	AlertConfigFileNotExistError    = "B0717"
	AlertTargetGroupNotExistError   = "B0718"
	AlertCheckRuleError             = "B0719"

	// AlertManagerReceiver Check
	AlertManagerReceiverAlreadyExistsError  = "B0721"
	AlertManagerReceiverNotExistsError      = "B0722"
	AlertManagerReceiverEmailHostMissing    = "B0723"
	AlertManagerReceiverEmailFromMissing    = "B0724"
	AlertManagerEmptyReceiver               = "B0725"
	AlertManagerDefaultReceiverCannotDelete = "B0726"

	// Alert Analyze
	AlertEventImpactError            = "B0727"
	AlertEventImpactMissingTag       = "B0728"
	AlertEventImpactNoMatchedService = "B0729"
	AlertEventIDMissing              = "B0730"

	AlertAnalyzeDescendantAnormalEventDeltaError = "B0731"
	GetAnomalySpanError                          = "B0732"
	MutationPQLCheckFailed                       = "B0733"
	AlertAnalyzeDescendantAnormalEventError      = "B0734"
	AlertAnalyzeDescendantAnormalContribution    = "B0735"
	DetectDefectsError                           = "B0736"
	DetectDefectsCreatAlertError                 = "B0737"
	GetDetectMutationExecListError               = "B0738"
	AddExecRecordError                           = "B0739"
	GetDetectMutationRuleListError               = "B0740"
	GetQuickMutationMetricError                  = "B0741"
	GetMetricPQLError                            = "B0742"

	GetMonitorStatusError = "B0801"

	// k8s api
	K8sGetResourceError = "B1001"

	// user api
	UserNotExistsError         = "B0901"
	UserPasswordIncorrectError = "B0902"
	UserLoginError             = "B0903"
	UserTokenExpireError       = "B0904"
	UserAlreadyExists          = "B0905"
	UserCreateError            = "B0906"
	UserUpdateError            = "B0907"
	UserConfirmPasswordError   = "B0908"
	GetUserInfoError           = "B0909"
	RemoveUserError            = "B0910"
	UserPasswordSimpleError    = "B0911"
	UserRemoveSelfError        = "B0912"
	UserEmailUsed              = "B0914"
	UserPhoneUsed              = "B0915"
	UserPhoneFormatError       = "B0916"
	UserGrantRoleError         = "B0917"
	UserGetRolesERROR          = "B0918"
	RoleNotExistsError         = "B0919"
	GetMenuConfigError         = "B0920"
	UpdateMenuConfigError      = "B0921"
	RoleGrantedError           = "B0922"
	GetFeatureError            = "B0923"
	AuthSubjectNotExistError   = "B0924"
	UserGrantPermissionError   = "B0925"
	ConfigureMenuError         = "B0926"
	PermissionNotExistError    = "B0927"
	CheckRouterError           = "B0949"

	// alertinput
	GetAlertsInputTargetTagsError     = "B1301"
	CreateAlertSourceFailed           = "B1302"
	AlertSourceAlreadyExisted         = "B1303"
	DeleteAlertSourceFailed           = "B1304"
	GetAlertSourceFailed              = "B1305"
	CreateClusterFailed               = "B1306"
	CreateSchemaFailed                = "B1307"
	ListSchemaFailed                  = "B1308"
	DeleteSchemaFailed                = "B1309"
	GetSchemaColumnsFailed            = "B1310"
	UpdateSchemaDataFailed            = "B1311"
	CheckSchemaUsedFailed             = "B1312"
	GetSchemaDataFailed               = "B1313"
	SetDefaultAlertEnrichRuleFailed   = "B1314"
	ClearDefaultAlertEnrichRuleFailed = "B1315"
	UpdateAlertEnrichRuleFailed       = "B1316"
	UpdateAlertSourceFailed           = "B1317"
	AcceptAlertEventFailed            = "B1318"
	ProcessAlertEventFailed           = "B1319"
	DeleteClusterFailed               = "B1320"
	ListAlertSourceFailed             = "B1321"
	ListClusterFailed                 = "B1322"
	GetAlertEnrichRuleFailed          = "B1323"
	AlertSourceNotExisted             = "B1324"
	GetDatasourceError                = "B0928"
	DataSourceNotExistError           = "B0929"
	CreateDataGroupError              = "B0930"
	DeleteDataGroupError              = "B0931"
	DataGroupNotExistError            = "B0932"
	GetDataGroupError                 = "B0933"
	UpdateDataGroupError              = "B0934"
	AllocateDatasourceError           = "B0935"
	GetGroupDatasourceError           = "B0936"
	DataGroupExistError               = "B0937"
	DatasourceNotExistError           = "B0938"
	AssignDataGroupError              = "B0939"
	RoleExistsError                   = "B0940"
	CreateRoleError                   = "B0941"
	UpdateRoleError                   = "B0942"
	DeleteRoleError                   = "B0943"
	APINotExist                       = "B0944"
	AuthError                         = "B0945"
	GetGroupSubsError                 = "B0946"
	UserNameError                     = "B0947"

	CreateTeamError       = "B1101"
	TeamAlreadyExistError = "B1102"
	TeamNotExistError     = "B1103"
	GetTeamError          = "B1104"
	UpdateTeamError       = "B1105"
	DeleteTeamError       = "B1106"
	AssignToTeamError     = "B1107"
	UnSupportedSubType    = "B1108"

	// integration api
	GetIntegrationInstallDocFailed        = "B1400"
	GetIntegrationInstallConfigFileFailed = "B1401"

	GetClusterIntegrationFailed = "B1402"

	GetAlertEventListError = "B1501"
)

func Text(lang string, code string) string {
	if lang == LANG_EN {
		return enText[code]
	}
	return zhCnText[code]
}

// Failure return structure
type Failure struct {
	Code    string `json:"code"`    // business code
	Message string `json:"message"` // error message
}
