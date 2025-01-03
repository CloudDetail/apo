/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */
import { get, post, headers } from '../utils/request'

/**
 * Get team list
 * @param {object} params
 * @param {string} params.teamName
 * @param {number[]} params.featureList
 * @param {number[]} params.dataGroupList
 * @param {number} params.currentPage
 * @param {number} params.pageSize
 * @returns {Promise<Object>}
 */
export const getTeamListApi = (params) => {
  return get('/api/team', params)
}

/**
 * create team
 * @param {object} params
 * @param {string} params.teamName
 * @param {number[]} params.featureList
 * @param {number[]} params.dataGroupList
 * @returns {Promise<Object>}
 */
export const createTeamApi = (params) => {
  return post('/api/team/create', params, headers.formUrlencoded)
}

/**
 * delete team
 * @param {object} params
 * @param {number} params.teamId
 * @returns {Promise<Object>}
 */
export const deleteTeamApi = (params) => {
  return post('/api/team/delete', params, headers.formUrlencoded)
}

/**
 * operation team
 * @param {object} params
 * @param {number} params.userId
 * @param {number} params.teamList
 * @returns {Promise<Object>}
 */
export const operationTeamApi = (params) => {
  return post('/api/team/operation', params, headers.formUrlencoded)
}

/**
 * update team info
 * @param {object} params
 * @param {number} params.teamId
 * @param {string} params.teamName
 * @param {number[]} params.featureList
 * @param {number[]} params.dataGroupList
 * @returns {Promise<Object>}
 */
export const updateTeamApi = (params) => {
  return post('/api/team/update', params, headers.formUrlencoded)
}

/**
 * get team's user list
 * @param {object} params
 * @param {number} params.teamId
 * @returns {Promise<Object>}
 */
export const getTeamUserListApi = (params) => {
  return get('/api/team/user', params)
}

/**
 * operation team user
 * @param {object} params
 * @param {number} params.teamId
 * @param {number[]} params.userList
 * @returns {Promise<Object>}
 */
export const operationTeamUserApi = (params) => {
  return post('/api/team/user/operation', params, headers.formUrlencoded)
}
