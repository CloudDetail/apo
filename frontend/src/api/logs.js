import { get, post } from 'src/utils/request'

// 获取故障现场分页日志
export const getLogPageListApi = (params) => {
  return post(`/api/log/fault/pagelist`, params)
}
// 获取故障现场日志内容
export const getLogContentApi = (params) => {
  return post(`/api/log/fault/content`, params)
}

// ——————全量日志

//获取全量日志
export const getFullLogApi = (params) => {
  return post(`/api/log/query`, params)
}

//获取全量日志直方图数据
export const getFullLogChartApi = (params) => {
  return post(`/api/log/chart`, params)
}

//查询当前日志规则
export const getLogRuleApi = (params) => {
  return get(`/api/log/rule/get`, params)
}
//更新当前日志规则
export const updateLogRuleApi = (params) => {
  return get(`/api/log/rule/update`, params)
}
