import { get } from 'src/core/utils/request'

export const getSystemHealthReportApi = (params) => {
  return get('/api/system-health/report', params, {
    timeout: 0,
  })
}

export const getSystemHealthReportRecordsApi = (params) => {
  return get('/api/system-health/report/records', params)
}

export const getSystemHealthReportRecordApi = (recordId) => {
  return get(`/api/system-health/report/records/${recordId}`)
}
