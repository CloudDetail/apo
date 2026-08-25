import { Button, Card, Drawer, Input, Select, Space, Switch, Table, Tag, Typography } from 'antd'
import { SettingOutlined } from '@ant-design/icons'
import {
  Activity,
  Cpu,
  Download,
  FileText,
  MemoryStick,
  Network,
  Radio,
  RefreshCw,
  Server,
  Settings,
} from 'lucide-react'
import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  createAgentCommandApi,
  getAgentCommandsApi,
  getAgentConfigNodesApi,
  getAgentConfigStatesApi,
} from 'src/core/api/integration'
import { notify } from 'src/core/utils/notify'
import styles from './probeManagement.module.scss'

type AgentNode = {
  id: string
  name?: string
  clusterId?: string
  environment?: string
  labels?: Record<string, string>
  lastSeenAt?: string
}

type ProbeRuntime = {
  childCpuPercent?: number
  childRssBytes?: number
  guardRssBytes?: number
  childRunning?: boolean
  supervisorHealthy?: boolean
  resourceUsage?: {
    windowSeconds: number
    sampleIntervalSeconds: number
    sampleCount: number
    cpuPercentAvg?: number
    memoryBytesAvg?: number
    sampledAt: string
  }
}

type ProbeState = {
  nodeId: string
  probeType: string
  status: string
  runtime?: ProbeRuntime
}

const getAveragedResourceUsage = (state: ProbeState) => {
  const usage = state.runtime?.resourceUsage
  return usage?.windowSeconds === 600 && usage.sampleIntervalSeconds === 60 ? usage : undefined
}

type NodeRow = {
  kind: 'node'
  key: string
  clusterId: string
  nodeId: string
  name: string
  environment: string
  states: ProbeState[]
  cpuPercent?: number
  memoryBytes?: number
  healthyCount: number
  probeCount: number
}

type ClusterRow = {
  kind: 'cluster'
  key: string
  clusterId: string
  name: string
  environment: string
  children: NodeRow[]
  cpuPercent?: number
  memoryBytes?: number
  healthyCount: number
  probeCount: number
}

type TableRow = ClusterRow | NodeRow
type LogFile = { path: string; name: string; size: number; modifiedAt: string }
type LogPreview = { file: LogFile; preview: string; lines: number; truncated: boolean }

const UNASSIGNED_CLUSTER = '未分配集群'
const STATUS_REFRESH_INTERVAL_MS = 60_000

const getClusterId = (node: AgentNode) =>
  String(node.clusterId || node.labels?.cluster_id || node.labels?.clusterId || '').trim() ||
  UNASSIGNED_CLUSTER

const isHealthy = (state: ProbeState) =>
  state.status === 'RUNNING' &&
  state.runtime?.childRunning !== false &&
  state.runtime?.supervisorHealthy !== false

const summarizeStates = (states: ProbeState[]) => {
  const usage = states.map(getAveragedResourceUsage)
  const cpuAvailable = usage.length > 0 && usage.every((item) => item?.cpuPercentAvg !== undefined)
  const memoryAvailable =
    usage.length > 0 && usage.every((item) => item?.memoryBytesAvg !== undefined)
  return {
    cpuPercent: cpuAvailable
      ? usage.reduce((total, item) => total + Number(item?.cpuPercentAvg), 0)
      : undefined,
    memoryBytes: memoryAvailable
      ? usage.reduce((total, item) => total + Number(item?.memoryBytesAvg), 0)
      : undefined,
    healthyCount: states.filter(isHealthy).length,
    probeCount: states.length,
  }
}

const groupNodesByCluster = (nodes: AgentNode[], states: ProbeState[]): ClusterRow[] => {
  const statesByNode = new Map<string, ProbeState[]>()
  states.forEach((state) => {
    const current = statesByNode.get(state.nodeId) || []
    current.push(state)
    statesByNode.set(state.nodeId, current)
  })

  const knownNodeIds = new Set(nodes.map((node) => node.id))
  const orphanNodes = [...statesByNode.keys()]
    .filter((nodeId) => !knownNodeIds.has(nodeId))
    .map((nodeId) => ({ id: nodeId, name: nodeId }))
  if (orphanNodes.length) {
    console.warn('[ProbeManagement] probe states reported before node registration', orphanNodes)
  }

  const grouped = new Map<string, NodeRow[]>()
  ;[...nodes, ...orphanNodes].forEach((node) => {
    const clusterId = getClusterId(node)
    const nodeStates = (statesByNode.get(node.id) || []).sort((a, b) =>
      a.probeType.localeCompare(b.probeType),
    )
    const row: NodeRow = {
      kind: 'node',
      key: `node:${node.id}`,
      clusterId,
      nodeId: node.id,
      name: node.name || node.id,
      environment: node.environment || '',
      states: nodeStates,
      ...summarizeStates(nodeStates),
    }
    grouped.set(clusterId, [...(grouped.get(clusterId) || []), row])
  })

  return [...grouped.entries()]
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([clusterId, children]) => {
      children.sort((a, b) => a.name.localeCompare(b.name))
      const cpuAvailable =
        children.some((node) => node.probeCount > 0) &&
        children.every((node) => !node.probeCount || node.cpuPercent !== undefined)
      const memoryAvailable =
        children.some((node) => node.probeCount > 0) &&
        children.every((node) => !node.probeCount || node.memoryBytes !== undefined)
      const summary = {
        cpuPercent: cpuAvailable
          ? children.reduce((total, node) => total + Number(node.cpuPercent), 0)
          : undefined,
        memoryBytes: memoryAvailable
          ? children.reduce((total, node) => total + Number(node.memoryBytes), 0)
          : undefined,
        healthyCount: children.reduce((total, node) => total + node.healthyCount, 0),
        probeCount: children.reduce((total, node) => total + node.probeCount, 0),
      }
      return {
        kind: 'cluster',
        key: `cluster:${clusterId}`,
        clusterId,
        name: clusterId,
        environment:
          [...new Set(children.map((node) => node.environment).filter(Boolean))][0] || '',
        children,
        ...summary,
      }
    })
}

const formatMemory = (bytes?: number) => {
  if (bytes === undefined) return '-'
  if (!bytes) return '0 MiB'
  if (bytes >= 1024 ** 3) return `${(bytes / 1024 ** 3).toFixed(1)} GiB`
  return `${Math.round(bytes / 1024 ** 2)} MiB`
}

const formatCPU = (value?: number) => {
  if (value === undefined) return '-'
  return value ? `${value.toFixed(value >= 10 ? 0 : 1)}%` : '0%'
}

export default function ProbeManagement() {
  const navigate = useNavigate()
  const [nodes, setNodes] = useState<AgentNode[]>([])
  const [states, setStates] = useState<ProbeState[]>([])
  const [loading, setLoading] = useState(true)
  const [expandedRowKeys, setExpandedRowKeys] = useState<React.Key[]>([])
  const [logNodes, setLogNodes] = useState<NodeRow[]>([])
  const [selectedLogNodeId, setSelectedLogNodeId] = useState('')
  const [selectedLogProbe, setSelectedLogProbe] = useState('')
  const [logFiles, setLogFiles] = useState<LogFile[]>([])
  const [selectedLogPath, setSelectedLogPath] = useState('')
  const [logPreview, setLogPreview] = useState<LogPreview>()
  const [logLoading, setLogLoading] = useState(false)
  const [autoRefresh, setAutoRefresh] = useState(true)
  const logRequestActive = useRef(false)

  const load = useCallback(async () => {
    try {
      const [nodeResult, stateResult] = await Promise.all([
        getAgentConfigNodesApi(),
        getAgentConfigStatesApi(),
      ])
      setNodes(Array.isArray((nodeResult as any)?.nodes) ? (nodeResult as any).nodes : [])
      const loadedStates = Array.isArray((stateResult as any)?.states)
        ? (stateResult as any).states
        : []
      const missingAverage = loadedStates.filter(
        (state: ProbeState) => state.runtime && !getAveragedResourceUsage(state),
      )
      if (missingAverage.length) {
        console.warn(
          '[ProbeManagement] 10-minute resource average is unavailable for reported probes',
          missingAverage.map((state: ProbeState) => ({
            nodeId: state.nodeId,
            probeType: state.probeType,
          })),
        )
      }
      setStates(loadedStates)
    } catch (error) {
      console.error('[ProbeManagement] failed to load registered probes', error)
      setNodes([])
      setStates([])
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    load()
    const timer = window.setInterval(load, STATUS_REFRESH_INTERVAL_MS)
    return () => window.clearInterval(timer)
  }, [load])

  const clusters = useMemo(() => groupNodesByCluster(nodes, states), [nodes, states])
  const clusterKeys = useMemo(() => clusters.map((cluster) => cluster.key), [clusters])

  useEffect(() => {
    setExpandedRowKeys((current) =>
      current.length ? current.filter((key) => clusterKeys.includes(String(key))) : clusterKeys,
    )
  }, [clusterKeys])

  const nodeCount = clusters.reduce((total, cluster) => total + cluster.children.length, 0)
  const selectedLogNode = logNodes.find((node) => node.nodeId === selectedLogNodeId)
  const logProbeOptions = selectedLogNode?.states.map((state) => state.probeType) || []

  useEffect(() => {
    setSelectedLogProbe((current) =>
      logProbeOptions.includes(current) ? current : logProbeOptions[0] || '',
    )
  }, [selectedLogNodeId, logProbeOptions.join('\u0000')])

  const waitForCommand = async (nodeId: string, probeType: string, commandId: string) => {
    for (let attempt = 0; attempt < 30; attempt += 1) {
      const response: any = await getAgentCommandsApi(nodeId, probeType)
      const command = (response?.commands || []).find((item) => item.id === commandId)
      if (command?.status === 'SUCCEEDED' || command?.status === 'FAILED') return command
      await new Promise((resolve) => window.setTimeout(resolve, 1000))
    }
    throw new Error('等待节点返回日志结果超时')
  }

  const runLogCommand = async (
    type: string,
    options: Record<string, unknown> = {},
    nodeId = selectedLogNodeId,
    probeType = selectedLogProbe,
  ) => {
    const created: any = await createAgentCommandApi(nodeId, probeType, type, options)
    const command = await waitForCommand(nodeId, probeType, created.id)
    const result = command.result ? JSON.parse(command.result) : {}
    if (command.status === 'FAILED') {
      throw new Error(result?.error?.message || command.result || '读取日志失败')
    }
    return result
  }

  const refreshLogFiles = async () => {
    if (!selectedLogNodeId || !selectedLogProbe) return
    setLogLoading(true)
    try {
      const result = await runLogCommand('LIST_LOG_FILES')
      const files = (result.files || []) as LogFile[]
      setLogFiles(files)
      setSelectedLogPath((current) =>
        files.some((file) => file.path === current) ? current : files[0]?.path || '',
      )
      if (result.errors?.length) {
        notify({ type: 'warning', message: result.errors[0].message })
      }
    } catch (error: any) {
      notify({ type: 'error', message: error.message || '读取日志文件列表失败' })
    } finally {
      setLogLoading(false)
    }
  }

  const refreshLogPreview = async (quiet = false) => {
    if (!selectedLogPath || logRequestActive.current) return
    logRequestActive.current = true
    if (!quiet) setLogLoading(true)
    try {
      setLogPreview(
        await runLogCommand('PREVIEW_LOG_FILE', { filePath: selectedLogPath, tailLines: 100 }),
      )
    } catch (error: any) {
      notify({ type: 'error', message: error.message || '读取实时日志失败' })
    } finally {
      logRequestActive.current = false
      if (!quiet) setLogLoading(false)
    }
  }

  useEffect(() => {
    if (!logNodes.length || !selectedLogNodeId || !selectedLogProbe) return
    setLogFiles([])
    setSelectedLogPath('')
    setLogPreview(undefined)
    refreshLogFiles()
  }, [selectedLogNodeId, selectedLogProbe, logNodes.length])

  useEffect(() => {
    if (!logNodes.length || !selectedLogPath) return
    refreshLogPreview()
  }, [selectedLogPath])

  useEffect(() => {
    if (!logNodes.length || !selectedLogPath || !autoRefresh) return
    const timer = window.setInterval(() => refreshLogPreview(true), 5000)
    return () => window.clearInterval(timer)
  }, [logNodes.length, selectedLogNodeId, selectedLogProbe, selectedLogPath, autoRefresh])

  const openRealtimeLogs = (targetNodes: NodeRow[]) => {
    const available = targetNodes.filter((node) => node.states.length)
    setLogNodes(available)
    setSelectedLogNodeId(available[0]?.nodeId || '')
    setSelectedLogProbe(available[0]?.states[0]?.probeType || '')
    setAutoRefresh(true)
  }

  const downloadLog = async () => {
    if (!selectedLogPath) return
    setLogLoading(true)
    try {
      const result = await runLogCommand('DOWNLOAD_LOG_FILE', { filePath: selectedLogPath })
      const binary = window.atob(result.content)
      const bytes = Uint8Array.from(binary, (char) => char.charCodeAt(0))
      const url = URL.createObjectURL(new Blob([bytes]))
      const anchor = document.createElement('a')
      anchor.href = url
      anchor.download = result.file?.name || selectedLogPath.split('/').pop() || 'probe.log'
      anchor.click()
      URL.revokeObjectURL(url)
    } catch (error: any) {
      notify({ type: 'error', message: error.message || '下载日志失败' })
    } finally {
      setLogLoading(false)
    }
  }

  const columns = [
    {
      title: '集群 / 节点',
      key: 'name',
      width: '30%',
      render: (_: unknown, record: TableRow) => (
        <div className={styles.identityCell}>
          <span className={record.kind === 'cluster' ? styles.clusterIcon : styles.nodeIcon}>
            {record.kind === 'cluster' ? <Network size={16} /> : <Server size={15} />}
          </span>
          <div>
            <Typography.Text strong>{record.name}</Typography.Text>
            <span className={styles.secondaryLine}>
              {record.kind === 'cluster'
                ? `${record.children.length} 个节点 · ${record.probeCount} 个探针`
                : `Node ID: ${record.nodeId}`}
            </span>
          </div>
        </div>
      ),
    },
    {
      title: '探针状态',
      key: 'probeStatus',
      width: '34%',
      render: (_: unknown, record: TableRow) => {
        if (record.kind === 'cluster') {
          const healthy = record.probeCount > 0 && record.healthyCount === record.probeCount
          return (
            <Tag color={healthy ? 'success' : record.probeCount ? 'warning' : 'default'}>
              {record.probeCount
                ? `${record.healthyCount} / ${record.probeCount} 运行正常`
                : '暂无探针上报'}
            </Tag>
          )
        }
        if (!record.states.length) {
          return <span className={styles.resourceEmpty}>等待探针上报</span>
        }
        return (
          <div className={styles.probeTags}>
            {record.states.map((state) => (
              <Tag key={state.probeType} color={isHealthy(state) ? 'success' : 'warning'}>
                {state.probeType}
              </Tag>
            ))}
          </div>
        )
      },
    },
    {
      title: '探针资源开销（近 10 分钟均值）',
      key: 'resourceUsage',
      width: '24%',
      render: (_: unknown, record: TableRow) => (
        <div className={styles.metrics}>
          <div className={styles.metric}>
            <span className={`${styles.metricIcon} ${styles.cpuIcon}`}>
              <Cpu size={15} />
            </span>
            <span>CPU</span>
            <strong>{formatCPU(record.cpuPercent)}</strong>
          </div>
          <div className={styles.metric}>
            <span className={`${styles.metricIcon} ${styles.memoryIcon}`}>
              <MemoryStick size={15} />
            </span>
            <span>内存</span>
            <strong>{formatMemory(record.memoryBytes)}</strong>
          </div>
        </div>
      ),
    },
    {
      title: '操作',
      key: 'operation',
      width: '22%',
      render: (_: unknown, record: TableRow) =>
        record.kind === 'cluster' ? (
          <Space size={0}>
            <Button
              type="link"
              icon={<SettingOutlined />}
              onClick={() =>
                navigate(
                  `/probe-management/config?clusterId=${encodeURIComponent(record.clusterId)}&clusterType=${encodeURIComponent(record.environment)}`,
                )
              }
            >
              集群配置
            </Button>
          </Space>
        ) : (
          <Space size={0}>
            <Button
              type="link"
              icon={<SettingOutlined />}
              disabled={!record.states.length}
              onClick={() =>
                navigate(
                  `/probe-management/config?clusterId=${encodeURIComponent(record.clusterId)}&nodeId=${encodeURIComponent(record.nodeId)}&clusterType=${encodeURIComponent(record.environment)}`,
                )
              }
            >
              节点配置
            </Button>
            <Button
              type="link"
              icon={<Radio size={15} />}
              disabled={!record.states.length}
              onClick={() => openRealtimeLogs([record])}
            >
              实时日志
            </Button>
          </Space>
        ),
    },
  ]

  return (
    <div className={styles.managementPage}>
      <div className={styles.sectionHeader}>
        <div>
          <Typography.Title level={4}>探针管理</Typography.Title>
          <Typography.Text type="secondary">按集群查看节点及其探针运行状态</Typography.Text>
        </div>
        <Space className={styles.headerActions} wrap>
          <Button
            icon={<Settings size={15} />}
            onClick={() => navigate('/probe-management/config?scope=global')}
          >
            默认配置
          </Button>
          <Button
            type="primary"
            icon={<Network size={15} />}
            onClick={() => navigate('/probe-management/config')}
          >
            批量配置
          </Button>
          <Tag color="processing">
            {clusters.length} 个集群 · {nodeCount} 个节点
          </Tag>
        </Space>
      </div>
      <Card className={styles.tableCard} styles={{ body: { padding: 0 } }}>
        <Table
          columns={columns}
          dataSource={clusters}
          rowKey="key"
          loading={loading}
          pagination={false}
          size="small"
          scroll={{ x: 860 }}
          rowClassName={(record) =>
            record.kind === 'cluster' ? styles.clusterRow : styles.nodeRow
          }
          locale={{
            emptyText: (
              <span className={styles.resourceEmpty}>
                <Activity size={15} />
                暂无已注册探针
              </span>
            ),
          }}
          expandable={{
            expandedRowKeys,
            onExpandedRowsChange: (keys) => setExpandedRowKeys([...keys]),
            rowExpandable: (record) => record.kind === 'cluster' && record.children.length > 0,
          }}
        />
      </Card>
      <Drawer
        open={logNodes.length > 0}
        title="实时日志"
        width="min(960px, 92vw)"
        placement="right"
        onClose={() => {
          setLogNodes([])
          setLogPreview(undefined)
        }}
      >
        <div className={styles.logToolbar}>
          <label>
            <span>节点</span>
            <Select
              value={selectedLogNodeId || undefined}
              onChange={setSelectedLogNodeId}
              options={logNodes.map((node) => ({ value: node.nodeId, label: node.name }))}
            />
          </label>
          <label>
            <span>探针</span>
            <Select
              value={selectedLogProbe || undefined}
              onChange={setSelectedLogProbe}
              options={logProbeOptions.map((probe) => ({ value: probe, label: probe }))}
            />
          </label>
          <Space>
            <Switch checked={autoRefresh} onChange={setAutoRefresh} />
            <Typography.Text>自动刷新</Typography.Text>
          </Space>
          <Button icon={<RefreshCw size={15} />} loading={logLoading} onClick={refreshLogFiles}>
            刷新文件
          </Button>
        </div>
        <div className={styles.logWorkspace}>
          <div className={styles.logFileList}>
            <div className={styles.logListTitle}>
              <FileText size={16} />
              <Typography.Text strong>日志文件</Typography.Text>
            </div>
            {logFiles.map((file) => (
              <button
                type="button"
                key={file.path}
                className={selectedLogPath === file.path ? styles.logFileActive : ''}
                onClick={() => setSelectedLogPath(file.path)}
              >
                <strong>{file.name}</strong>
                <span>{file.path}</span>
              </button>
            ))}
            {!logFiles.length && !logLoading && (
              <Typography.Text type="secondary">没有可查看的日志文件</Typography.Text>
            )}
          </div>
          <div className={styles.logPreview}>
            <div className={styles.logPreviewHeader}>
              <Typography.Text strong>{logPreview?.file?.name || '选择日志文件'}</Typography.Text>
              <Space>
                <Button
                  icon={<RefreshCw size={15} />}
                  disabled={!selectedLogPath}
                  onClick={() => refreshLogPreview()}
                />
                <Button
                  icon={<Download size={15} />}
                  disabled={!selectedLogPath}
                  onClick={downloadLog}
                >
                  下载
                </Button>
              </Space>
            </div>
            <Input.TextArea
              value={logPreview?.preview || ''}
              readOnly
              placeholder={logLoading ? '正在读取日志...' : '请选择日志文件'}
              autoSize={{ minRows: 18, maxRows: 28 }}
            />
          </div>
        </div>
      </Drawer>
    </div>
  )
}
