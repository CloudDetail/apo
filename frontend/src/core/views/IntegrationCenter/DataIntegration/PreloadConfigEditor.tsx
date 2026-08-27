import { Alert, Button, Modal, Switch, Typography } from 'antd'
import { Eye, FileText } from 'lucide-react'
import { useEffect, useMemo, useState } from 'react'

import {
  PRELOAD_RUNTIMES,
  getPreloadRuntimeSection,
  parsePreloadConfig,
  setPreloadRuntimeEnabled,
  type PreloadRuntimeId,
} from './preloadConfig'
import styles from './probeConfigPage.module.scss'

type Props = {
  content: string
  onChange: (content: string) => void
  readOnly?: boolean
}

export function PreloadConfigEditor({ content, onChange, readOnly = false }: Props) {
  const [previewOpen, setPreviewOpen] = useState(false)
  const state = useMemo(() => parsePreloadConfig(content), [content])
  const hasIssues = state.issues.length > 0

  useEffect(() => {
    if (hasIssues) {
      console.error('[PreloadConfigEditor] failed to parse injection config', state.issues)
    }
  }, [hasIssues, state.issues])

  const toggleRuntime = (id: PreloadRuntimeId, enabled: boolean) => {
    const result = setPreloadRuntimeEnabled(content, id, enabled)
    if (!result.issues.length) onChange(result.content)
  }

  return (
    <div className={styles.preloadEditor}>
      <div className={styles.preloadHeading}>
        <div>
          <Typography.Text strong>进程注入配置</Typography.Text>
          <Typography.Text type="secondary">
            通过系统 preload 为新启动的运行时写入探针参数，已运行进程不受影响。
          </Typography.Text>
        </div>
      </div>

      <div className={styles.runtimeControlList}>
        {PRELOAD_RUNTIMES.map((runtime) => {
          const section = getPreloadRuntimeSection(content, runtime.id).content
          return (
            <div className={styles.runtimeControlRow} key={runtime.id}>
              <div className={styles.runtimeHeader}>
                <div className={styles.runtimeCopy}>
                  <Typography.Text strong>{runtime.name}</Typography.Text>
                  <Typography.Text type="secondary">
                    {runtime.injection} · {runtime.timing}
                  </Typography.Text>
                </div>
                <Switch
                  aria-label={`${runtime.name} 注入`}
                  checked={state.enabled[runtime.id]}
                  checkedChildren="启用"
                  unCheckedChildren="停用"
                  disabled={readOnly || hasIssues}
                  onChange={(enabled) => toggleRuntime(runtime.id, enabled)}
                />
              </div>
              {state.enabled[runtime.id] && (
                <div className={styles.runtimeConfig}>
                  <span>
                    <FileText size={14} />
                    <Typography.Text type="secondary">Config</Typography.Text>
                  </span>
                  <pre>{section}</pre>
                </div>
              )}
            </div>
          )
        })}
      </div>

      {hasIssues && (
        <Alert
          type="error"
          showIcon
          message="无法可靠解析语言配置"
          description={`${state.issues.join('；')}。原始文件已保留，请修正节点 preload 配置源后重试。`}
        />
      )}

      <div className={styles.preloadActions}>
        <Button icon={<Eye size={15} />} onClick={() => setPreviewOpen(true)}>
          配置预览
        </Button>
      </div>

      <Modal
        title="正式下发配置预览"
        open={previewOpen}
        width={820}
        footer={null}
        onCancel={() => setPreviewOpen(false)}
      >
        <pre className={styles.preloadFinalPreview}>{content}</pre>
      </Modal>
    </div>
  )
}
