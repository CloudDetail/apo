import assert from 'node:assert/strict'
import test from 'node:test'
import { shouldSubmitChatInput } from './chatInputPolicy.ts'

const keyboardState = (overrides = {}) => ({
  ctrlKey: false,
  metaKey: false,
  shiftKey: false,
  nativeEvent: { isComposing: false },
  ...overrides,
})

test('普通 Enter 发送消息', () => {
  assert.equal(shouldSubmitChatInput(keyboardState()), true)
})

test('Ctrl/Cmd/Shift + Enter 只换行', () => {
  assert.equal(shouldSubmitChatInput(keyboardState({ ctrlKey: true })), false)
  assert.equal(shouldSubmitChatInput(keyboardState({ metaKey: true })), false)
  assert.equal(shouldSubmitChatInput(keyboardState({ shiftKey: true })), false)
})

test('输入法组合态下 Enter 不会发送', () => {
  assert.equal(
    shouldSubmitChatInput(keyboardState({ nativeEvent: { isComposing: true } })),
    false,
  )
  assert.equal(shouldSubmitChatInput(keyboardState(), true), false)
  assert.equal(
    shouldSubmitChatInput(keyboardState({ nativeEvent: { keyCode: 229 } })),
    false,
  )
  assert.equal(
    shouldSubmitChatInput(keyboardState({ nativeEvent: { which: 229 } })),
    false,
  )
})
