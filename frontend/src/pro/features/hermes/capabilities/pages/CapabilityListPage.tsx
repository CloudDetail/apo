import { useCallback, useEffect, useMemo, useState } from 'react'
import { Alert, Button, Empty, Input, message, Modal, Select, Space, Table, Tag, Tooltip, Typography } from 'antd'
import { ArrowLeftOutlined, DeleteOutlined, EyeOutlined, ReloadOutlined, SearchOutlined } from '@ant-design/icons'
import { useNavigate, useSearchParams } from 'react-router-dom'
import type { CapabilityListParams, WorkCapability } from '../client/types.ts'
import { StatusTag, formatCount, formatDate } from '../components/CapabilityPanels.tsx'
import {
  useWorkCapabilityApi,
  useWorkCapabilityUserScopeKey,
} from '../useWorkCapabilityApi.ts'
import styles from './index.module.scss'

const { Title, Text } = Typography

export default function CapabilityListPage() {
  const userScopeKey = useWorkCapabilityUserScopeKey()
  return <CapabilityListView key={userScopeKey ? `user:${userScopeKey}` : 'pending-user'} />
}

function CapabilityListView() {
  const workCapabilityApi = useWorkCapabilityApi()
  const navigate = useNavigate()
  const [searchParams, setSearchParams] = useSearchParams()
  const [searchInput, setSearchInput] = useState(searchParams.get('search') || '')
  const [items, setItems] = useState<WorkCapability[]>([])
  const [selectedIds, setSelectedIds] = useState<string[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)
  const [deleting, setDeleting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const page = positiveInteger(searchParams.get('page'), 1)
  const pageSize = positiveInteger(searchParams.get('pageSize'), 20)
  const params = useMemo<CapabilityListParams>(() => ({
    page,
    pageSize,
    search: searchParams.get('search') || undefined,
    status: searchParams.get('status') || undefined,
    recentRun: searchParams.get('recentRun') || undefined,
  }), [page, pageSize, searchParams])

  const load = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const result = await workCapabilityApi.list(params)
      setItems(result.items)
      setTotal(result.total)
      setSelectedIds((current) => current.filter((id) => result.items.some((item) => item.id === id)))
    } catch (reason) {
      setItems([])
      setTotal(0)
      setSelectedIds([])
      setError(reason instanceof Error ? reason.message : '加载运维任务库失败。')
    } finally {
      setLoading(false)
    }
  }, [params, workCapabilityApi])

  useEffect(() => {
    void load()
  }, [load])

  useEffect(() => {
    setSearchInput(searchParams.get('search') || '')
  }, [searchParams])

  const updateParam = (key: string, value?: string) => {
    const next = new URLSearchParams(searchParams)
    if (value) next.set(key, value)
    else next.delete(key)
    if (key !== 'page' && key !== 'pageSize') next.set('page', '1')
    setSearchParams(next, { replace: true })
  }

  const submitSearch = () => updateParam('search', searchInput.trim())

  const requestDelete = (targets: WorkCapability[]) => {
    if (!targets.length || deleting) return
    Modal.confirm({
      title: `确认删除 ${targets.length} 个任务方案？`,
      content: (
        <div>
          <p>删除后不会再显示在任务库或参与匹配，且不能恢复。已有 Run、Receipt 和 Evidence 会继续保留。</p>
          <ul>{targets.slice(0, 5).map((item) => <li key={item.id}>{item.name}</li>)}</ul>
          {targets.length > 5 ? <p>另有 {targets.length - 5} 个已选任务方案。</p> : null}
        </div>
      ),
      okText: '确认删除',
      cancelText: '取消',
      okButtonProps: { danger: true, loading: deleting },
      onOk: async () => {
        setDeleting(true)
        try {
          const result = targets.length === 1
            ? await workCapabilityApi.delete(targets[0].id)
            : await workCapabilityApi.deleteMany(targets.map((item) => item.id))
          const deleted = new Set(result.deletedIds)
          setItems((current) => current.filter((item) => !deleted.has(item.id)))
          setSelectedIds((current) => current.filter((id) => !deleted.has(id)))
          setTotal((current) => Math.max(0, current - result.deletedCount))
          message.success(`已删除 ${result.deletedCount} 个任务方案；历史运行与执行证据仍保留。`)
        } catch (reason) {
          message.error(reason instanceof Error ? reason.message : '任务方案删除失败。')
          throw reason
        } finally {
          setDeleting(false)
        }
      },
    })
  }

  const columns = [
    {
      title: '任务方案',
      key: 'name',
      render: (_value: unknown, item: WorkCapability) => (
        <div className={styles.nameCell}>
          <Space size={6} wrap>
            <strong>{item.name}</strong>
            {item.pinned ? <Tag color="processing">已固定</Tag> : null}
          </Space>
          <small>{item.goal || 'Agent 未返回任务目标'}</small>
          <Text type="secondary">当前版本 v{item.currentVersion}</Text>
        </div>
      ),
    },
    {
      title: '状态',
      key: 'status',
      width: 120,
      render: (_value: unknown, item: WorkCapability) => <StatusTag value={item.status} />,
    },
    {
      title: '形成时间',
      key: 'createdAt',
      width: 170,
      render: (_value: unknown, item: WorkCapability) => formatDate(item.createdAt),
    },
    {
      title: '复用效果',
      key: 'runs',
      width: 250,
      render: (_value: unknown, item: WorkCapability) => (
        <div className={styles.tagStack}>
          <Tag>已复用 {formatCount(item.reuseCount)} 次</Tag>
          <Tag color="success">成功 {formatCount(item.successCount)} 次</Tag>
          {item.recentFailures ? <Tag color="error">近期异常 {formatCount(item.recentFailures)} 次</Tag> : null}
        </div>
      ),
    },
    {
      title: '最近使用',
      key: 'lastRun',
      width: 180,
      render: (_value: unknown, item: WorkCapability) => item.lastRun ? (
        <Space direction="vertical" size={0}>
          <StatusTag value={item.lastRun.status} />
          <Text type="secondary">{formatDate(item.lastUsedAt || item.lastRun.createdAt)}</Text>
        </Space>
      ) : <Text type="secondary">暂无使用记录</Text>,
    },
    {
      title: '操作',
      key: 'actions',
      width: 104,
      fixed: 'right' as const,
      render: (_value: unknown, item: WorkCapability) => (
        <Space size={2}>
          <Tooltip title="查看详情">
            <Button
              type="text"
              aria-label={`查看 ${item.name}`}
              icon={<EyeOutlined />}
              onClick={() => navigate(`/agent/capabilities/${encodeURIComponent(item.id)}`)}
            />
          </Tooltip>
          <Tooltip title="删除任务">
            <Button
              type="text"
              danger
              aria-label={`删除 ${item.name}`}
              icon={<DeleteOutlined />}
              disabled={deleting}
              onClick={() => requestDelete([item])}
            />
          </Tooltip>
        </Space>
      ),
    },
  ]

  const selectedItems = items.filter((item) => selectedIds.includes(item.id))

  return (
    <div className={styles.page}>
      <div className={styles.shell}>
        <div className={styles.header}>
          <div className={styles.headerTitle}>
            <Button type="text" icon={<ArrowLeftOutlined />} onClick={() => navigate('/agent')}>返回运维会话</Button>
            <Title level={2}>运维任务库</Title>
            <p>查看由成功运行自动沉淀的可复用任务方案。</p>
          </div>
          <Space wrap>
            <Tooltip title="重新加载任务库"><Button aria-label="重新加载任务库" icon={<ReloadOutlined />} onClick={() => void load()} /></Tooltip>
          </Space>
        </div>

        <div className={styles.toolbar}>
          <Input
            className={styles.search}
            allowClear
            prefix={<SearchOutlined />}
            placeholder="搜索任务名称、目标、提示或步骤"
            value={searchInput}
            onChange={(event) => setSearchInput(event.target.value)}
            onPressEnter={submitSearch}
          />
          <Select
            className={styles.filter}
            allowClear
            placeholder="任务状态"
            value={searchParams.get('status') || undefined}
            onChange={(value) => updateParam('status', value)}
            options={[
              { value: 'active', label: '自动匹配开启' },
              { value: 'disabled', label: '自动匹配关闭' },
              { value: 'archived', label: '已归档' },
            ]}
          />
          <Select
            className={styles.filter}
            allowClear
            placeholder="最近使用"
            value={searchParams.get('recentRun') || undefined}
            onChange={(value) => updateParam('recentRun', value)}
            options={[
              { value: '7d', label: '近 7 天使用' },
              { value: '30d', label: '近 30 天使用' },
              { value: 'never', label: '从未使用' },
            ]}
          />
          <Button
            danger
            icon={<DeleteOutlined />}
            disabled={!selectedItems.length || deleting}
            loading={deleting}
            onClick={() => requestDelete(selectedItems)}
          >
            删除选中{selectedItems.length ? ` (${selectedItems.length})` : ''}
          </Button>
        </div>

        {error ? <Alert className={styles.notice} type="error" showIcon message="无法读取 Agent 运维任务库" description={error} /> : null}

        <div className={styles.tableCard}>
          <Table
            rowKey={(item) => item.id}
            loading={loading}
            columns={columns}
            dataSource={items}
            rowSelection={{
              selectedRowKeys: selectedIds,
              onChange: (keys) => setSelectedIds(keys.map(String)),
            }}
            scroll={{ x: 1120 }}
            locale={{ emptyText: <Empty className={styles.emptyState} image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无符合条件的任务方案" /> }}
            pagination={{
              current: page,
              pageSize,
              total,
              showSizeChanger: true,
              onChange: (nextPage, nextPageSize) => {
                const next = new URLSearchParams(searchParams)
                next.set('page', String(nextPage))
                next.set('pageSize', String(nextPageSize))
                setSearchParams(next, { replace: true })
              },
            }}
            onRow={(item) => ({ onDoubleClick: () => navigate(`/agent/capabilities/${encodeURIComponent(item.id)}`) })}
          />
        </div>
      </div>
    </div>
  )
}

function positiveInteger(value: string | null, fallback: number): number {
  const candidate = Number(value)
  return Number.isInteger(candidate) && candidate > 0 ? candidate : fallback
}
