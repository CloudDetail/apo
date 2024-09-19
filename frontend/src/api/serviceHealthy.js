import { get } from 'src/utils/request'

export const getServiceHealthyApi = (params) => {
  return get('/api/service/ryglight', params)
}
