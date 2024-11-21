package code

var enText = map[string]string{
	ServerError:    "Internal server error",
	ParamBindError: "Parameter error",
	DbConnectError: "Failed to connect Database",

	MockCreateError: "Failed to create mock",
	MockListError:   "Failed to get mock list",
	MockDetailError: "Failed to get mock detail",
	MockDeleteError: "Failed to delete mock",

	GetServiceUrlRelationError:     "Failed to get service url relation",
	GetDescendantMetricsError:      "Failed to get descendant metrics",
	GetDescendantRelevanceError:    "Failed to get descendant relevance",
	GetPolarisInferError:           "Failed to get polaris infer",
	GetErrorInstanceError:          "Failed to get error instance",
	GetErrorInstanceLogsError:      "Failed to get error instance logs",
	GetLogMetricsError:             "Failed to get log metrics",
	GetLogLogsError:                "Failed to get log logs",
	GetTraceMetricsError:           "Failed to get trace metrics",
	GetTraceLogsError:              "Failed to get trace logs",
	GetServiceListError:            "Failed to get service list",
	GetServiceInstanceOptionsError: "Failed to get service instance list",
	GetServiceEntryEndpointsError:  "Failed to get service entry endpoint list",
	GetK8sEventError:               "Failed to get k8s events",
	GetServiceEndPointListError:    "Failed to get service endpoint list",
	GetServiceRYGLightError:        "Failed to get service RYG light",
	GetSQLMetricError:              "Failed to get sql metric",
	GetFaultLogContentError:        "Failed to get fault log content",
	GetMonitorStatusError:          "Failed to get monitor status",

	QueryLogContextError: "Failed to query log context",
	QueryLogError:        "Failed to query all logs",
	GetLogChartError:     "Failed to get log chart",
	GetLogIndexError:     "Failed to get log index",

	GetLogTableInfoError:    "Failed to get log table info",
	GetLogParseRuleError:    "Failed to get log parse rule",
	UpdateLogParseRuleError: "Failed to update log parse rule",

	GetTracePageListError:    "Failed to get trace pagelist",
	GetTraceFiltersError:     "Failed to get trace filters",
	GetTraceFilterValueError: "Failed to get trace filter value",
	GetOnOffCPUError:         "Failed to get on off cpu value",
	GetSingleTraceError:      "Failed to get single trace value",
	GetFlameGraphError:       "Failed to get flame graph",

	GetOverviewServiceInstanceListError: "Failed to get overview service instance list",
	GetServiceMoreUrlListError:          "Failed to get service more url list",
	GetThresholdError:                   "Failed to get threshold",
	GetTop3UrlListError:                 "Failed to get top3 url list",
	SetThresholdError:                   "Failed to set threshold",
	GetServicesAlertError:               "Failed to get services alert",
	SetTTLError:                         "Failed to set ttl",
	GetTTLError:                         "Failed to get ttl",
	SetSingleTableTTLError:              "Failed to set single table ttl",

	GetAlertEventsError:       "Failed to get alert events",
	GetAlertEventsSampleError: "Failed to get sample alert events",

	GetAlertRuleError:    "Failed to get alert rule",
	AddAlertRuleError:    "Failed to add alert rule",
	UpdateAlertRuleError: "Failed to update alert rule",
	DeleteAlertRuleError: "Failed to delete alert rule",

	UpdateAlertRuleValidateError:    "Failed to validate alertRule, usually expr is illegle",
	AlertGroupAndLabelMismatchError: "Group and group field in label mismatch",
	AlertKeepFiringForIllegalError:  "'keepFiringFor' illegal",
	AlertForIllegalError:            "'for' illegal",
	AlertOldGroupNotExistError:      "chosen group does not exist",
	AlertAlertNotExistError:         "chosen alert does not exist",
	AlertAlertAlreadyExistError:     "alert already exist",
	AlertConfigFileNotExistError:    "config file does not exist",
	AlertTargetGroupNotExistError:   "target group does not exist",

	GetAMConfigReceiverError:    "Failed to get alertManager config receiver",
	AddAMConfigReceiverError:    "Failed to add alertManager config receiver",
	UpdateAMConfigReceiverError: "Failed to update alertManager config receiver",
	DeleteConfigReceiverError:   "Failed to delete alertManager config receiver",

	AlertManagerReceiverAlreadyExistsError:  "alertManager receiver name already exists",
	AlertManagerReceiverNotExistsError:      "alertManager receiver name not exists",
	AlertManagerReceiverEmailHostMissing:    "alertManager receiver email 'smarthost' missing",
	AlertManagerReceiverEmailFromMissing:    "alertManager receiver email 'from' missing",
	AlertManagerEmptyReceiver:               "alertManager receiver neither set webhook nor set email config",
	AlertManagerDefaultReceiverCannotDelete: "alertManager default receiver cannot be deleted",

	AlertEventImpactError:                        "Failed to get alert event impact",
	AlertEventImpactMissingTag:                   "Failed to get alert event impact, event missing tag ",
	AlertEventImpactNoMatchedService:             "Failed to get alert event impact, no matched service for event ",
	AlertEventIDMissing:                          "Failed to get alert event impact, can not find event by id within the search time range.",
	AlertAnalyzeDescendantAnormalEventDeltaError: "Failed to analyze descendant anormal event",
	GetAnomalySpanError:                          "get anomaly span failed",
	GetDetectMutationExecListError:               "Failed to get detect mutation exec list",
	GetDetectMutationRuleListError:               "Failed to get detect mutation rule list",
	GetQuickMutationMetricError:                  "Failed to get quick mutation metric",
	GetMetricPQLError:                            "Failed to get quick alert rule metrics",

	MutationPQLCheckFailed: "Failed to check mutation by PQL",

	AlertAnalyzeDescendantAnormalEventError:   "Failed to analyze descendant anormal event",
	AlertAnalyzeDescendantAnormalContribution: "Failed to analyze descendant anormal contribution",
	DetectDefectsError:                        "Failed to detect defects",
	DetectDefectsCreatAlertError:              "Failed to detect defects when create alert",
	AddExecRecordError:                        "Failed to add exec record",

	K8sGetResourceError: "Failed to get k8s resource",
}
