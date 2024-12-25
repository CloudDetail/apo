/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { get, post } from 'src/core/utils/request'

// 获取TTL
export const getTTLApi = () => {
  return get(`/api/config/getTTL`)
}
//设置TTL
export const setTTLApi = (params) => {
  return post(`/api/config/setTTL`, params)
}
//设置单个TTL
export const setSingleTableTTLApi = (params) => {
  return post(`/api/config/setSingleTableTTL`, params)
}
