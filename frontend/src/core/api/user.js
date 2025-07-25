/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { post, get, headers } from '../utils/request'

/**
 * 登录
 * @param {Object} params
 * @param {string} params.username
 * @param {string} params.password
 * @returns {Promise<Object>}
 */
const loginApi = (params) => {
  return post(`/api/user/login`, params, headers.formUrlencoded)
}

/**
 * 登出
 * @param {Object} params
 * @param {string} params.accessToken
 * @param {string} params.refreshToken
 * @returns {Promise<Object>}
 */
const logoutApi = (params) => {
  return post(`/api/user/logout`, params, headers.formUrlencoded)
}

/**
 * 更新邮件
 * @param {Object} params
 * @param {string} params.username
 * @param {string} params.email
 * @returns {Promise<Object>}
 */
const updateEmailApi = (params) => {
  return post(`/api/user/update/email`, params, headers.formUrlencoded)
}

/**
 * 更新用户信息
 * @param {Object} params
 * @param {string} params.userId
 * @param {string} params.corporation
 * @returns {Promise<Object>}
 */
const updateCorporationApi = (params) => {
  return post(`/api/user/update/info`, params, headers.formUrlencoded)
}

/**
 * 更新密码
 * @param {Object} params
 * @param {string} params.userId
 * @param {string} params.oldPassword
 * @param {string} params.newPassword
 * @param {string} params.confirmPassword
 * @returns {Promise<Object>}
 */
const updatePasswordApi = (params) => {
  return post(`/api/user/update/password`, params, headers.formUrlencoded)
}

/**
 * 更新手机号码
 * @param {Object} params
 * @param {string} params.username
 * @param {string} params.phone
 * @returns {Promise<Object>}
 */
const updatePhoneApi = (params) => {
  return post(`/api/user/update/phone`, params, headers.formUrlencoded)
}

/**
 * 创建用户
 * @param {Object} params
 * @param {string} params.username
 * @param {string} params.password
 * @param {string} params.confirmPassword
 * @returns {Promise<Object>}
 */
const createUserApi = (params) => {
  return post(`/api/user/create`, params)
}

/**
 * 获取用户信息
 * @returns {Promise<Object>}
 */
const getUserInfoApi = () => {
  return get(`api/user/info`)
}

/**
 * 获取用户列表
 * @param {Object} params
 * @param {number} params.currentPage - 当前页码 (必填)
 * @param {number} params.pageSize - 每页数量 (必填)
 * @param {string} [params.username] - 用户名 (可选)
 * @param {string} [params.role] - 角色 (可选)
 * @param {string} [params.corporation] - 公司 (可选)
 * @param {*} signal
 * @returns {Promise<Object>}
 */
const getUserListApi = ({ currentPage, pageSize, ...optionalParams }, signal = null) => {
  const requiredParams = { currentPage, pageSize }
  const params = { ...requiredParams, ...optionalParams }
  return get(`/api/user/list`, params, { signal })
}

/**
 * 移除用户
 * @param {Object} params
 * @param {string} params.userId
 * @returns {Promise<Object>}
 */
const removeUserApi = (params) => {
  return post(`/api/user/remove`, params, headers.formUrlencoded)
}

/**
 * 重设密码
 * @param {Object} params
 * @param {string} params.userId
 * @param {string} params.newPassword
 * @returns
 */
const updatePasswordWithNoOldPwdApi = (params) => {
  return post(`/api/user/reset`, params, headers.formUrlencoded)
}

export {
  loginApi,
  logoutApi,
  updateEmailApi,
  updateCorporationApi,
  updatePasswordApi,
  updatePhoneApi,
  createUserApi,
  getUserInfoApi,
  getUserListApi,
  removeUserApi,
  updatePasswordWithNoOldPwdApi,
}
