/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { post } from "../utils/request"


export const workflowLoginApi = (params) =>{
    return post('/dify/console/api/login',params)
}