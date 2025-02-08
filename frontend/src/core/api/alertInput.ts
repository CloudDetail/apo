/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { AlertInputSourceParams, SaveAlertEnrichParams } from '../types/alertIntegration'
import { get, post } from '../utils/request'

export function getAlertInputSourceListApi() {
  return get('/api/alertinput/source/list')
}

export function creatAlertInputSourceApi(params: AlertInputSourceParams) {
  return post('/api/alertinput/source/create', params)
}
export function updateAlertsIntegrationApi(params: AlertInputSourceParams) {
  return post('/api/alertinput/source/update', params)
}
export function getAlertInputBaseInfoApi(params: AlertInputSourceParams) {
  return post('/api/alertinput/source/get', params)
}

export function getClusterListApi() {
  return get('/api/alertinput/cluster/list')
}

export function getTargetTagsListApi() {
  return get('/api/alertinput/enrich/tags/list')
}

export function getSchemaListApi() {
  return get('/api/alertinput/schema/list')
}
interface SchemaParams {
  schema: string
}
export function getSchemaColumnsApi(params: SchemaParams) {
  return get('/api/alertinput/schema/column/get', params)
}

export function saveAlertEnrichApi(params: SaveAlertEnrichParams) {
  return post('/api/alertinput/source/enrich/update', params)
}
//获取告警源标签关联配置
export function getAlertEnrichApi(params) {
  return post('/api/alertinput/source/enrich/get', params)
}

export function getAllSchemaApi() {
  return get('/api/alertinput/schema/listwithcolumns')
}

export function deleteAlertIntegrationApi(sourceId: string) {
  return post('/api/alertinput/source/delete', { sourceId })
}
