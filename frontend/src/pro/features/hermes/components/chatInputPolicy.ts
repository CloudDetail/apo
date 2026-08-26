interface ChatInputKeyboardState {
  ctrlKey: boolean
  metaKey: boolean
  shiftKey: boolean
  nativeEvent?: {
    isComposing?: boolean
    keyCode?: number
    which?: number
  }
}

export function shouldSubmitChatInput(
  event: ChatInputKeyboardState,
  composing = false,
): boolean {
  const nativeEvent = event.nativeEvent
  return (
    !event.ctrlKey &&
    !event.metaKey &&
    !event.shiftKey &&
    !composing &&
    !nativeEvent?.isComposing &&
    nativeEvent?.keyCode !== 229 &&
    nativeEvent?.which !== 229
  )
}
