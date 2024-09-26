import { get, post, del } from 'src/utils/request'

// 获取告警规则
export const getAlertRulesApi = (params) => {
  return post(`api/alerts/rule/list`, params)
}
// // 旧获取故障现场分页日志
// export const getAlertRulesApi = (params) => {
//   return get(`/alert/api/v1/rules`, params)
// }
// 重载
export const reloadAlertRulesApi = (params) => {
  return get(`/alert/-/reload`, params)
}

//更新或新增告警规则
export const updateRuleApi = (params) => {
  return post(`api/alerts/rule`, params)
}
//删除告警规则
export const deleteRuleApi = (params) => {
  return del(`api/alerts/rule`, params)
}
//获取group和对应的label
export const getRuleGroupLabelApi = (params) => {
  return get(`api/alerts/rule/groups`, params)
}
//获取告警规则中指标和PQl
export const getRuleMetricsApi = (params) => {
  return get(`api/alerts/rule/metrics`, params)
}
