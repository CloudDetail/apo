export const buildConversationMessagesRequestParams = (threadId: string) => ({
  threadId,
})

export const buildConversationHistoryListRequestParams = <T extends Record<string, unknown>>(
  params: T,
  userId?: string,
) => {
  if (!userId) {
    return params
  }

  return {
    ...params,
    userId,
  }
}
