/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { get } from '../utils/request'

/**
 * 获取应用部署模式
 * @returns {Promise<Object>} - 返回部署模式信息
 */
export const getDeployModeApi = () => {
    return get('/api/app/deployMode')
} 