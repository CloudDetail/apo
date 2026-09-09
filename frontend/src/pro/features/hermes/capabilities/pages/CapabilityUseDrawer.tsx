import { useEffect } from 'react'
import { Alert, Button, Drawer, Form, Input, message } from 'antd'
import { SendOutlined } from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import type { WorkflowInput, WorkflowLaunchContext, WorkCapability } from '../client/types.ts'
import { containsSensitiveValue, isSensitiveHint, redactSensitiveText } from '../client/sensitive.ts'
import styles from './index.module.scss'

interface CapabilityUseDrawerProps {
  open: boolean
  capability: WorkCapability
  onClose: () => void
  onNavigate?: () => void
}

interface UseValues {
  goal: string
  [key: string]: string | undefined
}

/** Collects this run's inputs and hands an explicit Workflow ID to the Agent page. */
export default function CapabilityUseDrawer({ open, capability, onClose, onNavigate }: CapabilityUseDrawerProps) {
  const navigate = useNavigate()
  const [form] = Form.useForm<UseValues>()
  const userInputs = capability.inputs.filter(isUserProvidedInput).filter((input) => !isGoalInput(input))
  const sensitiveInputs = userInputs.filter((input) => isSensitiveHint(input.name))
  const visibleInputs = userInputs.filter((input) => !isSensitiveHint(input.name))

  useEffect(() => {
    if (!open) return
    form.resetFields()
    form.setFieldsValue({ goal: capability.goal })
  }, [capability.goal, form, open])

  const submit = (values: UseValues) => {
    const inputValues = Object.fromEntries(
      visibleInputs
        .map((input) => [input.name, String(values[inputKey(input.name)] || '').trim()])
        .filter(([, value]) => value),
    ) as Record<string, string>
    const rawGoal = String(values.goal || '').trim()
    if (containsSensitiveValue(rawGoal) || Object.values(inputValues).some(containsSensitiveValue)) {
      message.error('检测到疑似敏感凭据。请使用 Agent 授权或安全输入流程，不要把秘密写入普通会话。')
      return
    }
    const goal = buildGoal(rawGoal, inputValues)
    if (!goal) return
    const launch: WorkflowLaunchContext = {
      workflowId: capability.id,
      workflowName: capability.name,
      version: capability.currentVersion,
      goal,
      source: 'workflow_library',
    }
    navigate('/agent', { state: { workflowLaunch: launch } })
    onNavigate?.()
  }

  return (
    <Drawer
      open={open}
      width={460}
      title="使用该任务"
      destroyOnClose
      onClose={onClose}
      extra={<Button type="text" onClick={onClose}>关闭</Button>}
    >
      <p className={styles.drawerHint}>填写本次任务目标和所需输入。提交后会回到运维会话，由 Agent 复用当前 Workflow 创建新的运行。不要在普通会话中输入密码、Token、密钥或凭据。</p>
      {sensitiveInputs.length ? (
        <Alert
          className={styles.drawerNotice}
          type="warning"
          showIcon
          message="敏感参数不会在此处收集"
          description={`该任务声明了${sensitiveInputs.map((item) => item.name).join('、')}等敏感字段。请按 Agent 的授权或安全输入流程提供。`}
        />
      ) : null}
      <Form form={form} layout="vertical" className={styles.drawerForm} onFinish={submit}>
        <Form.Item name="goal" label="本次任务目标" rules={[{ required: true, message: '请输入本次任务目标' }]}>
          <Input.TextArea autoSize={{ minRows: 3, maxRows: 7 }} placeholder="描述本次要排查或确认的问题" />
        </Form.Item>
        {visibleInputs.map((input) => (
          <Form.Item
            key={input.name}
            name={inputKey(input.name)}
            label={input.name}
            rules={input.required ? [{ required: true, whitespace: true, message: `请填写${input.name}` }] : undefined}
          >
            <Input placeholder={`填写${input.name}`} />
          </Form.Item>
        ))}
        <div className={styles.drawerFooter}>
          <Button onClick={onClose}>取消</Button>
          <Button type="primary" htmlType="submit" icon={<SendOutlined />}>带入运维会话</Button>
        </div>
      </Form>
    </Drawer>
  )
}

function isUserProvidedInput(input: WorkflowInput): boolean {
  return ['conversation_or_prompt', 'conversation', 'user'].includes(input.source)
}

function isGoalInput(input: WorkflowInput): boolean {
  return /^(request|prompt|goal|question|query|任务|问题|目标)$/i.test(input.name.trim())
}

function inputKey(name: string): string {
  return `input:${name}`
}

function buildGoal(goal: string, inputValues: Record<string, string>): string {
  const context = Object.entries(inputValues)
    .map(([key, value]) => `${key}：${sanitizePromptText(value)}`)
    .join('；')
  const safeGoal = sanitizePromptText(goal)
  return context ? `${safeGoal}\n本次输入：${context}` : safeGoal
}

function sanitizePromptText(value: string): string {
  return redactSensitiveText(value)
}
