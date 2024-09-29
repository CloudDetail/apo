package router

import (
	"github.com/CloudDetail/apo/backend/internal/api/mock"
	"github.com/CloudDetail/apo/backend/pkg/api/alerts"
	"github.com/CloudDetail/apo/backend/pkg/api/config"
	"github.com/CloudDetail/apo/backend/pkg/api/log"
	"github.com/CloudDetail/apo/backend/pkg/api/service"
	"github.com/CloudDetail/apo/backend/pkg/api/serviceoverview"
	"github.com/CloudDetail/apo/backend/pkg/api/trace"
	"github.com/CloudDetail/apo/backend/pkg/util"
	"github.com/CloudDetail/metadata/source"
)

func setApiRouter(r *resource) {
	api := r.mux.Group("/api")
	{
		mockHandler := mock.New(r.logger, r.internal_db)
		api.POST("/mock", mockHandler.Create())
		api.GET("/mock", mockHandler.List())
		api.GET("/mock/:id", mockHandler.Detail())
		api.DELETE("/mock/:id", mockHandler.Delete())
	}

	serviceApi := r.mux.Group("/api/service")
	{
		serviceOverviewHandler := serviceoverview.New(r.logger, r.ch, r.prom, r.pkg_db)
		serviceApi.GET("/endpoints", serviceOverviewHandler.GetEndPointsData())
		serviceApi.GET("/servicesAlert", serviceOverviewHandler.GetServicesAlert())
		serviceApi.GET("/moreUrl", serviceOverviewHandler.GetServiceMoreUrlList())
		serviceApi.GET("/getThreshold", serviceOverviewHandler.GetThreshold())
		serviceApi.POST("/setThreshold", serviceOverviewHandler.SetThreshold())
		serviceApi.GET("/ryglight", serviceOverviewHandler.GetRYGLight())

		serviceHandler := service.New(r.logger, r.ch, r.prom, r.pol, r.pkg_db)
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
		serviceApi.GET("/instance/options", serviceHandler.GetServiceInstanceOptions())
		serviceApi.GET("/endpoint/list", serviceHandler.GetServiceEndPointList())
		serviceApi.GET("/k8s/events/count", serviceHandler.CountK8sEvents())

		serviceApi.GET("/alert/events", serviceHandler.GetAlertEvents())
		serviceApi.GET("/alert/sample/events", serviceHandler.GetAlertEventsSample())

		serviceApi.GET("/sql/metrics", serviceHandler.GetSQLMetrics())
	}

	logApi := r.mux.Group("/api/log")
	{
		logHandler := log.New(r.logger, r.ch)
		logApi.POST("/fault/pagelist", logHandler.GetFaultLogPageList())
		logApi.POST("/fault/content", logHandler.GetFaultLogContent())
	}

	traceApi := r.mux.Group("/api/trace")
	{
		traceHandler := trace.New(r.logger, r.ch)
		traceApi.POST("/pagelist", traceHandler.GetTracePageList())
		traceApi.GET("/pagelist/filters", traceHandler.GetTraceFilters())
		traceApi.POST("/pagelist/filter/value", traceHandler.GetTraceFilterValue())
	}

	alertApi := r.mux.Group("/api/alerts")
	{
		alertHandler := alerts.New(r.logger, r.ch, r.k8sApi)
		alertApi.POST("/inputs/alertmanager", alertHandler.InputAlertManager())
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

	configApi := r.mux.Group("/api/config")
	{
		configHandler := config.New(r.logger, r.ch)
		configApi.POST("/setTTL", configHandler.SetTTL())
		configApi.POST("/setSingleTableTTL", configHandler.SetSingleTableTTL())
		configApi.GET("/getTTL", configHandler.GetTTL())
	}

}

func SetMetaServerRouter(srv *Server, meta source.MetaSource) {
	api := srv.Mux.Group("/metadata")
	for path, handler := range meta.Handlers() {
		// 这组API同时支持GET和POST
		api.POST_Gin(path, util.WrapHandlerFunctions(handler))
		api.GET_Gin(path, util.WrapHandlerFunctions(handler))
	}
}
