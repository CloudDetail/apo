import { post } from 'src/utils/request'

// 获取故障现场分页日志
export const getLogPageListApi = (params) => {
  return post(`/api/log/fault/pagelist`, params,)
}
// 获取故障现场日志内容
export const getLogContentApi = (params) => {
  return post(`/api/log/fault/content`, params,)
}
