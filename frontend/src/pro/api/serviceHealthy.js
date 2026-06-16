import { get } from 'src/core/utils/request'

export const getSystemHealthReportApi = (params) => {
  return get('/api/system-health/report', params)
}
