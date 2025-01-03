import { get, post, headers } from '../utils/request'

/**
 * 获取团队列表
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
 * 创建团队
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
 * 删除团队
 * @param {object} params
 * @param {number} params.teamId
 * @returns {Promise<Object>}
 */
export const deleteTeamApi = (params) => {
  return post('/api/team/delete', params, headers.formUrlencoded)
}

/**
 * 操作团队
 * @param {object} params
 * @param {number} params.userId
 * @param {number} params.teamList
 * @returns {Promise<Object>}
 */
export const operationTeamApi = (params) => {
  return post('/api/team/operation', params, headers.formUrlencoded)
}

/**
 * 更新团队信息
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
 * 获取团队用户列表
 * @param {object} params
 * @param {number} params.teamId
 * @returns {Promise<Object>}
 */
export const getTeamUserListApi = (params) => {
  return get('/api/team/user', params)
}

/**
 * 操作团队用户
 * @param {object} params
 * @param {number} params.teamId
 * @param {number[]} params.userList
 * @returns {Promise<Object>}
 */
export const operationTeamUserApi = (params) => {
  return post('/api/team/user/operation', params, headers.formUrlencoded)
}
