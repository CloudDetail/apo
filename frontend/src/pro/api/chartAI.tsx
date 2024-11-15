import { post } from 'src/core/utils/request'

// ai对话
export function alertchatApi(params: any) {
  return post(`/api/alertchat`, params)
}
