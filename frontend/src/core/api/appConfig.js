/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { get, post } from 'src/core/utils/request'

// Upsert接口 - Bot安装成功后添加TeamId
export const upsertAppConfigApi = (params) => {
    return post(`/api/v1/app-config/upsert`, params)
}

// 获取Slack状态接口
export const getSlackStatusApi = () => {
    return get(`/api/v1/app-config/slack-status`)
}
