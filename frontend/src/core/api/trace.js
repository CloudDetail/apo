/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { post } from 'src/core/utils/request'

// 获取trace日志
export const getTracePageListApi = (params) => {
  return post(`/api/trace/pagelist`, params)
}
