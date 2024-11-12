import { get } from 'src/core/utils/request'

/**
 * 告警分析页面->获取告警事件
 * @param {Object} params - 包含开始和结束时间的参数对象
 * @param {number} params.startTime - 开始微秒级时间戳
 * @param {number} params.endTime - 结束微秒级时间戳
 * @param {string} params.currentPage - 当前页数
 * @param {string} params.pageSize - 每页数量
 *
 * @param {string} params.service - 查询服务名
 * @param {string} params.source - 查询告警来源
 * @param {string} params.group - 查询告警类型
 * @param {string} params.name - 查询告警名
 * @param {string} params.id - 查询告警id
 * @param {string} params.status - 查询告警状态
 *
 * @returns {Promise<Object>} - 返回结果
 */
export const getServiceAlertEventsApi = (params) => {
  return get(`/api/service/alert/events`, params)
}
/**
 * 告警分析页面->获取服务上下游调用关系
 * @param {Object} params - 包含开始和结束时间的参数对象
 * @param {number} params.startTime - 开始微秒级时间戳
 * @param {number} params.endTime - 结束微秒级时间戳
 * @param {string} params.service - 查询服务名
 * @param {string} params.endpoint - 查询Endpoint
 * @param {string} params.entryService - 查询入口服务名
 * @param {string} params.entryEndpoint - 查询入口Endpoint
 * @param {boolean} params.withTopologyLevel - 是否返回层级
 *
 * @returns {Promise<Object>} - 返回结果
 */
export const getServiceRelationApi = (params) => {
  return get(`/api/service/relation`, params)
}
