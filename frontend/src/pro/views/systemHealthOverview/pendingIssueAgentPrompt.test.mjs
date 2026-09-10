import test from 'node:test'
import assert from 'node:assert/strict'

import { buildPendingIssueAgentPrompt } from './pendingIssueAgentPrompt.mjs'

test('buildPendingIssueAgentPrompt carries investigation context and asks Agent to verify evidence', () => {
  const prompt = buildPendingIssueAgentPrompt(
    {
      groupId: 1001,
      systemName: '订单系统',
      serviceName: 'order-api',
      metricKey: 'errorRate',
      metricLabel: '错误率',
      status: 'red',
      value: '5.70%',
      note: '错误率超过异常阈值',
    },
    { relativeTime: '15m' },
  )

  assert.match(prompt, /系统范围：订单系统，数据组 ID：1001/)
  assert.match(prompt, /服务：order-api/)
  assert.match(prompt, /异常指标：错误率（errorRate）/)
  assert.match(prompt, /异常等级：异常（red）/)
  assert.match(prompt, /当前观测值：5\.70%/)
  assert.match(prompt, /指标含义：错误率超过异常阈值/)
  assert.match(prompt, /分析时间窗口：最近 15 分钟/)
  assert.match(prompt, /围绕上述异常和时间窗口，查询有助于排查故障的数据，只选择有助于验证结论的证据。/)
  assert.doesNotMatch(prompt, /日志、Trace、调用链、告警、实例和资源数据/)
  assert.match(prompt, /不要编造不存在的日志、Trace、告警、变更记录或根因/)
  assert.doesNotMatch(prompt, /健康度页面/)
  assert.doesNotMatch(prompt, /健康分/)
})

test('buildPendingIssueAgentPrompt preserves all systems for an aggregated queue row', () => {
  const prompt = buildPendingIssueAgentPrompt(
    {
      systemName: '订单系统、支付系统',
      serviceName: 'gateway',
      metricKey: 'latency',
      metricLabel: '响应延时',
      status: 'yellow',
      value: '820ms、960ms',
      note: '响应延时升高',
      issueCount: 2,
      systems: [
        { groupId: 1001, name: '订单系统' },
        { groupId: 1002, name: '支付系统' },
      ],
    },
    { relativeTime: '1h' },
  )

  assert.match(prompt, /订单系统，数据组 ID：1001；支付系统，数据组 ID：1002/)
  assert.match(prompt, /异常等级：关注（yellow）/)
  assert.doesNotMatch(prompt, /同类异常涉及事项数/)
  assert.match(prompt, /分析时间窗口：最近 1 小时/)
})
