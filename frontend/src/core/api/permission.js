/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { get, headers, post } from '../utils/request'

//获取用户权限（菜单+路由）
export function getUserPermissionApi(params) {
  return get('/api/permission/config', params)
}
//获取所有可配置权限
export function getAllPermissionApi(params) {
  return get('/api/permission/feature', params)
}
//配置全局菜单
export function configMenuApi(params) {
  return post('/api/permission/menu/configure', params, headers.formUrlencoded)
}
//获取所有角色列表
export function getAllRoleList() {
  return get('/api/permission/roles')
}

//获取角色或用户的所有权限列表
export function getSubjectPermissionApi(params) {
  return get('/api/permission/sub/feature', params)
}
