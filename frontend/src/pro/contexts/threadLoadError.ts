export type ThreadLoadErrorKind = 'not_found' | 'transient' | 'unknown'

const readErrorMessage = (error: unknown) =>
  error instanceof Error && error.message.trim().length > 0 ? error.message.trim() : ''

export const classifyThreadLoadError = (error: unknown): ThreadLoadErrorKind => {
  const message = readErrorMessage(error).toLowerCase()

  if (!message) {
    return 'unknown'
  }

  if (
    message.includes('thread not found') ||
    message.includes('conversation not found') ||
    message.includes('thread_not_found') ||
    message.includes('conversation_not_found') ||
    message.includes('not found') ||
    message.includes('404')
  ) {
    return 'not_found'
  }

  if (
    message.includes('timeout') ||
    message.includes('network') ||
    message.includes('status code 5')
  ) {
    return 'transient'
  }

  return 'unknown'
}

export const formatThreadLoadErrorMessage = ({
  error,
  messages,
}: {
  error: unknown
  messages: {
    loadFailed: string
    unknownError: string
  }
}) => {
  const detail = readErrorMessage(error) || messages.unknownError

  return `${messages.loadFailed}: ${detail}`
}
