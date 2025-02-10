import { get, post } from 'src/core/utils/request'

// 告警分析-》影响面分析,基于用户点击的告警,查询受此告警影响的业务入口
export const getAlertEventImpactApi = (params) => {
  return get(`/api/alerts/event/impact`, params, {
    timeout: 0, // 单独设置超时时间为 0，表示不限制
  })
}
// 告警分析-》下游服务告警/异常事件叠加
export const getAlertDescendantApi = (params) => {
  return get(`/api/alerts/descendant/anormal`, params)
}
// 告警分析-》下游服务告警/下游异常变化查询
export const getAlertDescendantDeltaApi = (params) => {
  return post(`/api/alerts/descendant/anormal/delta`, params)
}
// 告警分析-》下游服务告警/下游异常 折线图数据
export const getAlertDescendantTrendApi = (params) => {
  return get(`/api/alerts/descendant/anormal/trend`, params)
}
// 告警分析-》贡献度
export const getAlertDescendantContributionApi = (params) => {
  return get(`/api/alerts/descendant/anormal/contribution`, params)
}
// 告警分析-》查看某个业务入口的告警分析
export const getAlertDescendantEventApi = (params) => {
  return get(`/api/alerts/descendant/anormal/list`, params)
}

/**
 * 告警分析页面->基于用户选择的下游接口和根因类型,查询故障报告
 * @param {object} params - 包含开始和结束时间的参数对象
 * @param {number} params.startTime - 开始微秒级时间戳
 * @param {number} params.endTime - 结束微秒级时间戳
 * @param {string} params.service - 查询服务名
 * @param {string} params.endpoint - 查询Endpoint
 * @param {string} params.reason - 初因类型
 * @param {string} params.isSlow - span是否为慢请求
 * @param {string} params.isError - span是否出错
 * @param {number} params.currentPage - 当前页
 * @param {number} params.pageSize - 页大小
 * @returns {Promise<Object>} - 返回结果
 */

export const getFailureReportApi = (params) => {
  return post(`/api/alerts/anomaly-span/list`, params)
}

/**
 * 突变指标快速指标获取
 * @returns {Promise<Object>} - 返回结果
 */
export const getMetricsApi = () => {
  return get('/api/alerts/detect/mutation/metrics')
}

/**
 * 添加突变检测规则
 * @param {*} params
 * @returns {Promise<Object>} - 返回结果
 */
export const addMutationRule = (params) => {
  return post('api/alerts/detect/mutation/add', params)
}

/**
 * 获取异常检测执行记录
 * @param {*} params
 * @returns {Promise<Object>} - 返回结果
 */
export const getDetectionRecordApi = (params) => {
  return get('/api/alerts/detect/mutation/exec-list', params)
}

/**
 * 获取异常检测规则
 * @param {*} params
 * @returns {Promise<Object>} - 返回结果
 */
export const getMutationList = (params) => {
  return get('/api/alerts/detect/mutation/rule-list', params)
}
