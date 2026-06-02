/**
 * Copyright 2026 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { del, get, post } from 'src/core/utils/request'

const prefix = '/api/custom-dashboard'

export const listCustomDashboards = (params = {}) => get(`${prefix}/list`, params)
export const getCustomDashboard = (id) => get(`${prefix}/detail`, { id })
export const createCustomDashboard = (params) => post(`${prefix}/create`, params)
export const updateCustomDashboard = (params) => post(`${prefix}/update`, params)
export const copyCustomDashboard = (id) => post(`${prefix}/copy`, { id })
export const deleteCustomDashboard = (id) => del(`${prefix}/delete?id=${id}`)
export const getCustomDashboardTemplates = () => get(`${prefix}/templates`)
export const getCustomDashboardPresetQueries = () => get(`${prefix}/preset-queries`)
export const getCustomDashboardVariableCandidates = (params, config = {}) => post(`${prefix}/variables/query`, params, config)
export const validateCustomDashboardQuery = (params) => post(`${prefix}/query/validate`, params)
export const previewCustomDashboardQuery = (params) => post(`${prefix}/query/preview`, params)
