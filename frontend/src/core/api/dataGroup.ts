/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import {
  DataGroupSubsParams,
  GetDataGroupsParams,
  GetSubsDataGroupParams,
  SaveDataGroupParams,
  SubsDataGroupParams,
} from '../types/dataGroup'
import { get, headers, post } from '../utils/request'

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

export function getDataGroupPermissionSubsApi(groupId: string) {
  return get('/api/data/subs', { groupId })
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
  return get('/api/data/group/data', params)
}
