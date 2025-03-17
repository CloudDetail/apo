/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { get, post } from "../utils/request"


export const workflowLoginApi = (params) =>{
    return post('/dify/console/api/login/apo',params)
}
export const workflowLogoutApi = () =>{
    return get('/dify/console/api/logout')
}
export const workflowAnonymousLoginApi = (params) =>{
    return post('/dify/console/api/login/apo/anonymous',params)
}