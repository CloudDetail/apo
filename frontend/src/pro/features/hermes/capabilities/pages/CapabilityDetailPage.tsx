import { useCallback, useState, useEffect } from 'react'
import {
  Alert,
  Button,
  Drawer,
  Empty,
  message,
  Modal,
  Spin,
  Tag,
  Tooltip,
  Typography,
} from 'antd'
import {
  ArrowLeftOutlined,
  DeleteOutlined,
  EditOutlined,
  InboxOutlined,
  PauseCircleOutlined,
  PlayCircleOutlined,
  PushpinOutlined,
  ReloadOutlined,
  SendOutlined,
  UndoOutlined,
} from '@ant-design/icons'
import { useNavigate, useParams } from 'react-router-dom'
import {
  WorkCapabilityApiError,
  workCapabilityApi,
} from '../client/workCapabilityApi.ts'
import type {
  CapabilityProjection,
  WorkflowRun,
  WorkflowUpdateInput,
  WorkflowVersion,
  WorkCapability,
} from '../client/types.ts'
import {
  blockerText,
  formatCount,
  formatDate,
  RunProjectionPanel,
  RunSummary,
  SectionCard,
  StatusTag,
  StepList,
} from '../components/CapabilityPanels.tsx'
import CapabilityUseDrawer from './CapabilityUseDrawer.tsx'
import CapabilityEditDrawer from './CapabilityEditDrawer.tsx'
import styles from './index.module.scss'

const { Title, Text } = Typography

export interface CapabilityDetailViewProps {
  /** Historical version routes show only metadata because Hermes has no version GET route. */
  forcedVersion?: number
}

export default function CapabilityDetailPage() {
  return <CapabilityDetailView />
}

export function CapabilityDetailView({ forcedVersion }: CapabilityDetailViewProps) {
  const { id = '' } = useParams()
  const navigate = useNavigate()
  const [capability, setCapability] = useState<WorkCapability | null>(null)
  const [versions, setVersions] = useState<WorkflowVersion[]>([])
  const [runs, setRuns] = useState<WorkflowRun[]>([])
  const [loading, setLoading] = useState(true)
  const [detailError, setDetailError] = useState<string | null>(null)
  const [useOpen, setUseOpen] = useState(false)
  const [editOpen, setEditOpen] = useState(false)
  const [editSaving, setEditSaving] = useState(false)
  const [actionBusy, setActionBusy] = useState(false)
  const [projectionRun, setProjectionRun] = useState<WorkflowRun | null>(null)
  const [projection, setProjection] = useState<CapabilityProjection | null>(null)
  const [projectionLoading, setProjectionLoading] = useState(false)
  const [projectionError, setProjectionError] = useState<string | null>(null)
  const historical = forcedVersion !== undefined

  const replaceCapability = useCallback((updated: WorkCapability) => {
    setCapability(updated)
    setVersions(updated.versions)
    setRuns(updated.runs)
  }, [])

  const load = useCallback(async () => {
    if (!id) {
      setDetailError('缺少任务方案 ID。')
      setLoading(false)
      return
    }
    setLoading(true)
    setCapability(null)
    setProjectionRun(null)
    setProjection(null)
    setProjectionError(null)
    setDetailError(null)
    try {
      replaceCapability(await workCapabilityApi.get(id))
    } catch (reason) {
      setCapability(null)
      setVersions([])
      setRuns([])
      setDetailError(errorText(reason, '加载任务方案详情失败。'))
    } finally {
      setLoading(false)
    }
  }, [id, replaceCapability])

  useEffect(() => {
    void load()
  }, [load])

  const openProjection = async (run: WorkflowRun) => {
    setProjectionRun(run)
    setProjection(null)
    setProjectionError(null)
    setProjectionLoading(true)
    try {
      setProjection(await workCapabilityApi.projection(run.runId))
    } catch (reason) {
      setProjectionError(errorText(reason, '运行投影暂不可用。'))
    } finally {
      setProjectionLoading(false)
    }
  }

  const performToggle = () => {
    if (!capability || capability.enabled === null || actionBusy || historical || capability.status === 'archived') return
    const enabling = !capability.enabled
    if (!capability.allowedActions.includes(enabling ? 'enable' : 'disable')) return
    Modal.confirm({
      title: enabling ? '开启自动匹配？' : '关闭自动匹配？',
      content: enabling
        ? '开启后，该 Workflow 可以再次参与匹配并从任务库使用。'
        : '关闭后，该 Workflow 不再参与匹配，也不能手动复用；重新开启后可继续使用。',
      okText: enabling ? '开启自动匹配' : '关闭自动匹配',
      cancelText: '取消',
      okButtonProps: enabling ? undefined : { danger: true },
      onOk: async () => {
        setActionBusy(true)
        try {
          replaceCapability(await workCapabilityApi.setEnabled(capability.id, enabling))
          message.success(enabling ? '已开启自动匹配。' : '已关闭自动匹配。')
        } catch (reason) {
          message.error(errorText(reason, '自动匹配状态更新失败。'))
          throw reason
        } finally {
          setActionBusy(false)
        }
      },
    })
  }

  const performPin = async () => {
    if (!capability || actionBusy || historical || capability.status === 'archived') return
    if (!capability.allowedActions.includes(capability.pinned ? 'unpin' : 'pin')) return
    setActionBusy(true)
    try {
      replaceCapability(await workCapabilityApi.setPinned(capability.id, !capability.pinned))
      message.success(capability.pinned ? '已取消固定。' : '已固定该任务方案。')
    } catch (reason) {
      message.error(errorText(reason, '固定状态更新失败。'))
    } finally {
      setActionBusy(false)
    }
  }

  const performArchive = () => {
    if (!capability || actionBusy || historical) return
    const restoring = capability.status === 'archived'
    if (!capability.allowedActions.includes(restoring ? 'restore' : 'archive')) return
    Modal.confirm({
      title: restoring ? '恢复该任务方案？' : '归档该任务方案？',
      content: restoring
        ? '恢复后会重新开启自动匹配。Hermes 最多保留 20 个 active Workflow，必要时可能自动归档一个未固定方案。'
        : '归档后不会参与匹配，也不能编辑或使用；历史版本和运行记录仍保留，可稍后恢复。',
      okText: restoring ? '恢复任务' : '归档任务',
      cancelText: '取消',
      onOk: async () => {
        setActionBusy(true)
        try {
          replaceCapability(await workCapabilityApi.setArchived(capability.id, !restoring))
          message.success(restoring ? '任务方案已恢复。' : '任务方案已归档。')
        } catch (reason) {
          message.error(errorText(reason, restoring ? '任务方案恢复失败。' : '任务方案归档失败。'))
          throw reason
        } finally {
          setActionBusy(false)
        }
      },
    })
  }

  const performDelete = () => {
    if (!capability || actionBusy || historical) return
    if (!capability.allowedActions.includes('delete')) return
    Modal.confirm({
      title: `确认删除“${capability.name}”？`,
      content: '删除后不会再显示在任务库或参与匹配，且不能恢复。已有 Run、Receipt 和 Evidence 会继续保留；运行中的任务不能删除。',
      okText: '确认删除',
      cancelText: '取消',
      okButtonProps: { danger: true },
      onOk: async () => {
        setActionBusy(true)
        try {
          await workCapabilityApi.delete(capability.id)
          message.success('任务方案已删除；历史运行与执行证据仍保留。')
          navigate('/agent/capabilities', { replace: true })
        } catch (reason) {
          message.error(errorText(reason, '任务方案删除失败。'))
          throw reason
        } finally {
          setActionBusy(false)
        }
      },
    })
  }

  const saveEdit = async (input: WorkflowUpdateInput) => {
    if (!capability || editSaving) return
    if (!capability.allowedActions.includes('edit')) return
    setEditSaving(true)
    try {
      replaceCapability(await workCapabilityApi.update(capability.id, input))
      setEditOpen(false)
      message.success('任务内容已保存为新版本。')
    } catch (reason) {
      message.error(errorText(reason, '任务内容保存失败。'))
      throw reason
    } finally {
      setEditSaving(false)
    }
  }

  if (loading) {
    return <div className={styles.page}><div className={styles.emptyState}><Spin tip="正在读取任务方案…" /></div></div>
  }

  if (!capability) {
    return (
      <div className={styles.page}>
        <div className={styles.shell}>
          <div className={styles.header}>
            <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/agent/capabilities')}>返回任务库</Button>
          </div>
          <Alert type="error" showIcon message="无法读取 Hermes 任务方案" description={detailError || 'Hermes 未返回任务方案。'} action={<Button icon={<ReloadOutlined />} onClick={() => void load()}>重试</Button>} />
        </div>
      </div>
    )
  }

  if (historical) {
    const selectedVersion = versions.find((item) => item.version === forcedVersion)
    return (
      <div className={styles.page}>
        <div className={styles.shell}>
          <div className={styles.header}>
            <div className={styles.headerTitle}>
              <Button type="text" icon={<ArrowLeftOutlined />} onClick={() => navigate(`/agent/capabilities/${encodeURIComponent(capability.id)}`)}>返回任务详情</Button>
              <Title level={2}>{capability.name} · v{forcedVersion}</Title>
            </div>
          </div>
          {!selectedVersion ? (
            <Alert type="error" showIcon message="版本不存在" description={`Hermes 未返回版本 v${forcedVersion} 的版本记录。`} />
          ) : (
            <>
              <Alert
                className={styles.historyBanner}
                type="info"
                showIcon
                message="Hermes 当前只提供历史版本摘要"
                description="历史版本的完整目标和步骤没有读取接口，因此本页只展示真实版本元数据，不使用当前版本内容代替。"
              />
              <SectionCard title={`版本 v${selectedVersion.version}`}>
                <dl className={styles.facts}>
                  <div><dt>版本状态</dt><dd>{selectedVersion.version === capability.currentVersion ? '当前版本' : '历史版本'}</dd></div>
                  <div><dt>创建时间</dt><dd>{formatDate(selectedVersion.createdAt)}</dd></div>
                  <div><dt>形成方式</dt><dd>{selectedVersion.createdBy || 'Hermes 未返回'}</dd></div>
                  <div><dt>来源运行</dt><dd>{selectedVersion.sourceRunId || 'Hermes 未返回'}</dd></div>
                  <div><dt>内容摘要</dt><dd><Text code>{selectedVersion.digest || 'Hermes 未返回'}</Text></dd></div>
                </dl>
              </SectionCard>
            </>
          )}
        </div>
      </div>
    )
  }

  const archived = capability.status === 'archived'
  const canUse = capability.allowedActions.includes('use')
  const canEdit = capability.allowedActions.includes('edit')
  const canArchive = capability.allowedActions.includes(archived ? 'restore' : 'archive')
  const canDelete = capability.allowedActions.includes('delete')
  const useBlocker = capability.actionBlockers.find((item) => typeof item !== 'string' && item.action === 'use')
  const useBlockerReason = useBlocker
    ? blockerText(useBlocker)
    : archived
      ? '请先恢复该任务'
      : capability.status === 'disabled'
        ? '请先开启自动匹配'
        : 'Hermes 未返回可使用状态'
  const sourceRun = capability.sourceRun

  return (
    <div className={styles.page}>
      <div className={styles.shell}>
        <div className={styles.detailHeader}>
          <div>
            <Button type="text" icon={<ArrowLeftOutlined />} onClick={() => navigate('/agent/capabilities')}>返回任务库</Button>
            <div className={styles.detailMeta}>
              <StatusTag value={capability.status} />
              {capability.pinned ? <Tag color="processing">已固定</Tag> : null}
              <span>当前版本 v{capability.currentVersion}</span>
              <span>最近更新 {formatDate(capability.updatedAt)}</span>
            </div>
            <Title level={1}>{capability.name}</Title>
          </div>
          <div className={styles.headerActions}>
            <Tooltip title={archived ? '归档任务不能编辑' : undefined}>
              <span><Button icon={<EditOutlined />} disabled={actionBusy || !canEdit} onClick={() => setEditOpen(true)}>编辑任务</Button></span>
            </Tooltip>
            <Tooltip title={canUse ? undefined : useBlockerReason}>
              <span><Button type="primary" icon={<SendOutlined />} disabled={actionBusy || !canUse} onClick={() => setUseOpen(true)}>使用该任务</Button></span>
            </Tooltip>
            {!archived && capability.enabled !== null ? (
              <Button
                danger={Boolean(capability.enabled)}
                icon={capability.enabled ? <PauseCircleOutlined /> : <PlayCircleOutlined />}
                loading={actionBusy}
                onClick={performToggle}
              >
                {capability.enabled ? '关闭自动匹配' : '开启自动匹配'}
              </Button>
            ) : null}
            {!archived ? (
              <Button icon={<PushpinOutlined />} disabled={actionBusy} onClick={() => void performPin()}>
                {capability.pinned ? '取消固定' : '固定'}
              </Button>
            ) : null}
            <Button
              icon={archived ? <UndoOutlined /> : <InboxOutlined />}
              disabled={actionBusy || !canArchive}
              onClick={performArchive}
            >
              {archived ? '恢复' : '归档'}
            </Button>
            <Button danger icon={<DeleteOutlined />} disabled={actionBusy || !canDelete} onClick={performDelete}>删除</Button>
          </div>
        </div>

        {detailError ? <Alert className={styles.notice} type="warning" showIcon message="详情数据不完整" description={detailError} /> : null}

        <div className={styles.detailLayout}>
          <main className={styles.mainColumn}>
            <SectionCard title="任务目标">
              <p className={styles.objective}>{capability.goal || 'Hermes 未返回任务目标。'}</p>
            </SectionCard>

            <SectionCard title="典型提问">
              {capability.examples.length ? (
                <div className={styles.chipList}>{capability.examples.map((example) => <span className={styles.chip} key={example}>{example}</span>)}</div>
              ) : <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="Hermes 未返回典型提问" />}
            </SectionCard>

            <SectionCard title="输入提示">
              {capability.inputs.length ? (
                <div className={styles.chipList}>{capability.inputs.map((input) => <span className={styles.chip} key={input.name}>{input.name}{input.required ? ' · 必填' : ''}</span>)}</div>
              ) : <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="Hermes 未声明输入项" />}
            </SectionCard>

            <SectionCard title="稳定步骤">
              <StepList steps={capability.steps} />
            </SectionCard>

            {capability.notes.length ? (
              <SectionCard title="经验和注意事项">
                <ul className={styles.noteList}>{capability.notes.map((note) => <li key={note}>{note}</li>)}</ul>
              </SectionCard>
            ) : null}

            <SectionCard title="最近运行" extra={<Tag>返回 {runs.length} 条</Tag>}>
              {runs.length ? <div>{runs.map((run) => <RunSummary key={run.runId} run={run} onOpen={(item) => void openProjection(item)} />)}</div> : <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无运行记录" />}
            </SectionCard>
          </main>

          <aside className={styles.asideColumn}>
            <SectionCard title="任务状态">
              <dl className={styles.facts}>
                <div><dt>自动匹配</dt><dd><StatusTag value={capability.status} /></dd></div>
                <div><dt>固定状态</dt><dd>{capability.pinned ? '已固定' : '未固定'}</dd></div>
                <div><dt>当前版本</dt><dd>v{capability.currentVersion}</dd></div>
                <div><dt>复用次数</dt><dd>{formatCount(capability.reuseCount)}</dd></div>
                <div><dt>成功次数</dt><dd>{formatCount(capability.successCount)}</dd></div>
                <div><dt>近期异常</dt><dd>{formatCount(capability.recentFailures)}</dd></div>
                <div><dt>最近使用</dt><dd>{formatDate(capability.lastUsedAt)}</dd></div>
                <div><dt>形成时间</dt><dd>{formatDate(capability.createdAt)}</dd></div>
              </dl>
            </SectionCard>

            <SectionCard title="预期结果">
              <p className={styles.objective}>{capability.expectedResult || 'Hermes 未返回预期结果。'}</p>
            </SectionCard>

            <SectionCard title="来源运行">
              {sourceRun ? (
                <div className={styles.runMain}>
                  <strong>{sourceRun.title || sourceRun.runId}</strong>
                  <Text type="secondary">运行 ID：{sourceRun.runId}</Text>
                  <Text type="secondary">{formatDate(sourceRun.createdAt || sourceRun.updatedAt)}</Text>
                  {sourceRun.status ? <StatusTag value={sourceRun.status} /> : null}
                  <Button type="link" size="small" onClick={() => void openProjection(sourceRun)}>查看来源运行</Button>
                </div>
              ) : <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="Hermes 未返回来源运行" />}
            </SectionCard>

            <SectionCard title="版本历史">
              {versions.length ? <div className={styles.versionList}>{versions.map((version) => <div className={styles.versionItem} key={version.version}><div className={styles.versionMeta}><strong>v{version.version} {version.isCurrent ? <Tag color="processing">当前</Tag> : null}</strong><Text type="secondary">{formatDate(version.createdAt)}{version.createdBy ? ` · ${version.createdBy}` : ''}</Text></div><Button type="link" size="small" onClick={() => navigate(`/agent/capabilities/${encodeURIComponent(capability.id)}/versions/${version.version}`)}>查看</Button></div>)}</div> : <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="Hermes 未返回版本历史" />}
            </SectionCard>
          </aside>
        </div>
      </div>

      <CapabilityUseDrawer
        open={useOpen}
        capability={capability}
        onClose={() => setUseOpen(false)}
        onNavigate={() => setUseOpen(false)}
      />

      <CapabilityEditDrawer
        open={editOpen}
        capability={capability}
        saving={editSaving}
        onClose={() => setEditOpen(false)}
        onSave={saveEdit}
      />

      <Drawer
        open={Boolean(projectionRun)}
        width={640}
        title={projectionRun ? `运行详情 · ${projectionRun.runId}` : '运行详情'}
        onClose={() => { setProjectionRun(null); setProjection(null); setProjectionError(null) }}
      >
        {projectionLoading ? <div className={styles.emptyState}><Spin tip="正在读取运行投影…" /></div> : projectionError ? <Alert type="warning" showIcon message="运行投影暂不可用" description={projectionError} /> : <RunProjectionPanel projection={projection} />}
      </Drawer>
    </div>
  )
}

function errorText(reason: unknown, fallback: string): string {
  if (reason instanceof WorkCapabilityApiError) return reason.message
  return reason instanceof Error ? reason.message : fallback
}
