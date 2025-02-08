/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

export interface GetDataGroupsParams {
  groupName?: string
  currentPage: number
  pageSize: number
}
export type DatasourceType = 'service' | 'namespace'
export type DatasourceCategory = 'normal' | 'apm'
interface DatasourceItem {
  datasource: string
  type: DatasourceType
  category?: DatasourceCategory
}
export interface SaveDataGroupParams {
  groupId?: string
  groupName: string
  description?: string
  datasourceList: DatasourceItem[]
}
export type PermissionType = 'view' | 'edit'
export interface PermissionSub {
  subjectId: string
  type: PermissionType
}
export interface DataGroupSubsParams {
  groupId: string
  userList: PermissionSub[]
  teamList: PermissionSub[]
}
export type subjectType = 'user' | 'team'
export interface DataGroupPermission {
  dataGroupId: string
  type: PermissionType
}
export interface GetSubsDataGroupParams {
  subjectId: string
  subjectType: subjectType
  category?: DatasourceCategory
}
export interface SubsDataGroupParams {
  subjectId: string
  subjectType: subjectType
  dataGroupPermission: DataGroupPermission[]
}
