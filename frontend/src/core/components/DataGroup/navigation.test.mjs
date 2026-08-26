import assert from 'node:assert/strict'
import test from 'node:test'
import { buildDataGroupNavigation } from './navigation.ts'

test('补写数据组参数时保留当前路由状态和 hash', () => {
  const workflowLaunch = { workflowId: 'workflow-1', goal: '检查订单服务' }
  const location = {
    pathname: '/agent',
    hash: '#result',
    state: { workflowLaunch },
  }
  const params = new URLSearchParams('groupId=7')

  assert.deepEqual(buildDataGroupNavigation(location, params), {
    to: {
      pathname: '/agent',
      search: '?groupId=7',
      hash: '#result',
    },
    options: {
      replace: true,
      state: { workflowLaunch },
    },
  })
})

test('更新数据组参数时不复用旧 search 字符串', () => {
  const params = new URLSearchParams('page=2&groupId=3')
  params.set('groupId', '8')

  const result = buildDataGroupNavigation(
    { pathname: '/agent/capabilities', hash: '', state: null },
    params,
  )

  assert.equal(result.to.search, '?page=2&groupId=8')
  assert.equal(result.options.state, null)
})
