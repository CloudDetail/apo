import { get, post } from 'src/core/utils/request'

export const getSystemHealthReportApi = (params) => {
  return get('/api/system-health/report', params, {
    timeout: 0,
  })
}

export const getSystemHealthOverviewApi = (params) => {
  return get('/api/system-health/overview', params)
}

export const getSystemHealthSystemsApi = (params) => {
  return get('/api/system-health/systems', params)
}

export const getSystemHealthServicesApi = (groupId, params) => {
  return get(`/api/system-health/overview/${groupId}/services`, params)
}

export const getSystemHealthPendingIssuesApi = (params) => {
  return get('/api/system-health/pending-issues', params)
}

export const getSystemHealthTrendApi = (params) => {
  return get('/api/system-health/trend', { judgement: 'snapshot', ...params })
}

export const getSystemHealthSystemsTrendApi = (data) => {
  return post('/api/system-health/trend/systems', { judgement: 'snapshot', ...data })
}

export const getSystemHealthSystemDetailApi = (groupId, params) => {
  return get(`/api/system-health/overview/${groupId}`, params)
}

export const getSystemHealthConfigApi = () => {
  return get('/api/system-health/config')
}

export const setSystemHealthConfigApi = (data) => {
  return post('/api/system-health/config', data)
}

export const analyzeSystemHealthPendingIssueApi = (data) => {
  return post('/api/system-health/pending-issues/analyze', data)
}

export const getSystemHealthReportRecordsApi = (params) => {
  return get('/api/system-health/report/records', params)
}

export const getSystemHealthReportRecordApi = (recordId) => {
  return get(`/api/system-health/report/records/${recordId}`)
}
