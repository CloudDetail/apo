import { get, post } from "src/core/utils/request"


export function getModelListApi() {
    return get('/api/v1/model-setting/list')
}
export function setDefaultModelApi(params: any) {
    return post('/api/v1/model-setting/default', params)
}
export function addModelInfoApi(params: any) {
    return post(`/api/v1/model-setting/add`, params)
}
export function updateModelInfoApi(params: any) {
    return post(`/api/v1/model-setting/update`, params)
}
export function deleteModelInfoApi(params: any) {
    return post(`/api/v1/model-setting/remove`, params)
}