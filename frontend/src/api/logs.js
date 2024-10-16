import { del, get, post } from 'src/utils/request'

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
//索引分析(百分比)
export const getLogIndexApi = (params) => {
  return post(`/api/log/index`, params)
}

//获取日志表信息
export const getLogTableInfoAPi = (params) => {
  return get(`/api/log/table`, params)
}

// ————全量日志->日志规则
//添加
export const addLogRuleApi = (params) => {
  return post(`/api/log/rule/add`, params)
}
//更新
export const updateLogRuleApi = (params) => {
  return post(`/api/log/rule/update`, params)
}
//删除
export const deleteLogRuleApi = (params) => {
  return del(`/api/log/rule/delete`, params)
}
//获取指定service的route map
export const getLogRuleServiceRouteRuleApi = (params) => {
  return get(`/api/log/rule/service`, params)
}

//——————全量日志->接入外部表
// 获取所有外部表
export const getLogOtherTableListApi = (params) => {
  return get(`/api/log/other`, params)
}
//获取外部日志表信息
export const getLogOtherTableInfoApi = (params) => {
  return get(`/api/log/other/table`, params)
}
//新增外部日志表信息
export const addLogOtherTableApi = (params) => {
  return post(`/api/log/other/add`, params)
}
//删除外部日志表信息
export const deleteLogOtherTableApi = (params) => {
  return del(`/api/log/other/delete`, params)
}
