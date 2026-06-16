// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package router

import (
	appconfig "github.com/CloudDetail/apo/backend/config"
	"github.com/CloudDetail/apo/backend/pkg/api/alertanalyze"
	alertinput "github.com/CloudDetail/apo/backend/pkg/api/alertinput"
	"github.com/CloudDetail/apo/backend/pkg/api/alerts"
	"github.com/CloudDetail/apo/backend/pkg/api/app"
	"github.com/CloudDetail/apo/backend/pkg/api/config"
	"github.com/CloudDetail/apo/backend/pkg/api/custom"
	"github.com/CloudDetail/apo/backend/pkg/api/data"
	"github.com/CloudDetail/apo/backend/pkg/api/dataplane"
	"github.com/CloudDetail/apo/backend/pkg/api/health"
	"github.com/CloudDetail/apo/backend/pkg/api/integration"
	"github.com/CloudDetail/apo/backend/pkg/api/k8s"
	"github.com/CloudDetail/apo/backend/pkg/api/license"
	"github.com/CloudDetail/apo/backend/pkg/api/log"
	"github.com/CloudDetail/apo/backend/pkg/api/metric"
	networkapi "github.com/CloudDetail/apo/backend/pkg/api/network"
	"github.com/CloudDetail/apo/backend/pkg/api/permission"
	"github.com/CloudDetail/apo/backend/pkg/api/role"
	"github.com/CloudDetail/apo/backend/pkg/api/service"
	"github.com/CloudDetail/apo/backend/pkg/api/serviceoverview"
	"github.com/CloudDetail/apo/backend/pkg/api/systemhealth"
	"github.com/CloudDetail/apo/backend/pkg/api/team"
	"github.com/CloudDetail/apo/backend/pkg/api/trace"
	"github.com/CloudDetail/apo/backend/pkg/api/user"
	"github.com/CloudDetail/apo/backend/pkg/middleware"
)

func setApiRouter(r *resource) {
	middlewares := middleware.New(r.cache, r.pkg_db, r.dify, r.prom, r.logger)

	serviceApi := r.mux.Group("/api/service").Use(middlewares.AuthMiddleware()).Use(middlewares.LicMiddleware())
	{
		serviceOverviewHandler := serviceoverview.New(r.logger, r.ch, r.prom, r.pkg_db, r.k8sApi)
		serviceApi.Any("/endpoints", serviceOverviewHandler.GetEndPointsData())
		serviceApi.Any("/servicesAlert", serviceOverviewHandler.GetServicesAlert())
		serviceApi.Any("/moreUrl", serviceOverviewHandler.GetServiceMoreUrlList())
		serviceApi.GET("/getThreshold", serviceOverviewHandler.GetThreshold())
		serviceApi.POST("/setThreshold", serviceOverviewHandler.SetThreshold())
		serviceApi.GET("/ryglight", serviceOverviewHandler.GetRYGLight())
		serviceApi.GET("/monitor/status", serviceOverviewHandler.GetMonitorStatus())

		serviceHandler := service.New(r.logger, r.ch, r.prom, r.pol, r.pkg_db, r.k8sApi)
		serviceApi.Any("/entry/endpoints", serviceHandler.GetServiceEntryEndpoints())
		serviceApi.Any("/relation", serviceHandler.GetServiceEndpointRelation())
		serviceApi.Any("/topology", serviceHandler.GetServiceEndpointTopology())
		serviceApi.Any("/descendant/metrics", serviceHandler.GetDescendantMetrics())
		serviceApi.Any("/descendant/relevance", serviceHandler.GetDescendantRelevance())
		serviceApi.Any("/polaris/infer", serviceHandler.GetPolarisInfer())
		serviceApi.Any("/error/instance", serviceHandler.GetErrorInstance())
		serviceApi.Any("/errorinstance/logs", serviceHandler.GetErrorInstanceLogs())
		serviceApi.Any("/log/metrics", serviceHandler.GetLogMetrics())
		serviceApi.Any("/log/logs", serviceHandler.GetLogLogs())
		serviceApi.Any("/trace/metrics", serviceHandler.GetTraceMetrics())
		serviceApi.Any("/trace/logs", serviceHandler.GetTraceLogs())

		serviceApi.Any("/list", serviceHandler.GetServiceList())
		serviceApi.Any("/instances", serviceHandler.GetServiceInstance())
		serviceApi.Any("/instance/list", serviceHandler.GetServiceInstanceList())
		serviceApi.Any("/instanceinfo/list", serviceHandler.GetServiceInstanceInfoList())
		serviceApi.Any("/instance/options", serviceHandler.GetServiceInstanceOptions())
		serviceApi.Any("/endpoint/list", serviceHandler.GetServiceEndPointList())
		serviceApi.Any("/k8s/events/count", serviceHandler.CountK8sEvents())
		serviceApi.Any("/namespace/list", serviceHandler.GetNamespaceList())

		serviceApi.Any("/alert/events", serviceHandler.GetAlertEvents())
		serviceApi.Any("/alert/sample/events", serviceHandler.GetAlertEventsSample())

		serviceApi.Any("/sql/metrics", serviceHandler.GetSQLMetrics())
		serviceApi.POST("/redcharts", serviceHandler.GetServiceREDCharts())
	}

	systemHealthApi := r.mux.Group("/api/system-health").Use(middlewares.AuthMiddleware()).Use(middlewares.LicMiddleware())
	{
		systemHealthHandler := systemhealth.New(r.ch, r.pkg_db, r.dify, r.k8sApi, r.prom, appconfig.Get().AlertManager.Address, appconfig.Get().VMAlert.Address)
		systemHealthApi.GET("/report", systemHealthHandler.GetReport())
	}

	logApi := r.mux.Group("/api/log").Use(middlewares.AuthMiddleware()).Use(middlewares.LicMiddleware())
	{
		logHandler := log.New(r.logger, r.ch, r.pkg_db, r.k8sApi, r.prom)
		logApi.POST("/fault/pagelist", logHandler.GetFaultLogPageList())
		logApi.POST("/fault/content", logHandler.GetFaultLogContent())

		logApi.POST("/context", logHandler.QueryLogContext())

		logApi.POST("/query", logHandler.QueryLog())
		logApi.POST("/chart", logHandler.GetLogChart())
		logApi.POST("/index", logHandler.GetLogIndex())

		logApi.GET("/table", logHandler.GetLogTableInfo())

		logApi.GET("/rule/service", logHandler.GetServiceRoute())

		logApi.GET("/rule/get", logHandler.GetLogParseRule())
		logApi.POST("/rule/update", logHandler.UpdateLogParseRule())
		logApi.POST("/rule/add", logHandler.AddLogParseRule())
		logApi.DELETE("/rule/delete", logHandler.DeleteLogParseRule())

		logApi.GET("/other", logHandler.OtherTable())
		logApi.GET("/other/table", logHandler.OtherTableInfo())
		logApi.POST("/other/add", logHandler.AddOtherTable())
		logApi.DELETE("/other/delete", logHandler.DeleteOtherTable())
	}

	traceApi := r.mux.Group("/api/trace").Use(middlewares.LicMiddleware())
	{
		traceHandler := trace.New(r.logger, r.pkg_db, r.ch, r.jaegerRepo, r.prom, r.k8sApi)
		// These APIs are used by another project. DO NOT Auth.
		traceApi.GET("/onoffcpu", traceHandler.GetOnOffCPU())
		traceApi.GET("/flame", traceHandler.GetFlameGraphData())
		traceApi.GET("/flame/process", traceHandler.GetProcessFlameGraph())

		traceApi.Use(middlewares.AuthMiddleware())
		traceApi.POST("/pagelist", traceHandler.GetTracePageList())
		traceApi.GET("/pagelist/filters", traceHandler.GetTraceFilters())
		traceApi.POST("/pagelist/filter/value", traceHandler.GetTraceFilterValue())
		traceApi.GET("/info", traceHandler.GetSingleTraceInfo())
	}

	alertApi := r.mux.Group("/api/alerts").Use(middlewares.LicMiddleware())
	{
		alertHandler := alerts.New(r.logger, r.ch, r.pkg_db, r.k8sApi, r.prom, r.dify, r.receivers)
		alertApi.POST("/event/list", alertHandler.AlertEventList())
		alertApi.POST("/event/detail", alertHandler.AlertEventDetail())
		alertApi.GET("/events/classify", alertHandler.AlertEventClassify())
		alertApi.POST("/events/report/get", alertHandler.GetAlertEventReport())
		alertApi.POST("/events/report/add", alertHandler.AddAlertEventReport())
		alertApi.POST("/inputs/alertmanager", alertHandler.InputAlertManager())
		alertApi.POST("/outputs/dingtalk/:uuid", alertHandler.ForwardToDingTalk())

		alertApi.Use(middlewares.AuthMiddleware())
		alertApi.GET("/rules/file", alertHandler.GetAlertRuleFile())
		alertApi.POST("/rules/file", alertHandler.UpdateAlertRuleFile())

		alertApi.GET("/rule/groups", alertHandler.GetGroupList())
		alertApi.GET("/rule/metrics", alertHandler.GetMetricPQL())

		alertApi.POST("/rule/list", alertHandler.GetAlertRules())
		alertApi.POST("/rule", alertHandler.UpdateAlertRule())
		alertApi.DELETE("/rule", alertHandler.DeleteAlertRule())
		alertApi.POST("/rule/add", alertHandler.AddAlertRule())
		alertApi.GET("/rule/available", alertHandler.CheckAlertRule())

		alertApi.POST("/alertmanager/receiver/list", alertHandler.GetAlertManagerConfigReceiver())
		alertApi.POST("/alertmanager/receiver/add", alertHandler.AddAlertManagerConfigReceiver())
		alertApi.POST("/alertmanager/receiver", alertHandler.UpdateAlertManagerConfigReceiver())
		alertApi.DELETE("/alertmanager/receiver", alertHandler.DeleteAlertManagerConfigReceiver())

		alertanalyzeHandler := alertanalyze.New(r.logger, r.ch, r.pkg_db, r.prom, r.pol, r.k8sApi)
		alertApi.POST("/alert/events", alertanalyzeHandler.GetAlertEvents())
		alertApi.POST("/event/impact", alertanalyzeHandler.GetAlertImpact())
		alertApi.Any("/descendant/anormal/contribution", alertanalyzeHandler.GetDescendantContribution())
		alertApi.Any("/descendant/anormal/list", alertanalyzeHandler.GetDescendantAnormal())
		alertApi.Any("/descendant/anormal/trend", alertanalyzeHandler.GetAnormalTrendByEntry())
		alertApi.POST("/descendant/anormal/delta", alertanalyzeHandler.GetDescendantAnormalDelta())
		alertApi.POST("/anomaly-span/list", alertanalyzeHandler.GetAnomalySpan())
		alertApi.GET("/detect/mutation/metrics", alertanalyzeHandler.GetPredefinedDetectExpr())
		alertApi.POST("/detect/mutation/add", alertanalyzeHandler.AddDefectsDetect())
		alertApi.GET("/detect/mutation/exec-list", alertanalyzeHandler.GetDefectDetectExecList())
		alertApi.GET("/detect/mutation/rule-list", alertanalyzeHandler.GetDefectDetectRuleList())
		alertApi.GET("/slient", alertHandler.GetAlertSlienceConfig())
		alertApi.GET("/slient/list", alertHandler.ListAlertSlienceConfig())
		alertApi.DELETE("/slient", alertHandler.RemoveAlertSlienceConfig())
		alertApi.POST("/slient", alertHandler.SetAlertSlienceConfig())

		alertApi.POST("/resolve", alertHandler.MarkAlertResolvedManually())

		alertApi.GET("/filter/keys", alertHandler.GetAlertEventStaticFilters())
		alertApi.POST("/filter/labelkeys", alertHandler.GetAlertEventLabelFilterKeys())
		alertApi.POST("/filter/values", alertHandler.SearchAlertEventFilterValues())
	}

	configApi := r.mux.Group("/api/config").Use(middlewares.AuthMiddleware()).Use(middlewares.LicMiddleware())
	{
		configHandler := config.New(r.logger, r.ch)
		configApi.POST("/setTTL", configHandler.SetTTL())
		configApi.POST("/setSingleTableTTL", configHandler.SetSingleTableTTL())
		configApi.GET("/getTTL", configHandler.GetTTL())
	}

	userApi := r.mux.Group("/api/user")
	{
		userHandler := user.New(r.logger, r.pkg_db, r.cache, r.dify)
		userApi.POST("/login", userHandler.Login())
		userApi.POST("/logout", userHandler.Logout())
		userApi.GET("/refresh", userHandler.RefreshToken())
		userApi.Use(middlewares.AuthMiddleware())
		userApi.GET("/info", userHandler.GetUserInfo())
		userApi.GET("/list", userHandler.GetUserList())
		userApi.GET("/team", userHandler.GetUserTeam())
		userApi.Use(middlewares.LicMiddleware())
		userApi.POST("/create", userHandler.CreateUser())
		userApi.POST("/update/password", userHandler.UpdateUserPassword())
		userApi.POST("/update/phone", userHandler.UpdateUserPhone())
		userApi.POST("/update/email", userHandler.UpdateUserEmail())
		userApi.POST("/update/info", userHandler.UpdateUserInfo())
		userApi.POST("/update/self", userHandler.UpdateSelfInfo())
		userApi.POST("/remove", userHandler.RemoveUser())
		userApi.POST("/reset", userHandler.ResetPassword())
	}

	permissionApi := r.mux.Group("/api/permission").Use(middlewares.AuthMiddleware())
	{
		permissionHandler := permission.New(r.logger, r.pkg_db)
		permissionApi.GET("/config", permissionHandler.GetUserConfig())
		permissionApi.GET("/feature", permissionHandler.GetFeature())
		permissionApi.GET("/router", permissionHandler.CheckRouterPermission())
		permissionApi.GET("/sub/feature", permissionHandler.GetSubjectFeature())
		permissionApi.Use(middlewares.LicMiddleware())
		permissionApi.POST("/operation", permissionHandler.PermissionOperation())
		permissionApi.POST("/menu/configure", permissionHandler.ConfigureMenu())
	}

	roleApi := r.mux.Group("/api/role").Use(middlewares.AuthMiddleware())
	{
		roleHandler := role.New(r.logger, r.pkg_db)
		roleApi.GET("/roles", roleHandler.GetRole())
		roleApi.GET("/user", roleHandler.GetUserRole())
		roleApi.Use(middlewares.LicMiddleware())
		roleApi.POST("/operation", roleHandler.RoleOperation())
		roleApi.POST("/create", roleHandler.CreateRole())
		roleApi.POST("/update", roleHandler.UpdateRole())
		roleApi.POST("/delete", roleHandler.DeleteRole())
	}

	dataApi := r.mux.Group("/api/data").Use(middlewares.AuthMiddleware()).Use(middlewares.LicMiddleware())
	dataApiV2 := r.mux.Group("/api/v2/data").Use(middlewares.AuthMiddleware()).Use(middlewares.LicMiddleware())
	{
		dataHandler := data.New(r.logger, r.pkg_db, r.prom, r.ch, r.k8sApi)
		dataApiV2.GET("/group", dataHandler.GetDataGroupV2())
		dataApiV2.GET("/group/datasource/list", dataHandler.GetDGScopeList())
		dataApiV2.GET("/group/detail", dataHandler.GetDGDetailV2())
		dataApiV2.POST("/group/add", dataHandler.CreateDataGroupV2())
		dataApiV2.POST("/group/update", dataHandler.UpdateDataGroupV2())
		dataApiV2.DELETE("/group/delete", dataHandler.DeleteDataGroupV2())
		dataApiV2.POST("/group/filter", dataHandler.GetFilterByGroupIDV2())
		dataApiV2.Any("/group/datasource/refresh", dataHandler.CleanExpiredDataScope())

		dataApi.GET("/datasource", dataHandler.GetDatasource())
		dataApi.POST("/group", dataHandler.GetDataGroup())
		// 旧版本使用GET,新版本使用POST
		dataApi.Any("/group/data", dataHandler.GetGroupDatasource())
		//dataApi.POST("/group/update", dataHandler.UpdateDataGroup())
		//dataApi.POST("/group/create", dataHandler.CreateDataGroup())
		dataApi.GET("/sub/group", dataHandler.GetSubjectDataGroup())
		dataApi.GET("/user/group", dataHandler.GetUserDataGroup())
		dataApi.POST("/group/operation", dataHandler.DataGroupOperation())
		dataApi.GET("/subs", dataHandler.GetGroupSubs())
		dataApi.POST("/subs/operation", dataHandler.GroupSubsOperation())
	}

	teamApi := r.mux.Group("/api/team").Use(middlewares.AuthMiddleware())
	{
		teamHandler := team.New(r.logger, r.pkg_db)
		teamApi.GET("", teamHandler.GetTeam())
		teamApi.GET("/user", teamHandler.GetTeamUser())
		teamApi.Use(middlewares.LicMiddleware())
		teamApi.POST("/create", teamHandler.CreateTeam())
		teamApi.POST("/update", teamHandler.UpdateTeam())
		teamApi.POST("/delete", teamHandler.DeleteTeam())
		teamApi.POST("/operation", teamHandler.TeamOperation())
		teamApi.POST("/user/operation", teamHandler.TeamUserOperation())
	}

	k8sApi := r.mux.Group("/api/k8s").Use(middlewares.LicMiddleware())
	{
		k8sHandler := k8s.New(r.k8sApi)
		// These APIs are used by another project. DO NOT Auth.
		k8sApi.GET("/namespaces", k8sHandler.GetNamespaceList())
		k8sApi.GET("/namespace/info", k8sHandler.GetNamespaceInfo())
		k8sApi.GET("/pods", k8sHandler.GetPodList())
		k8sApi.GET("/pod/info", k8sHandler.GetPodInfo())
	}
	networkApi := r.mux.Group("/api/network/").Use(middlewares.LicMiddleware())
	{
		handler := networkapi.New(r.logger, r.pkg_db, r.deepflowClickhouse, r.prom, r.k8sApi)
		networkApi.GET("/podmap", handler.GetPodMap())
		networkApi.GET("/segments", handler.GetSpanSegmentsMetrics())
	}

	healthApi := r.mux.Group("/api/health")
	{
		handler := health.New()
		healthApi.GET("", handler.HealthCheck())
	}
	customApi := r.mux.Group("/api/custom").Use(middlewares.LicMiddleware())
	{
		handler := custom.New(r.logger, r.pkg_db)
		customApi.POST("/create", handler.CreateCustomQuery())
		customApi.POST("/update", handler.UpdateCustomQuery())
		customApi.GET("/get", handler.GetCustomQuery())
		customApi.DELETE("/delete", handler.DeleteCustomQuery())
	}

	alertInputApi := r.mux.Group("/api/alertinput").Use(middlewares.LicMiddleware())
	{
		handler := alertinput.New(r.logger, r.ch, r.prom, r.pkg_db, r.dify)
		alertInputApi.POST("/event/source", handler.SourceHandler())
		alertInputApi.POST("/event/json", handler.JsonHandler())
		alertInputApi.POST("/source/create", handler.CreateAlertSource())
		alertInputApi.GET("/source/paramspec", handler.GetAlertProviderParamsSpec())
		alertInputApi.POST("/source/setup/webhook", handler.SetupAlertProviderWebhook())
		alertInputApi.POST("/source/update", handler.UpdateAlertSource())
		alertInputApi.POST("/source/get", handler.GetAlertSource())
		alertInputApi.POST("/source/delete", handler.DeleteAlertSource())
		alertInputApi.GET("/source/list", handler.ListAlertSource())
		alertInputApi.POST("/source/enrich/update", handler.UpdateAlertSourceEnrichRule())
		alertInputApi.POST("/source/enrich/get", handler.GetAlertSourceEnrichRule())
		alertInputApi.GET("/enrich/tags/list", handler.ListTargetTags())

		alertInputApi.POST("/cluster/create", handler.CreateCluster())
		alertInputApi.GET("/cluster/list", handler.ListCluster())
		alertInputApi.POST("/cluster/update", handler.UpdateCluster())
		alertInputApi.POST("/cluster/delete", handler.DeleteCluster())

		alertInputApi.POST("/schema/create", handler.CreateSchema())
		alertInputApi.GET("/schema/delete", handler.DeleteSchema())
		alertInputApi.GET("/schema/used/check", handler.CheckSchemaIsUsed())
		alertInputApi.GET("/schema/list", handler.ListSchema())
		alertInputApi.GET("/schema/listwithcolumns", handler.ListSchemaWithColumns())
		alertInputApi.GET("/schema/column/get", handler.GetSchemaColumns())
		alertInputApi.POST("/schema/data/update", handler.UpdateSchemaData())
		alertInputApi.GET("/schema/data/get", handler.GetSchemaData())

		alertInputApi.GET("/source/enrich/default/clear", handler.ClearDefaultAlertEnrichRule())
		alertInputApi.GET("/source/enrich/default/get", handler.GetDefaultAlertEnrichRule())
		alertInputApi.POST("/source/enrich/default/set", handler.SetDefaultAlertEnrichRule())
	}

	integrationAPI := r.mux.Group("/api/integration").Use(middlewares.LicMiddleware())
	{
		handler := integration.New(r.pkg_db)
		integrationAPI.GET("/configuration", handler.GetStaticIntegration())
		integrationAPI.GET("/vars/:variable", handler.GetIntegrationVar())

		integrationAPI.GET("/cluster/list", handler.ListCluster())
		integrationAPI.GET("/cluster/get", handler.GetCluster())
		integrationAPI.POST("/cluster/create", handler.CreateCluster())
		integrationAPI.POST("/cluster/update", handler.UpdateCluster())
		integrationAPI.GET("/cluster/delete", handler.DeleteCluster())

		integrationAPI.GET("/cluster/install/config", handler.GetIntegrationInstallConfigFile())
		integrationAPI.GET("/cluster/install/cmd", handler.GetIntegrationInstallDoc())
		integrationAPI.GET("/adapter/update", handler.TriggerAdapterUpdate())
	}

	metricAPI := r.mux.Group("/api/metric").Use(middlewares.LicMiddleware())
	{
		handler := metric.New(r.logger, r.prom)
		metricAPI.GET("/list", handler.ListMetrics())
		metricAPI.POST("/query", handler.QueryMetrics())
		metricAPI.POST("/queryPods", handler.QueryPods())
	}

	dataplaneAPI := r.mux.Group("/api/dataplane").Use(middlewares.LicMiddleware())
	{
		handler := dataplane.New(r.logger, r.ch, r.prom, r.pkg_db, r.dataplaneRepo)
		dataplaneAPI.GET("/services", handler.QueryServices())
		dataplaneAPI.GET("/redcharts", handler.QueryServiceRedCharts())
		dataplaneAPI.GET("/redcharts/list", handler.ListServiceRedCharts())
		dataplaneAPI.GET("/endpoints", handler.QueryServiceEndpoints())
		dataplaneAPI.GET("/instances", handler.QueryServiceInstances())
		dataplaneAPI.POST("/servicename", handler.QueryServiceName())
		dataplaneAPI.GET("/topology", handler.QueryTopology())
		dataplaneAPI.POST("/checkDataSource", handler.CheckDataSource())

		dataplaneAPI.POST("/customtopology/create", handler.CreateCustomTopology())
		dataplaneAPI.GET("/customtopology/list", handler.ListCustomTopology())
		dataplaneAPI.POST("/customtopology/delete", handler.DeleteCustomTopology())
		dataplaneAPI.POST("/servicename/checkRule", handler.CheckServiceNameRule())
		dataplaneAPI.POST("/servicename/upsertRule", handler.SetServiceNameRule())
		dataplaneAPI.GET("/servicename/listRule", handler.ListServiceNameRule())
		dataplaneAPI.POST("/servicename/deleteRule", handler.DeleteServiceNameRule())
		dataplaneAPI.POST("/servicename/cleanRuleServices", handler.CleanRuleServices())
		dataplaneAPI.POST("/servicename/cleanDataSourceServices", handler.CleanDataSourceServices())
		dataplaneAPI.GET("/type/list", handler.ListDataPlaneType())
		dataplaneAPI.GET("/list", handler.ListDataPlane())
		dataplaneAPI.POST("/create", handler.CreateDataPlane())
		dataplaneAPI.POST("/update", handler.UpdateDataPlane())
		dataplaneAPI.POST("/delete", handler.DeleteDataPlane())
		dataplaneAPI.GET("/data/check", handler.CheckDataPlane())

		dataplaneAPI.GET("/appinfo/tags", handler.QueryAPPInfoTags())
		dataplaneAPI.GET("/appinfo/tags/values", handler.QueryAPPInfoTagValues())
		dataplaneAPI.GET("datasources", handler.GetDatasources())
		dataplaneAPI.GET("servicesWithInstances", handler.GetServiceWithInstances())
		dataplaneAPI.GET("capabilities", handler.GetIntegratedCapabilities())
	}
	licenseApi := r.mux.Group("/api/license").Use(middlewares.AuthMiddleware())
	{
		licenseHandler := license.New(r.logger, r.prom, r.pkg_db)
		licenseApi.POST("/upload", licenseHandler.UploadLicense())
		licenseApi.GET("", licenseHandler.GetLicenses())
		licenseApi.GET("/current", licenseHandler.GetCurrentLicense())
	}

	appApi := r.mux.Group("/api/app").Use(middlewares.AuthMiddleware())
	{
		appHandler := app.New(r.pkg_db)
		appApi.POST("/upload", middlewares.LicMiddleware(), appHandler.UploadLogo())
		appApi.GET("/logo", appHandler.GetLogo())
		appApi.GET("/deployMode", appHandler.GetDeployMode())
	}
}
