import { get } from 'src/core/utils/request'

export const getServiceHealthyApi = (params) => {
  return get('/api/service/ryglight', params)
}
