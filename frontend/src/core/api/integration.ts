/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { get, post, put } from '../utils/request'

export function getIntegrationConfigApi() {
  return get('/api/integration/configuration')
}

export function createDataIntegrationApi(params) {
  return post('/api/integration/cluster/create', params)
}

export function getIntegrationClusterListApi() {
  return get('/api/integration/cluster/list')
}

export function getReportedClusterIdsApi() {
  return get('/api/dataplane/data/check')
}

export function getAgentConfigNodesApi() {
  return get('/api/agent-config/nodes')
}

export function getAgentConfigStatesApi() {
  return get('/api/agent-config/states')
}

export function getAgentConfigTemplatesApi(probeType?: string) {
  return get('/api/agent-config/templates', { probeType })
}

export function createAgentConfigTemplateApi(data: {
  probeType: string
  name: string
  content: string
  files?: Array<{ path: string; content: string }>
  format: string
}) {
  return post('/api/agent-config/templates', data)
}

export function updateAgentConfigTemplateApi(templateId: string, data: unknown) {
  return put(`/api/agent-config/templates/${encodeURIComponent(templateId)}`, data)
}

export function getGlobalCircuitBreakerApi() {
  return get('/api/agent-config/defaults/circuit-breaker')
}

export function updateGlobalCircuitBreakerApi(data: unknown) {
  return put('/api/agent-config/defaults/circuit-breaker', data)
}

export function getNodeCurrentConfigsApi(nodeId: string) {
  return get(`/api/agent-config/nodes/${encodeURIComponent(nodeId)}/current-configs`)
}

export function updateNodeCurrentConfigsApi(nodeId: string, data: unknown) {
  return put(`/api/agent-config/nodes/${encodeURIComponent(nodeId)}/current-configs`, data)
}

export function getNodeConfigHistoryApi(nodeId: string, probeType: string, limit = 20) {
  return get(
    `/api/agent-config/nodes/${encodeURIComponent(nodeId)}/probes/${encodeURIComponent(probeType)}/releases`,
    { limit },
  )
}

export function createAgentCommandApi(
  nodeId: string,
  probeType: string,
  type: string,
  options: Record<string, unknown> = {},
) {
  return post(
    `/api/agent-config/nodes/${encodeURIComponent(nodeId)}/probes/${encodeURIComponent(probeType)}/commands`,
    { type, ...options },
  )
}

export function getAgentCommandsApi(nodeId: string, probeType: string, limit = 20) {
  return get(
    `/api/agent-config/nodes/${encodeURIComponent(nodeId)}/probes/${encodeURIComponent(probeType)}/commands`,
    { limit },
  )
}

export function getNodeCircuitBreakerApi(nodeId: string) {
  return get(`/api/agent-config/nodes/${encodeURIComponent(nodeId)}/circuit-breaker`)
}

export function updateNodeCircuitBreakerApi(nodeId: string, data: unknown) {
  return put(`/api/agent-config/nodes/${encodeURIComponent(nodeId)}/circuit-breaker`, data)
}

export function batchAssignAgentConfigApi(data: unknown) {
  return post('/api/agent-config/batch-assignments', data)
}

export function batchUpdateCircuitBreakerApi(data: unknown) {
  return post('/api/agent-config/circuit-breaker/batch', data)
}

export function getNodeLogSourcesApi(nodeId: string, probeType: string) {
  return get(
    `/api/agent-config/nodes/${encodeURIComponent(nodeId)}/probes/${encodeURIComponent(probeType)}/log-sources`,
  )
}

export function updateNodeLogSourcesApi(nodeId: string, probeType: string, data: unknown) {
  return put(
    `/api/agent-config/nodes/${encodeURIComponent(nodeId)}/probes/${encodeURIComponent(probeType)}/log-sources`,
    data,
  )
}

export function batchUpdateLogSourcesApi(data: unknown) {
  return post('/api/agent-config/log-sources/batch', data)
}

export function getAgentInitialConfigApi(nodeId: string, probeType: string) {
  return get(
    `/api/agent-config/nodes/${encodeURIComponent(nodeId)}/probes/${encodeURIComponent(probeType)}/initial-config`,
  )
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
