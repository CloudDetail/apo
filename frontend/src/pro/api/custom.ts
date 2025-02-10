import { del, get, post } from 'src/core/utils/request'

//获取所有自定义类型列表
export function getCustomList() {
  return get(`/api/custom/get`)
}

interface customDataViewParams {
  customName: string
  type: string[]
}
// 创建自定义数据类型
export function creatCustomDataView(params: customDataViewParams) {
  return post(`/api/custom/create`, params)
}

// 更新自定义数据类型
export function updateCustomDataView(params: customDataViewParams) {
  return post(`/api/custom/update`, params)
}

// 删除自定义数据类型
export function deleteCustomDataView(customName: string) {
  return del(`/api/custom/delete?customName=${customName}`)
}
