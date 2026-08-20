import { Alert } from 'antd'
import { useParams } from 'react-router-dom'
import { CapabilityDetailView } from './CapabilityDetailPage.tsx'

/** Historical Workflow versions are rendered through the same read-only view. */
export default function CapabilityVersionPage() {
  const { version } = useParams()
  const parsed = Number(version)
  if (!Number.isInteger(parsed) || parsed < 1) {
    return <Alert type="error" showIcon message="版本号无效" description="请从任务方案详情中的版本历史进入。" />
  }
  return <CapabilityDetailView forcedVersion={parsed} />
}
