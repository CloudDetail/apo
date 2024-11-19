import { get, post } from 'src/core/utils/request'

// ai对话
export function alertchatApi(params: any) {
  return post(`/api/alertchat`, params)
}
// 用户输入对话
export function chatApi(params: any) {
  return post(`/api/chat`, params)
}

//ai对话中文件下载
export function downloadFileApi(params: any) {
  return get(`/api/download`, params)
}
