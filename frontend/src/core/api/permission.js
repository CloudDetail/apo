import { get, headers, post } from '../utils/request'

//获取用户权限（菜单+路由）
export function getUserPermissionApi(params) {
  return get('/api/permission/config', params)
}
//获取所有可配置权限
export function getAllPermissionApi() {
  return get('/api/permission/feature')
}
//配置全局菜单
export function configMenuApi(params) {
  return post('/api/permission/menu/configure', params, headers.formUrlencoded)
}
//获取所有角色列表
export function getAllRoleList() {
  return get('/api/permission/roles')
}
