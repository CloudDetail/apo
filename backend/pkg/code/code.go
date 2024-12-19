package code

import "github.com/CloudDetail/apo/backend/config"

const (
	ServerError    = "A0001"
	ParamBindError = "A0002"
	DbConnectError = "A0003"
	UnAuth         = "A0004"
	InValidToken   = "A0005"

	MockCreateError = "B0101"
	MockListError   = "B0102"
	MockDetailError = "B0103"
	MockDeleteError = "B0104"

	GetServiceUrlRelationError          = "B0301"
	GetServiceUrlTopologyError          = "B0301" // TODO 待删除
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
	GetServiceInstanceListError         = "B0313" // TODO 待删除.
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
	UserNoPermissionError      = "B0913"
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
)

func Text(code string) string {
	lang := config.Get().Language.Local

	if lang == config.LANG_EN {
		return enText[code]
	}
	return zhCnText[code]
}

// Failure 返回结构
type Failure struct {
	Code    string `json:"code"`    // 业务码
	Message string `json:"message"` // 错误信息
}
