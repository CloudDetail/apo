import { get } from 'src/core/utils/request'

// 获取对话历史消息
export function getConversationMessagesApi(threadId: string) {
    return get(`/ai/api/v1/conversations/messages?threadId=${threadId}`)
}

// 获取选定服务的 findings 数据
export function getFindingsDataApi(params: {
    threadId: string        // 必需：对话线程 ID
    serviceName?: string   // 可选：服务名称，用于过滤特定服务的 findings 数据
    [key: string]: any
}) {
    const queryParams = new URLSearchParams(params).toString()
    return get(`/ai/api/v1/conversations/findings-data?${queryParams}`)
} 