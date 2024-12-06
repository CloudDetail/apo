import { get, post } from 'src/core/utils/request'

export const getServiceTableApi = (params) => {
  return get(`api/service/top3`, params)
}
export const getEndpointTableApi = (params) => {
  return get(`api/service/moreUrl`, params)
}
/**
 * 获取所有服务列表（通常用于下拉选择）
 * @param {Object} params - 包含开始和结束时间的参数对象
 * @param {number} params.startTime - 开始微秒级时间戳
 * @param {number} params.endTime - 结束微秒级时间戳
 * @returns {Promise<Object>} - 包含服务列表的Promise对象
 */
export const getServiceListApi = (params) => {
  return get(`/api/service/list`, params)
}
/**
 * 获取指定服务的所有实例列表（通常用于下拉选择 与service 下拉联动）
 * @param {Object} params - 包含开始和结束时间的参数对象
 * @param {number} params.startTime - 开始微秒级时间戳
 * @param {number} params.endTime - 结束微秒级时间戳
 * @param {string} params.service - service name
 * @returns {Promise<Object>} - 返回结果
 */
export const getServiceInstanceListApi = (params) => {
  return get(`/api/service/instance/list`, params)
}
/**
 * 获取指定服务的所有实例map列表（通常用于下拉选择 与service 下拉联动）
 * @param {Object} params - 包含开始和结束时间的参数对象
 * @param {number} params.startTime - 开始微秒级时间戳
 * @param {number} params.endTime - 结束微秒级时间戳
 * @param {string} params.service - service name
 * @returns {Promise<Object>} - 返回结果
 */
export const getServiceInstancOptionsListApi = (params) => {
  return get(`/api/service/instance/options`, params)
}

/**
 * 获取应用table前半部分的数据
 * @param {Object} params - 包含开始和结束时间的参数对象
 * @param {number} params.startTime - 开始微秒级时间戳
 * @param {number} params.endTime - 结束微秒级时间戳
 * @param {string} params.step - 步长
 * @param {string} params.serviceName - serviceName
 * @param {number} params.sortRule -1 默认是按是否超出阈值排序，参数传递值为1
 * @returns {Promise<Object>} - 返回结果
 */
export const getServicesEndpointsApi = (params) => {
  return get(`/api/service/endpoints`, params)
}
/**
 * 获取日志告警或指示灯接口
 * @param {Object} params - 包含开始和结束时间的参数对象
 * @param {number} params.startTime - 开始微秒级时间戳
 * @param {number} params.endTime - 结束微秒级时间戳
 * @param {string} params.step - 步长
 * @param {string} params.returnData - returnData
 * @param {string} params.serviceNames - serviceNames
 * @returns {Promise<Object>} - 返回结果
 */
export const getServicesAlertApi = (params) => {
  return get(`/api/service/servicesAlert`, params)
}

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

/**
 * 告警分析页面->获取告警分析服务端点数据
 * @param {*} params 
 * @returns {Promise<Object>}
 */
export const getServiceEndpointNameApi = (params) => {
  return get(`/api/service/moreUrl`, params)
}


/**
 * 获取所有命名空间
 * @param {*} params 
 * @returns {Promise<Object>}
 */
export const getNamespacesApi = (params) => {
  return get(`/api/k8s/namespaces`, params)
}