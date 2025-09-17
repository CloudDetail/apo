import { get } from 'src/core/utils/request'

// 获取对话历史消息
export function getConversationMessagesApi(threadId: string) {
    return get(`/api/v1/conversations/messages?threadId=${threadId}`)
} 