/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { get, headers, post } from '../utils/request'

export interface Role {
  roleId: number; // Role ID
  roleName: string; // Role Name
  description: string; // Role Description
}

export function getAllRolesApi() {
  return get('/api/role/roles');
}

export function revokeUserRoleApi(params) {
  return post(`/api/role/operation`, params, headers.formUrlencoded)
}

export function deleteRoleApi(params) {
  return post(`/api/role/delete`, params, headers.formUrlencoded)
}

export function createRoleApi(params) {
  return post(`/api/role/create`, params, headers.formUrlencoded)
}

export function updateRoleApi(params) {
  console.log('updateRoleApi: ', params)
  return post(`/api/role/update`, params, headers.formUrlencoded)
}