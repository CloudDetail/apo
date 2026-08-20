// @ts-nocheck
import assert from 'node:assert/strict'
import test from 'node:test'
import { containsSensitiveValue, isSensitiveHint, redactSensitiveText } from './sensitive.ts'

test('detects assignment-style and long-token credential values', () => {
  assert.equal(containsSensitiveValue('password: p@ssword-value'), true)
  assert.equal(containsSensitiveValue('token is abcdefghijklmnop'), true)
  assert.equal(containsSensitiveValue('检查 token 过期'), false)
  assert.equal(containsSensitiveValue('密码是过期'), false)
})

test('redacts credential values without removing the field label', () => {
  assert.equal(redactSensitiveText('password: p@ssword-value'), 'password: [已隐藏]')
  assert.equal(redactSensitiveText('Authorization Bearer abcdefghijklmnop'), 'Authorization Bearer [已隐藏]')
  assert.equal(isSensitiveHint('credentialRef'), true)
})
