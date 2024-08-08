import { get } from 'src/utils/request'

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
