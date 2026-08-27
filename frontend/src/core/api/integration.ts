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

export function getNodeVariablesApi(nodeId: string) {
  return get(`/api/agent-config/nodes/${encodeURIComponent(nodeId)}/variables`)
}

export function updateNodeVariablesApi(nodeId: string, overrides: Record<string, string>) {
  return put(`/api/agent-config/nodes/${encodeURIComponent(nodeId)}/variables`, { overrides })
}

export function getNodeClusterIdentityApi(nodeId: string) {
  return get(`/api/agent-config/nodes/${encodeURIComponent(nodeId)}/cluster-identity`)
}

export function updateNodeClusterIdentityApi(nodeId: string, data: { clusterId: string; clusterIp?: string }) {
  return put(`/api/agent-config/nodes/${encodeURIComponent(nodeId)}/cluster-identity`, data)
}

export function restoreNodeClusterIdApi(nodeId: string) {
  return post(`/api/agent-config/nodes/${encodeURIComponent(nodeId)}/cluster-identity/restore`, {})
}

export function getAgentConfigStatesApi() {
  return get('/api/agent-config/states')
}

export function getAgentConfigDeliveriesApi(limit = 50) {
  return get('/api/agent-config/deliveries', { limit })
}

export function getAgentConfigTemplatesApi() {
  return get('/api/agent-config/templates')
}

export function createAgentConfigTemplateApi(data: {
  name: string
  description?: string
  clusterType: 'kubernetes' | 'vm'
  variables?: Array<{
    name: string
    required?: boolean
    defaultValue?: string
  }>
  configs: Array<{
    probeType: string
    files: Array<{ path: string; content: string }>
  }>
}) {
  return post('/api/agent-config/templates', data)
}

export function importAgentConfigTemplateApi(data: { yaml: string }) {
  return post('/api/agent-config/templates/import', data)
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

export function updateNodeProbeConfigsApi(nodeId: string, data: unknown) {
  return put(`/api/agent-config/nodes/${encodeURIComponent(nodeId)}/probe-configs`, data)
}

export function getNodeInjectionControlApi(nodeId: string) {
  return get(`/api/agent-config/nodes/${encodeURIComponent(nodeId)}/injection-control`)
}

export function updateNodeInjectionControlApi(nodeId: string, data: unknown) {
  return put(`/api/agent-config/nodes/${encodeURIComponent(nodeId)}/injection-control`, data)
}

export function batchUpdateNodeInjectionControlApi(data: unknown) {
  return post('/api/agent-config/injection-control/batch', data)
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

export function getClusterInjectionControlApi(clusterId: string) {
  return get(`/api/agent-config/clusters/${encodeURIComponent(clusterId)}/injection-control`)
}

export function updateClusterInjectionControlApi(clusterId: string, data: unknown) {
  return put(`/api/agent-config/clusters/${encodeURIComponent(clusterId)}/injection-control`, data)
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

export function getLogCollectionDefaultsApi() {
  return get('/api/agent-config/defaults/log-collection')
}

export function updateLogCollectionDefaultsApi(data: unknown) {
  return put('/api/agent-config/defaults/log-collection', data)
}

export function getHostFileCollectionApi(nodeId: string) {
  return get(`/api/agent-config/log-collection/nodes/${encodeURIComponent(nodeId)}/host-files`)
}

export function updateHostFileCollectionApi(nodeId: string, data: unknown) {
  return put(
    `/api/agent-config/log-collection/nodes/${encodeURIComponent(nodeId)}/host-files`,
    data,
  )
}

export function previewHostFileCollectionApi(nodeId: string, data: unknown) {
  return post(
    `/api/agent-config/log-collection/nodes/${encodeURIComponent(nodeId)}/host-files/preview`,
    data,
  )
}

export function getContainerStdoutCollectionApi(clusterId: string) {
  return get(
    `/api/agent-config/log-collection/clusters/${encodeURIComponent(clusterId)}/container-stdout`,
  )
}

export function updateContainerStdoutCollectionApi(clusterId: string, data: unknown) {
  return put(
    `/api/agent-config/log-collection/clusters/${encodeURIComponent(clusterId)}/container-stdout`,
    data,
  )
}

export function previewContainerStdoutCollectionApi(clusterId: string, data: unknown) {
  return post(
    `/api/agent-config/log-collection/clusters/${encodeURIComponent(clusterId)}/container-stdout/preview`,
    data,
  )
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

export function updateClusterDisplayNameApi(id: string, name?: string) {
  return post('/api/integration/cluster/display-name', { id, name: name || '' })
}

export function deleteClusterIntegrationApi(id: string) {
  return get('/api/integration/cluster/delete', { id })
}

export function getIntegrationVarApi(variable: string) {
  return get(`/api/integration/vars/${variable}`)
}
