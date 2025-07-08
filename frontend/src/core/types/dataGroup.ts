/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

export interface GetDataGroupsParams {
  groupName?: string
  currentPage: number
  pageSize: number
}
export type DatasourceType = 'system' | 'cluster' | 'namespace' | 'service'
export type DatasourceCategory = 'normal' | 'apm'
interface DatasourceItem {
  datasource: string
  type: DatasourceType
  category?: DatasourceCategory
}
export interface SaveDataGroupParams {
  groupId?: number
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
  groupId: number
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
export interface DataGroupItem {
  groupId: number
  groupName: string
  subGroups: DataGroupItem[]
  description: string
  disabled?: boolean
}
export const DatasourceTypes: DatasourceType[] = ['system', 'cluster', 'namespace', 'service']
export interface DataGroupPermissionInfo {
  groupId: number
  groupName: string
  description: string
  permissionType: 'known' | 'view' | 'edit'
  subGroups?: DataGroupPermissionInfo[]
}