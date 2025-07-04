/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import {
  DataGroupSubsParams,
  DatasourceCategory,
  GetDataGroupsParams,
  GetSubsDataGroupParams,
  SaveDataGroupParams,
  SubsDataGroupParams,
} from '../types/dataGroup'
import { del, get, headers, post } from '../utils/request'

export function getDataGroupsApi(params: GetDataGroupsParams) {
  return post('/api/data/group', params)
}

export function getAllDatasourceApi() {
  return get('/api/data/datasource')
}

export function creatDataGroupApi(params: SaveDataGroupParams) {
  return post('/api/data/group/create', params)
}
export function updateDataGroupApi(params: SaveDataGroupParams) {
  return post('/api/data/group/update', params)
}

export function deleteDataGroupApi(groupId: string) {
  return post('/api/data/group/delete', { groupId }, headers.formUrlencoded)
}

export function getDataGroupPermissionSubsApi(groupId: string, subjectType?: 'user' | 'team') {
  return get('/api/data/subs', { groupId, subjectType })
}

export function updateDataGroupSubsApi(params: DataGroupSubsParams) {
  return post('/api/data/subs/operation', params)
}

export function updateSubsDataGroupApi(params: SubsDataGroupParams) {
  return post('/api/data/group/operation', params)
}
export function getSubsDataGroupApi(params: GetSubsDataGroupParams) {
  return get('/api/data/sub/group', params)
}

export function getDatasourceByGroupApi(params) {
  return post('/api/data/group/data', params)
}

export function getUserGroupApi(userId: string, category: DatasourceCategory) {
  return get('/api/data/user/group', { userId, category })
}

//v2
export function getDatasourceByGroupApiV2() {
  return get('/api/v2/data/group')
}
export function getCheckableDatasourceApi(groupId: string, skipNotChecked?: boolean | null) {
  return get('/api/v2/data/group/datasource/list', { groupId, skipNotChecked })
}
export function addDataGroupApi(params) {
  return post('/api/v2/data/group/add', params)
}
export function updateDataGroupApiV2(params) {
  return post('/api/v2/data/group/update', params)
}
export function getSubGroupsApiV2(groupId: string) {
  return get('/api/v2/data/group/detail', { groupId })
}
export function deleteDataGroupApiV2(groupId: string) {
  return del('/api/v2/data/group/delete', { groupId })
}
export function getDatasourceFilterApiV2(params) {
  return post('/api/v2/data/group/filter', params)
}