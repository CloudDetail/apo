import { Button, Input, Tooltip } from 'antd'
import { SendOutlined, StopOutlined } from '@ant-design/icons'
import { useRef, useState } from 'react'
import { shouldSubmitChatInput } from './chatInputPolicy.ts'
import styles from '../HermesAgentWorkspace.module.scss'

export default function ChatInput({
  value,
  busy,
  stopping,
  stopRetryAllowed,
  disabled,
  onChange,
  onSend,
  onStop,
}: {
  value: string
  busy: boolean
  stopping?: boolean
  stopRetryAllowed?: boolean
  disabled?: boolean
  onChange: (value: string) => void
  onSend: (text: string) => Promise<void>
  onStop: () => Promise<void>
}) {
  const [submitting, setSubmitting] = useState(false)
  const composingRef = useRef(false)

  const submit = async () => {
    const text = value.trim()
    if (!text || disabled || submitting) return
    setSubmitting(true)
    onChange('')
    try {
      await onSend(text)
    } catch {
      onChange(text)
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className={styles.chatComposer}>
      <Input.TextArea
        value={value}
        disabled={disabled}
        autoSize={{ minRows: 1, maxRows: 5 }}
        placeholder={
          stopping
            ? '正在停止当前任务…'
            : busy
              ? '可补充分析对象、时间范围或排除条件…'
              : '描述对象、时间范围和异常现象，例如：分析最近 30 分钟订单服务错误率升高'
        }
        onChange={(event) => onChange(event.target.value)}
        onCompositionStart={() => {
          composingRef.current = true
        }}
        onCompositionEnd={() => {
          composingRef.current = false
        }}
        onPressEnter={(event) => {
          if (shouldSubmitChatInput(event, composingRef.current)) {
            event.preventDefault()
            void submit()
          }
        }}
      />
      <div className={styles.composerActions}>
        <div>
          {busy && (
            <Tooltip title={stopRetryAllowed ? '再次发送停止请求' : '停止当前任务'}>
              <Button
                danger
                icon={<StopOutlined />}
                disabled={Boolean(stopping && !stopRetryAllowed)}
                loading={Boolean(stopping && !stopRetryAllowed)}
                onClick={() => void onStop()}
              >
                {stopRetryAllowed ? '重试停止' : stopping ? '正在停止' : '停止'}
              </Button>
            </Tooltip>
          )}
          <Button
            type="primary"
            icon={<SendOutlined />}
            disabled={!value.trim() || disabled}
            loading={submitting}
            onClick={() => void submit()}
          >
            {busy ? '补充要求' : '发送'}
          </Button>
        </div>
      </div>
    </div>
  )
}
