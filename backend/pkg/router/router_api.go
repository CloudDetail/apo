// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package router

import (
	alertinput "github.com/CloudDetail/apo/backend/pkg/api/alertinput"
	"github.com/CloudDetail/apo/backend/pkg/api/alerts"
	"github.com/CloudDetail/apo/backend/pkg/api/config"
	"github.com/CloudDetail/apo/backend/pkg/api/data"
	"github.com/CloudDetail/apo/backend/pkg/api/dataplane"
	"github.com/CloudDetail/apo/backend/pkg/api/health"
	"github.com/CloudDetail/apo/backend/pkg/api/integration"
	"github.com/CloudDetail/apo/backend/pkg/api/k8s"
	"github.com/CloudDetail/apo/backend/pkg/api/log"
	"github.com/CloudDetail/apo/backend/pkg/api/metric"
	networkapi "github.com/CloudDetail/apo/backend/pkg/api/network"
	"github.com/CloudDetail/apo/backend/pkg/api/permission"
	"github.com/CloudDetail/apo/backend/pkg/api/role"
	"github.com/CloudDetail/apo/backend/pkg/api/service"
	"github.com/CloudDetail/apo/backend/pkg/api/serviceoverview"
	"github.com/CloudDetail/apo/backend/pkg/api/team"
	"github.com/CloudDetail/apo/backend/pkg/api/trace"
	"github.com/CloudDetail/apo/backend/pkg/api/user"
	"github.com/CloudDetail/apo/backend/pkg/middleware"
)

func setApiRouter(r *resource) {
	middlewares := middleware.New(r.cache, r.pkg_db, r.dify)

	serviceApi := r.mux.Group("/api/service").Use(middlewares.AuthMiddleware())
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

	logApi := r.mux.Group("/api/log").Use(middlewares.AuthMiddleware())
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

	traceApi := r.mux.Group("/api/trace")
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

	alertApi := r.mux.Group("/api/alerts")
	{
		alertHandler := alerts.New(r.logger, r.ch, r.pkg_db, r.k8sApi, r.prom, r.dify, r.receivers)
		alertApi.POST("/event/list", alertHandler.AlertEventList())
		alertApi.POST("/event/detail", alertHandler.AlertEventDetail())
		alertApi.GET("/events/classify", alertHandler.AlertEventClassify())
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

		alertApi.GET("/slient", alertHandler.GetAlertSlienceConfig())
		alertApi.GET("/slient/list", alertHandler.ListAlertSlienceConfig())
		alertApi.DELETE("/slient", alertHandler.RemoveAlertSlienceConfig())
		alertApi.POST("/slient", alertHandler.SetAlertSlienceConfig())

		alertApi.POST("/resolve", alertHandler.MarkAlertResolvedManually())

		alertApi.GET("/filter/keys", alertHandler.GetAlertEventStaticFilters())
		alertApi.POST("/filter/labelkeys", alertHandler.GetAlertEventLabelFilterKeys())
		alertApi.POST("/filter/values", alertHandler.SearchAlertEventFilterValues())
	}

	configApi := r.mux.Group("/api/config").Use(middlewares.AuthMiddleware())
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
		userApi.POST("/create", userHandler.CreateUser())
		userApi.POST("/update/password", userHandler.UpdateUserPassword())
		userApi.POST("/update/phone", userHandler.UpdateUserPhone())
		userApi.POST("/update/email", userHandler.UpdateUserEmail())
		userApi.POST("/update/info", userHandler.UpdateUserInfo())
		userApi.POST("/update/self", userHandler.UpdateSelfInfo())
		userApi.GET("/info", userHandler.GetUserInfo())
		userApi.GET("/list", userHandler.GetUserList())
		userApi.POST("/remove", userHandler.RemoveUser())
		userApi.POST("/reset", userHandler.ResetPassword())
		userApi.GET("/team", userHandler.GetUserTeam())
	}

	permissionApi := r.mux.Group("/api/permission").Use(middlewares.AuthMiddleware())
	{
		permissionHandler := permission.New(r.logger, r.pkg_db)
		permissionApi.GET("/config", permissionHandler.GetUserConfig())
		permissionApi.GET("/feature", permissionHandler.GetFeature())
		permissionApi.GET("/sub/feature", permissionHandler.GetSubjectFeature())
		permissionApi.POST("/operation", permissionHandler.PermissionOperation())
		permissionApi.POST("/menu/configure", permissionHandler.ConfigureMenu())
		permissionApi.GET("/router", permissionHandler.CheckRouterPermission())
	}

	roleApi := r.mux.Group("/api/role").Use(middlewares.AuthMiddleware())
	{
		roleHandler := role.New(r.logger, r.pkg_db)
		roleApi.GET("/roles", roleHandler.GetRole())
		roleApi.GET("/user", roleHandler.GetUserRole())
		roleApi.POST("/operation", roleHandler.RoleOperation())
		roleApi.POST("/create", roleHandler.CreateRole())
		roleApi.POST("/update", roleHandler.UpdateRole())
		roleApi.POST("/delete", roleHandler.DeleteRole())
	}

	dataApi := r.mux.Group("/api/data").Use(middlewares.AuthMiddleware())
	dataApiV2 := r.mux.Group("/api/v2/data").Use(middlewares.AuthMiddleware())
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
		teamApi.POST("/create", teamHandler.CreateTeam())
		teamApi.POST("/update", teamHandler.UpdateTeam())
		teamApi.GET("", teamHandler.GetTeam())
		teamApi.POST("/delete", teamHandler.DeleteTeam())
		teamApi.POST("/operation", teamHandler.TeamOperation())
		teamApi.POST("/user/operation", teamHandler.TeamUserOperation())
		teamApi.GET("/user", teamHandler.GetTeamUser())
	}

	k8sApi := r.mux.Group("/api/k8s")
	{
		k8sHandler := k8s.New(r.k8sApi)
		// These APIs are used by another project. DO NOT Auth.
		k8sApi.GET("/namespaces", k8sHandler.GetNamespaceList())
		k8sApi.GET("/namespace/info", k8sHandler.GetNamespaceInfo())
		k8sApi.GET("/pods", k8sHandler.GetPodList())
		k8sApi.GET("/pod/info", k8sHandler.GetPodInfo())
	}
	networkApi := r.mux.Group("/api/network/")
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

	alertInputApi := r.mux.Group("/api/alertinput")
	{
		handler := alertinput.New(r.logger, r.ch, r.prom, r.pkg_db, r.dify)
		alertInputApi.POST("/event/source", handler.SourceHandler())
		alertInputApi.POST("/event/json", handler.JsonHandler())
		alertInputApi.POST("/source/create", handler.CreateAlertSource())
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

	integrationAPI := r.mux.Group("/api/integration")
	{
		handler := integration.New(r.pkg_db)
		integrationAPI.GET("/configuration", handler.GetStaticIntegration())

		integrationAPI.GET("/cluster/list", handler.ListCluster())
		integrationAPI.GET("/cluster/get", handler.GetCluster())
		integrationAPI.POST("/cluster/create", handler.CreateCluster())
		integrationAPI.POST("/cluster/update", handler.UpdateCluster())
		integrationAPI.GET("/cluster/delete", handler.DeleteCluster())

		integrationAPI.GET("/cluster/install/config", handler.GetIntegrationInstallConfigFile())
		integrationAPI.GET("/cluster/install/cmd", handler.GetIntegrationInstallDoc())
		integrationAPI.GET("/adapter/update", handler.TriggerAdapterUpdate())
	}

	metricAPI := r.mux.Group("/api/metric")
	{
		handler := metric.New(r.logger, r.prom)
		metricAPI.GET("/list", handler.ListMetrics())
		metricAPI.POST("/query", handler.QueryMetrics())
		metricAPI.POST("/queryPods", handler.QueryPods())
	}

	dataplaneAPI := r.mux.Group("/api/dataplane")
	{
		handler := dataplane.New(r.logger, r.ch, r.prom, r.pkg_db)
		dataplaneAPI.GET("/services", handler.QueryServices())
		dataplaneAPI.GET("/redcharts", handler.QueryServiceRedCharts())
		dataplaneAPI.GET("/endpoints", handler.QueryServiceEndpoints())
		dataplaneAPI.GET("/instances", handler.QueryServiceInstances())
		dataplaneAPI.POST("/servicename", handler.QueryServiceName())
		dataplaneAPI.GET("/topology", handler.QueryTopology())

		dataplaneAPI.POST("/customtopology/create", handler.CreateCustomTopology())
		dataplaneAPI.GET("/customtopology/list", handler.ListCustomTopology())
		dataplaneAPI.POST("/customtopology/delete", handler.DeleteCustomTopology())
		dataplaneAPI.POST("/servicename/checkRule", handler.CheckServiceNameRule())
		dataplaneAPI.POST("/servicename/upsertRule", handler.SetServiceNameRule())
		dataplaneAPI.GET("/servicename/listRule", handler.ListServiceNameRule())
		dataplaneAPI.POST("/servicename/deleteRule", handler.DeleteServiceNameRule())
	}
}
