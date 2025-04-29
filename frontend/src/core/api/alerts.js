/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { get, post, del } from 'src/core/utils/request'

// 获取告警规则
export const getAlertRulesApi = (params) => {
  return post(`api/alerts/rule/list`, params)
}
// 获取远程告警规则状态信息
export const getAlertRulesStatusApi = (params) => {
  return get(`/alert/api/v1/rules`, params)
}
// 重载
export const reloadAlertRulesApi = (params) => {
  return get(`/alert/-/reload`, params)
}

//更新告警规则
export const updateRuleApi = (params) => {
  return post(`api/alerts/rule`, params)
}
//新增告警规则
export const addRuleApi = (params) => {
  return post(`/api/alerts/rule/add`, params)
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

// 通知类
//告警通知规则获取
export const getAlertmanagerListApi = (params) => {
  return post(`/api/alerts/alertmanager/receiver/list`, params)
}
//添加告警通知规则获取
export const addAlertNotifyApi = (params) => {
  return post(`/api/alerts/alertmanager/receiver/add`, params)
}
//更新告警通知规则获取
export const updateAlertNotifyApi = (params) => {
  return post(`/api/alerts/alertmanager/receiver`, params)
}
//删除告警通知规则
export const deleteAlertNotifyApi = (params) => {
  return del(`/api/alerts/alertmanager/receiver`, params)
}

export const getAlertEventsApi = (params) => {
  return post('/api/alerts/event/list', params)
}

export const getAlertWorkflowIdApi = (params) => {
  return get('/api/alerts/events/classify', params)
}
