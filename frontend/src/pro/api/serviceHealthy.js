import { get, post } from 'src/core/utils/request'

export const getSystemHealthReportApi = (params) => {
  return get('/api/system-health/report', params, {
    timeout: 0,
  })
}

export const getSystemHealthOverviewApi = (params) => {
  return get('/api/system-health/overview', params)
}

export const getSystemHealthSystemDetailApi = (groupId, params) => {
  return get(`/api/system-health/overview/${groupId}`, params)
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
