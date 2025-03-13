// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package router

import (
	"github.com/CloudDetail/apo/backend/internal/api/mock"
	alertinput "github.com/CloudDetail/apo/backend/pkg/api/alertinput"
	"github.com/CloudDetail/apo/backend/pkg/api/alerts"
	"github.com/CloudDetail/apo/backend/pkg/api/config"
	"github.com/CloudDetail/apo/backend/pkg/api/data"
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
	"github.com/CloudDetail/apo/backend/pkg/util"
	"github.com/CloudDetail/metadata/source"
)

func setApiRouter(r *resource) {
	middlewares := middleware.New(r.cache, r.pkg_db)
	api := r.mux.Group("/api")
	{
		mockHandler := mock.New(r.logger, r.internal_db)
		api.POST("/mock", mockHandler.Create())
		api.GET("/mock", mockHandler.List())
		api.GET("/mock/:id", mockHandler.Detail())
		api.DELETE("/mock/:id", mockHandler.Delete())
	}

	serviceApi := r.mux.Group("/api/service").Use(middlewares.AuthMiddleware())
	{
		serviceOverviewHandler := serviceoverview.New(r.logger, r.ch, r.prom, r.pkg_db, r.k8sApi)
		serviceApi.GET("/endpoints", serviceOverviewHandler.GetEndPointsData())
		serviceApi.GET("/servicesAlert", serviceOverviewHandler.GetServicesAlert())
		serviceApi.GET("/moreUrl", serviceOverviewHandler.GetServiceMoreUrlList())
		serviceApi.GET("/getThreshold", serviceOverviewHandler.GetThreshold())
		serviceApi.POST("/setThreshold", serviceOverviewHandler.SetThreshold())
		serviceApi.GET("/ryglight", serviceOverviewHandler.GetRYGLight())
		serviceApi.GET("/monitor/status", serviceOverviewHandler.GetMonitorStatus())

		serviceHandler := service.New(r.logger, r.ch, r.prom, r.pol, r.pkg_db, r.k8sApi)
		serviceApi.GET("/entry/endpoints", serviceHandler.GetServiceEntryEndpoints())
		serviceApi.GET("/relation", serviceHandler.GetServiceEndpointRelation())
		serviceApi.GET("/topology", serviceHandler.GetServiceEndpointTopology())
		serviceApi.GET("/descendant/metrics", serviceHandler.GetDescendantMetrics())
		serviceApi.GET("/descendant/relevance", serviceHandler.GetDescendantRelevance())
		serviceApi.GET("/polaris/infer", serviceHandler.GetPolarisInfer())
		serviceApi.GET("/error/instance", serviceHandler.GetErrorInstance())
		serviceApi.GET("/errorinstance/logs", serviceHandler.GetErrorInstanceLogs())
		serviceApi.GET("/log/metrics", serviceHandler.GetLogMetrics())
		serviceApi.GET("/log/logs", serviceHandler.GetLogLogs())
		serviceApi.GET("/trace/metrics", serviceHandler.GetTraceMetrics())
		serviceApi.GET("/trace/logs", serviceHandler.GetTraceLogs())

		serviceApi.GET("/list", serviceHandler.GetServiceList())
		serviceApi.GET("/instances", serviceHandler.GetServiceInstance())
		serviceApi.GET("/instance/list", serviceHandler.GetServiceInstanceList())
		serviceApi.GET("/instanceinfo/list", serviceHandler.GetServiceInstanceInfoList())
		serviceApi.GET("/instance/options", serviceHandler.GetServiceInstanceOptions())
		serviceApi.GET("/endpoint/list", serviceHandler.GetServiceEndPointList())
		serviceApi.GET("/k8s/events/count", serviceHandler.CountK8sEvents())
		serviceApi.GET("/namespace/list", serviceHandler.GetNamespaceList())

		serviceApi.GET("/alert/events", serviceHandler.GetAlertEvents())
		serviceApi.GET("/alert/sample/events", serviceHandler.GetAlertEventsSample())

		serviceApi.GET("/sql/metrics", serviceHandler.GetSQLMetrics())
		serviceApi.GET("/redcharts", serviceHandler.GetServiceREDCharts())
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
		alertHandler := alerts.New(r.logger, r.ch, r.pkg_db, r.k8sApi, r.prom, r.alertWorkflow)
		alertApi.POST("/event/list", alertHandler.AlertEventList())
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
		userHandler := user.New(r.logger, r.pkg_db, r.cache)
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
	{
		dataHandler := data.New(r.logger, r.pkg_db, r.prom, r.k8sApi)
		dataApi.GET("/datasource", dataHandler.GetDatasource())
		dataApi.POST("/group", dataHandler.GetDataGroup())
		dataApi.GET("/group/data", dataHandler.GetGroupDatasource())
		dataApi.POST("/group/update", dataHandler.UpdateDataGroup())
		dataApi.POST("/group/create", dataHandler.CreateDataGroup())
		dataApi.POST("/group/delete", dataHandler.DeleteDataGroup())
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
		handler := alertinput.New(r.logger, r.ch, r.prom, r.pkg_db, r.alertWorkflow)
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
	}
}

func SetMetaServerRouter(srv *Server, meta source.MetaSource) {
	api := srv.Mux.Group("/metadata")
	for path, handler := range meta.Handlers() {
		// This set of APIs supports both GET and POST
		api.POST_Gin(path, util.WrapHandlerFunctions(handler))
		api.GET_Gin(path, util.WrapHandlerFunctions(handler))
	}
}
