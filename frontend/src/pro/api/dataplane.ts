import { get, post } from "src/core/utils/request";

// 数据平面类型列表
export function getDataplaneTypeListApi() {
    return get('/api/dataplane/type/list');
}

// 数据平面列表
export function getDataplaneListApi() {
    return get('/api/dataplane/list');
}

// 创建数据平面
export function createDataplaneApi(params: any) {
    return post('/api/dataplane/create', params);
}

// 更新数据平面定义
export function updateDataplaneApi(params: any) {
    return post('/api/dataplane/update', params);
}

// 删除数据平面
export function deleteDataplaneApi(id: number) {
    return post('/api/dataplane/delete', { id: id });
}

// 获取服务列表
export function getServicesApi(params: any) {
    return get('/api/dataplane/services', params);
}

// 配置或更新服务匹配规则
export function upsertServiceNameRuleApi(params: any) {
    return post('/api/dataplane/servicename/upsertRule', params);
}

// 获取服务匹配规则列表
export function listServiceNameRuleApi(params: any) {
    return get('/api/dataplane/servicename/listRule', params);
}

// 删除服务匹配规则
export function deleteServiceNameRuleApi(params: any) {
    return post('/api/dataplane/servicename/deleteRule', params);
}

// 查询可选择实例标签
export function getAppinfoTagsApi(params: any) {
    return get('/api/dataplane/appinfo/tags', params);
}

// 获取有效标签值
export function getAppinfoTagValuesApi(params: any) {
    return get('/api/dataplane/appinfo/tags/values', params);
}

// 检查服务匹配规则当前匹配情况
export function checkServiceNameRuleApi(params: any) {
    return post('/api/dataplane/servicename/checkRule', params);
}

// 配置自定义服务拓扑
export function createCustomTopologyApi(params: any) {
    return post('/api/dataplane/customtopology/create', params);
}

// 获取整体服务拓扑
export function getTopologyApi(params: any) {
    return get('/api/dataplane/topology', params);
}

// 获取自定义服务拓扑列表
export function getCustomTopologyListApi(params: any) {
    return get('/api/dataplane/customtopology/list', params);
}

// 删除自定义服务拓扑
export function deleteCustomTopologyApi(params: any) {
    return post('/api/dataplane/customtopology/delete', params);
}


export function checkDataplaneDataApi() {
    return get('/api/dataplane/data/check');
}


export function getServiceNameRuleApi(params: any) {
    return get('/api/dataplane/appinfo/tags', params);
}

export function getServiceValuesRuleApi(params: any) {
    return get('/api/dataplane/appinfo/tags/values', params);
}
export function saveServiceRuleApi(params: any) {
    return post('/api/dataplane/servicename/upsertRule', params);
}
export function deleteServiceRuleApi(params: any) {
    return post('/api/dataplane/servicename/deleteRule', params);
}


export function getServiceWithIstanceApi(params: any) {
    return get('/api/dataplane/servicesWithInstances', params);
}
export function getDataplaneDatasourcesApi(params: any) {
    return get('/api/dataplane/datasources', params);
}

export function checkDataplaneDatasourceApi(params: any) {
    return post('/api/dataplane/checkDataSource', params);
}