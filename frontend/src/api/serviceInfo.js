import { get } from 'src/utils/request'

// 服务直接上下游拓扑

export const getServiceTopologyApi = (params) => {
  return get(`api/service/topology`, params)
}
// 获取更多下游拓扑
export const getServiceRelationApi = (params) => {
  return get(`/api/service/relation`, params)
}

//获取所有下游服务的延时曲线
export const getServiceDsecendantMetricsApi = (params) => {
  return get(`api/service/descendant/metrics`, params)
}
//获取service对应url实例
export const getServiceInstancesApi = (params) => {
  return get(`/api/service/instances`, params)
}
//获取service对应url错误实例
export const getServiceErrorInstancesApi = (params) => {
  return get(`/api/service/error/instance`, params)
}
//获取service对应url错误实例日志切片
export const getServiceErrorInstancesLogsApi = (params) => {
  return get(`/api/service/errorinstance/logs`, params)
}

//获取service依赖节点延时关联度
export const getServiceDsecendantRelevanceApi = (params) => {
  return get(`/api/service/descendant/relevance`, params)
}
//获取北极星指标分析情况
export const getPolarisInferApi = (params) => {
  return get(`/api/service/polaris/infer`, params)
}
//获取日志tab table数据
export const getLogMetricsApi = (params) => {
  return get(`/api/service/log/metrics`, params)
}
//获取Trace tab table数据
export const getTraceMetricsApi = (params) => {
  return get(`/api/service/trace/metrics`, params)
}
//获取K8s事件
export const getK8sEventApi = (params) => {
  return get(`/api/service/k8s/events/count`, params)
}

//日志TAB故障现场日志列表切片
export const getServiceLogLogsApi = (params) => {
  return get(`/api/service/log/logs`, params)
}

//获取Trace TAB故障现场获取Trace列表切片
export const getServiceTraceLogsApi = (params) => {
  return get(`/api/service/trace/logs`, params)
}

//告警列表
export const getServiceAlertEventApi = (params) => {
  return get(`/api/service/alert/sample/events`, params)
}

//获取SQL指标Tab
export const getServiceSqlMetrics = (params) => {
  return get(`/api/service/sql/metrics`, params)
}
