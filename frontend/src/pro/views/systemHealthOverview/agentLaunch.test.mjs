import test from 'node:test'
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { URL } from 'node:url'

const sourceFor = (path) => readFileSync(new URL(path, import.meta.url), 'utf8')

const healthViewSource = sourceFor('./index.jsx')
const pendingQueueSource = sourceFor('./components/PendingIssueQueue.jsx')
const serviceAssistantSource = sourceFor('./components/ServiceHealthAssistantModal.jsx')
const agentWorkspaceSource = sourceFor('../../features/hermes/HermesAgentWorkspace.tsx')

test('pending and service health analysis launch independent Agent questions', () => {
  assert.ok(healthViewSource.includes("navigate('/agent', {"))
  assert.ok(healthViewSource.includes('buildPendingIssueAgentPrompt(issue, queryParams || {})'))
  assert.ok(healthViewSource.includes("openSystemHealthIssueAgent(issue, 'system_health_pending_issue')"))
  assert.ok(healthViewSource.includes("openSystemHealthIssueAgent(issue, 'system_health_service_issue')"))
  assert.ok(pendingQueueSource.includes('onClick={() => onOpenAgent(issue)}'))
  assert.ok(serviceAssistantSource.includes('onClick={() => onAnalyzeIssue(issue)}'))
  assert.ok(serviceAssistantSource.includes('在 Agent 中分析'))
})

test('Agent workspace consumes launch state, clears it, and auto-sends once', () => {
  assert.ok(agentWorkspaceSource.includes('agentLaunch?: AgentLaunchContext'))
  assert.ok(agentWorkspaceSource.includes('state: null'))
  assert.ok(agentWorkspaceSource.includes('autoSubmittedLaunchRef'))
  assert.ok(agentWorkspaceSource.includes('workspace.sendMessage(text)'))
  assert.ok(agentWorkspaceSource.includes('setAgentLaunch(null)'))
  assert.ok(agentWorkspaceSource.includes('setDraft(text)'))
})
