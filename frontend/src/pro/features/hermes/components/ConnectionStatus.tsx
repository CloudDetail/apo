import { Badge } from 'antd'
import type { HermesConnectionStatus } from '../client/types.ts'

const statusConfig: Record<
  HermesConnectionStatus,
  { status: 'default' | 'processing' | 'warning' | 'error'; text: string } | null
> = {
  idle: null,
  connecting: { status: 'processing', text: '正在连接 Agent' },
  connected: null,
  reconnecting: { status: 'warning', text: '连接中断，正在重连' },
  disconnected: { status: 'warning', text: 'Agent 连接已断开' },
  error: { status: 'error', text: 'Agent 连接异常' },
}

export default function ConnectionStatus({ status }: { status: HermesConnectionStatus }) {
  const config = statusConfig[status]
  if (!config) return null
  return <Badge status={config.status} text={config.text} />
}
