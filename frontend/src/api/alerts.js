import { get } from 'src/utils/request'

// 获取故障现场分页日志
export const getAlertRulesApi = (params) => {
  return get(`/alert/api/v1/rules`, params)
}
// 获取故障现场分页日志
export const reloadAlertRulesApi = (params) => {
  return get(`/alert/-/reload`, params)
}
