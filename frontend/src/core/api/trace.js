import { post } from 'src/core/utils/request'

// 获取trace日志
export const getTracePageListApi = (params) => {
  return post(`/api/trace/pagelist`, params)
}
