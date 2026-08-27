import {
  Button,
  Card,
  Drawer,
  Input,
  Popconfirm,
  Select,
  Space,
  Switch,
  Table,
  Tag,
  Tooltip,
  Typography,
} from 'antd'
import { SettingOutlined } from '@ant-design/icons'
import {
  Activity,
  Cpu,
  Download,
  FileText,
  MemoryStick,
  Network,
  Play,
  Radio,
  RefreshCw,
  Server,
  Settings,
  Square,
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

type CircuitStatus = {
  state?: string
  epoch?: number
  reason?: string
}

type ProbeRuntime = {
  childCpuPercent?: number
  childRssBytes?: number
  guardRssBytes?: number
  childRunning?: boolean
  supervisorHealthy?: boolean
  preloadEnabled?: boolean
  preloadConfigPath?: string
  circuit?: CircuitStatus
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
  reportedAt?: string
  runtime?: ProbeRuntime
  circuit?: CircuitStatus
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
  healthyNodeCount: number
  nodeCount: number
}

type TableRow = ClusterRow | NodeRow
type NodeStatusFilter = 'all' | 'healthy' | 'abnormal' | 'waiting'
type LogFile = { path: string; name: string; size: number; modifiedAt: string }
type LogPreview = { file: LogFile; preview: string; lines: number; truncated: boolean }

const UNASSIGNED_CLUSTER = '未分配集群'
const STATUS_REFRESH_INTERVAL_MS = 60_000
const MAX_VISIBLE_ABNORMAL_PROBES = 2

const getClusterId = (node: AgentNode) =>
  String(node.clusterId || node.labels?.cluster_id || node.labels?.clusterId || '').trim() ||
  UNASSIGNED_CLUSTER

const isKubernetesEnvironment = (environment?: string) =>
  ['', 'k8s', 'kubernetes', 'helm'].includes(
    String(environment || '')
      .trim()
      .toLowerCase(),
  )

const isHealthy = (state: ProbeState) => {
  if (state.runtime?.preloadEnabled !== undefined) return state.status === 'RUNNING'
  return (
    state.status === 'RUNNING' &&
    state.runtime?.childRunning !== false &&
    state.runtime?.supervisorHealthy !== false
  )
}

const isFused = (state: ProbeState) => {
  const circuitState = String(state.circuit?.state || '').toUpperCase()
  return state.status === 'FUSED' || circuitState === 'FUSED' || circuitState === 'OPEN'
}

const isLifecycleControllable = (state: ProbeState) =>
  state.status !== 'OFFLINE' &&
  state.runtime?.preloadEnabled === undefined &&
  Boolean(state.circuit?.state)

const probeTagColor = (state: ProbeState) => {
  if (isHealthy(state)) return 'success'
  return state.status === 'OFFLINE' ? 'default' : 'warning'
}

const getProbeStatusText = (state: ProbeState) => {
  if (state.runtime?.preloadEnabled !== undefined) {
    return state.runtime.preloadEnabled ? '注入已启用' : '注入未启用'
  }
  if (isHealthy(state)) return '运行正常'
  if (state.status === 'OFFLINE') return '上报中断'
  if (state.runtime?.supervisorHealthy === false) return '守护异常'
  if (state.runtime?.childRunning === false) return '未运行'
  return state.status || '状态异常'
}

const getNodeStatus = (node: NodeRow): Exclude<NodeStatusFilter, 'all'> => {
  if (!node.states.length) return 'waiting'
  return node.healthyCount === node.probeCount ? 'healthy' : 'abnormal'
}

const summarizeStates = (states: ProbeState[]) => {
  const usage = states.map((state) =>
    isHealthy(state) ? getAveragedResourceUsage(state) : undefined,
  )
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
    .map((nodeId): AgentNode => ({ id: nodeId, name: nodeId }))
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
        healthyNodeCount: children.filter((node) => getNodeStatus(node) === 'healthy').length,
        nodeCount: children.length,
        ...summary,
      }
    })
}

const filterClusters = (
  clusters: ClusterRow[],
  query: string,
  statusFilter: NodeStatusFilter,
): ClusterRow[] => {
  const normalizedQuery = query.trim().toLocaleLowerCase()
  return clusters.flatMap((cluster) => {
    const clusterMatches = cluster.name.toLocaleLowerCase().includes(normalizedQuery)
    const children = cluster.children.filter((node) => {
      const nodeMatches =
        !normalizedQuery ||
        clusterMatches ||
        node.name.toLocaleLowerCase().includes(normalizedQuery) ||
        node.nodeId.toLocaleLowerCase().includes(normalizedQuery) ||
        node.states.some((state) => state.probeType.toLocaleLowerCase().includes(normalizedQuery))
      return nodeMatches && (statusFilter === 'all' || getNodeStatus(node) === statusFilter)
    })
    if (!children.length) return []

    const cpuAvailable =
      children.some((node) => node.probeCount > 0) &&
      children.every((node) => !node.probeCount || node.cpuPercent !== undefined)
    const memoryAvailable =
      children.some((node) => node.probeCount > 0) &&
      children.every((node) => !node.probeCount || node.memoryBytes !== undefined)
    return [
      {
        ...cluster,
        children,
        cpuPercent: cpuAvailable
          ? children.reduce((total, node) => total + Number(node.cpuPercent), 0)
          : undefined,
        memoryBytes: memoryAvailable
          ? children.reduce((total, node) => total + Number(node.memoryBytes), 0)
          : undefined,
        healthyCount: children.reduce((total, node) => total + node.healthyCount, 0),
        probeCount: children.reduce((total, node) => total + node.probeCount, 0),
        healthyNodeCount: children.filter((node) => getNodeStatus(node) === 'healthy').length,
        nodeCount: children.length,
      },
    ]
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
  const [query, setQuery] = useState('')
  const [statusFilter, setStatusFilter] = useState<NodeStatusFilter>('all')
  const [currentPage, setCurrentPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [logNodes, setLogNodes] = useState<NodeRow[]>([])
  const [selectedLogNodeId, setSelectedLogNodeId] = useState('')
  const [selectedLogProbe, setSelectedLogProbe] = useState('')
  const [logFiles, setLogFiles] = useState<LogFile[]>([])
  const [selectedLogPath, setSelectedLogPath] = useState('')
  const [logPreview, setLogPreview] = useState<LogPreview>()
  const [logLoading, setLogLoading] = useState(false)
  const [autoRefresh, setAutoRefresh] = useState(true)
  const [preloadUpdatingNodes, setPreloadUpdatingNodes] = useState<Set<string>>(new Set())
  const [lifecycleUpdatingNodes, setLifecycleUpdatingNodes] = useState<Set<string>>(new Set())
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
  const filteredClusters = useMemo(
    () => filterClusters(clusters, query, statusFilter),
    [clusters, query, statusFilter],
  )
  const clusterKeys = useMemo(
    () => filteredClusters.map((cluster) => cluster.key),
    [filteredClusters],
  )

  useEffect(() => {
    setExpandedRowKeys((current) =>
      current.length ? current.filter((key) => clusterKeys.includes(String(key))) : clusterKeys,
    )
  }, [clusterKeys])

  useEffect(() => {
    const lastPage = Math.max(1, Math.ceil(filteredClusters.length / pageSize))
    setCurrentPage((current) => Math.min(current, lastPage))
  }, [filteredClusters.length, pageSize])

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
    throw new Error('等待节点执行命令超时')
  }

  const setNodePreloadEnabled = async (record: NodeRow, enabled: boolean) => {
    setPreloadUpdatingNodes((current) => new Set(current).add(record.nodeId))
    try {
      const created: any = await createAgentCommandApi(
        record.nodeId,
        'preload',
        enabled ? 'ENABLE_PRELOAD' : 'DISABLE_PRELOAD',
      )
      const command = await waitForCommand(record.nodeId, 'preload', created.id)
      if (command.status !== 'SUCCEEDED') {
        throw new Error(command.result || 'Preload 状态调整失败')
      }
      await load()
      notify({ type: 'success', message: enabled ? 'Preload 注入已启用' : 'Preload 注入已禁用' })
    } catch (error: any) {
      notify({ type: 'error', message: error.message || 'Preload 状态调整失败' })
    } finally {
      setPreloadUpdatingNodes((current) => {
        const next = new Set(current)
        next.delete(record.nodeId)
        return next
      })
    }
  }

  const setNodesRunning = async (records: NodeRow[], running: boolean, scope: '节点' | '集群') => {
    const targets = records.flatMap((record) => {
      const controllableStates = record.states.filter(isLifecycleControllable)
      const states = running
        ? controllableStates.filter(
            (state) => isFused(state) || state.runtime?.childRunning === false,
          )
        : controllableStates.filter((state) => !isFused(state))
      return states.map((state) => ({ nodeId: record.nodeId, state }))
    })
    if (!targets.length) return

    const nodeIds = new Set(records.map((record) => record.nodeId))
    setLifecycleUpdatingNodes((current) => new Set([...current, ...nodeIds]))
    const outcomes = await Promise.all(
      targets.map(async ({ nodeId, state }) => {
        try {
          const commandType = running
            ? isFused(state)
              ? 'RESET_CIRCUIT'
              : 'RESTART'
            : 'OPEN_CIRCUIT'
          const options = commandType.includes('CIRCUIT')
            ? {
                expectedCircuitEpoch: state.circuit?.epoch ?? 0,
                ...(running ? {} : { reason: 'manual_node_stop' }),
              }
            : {}
          const created: any = await createAgentCommandApi(
            nodeId,
            state.probeType,
            commandType,
            options,
          )
          const command = await waitForCommand(nodeId, state.probeType, created.id)
          if (command.status !== 'SUCCEEDED') {
            throw new Error(command.result || '节点未完成操作')
          }
          let runtime: ProbeRuntime | undefined
          try {
            runtime = command.result ? JSON.parse(command.result) : undefined
          } catch {
            runtime = undefined
          }
          return { nodeId, probeType: state.probeType, runtime }
        } catch (error) {
          return { nodeId, probeType: state.probeType, error }
        }
      }),
    )

    const succeeded = outcomes.filter((outcome) => !outcome.error)
    const failed = outcomes.filter((outcome) => outcome.error)
    if (succeeded.length) {
      const succeededByTarget = new Map(
        succeeded.map((outcome) => [
          `${outcome.nodeId}\u0000${outcome.probeType}`,
          outcome.runtime,
        ]),
      )
      setStates((current) =>
        current.map((state) => {
          const targetKey = `${state.nodeId}\u0000${state.probeType}`
          if (!succeededByTarget.has(targetKey)) return state
          const runtime = succeededByTarget.get(targetKey)
          if (!runtime) return state
          const circuit = runtime?.circuit || {
            ...state.circuit,
            state: running ? 'NORMAL' : 'FUSED',
            epoch: (state.circuit?.epoch ?? 0) + 1,
          }
          return {
            ...state,
            status: running ? 'RUNNING' : 'FUSED',
            runtime: runtime ? { ...state.runtime, ...runtime } : state.runtime,
            circuit,
          }
        }),
      )
    }
    if (failed.length) {
      console.error('[ProbeManagement] node lifecycle command partially failed', {
        nodeIds: [...nodeIds],
        action: running ? 'start' : 'stop',
        failures: failed,
      })
      notify({
        type: 'error',
        message: `${scope}${running ? '启动' : '停止'}未完全完成，请刷新状态后重试`,
      })
    } else {
      notify({ type: 'success', message: `${scope}已${running ? '启动' : '停止'}` })
    }
    setLifecycleUpdatingNodes((current) => {
      const next = new Set(current)
      nodeIds.forEach((nodeId) => next.delete(nodeId))
      return next
    })
  }

  const setNodeRunning = (record: NodeRow, running: boolean) =>
    setNodesRunning([record], running, '节点')

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
      width: '28%',
      render: (_: unknown, record: TableRow) => (
        <div className={styles.identityCell}>
          <span className={record.kind === 'cluster' ? styles.clusterIcon : styles.nodeIcon}>
            {record.kind === 'cluster' ? <Network size={16} /> : <Server size={15} />}
          </span>
          <div>
            <Typography.Text strong>{record.name}</Typography.Text>
            {record.kind === 'node' && (
              <span className={styles.secondaryLine}>Node ID: {record.nodeId}</span>
            )}
          </div>
        </div>
      ),
    },
    {
      title: '运行状态',
      key: 'probeStatus',
      width: '30%',
      render: (_: unknown, record: TableRow) => {
        if (record.kind === 'cluster') {
          const healthy = record.nodeCount > 0 && record.healthyNodeCount === record.nodeCount
          return (
            <Tag color={healthy ? 'success' : record.nodeCount ? 'warning' : 'default'}>
              {record.healthyNodeCount} / {record.nodeCount} 个节点运行正常
            </Tag>
          )
        }
        if (!record.states.length) {
          return <span className={styles.resourceEmpty}>等待探针上报</span>
        }
        const abnormalStates = record.states.filter((state) => !isHealthy(state))
        const visibleAbnormalStates = abnormalStates.slice(0, MAX_VISIBLE_ABNORMAL_PROBES)
        const hiddenAbnormalCount = abnormalStates.length - visibleAbnormalStates.length
        const statusDetails = (
          <div className={styles.probeStatusTooltip}>
            {record.states.map((state) => (
              <span key={state.probeType}>
                <span className={isHealthy(state) ? styles.healthyDot : styles.abnormalDot} />
                <span>{state.probeType}</span>
                <span>{getProbeStatusText(state)}</span>
              </span>
            ))}
          </div>
        )
        return (
          <div className={styles.probeStatusSummary}>
            <Tooltip title={statusDetails} placement="bottomLeft">
              <Tag color={abnormalStates.length ? 'warning' : 'success'}>
                {record.healthyCount} 个运行正常
              </Tag>
            </Tooltip>
            {visibleAbnormalStates.map((state) => (
              <Tag key={state.probeType} color={probeTagColor(state)}>
                {state.probeType} · {getProbeStatusText(state)}
              </Tag>
            ))}
            {hiddenAbnormalCount > 0 && (
              <Tooltip title={statusDetails} placement="bottomLeft">
                <Tag color="warning">另 {hiddenAbnormalCount} 个异常</Tag>
              </Tooltip>
            )}
          </div>
        )
      },
    },
    {
      title: '资源开销（近 10 分钟均值）',
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
      width: 250,
      render: (_: unknown, record: TableRow) =>
        record.kind === 'cluster'
          ? (() => {
              const clusterNodes =
                clusters.find((cluster) => cluster.clusterId === record.clusterId)?.children ||
                record.children
              const lifecycleStates = clusterNodes.flatMap((node) =>
                node.states.filter(isLifecycleControllable),
              )
              const lifecycleUpdating = clusterNodes.some((node) =>
                lifecycleUpdatingNodes.has(node.nodeId),
              )
              const canStart = lifecycleStates.some(
                (state) => isFused(state) || state.runtime?.childRunning === false,
              )
              const canStop = lifecycleStates.some((state) => !isFused(state))
              return (
                <div className={styles.actionCell}>
                  <Button
                    className={styles.primaryAction}
                    type="link"
                    size="small"
                    icon={<SettingOutlined />}
                    onClick={() =>
                      navigate(
                        `/probe-management/config?clusterId=${encodeURIComponent(record.clusterId)}&clusterType=${encodeURIComponent(record.environment)}`,
                      )
                    }
                  >
                    集群配置
                  </Button>
                  <div className={styles.actionTools}>
                    <Tooltip title={canStart ? '启动集群探针' : '当前没有可启动的探针'}>
                      <Button
                        className={styles.iconAction}
                        type="text"
                        size="small"
                        aria-label="启动集群探针"
                        icon={<Play size={14} />}
                        loading={lifecycleUpdating}
                        disabled={lifecycleUpdating || !canStart}
                        onClick={() => setNodesRunning(clusterNodes, true, '集群')}
                      />
                    </Tooltip>
                    <Popconfirm
                      title={`确认停止集群 ${record.name}？`}
                      description="集群内节点将批量停止，管理通道保持在线。"
                      okText="停止"
                      cancelText="取消"
                      okButtonProps={{ danger: true }}
                      disabled={lifecycleUpdating || !canStop}
                      onConfirm={() => setNodesRunning(clusterNodes, false, '集群')}
                    >
                      <Tooltip title={canStop ? '停止集群探针' : '当前没有可停止的探针'}>
                        <Button
                          className={styles.iconAction}
                          danger
                          type="text"
                          size="small"
                          aria-label="停止集群探针"
                          icon={<Square size={13} />}
                          loading={lifecycleUpdating}
                          disabled={lifecycleUpdating || !canStop}
                        />
                      </Tooltip>
                    </Popconfirm>
                  </div>
                </div>
              )
            })()
          : (() => {
              const preloadState = record.states.find((state) => state.probeType === 'preload')
              const preloadEnabled = preloadState?.runtime?.preloadEnabled
              const preloadControllable =
                !isKubernetesEnvironment(record.environment) && preloadEnabled !== undefined
              const lifecycleStates = record.states.filter(isLifecycleControllable)
              const lifecycleUpdating = lifecycleUpdatingNodes.has(record.nodeId)
              const canStart = lifecycleStates.some(
                (state) => isFused(state) || state.runtime?.childRunning === false,
              )
              const canStop = lifecycleStates.some((state) => !isFused(state))
              return (
                <div className={styles.actionCell}>
                  {preloadControllable && (
                    <span className={styles.preloadControl}>
                      <Typography.Text>Preload</Typography.Text>
                      <Switch
                        size="small"
                        checked={preloadEnabled}
                        loading={preloadUpdatingNodes.has(record.nodeId)}
                        disabled={preloadState?.status === 'OFFLINE'}
                        onChange={(enabled) => setNodePreloadEnabled(record, enabled)}
                      />
                    </span>
                  )}
                  <Button
                    className={styles.primaryAction}
                    type="link"
                    size="small"
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
                  <div className={styles.actionTools}>
                    <Tooltip title={canStart ? '启动节点探针' : '当前没有可启动的探针'}>
                      <Button
                        className={styles.iconAction}
                        type="text"
                        size="small"
                        aria-label="启动节点探针"
                        icon={<Play size={14} />}
                        loading={lifecycleUpdating}
                        disabled={lifecycleUpdating || !canStart}
                        onClick={() => setNodeRunning(record, true)}
                      />
                    </Tooltip>
                    <Popconfirm
                      title={`确认停止节点 ${record.name}？`}
                      description="停止后管理通道保持在线，可随时重新启动。"
                      okText="停止"
                      cancelText="取消"
                      okButtonProps={{ danger: true }}
                      disabled={lifecycleUpdating || !canStop}
                      onConfirm={() => setNodeRunning(record, false)}
                    >
                      <Tooltip title={canStop ? '停止节点探针' : '当前没有可停止的探针'}>
                        <Button
                          className={styles.iconAction}
                          danger
                          type="text"
                          size="small"
                          aria-label="停止节点探针"
                          icon={<Square size={13} />}
                          loading={lifecycleUpdating}
                          disabled={lifecycleUpdating || !canStop}
                        />
                      </Tooltip>
                    </Popconfirm>
                    <Tooltip title="查看实时日志">
                      <Button
                        className={styles.iconAction}
                        type="text"
                        size="small"
                        aria-label="查看节点实时日志"
                        icon={<Radio size={15} />}
                        disabled={!record.states.length}
                        onClick={() => openRealtimeLogs([record])}
                      />
                    </Tooltip>
                  </div>
                </div>
              )
            })(),
    },
  ]

  return (
    <div className={styles.managementPage}>
      <div className={styles.sectionHeader}>
        <div>
          <Typography.Title level={4}>探针管理</Typography.Title>
          <Typography.Text type="secondary">按集群查看节点运行状态</Typography.Text>
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
        </Space>
      </div>
      <div className={styles.tableFilters}>
        <Input.Search
          allowClear
          value={query}
          placeholder="搜索集群、节点或组件"
          onChange={(event) => {
            setQuery(event.target.value)
            setCurrentPage(1)
          }}
        />
        <Select<NodeStatusFilter>
          value={statusFilter}
          onChange={(value) => {
            setStatusFilter(value)
            setCurrentPage(1)
          }}
          options={[
            { value: 'all', label: '全部状态' },
            { value: 'healthy', label: '运行正常' },
            { value: 'abnormal', label: '存在异常' },
            { value: 'waiting', label: '等待上报' },
          ]}
        />
      </div>
      <Card className={styles.tableCard} styles={{ body: { padding: 0 } }}>
        <Table
          columns={columns}
          dataSource={filteredClusters}
          rowKey="key"
          loading={loading}
          pagination={{
            current: currentPage,
            pageSize,
            showSizeChanger: true,
            showQuickJumper: filteredClusters.length > pageSize,
            pageSizeOptions: [10, 20, 50],
            onChange: (page, size) => {
              setCurrentPage(page)
              setPageSize(size)
            },
          }}
          size="small"
          scroll={{ x: 860 }}
          rowClassName={(record) =>
            record.kind === 'cluster' ? styles.clusterRow : styles.nodeRow
          }
          locale={{
            emptyText: (
              <span className={styles.resourceEmpty}>
                <Activity size={15} />
                暂无已注册节点
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
