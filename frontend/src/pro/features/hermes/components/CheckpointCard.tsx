import { Button, Input, Tag } from 'antd'
import {
  CheckCircleOutlined,
  QuestionCircleOutlined,
  SafetyCertificateOutlined,
} from '@ant-design/icons'
import { useState } from 'react'
import ReactMarkdown from 'react-markdown'
import type { HermesCheckpoint } from '../client/types.ts'
import styles from '../HermesAgentWorkspace.module.scss'

export default function CheckpointCard({
  checkpoint,
  resolved = false,
  responseText,
  active = true,
  onRespond,
}: {
  checkpoint: HermesCheckpoint
  resolved?: boolean
  responseText?: string
  active?: boolean
  onRespond: (answer: string) => Promise<void>
}) {
  const [answer, setAnswer] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const plan = checkpoint.purpose === 'plan_confirmation'

  const submit = async (value: string) => {
    if (!active || resolved || !value.trim() || submitting) return
    setSubmitting(true)
    try {
      await onRespond(value)
      setAnswer('')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <section
      className={`${styles.checkpointCard} ${plan ? styles.checkpointPlan : ''} ${
        resolved ? styles.checkpointResolved : ''
      }`}
    >
      <div className={styles.checkpointIcon}>
        {resolved ? (
          <CheckCircleOutlined />
        ) : checkpoint.type === 'approval' ? (
          <SafetyCertificateOutlined />
        ) : (
          <QuestionCircleOutlined />
        )}
      </div>
      <div className={styles.checkpointContent}>
        <div className={styles.checkpointHeading}>
          <span>
            {plan ? '方案确认' : checkpoint.type === 'approval' ? '执行确认' : '补充信息'}
          </span>
          <strong>{checkpoint.title}</strong>
          <Tag bordered={false} color={resolved ? 'success' : 'warning'}>
            {resolved ? '已处理' : '等待回复'}
          </Tag>
        </div>
        {checkpoint.description &&
          (plan ? (
            <div className={styles.planDescription}>
              <ReactMarkdown>{checkpoint.description}</ReactMarkdown>
            </div>
          ) : (
            <p>{checkpoint.description}</p>
          ))}
        {resolved ? (
          <div className={styles.checkpointResponse}>
            <CheckCircleOutlined /> 已选择：{responseText || '已回复'}
          </div>
        ) : checkpoint.choices.length > 0 ? (
          <div className={styles.checkpointChoices}>
            {checkpoint.choices.map((choice, index) => (
              <Button
                key={choice.value}
                type={index === 0 ? 'primary' : 'default'}
                danger={/deny|reject|拒绝|不使用/i.test(choice.value)}
                disabled={!active}
                loading={active && submitting}
                title={choice.description}
                onClick={() => void submit(choice.value)}
              >
                {choice.label}
              </Button>
            ))}
          </div>
        ) : (
          <div className={styles.checkpointInput}>
            <Input.TextArea
              value={answer}
              disabled={!active}
              autoSize={{ minRows: 2, maxRows: 5 }}
              placeholder="请输入补充信息"
              onChange={(event) => setAnswer(event.target.value)}
            />
            <Button
              type="primary"
              disabled={!active || !answer.trim()}
              loading={submitting}
              onClick={() => void submit(answer)}
            >
              提交
            </Button>
          </div>
        )}
      </div>
    </section>
  )
}
