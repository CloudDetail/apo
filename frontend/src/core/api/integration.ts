/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { get, post } from '../utils/request'

export function getIntegrationConfigApi() {
  return get('/api/integration/configuration')
}

export function createDataIntegrationApi(params) {
  return post('/api/integration/cluster/create', params)
}

export function getIntegrationClusterListApi() {
  return get('/api/integration/cluster/list')
}
export function getClusterInstallCmdApi(clusterId: string) {
  return get('/api/integration/cluster/install/cmd', { clusterId })
}
export function getClusterInstallConfigApi(clusterId: string) {
  return get('/api/integration/cluster/install/config', { clusterId }, { responseType: 'blob' })
}
export function getClusterInstallPackageApi(clusterId: string, type: 'k8s' | 'vm') {
  return get(
    '/cluster/integration/cluster/install/package/' + type,
    { clusterId },
    { responseType: 'blob' },
  )
}

export function getClusterIntegrationInfoApi(id: string) {
  return get('/api/integration/cluster/get', { id })
}

export function updateDataIntegrationApi(params) {
  return post('/api/integration/cluster/update', params)
}

export function deleteClusterIntegrationApi(id: string) {
  return get('/api/integration/cluster/delete', { id })
}


export function getIntegrationVarApi(variable: string) {
  return get(`/api/integration/vars/${variable}`)
}
