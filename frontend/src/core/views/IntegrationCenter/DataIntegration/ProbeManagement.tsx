import {
  Button,
  Card,
  Drawer,
  Input,
  Modal,
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
  History,
  MemoryStick,
  Network,
  Pencil,
  Play,
  Plus,
  Radio,
  RefreshCw,
  Server,
  Settings,
  SlidersHorizontal,
  Square,
  TerminalSquare,
  Trash2,
} from 'lucide-react'
import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  createAgentCommandApi,
  deleteClusterIntegrationApi,
  getAgentCommandsApi,
  getAgentConfigNodesApi,
  getAgentConfigDeliveriesApi,
  getAgentConfigStatesApi,
  getIntegrationClusterListApi,
  updateClusterDisplayNameApi,
} from 'src/core/api/integration'
import { notify } from 'src/core/utils/notify'
import InstallCmd from './Integration/InstallCmd'
import styles from './probeManagement.module.scss'

type AgentNode = {
  id: string
  name?: string
  clusterId?: string
  environment?: string
  labels?: Record<string, string>
  lastSeenAt?: string
}

type IntegrationCluster = {
  id: string
  name?: string
  clusterType?: 'k8s' | 'vm'
  apoCollector?: Record<string, unknown>
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
  manualCluster?: IntegrationCluster
}

type TableRow = ClusterRow | NodeRow
type NodeStatusFilter = 'all' | 'healthy' | 'abnormal' | 'waiting'
type LogFile = { path: string; name: string; size: number; modifiedAt: string; probeType: string }
type LogPreview = { file: LogFile; preview: string; lines: number; truncated: boolean }
type DeliveryReceiptStatus = 'PENDING' | 'PARTIAL' | 'RECEIVED' | 'APPLIED' | 'FAILED'
type DeliveryProbeReceipt = {
  probeType: string
  targetGeneration: number
  desiredGeneration: number
  appliedGeneration: number
  status: DeliveryReceiptStatus
  agentStatus?: string
  reason?: string
  reportedAt?: string
}
type DeliveryTarget = {
  nodeId: string
  nodeName?: string
  clusterId?: string
  status: DeliveryReceiptStatus
  probes: DeliveryProbeReceipt[]
}
type DeliveryRecord = {
  id: string
  type: 'NODE_CONFIG' | 'BATCH_TEMPLATE' | 'NODE_INJECTION' | 'BATCH_INJECTION'
  scope: string
  templateName?: string
  requestedBy: number
  targetNodeCount: number
  appliedCount: number
  receivedCount: number
  failedCount: number
  pendingCount: number
  status: DeliveryReceiptStatus
  createdAt: string
  targets: DeliveryTarget[]
}
const UNASSIGNED_CLUSTER = '未分配集群'
const STATUS_REFRESH_INTERVAL_MS = 60_000
const MAX_VISIBLE_ABNORMAL_PROBES = 2
const DELIVERY_TYPE_LABELS: Record<DeliveryRecord['type'], string> = {
  NODE_CONFIG: '节点配置',
  BATCH_TEMPLATE: '批量配置',
  NODE_INJECTION: '节点注入配置',
  BATCH_INJECTION: '批量注入配置',
}
const DELIVERY_STATUS: Record<DeliveryReceiptStatus, { label: string; color: string }> = {
  PENDING: { label: '待接收', color: 'default' },
  PARTIAL: { label: '部分接收', color: 'processing' },
  RECEIVED: { label: '已接收', color: 'processing' },
  APPLIED: { label: '已应用', color: 'success' },
  FAILED: { label: '失败', color: 'error' },
}

const getClusterId = (node: AgentNode) =>
  String(node.clusterId || node.labels?.cluster_id || node.labels?.clusterId || '').trim() ||
  UNASSIGNED_CLUSTER

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

const groupNodesByCluster = (
  nodes: AgentNode[],
  states: ProbeState[],
  manualClusters: IntegrationCluster[],
): ClusterRow[] => {
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

  const manualClustersById = new Map(
    manualClusters
      .filter((cluster) => String(cluster.id || '').trim())
      .map((cluster) => [String(cluster.id).trim(), cluster]),
  )
  const clusterIds = [...new Set([...grouped.keys(), ...manualClustersById.keys()])]

  return clusterIds
    .sort((a, b) => a.localeCompare(b))
    .map((clusterId) => {
      const children = grouped.get(clusterId) || []
      const manualCluster = manualClustersById.get(clusterId)
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
        name: String(manualCluster?.name || '').trim() || clusterId,
        environment:
          manualCluster?.clusterType ||
          [...new Set(children.map((node) => node.environment).filter(Boolean))][0] ||
          '',
        children,
        healthyNodeCount: children.filter((node) => getNodeStatus(node) === 'healthy').length,
        nodeCount: children.length,
        manualCluster,
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
    const clusterMatches =
      cluster.name.toLocaleLowerCase().includes(normalizedQuery) ||
      cluster.clusterId.toLocaleLowerCase().includes(normalizedQuery)
    const children = cluster.children.filter((node) => {
      const nodeMatches =
        !normalizedQuery ||
        clusterMatches ||
        node.name.toLocaleLowerCase().includes(normalizedQuery) ||
        node.nodeId.toLocaleLowerCase().includes(normalizedQuery) ||
        node.states.some((state) => state.probeType.toLocaleLowerCase().includes(normalizedQuery))
      return nodeMatches && (statusFilter === 'all' || getNodeStatus(node) === statusFilter)
    })
    if (!children.length) {
      return clusterMatches && (statusFilter === 'all' || statusFilter === 'waiting')
        ? [{ ...cluster, children: [] }]
        : []
    }

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

const formatDeliveryTime = (value?: string) => {
  if (!value) return '-'
  const time = new Date(value)
  return Number.isNaN(time.getTime()) ? value : time.toLocaleString('zh-CN', { hour12: false })
}

const deliveryStatusTag = (status: DeliveryReceiptStatus) => {
  const display = DELIVERY_STATUS[status] || DELIVERY_STATUS.PENDING
  return <Tag color={display.color}>{display.label}</Tag>
}

export default function ProbeManagement() {
  const navigate = useNavigate()
  const [nodes, setNodes] = useState<AgentNode[]>([])
  const [states, setStates] = useState<ProbeState[]>([])
  const [manualClusters, setManualClusters] = useState<IntegrationCluster[]>([])
  const [loading, setLoading] = useState(true)
  const [expandedRowKeys, setExpandedRowKeys] = useState<React.Key[]>([])
  const [query, setQuery] = useState('')
  const [statusFilter, setStatusFilter] = useState<NodeStatusFilter>('all')
  const [currentPage, setCurrentPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [logNodes, setLogNodes] = useState<NodeRow[]>([])
  const [selectedLogNodeId, setSelectedLogNodeId] = useState('')
  const [logFiles, setLogFiles] = useState<LogFile[]>([])
  const [selectedLogFile, setSelectedLogFile] = useState<LogFile>()
  const [logPreview, setLogPreview] = useState<LogPreview>()
  const [logLoading, setLogLoading] = useState(false)
  const [autoRefresh, setAutoRefresh] = useState(true)
  const [lifecycleUpdatingNodes, setLifecycleUpdatingNodes] = useState<Set<string>>(new Set())
  const [displayNameCluster, setDisplayNameCluster] = useState<IntegrationCluster>()
  const [displayName, setDisplayName] = useState('')
  const [displayNameSaving, setDisplayNameSaving] = useState(false)
  const [installCluster, setInstallCluster] = useState<IntegrationCluster>()
  const [deliveryDrawerOpen, setDeliveryDrawerOpen] = useState(false)
  const [deliveries, setDeliveries] = useState<DeliveryRecord[]>([])
  const [deliveryLoading, setDeliveryLoading] = useState(false)
  const logRequestActive = useRef(false)

  const load = useCallback(async () => {
    try {
      const [nodeResult, stateResult, manualClusterResult] = await Promise.allSettled([
        getAgentConfigNodesApi(),
        getAgentConfigStatesApi(),
        getIntegrationClusterListApi(),
      ])

      if (nodeResult.status === 'fulfilled') {
        setNodes(Array.isArray((nodeResult.value as any)?.nodes) ? nodeResult.value.nodes : [])
      } else {
        console.error('[ProbeManagement] failed to load registered nodes', nodeResult.reason)
      }

      if (stateResult.status === 'fulfilled') {
        const loadedStates = Array.isArray((stateResult.value as any)?.states)
          ? stateResult.value.states
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
      } else {
        console.error('[ProbeManagement] failed to load probe states', stateResult.reason)
      }

      if (manualClusterResult.status === 'fulfilled') {
        setManualClusters(
          Array.isArray((manualClusterResult.value as any)?.clusters)
            ? manualClusterResult.value.clusters
            : [],
        )
      } else {
        console.error(
          '[ProbeManagement] failed to load manually registered clusters',
          manualClusterResult.reason,
        )
      }
    } catch (error) {
      console.error('[ProbeManagement] unexpected load failure', error)
    } finally {
      setLoading(false)
    }
  }, [])

  const loadDeliveries = useCallback(async () => {
    setDeliveryLoading(true)
    try {
      const result: any = await getAgentConfigDeliveriesApi()
      setDeliveries(Array.isArray(result?.deliveries) ? result.deliveries : [])
    } catch (error: any) {
      console.error('[ProbeManagement] failed to load delivery history', error)
      notify({ type: 'error', message: error?.message || '加载下发记录失败' })
    } finally {
      setDeliveryLoading(false)
    }
  }, [])

  const openDeliveryHistory = () => {
    setDeliveryDrawerOpen(true)
    void loadDeliveries()
  }

  useEffect(() => {
    load()
    const timer = window.setInterval(load, STATUS_REFRESH_INTERVAL_MS)
    return () => window.clearInterval(timer)
  }, [load])

  const clusters = useMemo(
    () => groupNodesByCluster(nodes, states, manualClusters),
    [nodes, states, manualClusters],
  )
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

  const waitForCommand = async (nodeId: string, probeType: string, commandId: string) => {
    for (let attempt = 0; attempt < 30; attempt += 1) {
      const response: any = await getAgentCommandsApi(nodeId, probeType)
      const command = (response?.commands || []).find((item) => item.id === commandId)
      if (command?.status === 'SUCCEEDED' || command?.status === 'FAILED') return command
      await new Promise((resolve) => window.setTimeout(resolve, 1000))
    }
    throw new Error('等待节点执行命令超时')
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
    setNodesRunning([record], running, '节点探针')

  const runLogCommand = async (
    type: string,
    options: Record<string, unknown> = {},
    nodeId = selectedLogNodeId,
    probeType = selectedLogFile?.probeType || '',
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
    const probeTypes = [...new Set(selectedLogNode?.states.map((state) => state.probeType) || [])]
    if (!selectedLogNodeId || !probeTypes.length) return
    setLogLoading(true)
    try {
      const responses = await Promise.all(
        probeTypes.map(async (probeType) => {
          try {
            const result = await runLogCommand('LIST_LOG_FILES', {}, selectedLogNodeId, probeType)
            return {
              files: ((result.files || []) as Omit<LogFile, 'probeType'>[]).map((file) => ({
                ...file,
                probeType,
              })),
              errors: result.errors || [],
            }
          } catch (error) {
            console.error('[ProbeManagement] failed to list log files for probe', {
              nodeId: selectedLogNodeId,
              probeType,
              error,
            })
            return { files: [], errors: [{ message: `${probeType}: 读取日志文件失败` }] }
          }
        }),
      )
      const files = responses
        .flatMap((response) => response.files)
        .sort((a, b) => Date.parse(b.modifiedAt) - Date.parse(a.modifiedAt))
      const errors = responses.flatMap((response) => response.errors)
      setLogFiles(files)
      setSelectedLogFile(
        (current) =>
          files.find(
            (file) => file.path === current?.path && file.probeType === current.probeType,
          ) || files[0],
      )
      if (errors.length) {
        notify({ type: 'warning', message: errors[0].message })
      }
    } catch (error: any) {
      notify({ type: 'error', message: error.message || '读取日志文件列表失败' })
    } finally {
      setLogLoading(false)
    }
  }

  const refreshLogPreview = async (quiet = false, file = selectedLogFile) => {
    if (!file || logRequestActive.current) return
    logRequestActive.current = true
    if (!quiet) setLogLoading(true)
    try {
      setLogPreview(
        await runLogCommand(
          'PREVIEW_LOG_FILE',
          { filePath: file.path, tailLines: 100 },
          selectedLogNodeId,
          file.probeType,
        ),
      )
    } catch (error: any) {
      notify({ type: 'error', message: error.message || '读取实时日志失败' })
    } finally {
      logRequestActive.current = false
      if (!quiet) setLogLoading(false)
    }
  }

  useEffect(() => {
    if (!logNodes.length || !selectedLogNodeId) return
    setLogFiles([])
    setSelectedLogFile(undefined)
    setLogPreview(undefined)
    refreshLogFiles()
  }, [selectedLogNodeId, logNodes.length])

  useEffect(() => {
    if (!logNodes.length || !selectedLogFile) return
    refreshLogPreview()
  }, [selectedLogFile])

  useEffect(() => {
    if (!logNodes.length || !selectedLogFile || !autoRefresh) return
    const timer = window.setInterval(() => refreshLogPreview(true), 5000)
    return () => window.clearInterval(timer)
  }, [logNodes.length, selectedLogNodeId, selectedLogFile, autoRefresh])

  const openRealtimeLogs = (targetNodes: NodeRow[]) => {
    const available = targetNodes.filter((node) => node.states.length)
    setLogNodes(available)
    setSelectedLogNodeId(available[0]?.nodeId || '')
    setAutoRefresh(true)
  }

  const downloadLog = async () => {
    if (!selectedLogFile) return
    setLogLoading(true)
    try {
      const result = await runLogCommand(
        'DOWNLOAD_LOG_FILE',
        { filePath: selectedLogFile.path },
        selectedLogNodeId,
        selectedLogFile.probeType,
      )
      const binary = window.atob(result.content)
      const bytes = Uint8Array.from(binary, (char) => char.charCodeAt(0))
      const url = URL.createObjectURL(new Blob([bytes]))
      const anchor = document.createElement('a')
      anchor.href = url
      anchor.download = result.file?.name || selectedLogFile.name || 'probe.log'
      anchor.click()
      URL.revokeObjectURL(url)
    } catch (error: any) {
      notify({ type: 'error', message: error.message || '下载日志失败' })
    } finally {
      setLogLoading(false)
    }
  }

  const openDisplayNameEditor = (cluster: IntegrationCluster) => {
    setDisplayNameCluster(cluster)
    setDisplayName(String(cluster.name || ''))
  }

  const saveDisplayName = async () => {
    if (!displayNameCluster) return
    setDisplayNameSaving(true)
    try {
      const updated: any = await updateClusterDisplayNameApi(
        displayNameCluster.id,
        displayName.trim(),
      )
      setManualClusters((current) =>
        current.map((cluster) =>
          cluster.id === displayNameCluster.id
            ? { ...cluster, name: updated?.name || displayNameCluster.id }
            : cluster,
        ),
      )
      notify({ type: 'success', message: '集群展示名已更新' })
      setDisplayNameCluster(undefined)
    } catch (error: any) {
      console.error('[ProbeManagement] failed to update cluster display name', error)
      notify({ type: 'error', message: error?.message || '更新集群展示名失败' })
    } finally {
      setDisplayNameSaving(false)
    }
  }

  const deleteManualCluster = async (cluster: IntegrationCluster) => {
    try {
      await deleteClusterIntegrationApi(cluster.id)
      setManualClusters((current) => current.filter((item) => item.id !== cluster.id))
      notify({ type: 'success', message: '手动接入记录已删除' })
    } catch (error: any) {
      console.error('[ProbeManagement] failed to delete manually registered cluster', error)
      notify({ type: 'error', message: error?.message || '删除手动接入记录失败' })
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
            {record.kind === 'node' ? (
              <span className={styles.secondaryLine}>Node ID: {record.nodeId}</span>
            ) : (
              <span className={styles.secondaryLine}>
                cluster_id: {record.clusterId}
                {record.manualCluster && <Tag color="blue">手动接入</Tag>}
              </span>
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
      width: 360,
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
                    {record.manualCluster && (
                      <>
                        <Tooltip title="编辑接入设置">
                          <Button
                            className={styles.iconAction}
                            type="text"
                            size="small"
                            aria-label="编辑接入设置"
                            icon={<SlidersHorizontal size={14} />}
                            onClick={() =>
                              navigate(
                                `/probe-management/settings?clusterId=${encodeURIComponent(record.clusterId)}&clusterType=${encodeURIComponent(record.manualCluster?.clusterType || record.environment)}`,
                              )
                            }
                          />
                        </Tooltip>
                        <Tooltip title="修改展示名">
                          <Button
                            className={styles.iconAction}
                            type="text"
                            size="small"
                            aria-label="修改集群展示名"
                            icon={<Pencil size={14} />}
                            onClick={() => openDisplayNameEditor(record.manualCluster!)}
                          />
                        </Tooltip>
                        <Tooltip title="查看安装配置">
                          <Button
                            className={styles.iconAction}
                            type="text"
                            size="small"
                            aria-label="查看安装配置"
                            icon={<TerminalSquare size={15} />}
                            onClick={() => setInstallCluster(record.manualCluster)}
                          />
                        </Tooltip>
                        <Popconfirm
                          title={`确认删除手动接入记录 ${record.name}？`}
                          description="已上报的探针节点仍会保留在管理列表中。"
                          okText="删除"
                          cancelText="取消"
                          okButtonProps={{ danger: true }}
                          onConfirm={() => deleteManualCluster(record.manualCluster!)}
                        >
                          <Tooltip title="删除手动接入记录">
                            <Button
                              className={styles.iconAction}
                              danger
                              type="text"
                              size="small"
                              aria-label="删除手动接入记录"
                              icon={<Trash2 size={14} />}
                            />
                          </Tooltip>
                        </Popconfirm>
                      </>
                    )}
                  </div>
                </div>
              )
            })()
          : (() => {
              const lifecycleStates = record.states.filter(isLifecycleControllable)
              const lifecycleUpdating = lifecycleUpdatingNodes.has(record.nodeId)
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
                      title={`确认停止节点 ${record.name} 上的探针？`}
                      description="操作后，该节点上的探针将停止数据采集与上报，节点管理通道不受影响；需要时可重新启动探针。"
                      okText="停止探针"
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

  const deliveryColumns = [
    {
      title: '下发时间',
      dataIndex: 'createdAt',
      width: 170,
      render: (value: string) => formatDeliveryTime(value),
    },
    {
      title: '下发范围',
      key: 'scope',
      width: 210,
      render: (_: unknown, record: DeliveryRecord) => (
        <div className={styles.deliveryScope}>
          <Typography.Text strong>
            {DELIVERY_TYPE_LABELS[record.type] || record.type}
          </Typography.Text>
          <Typography.Text type="secondary" ellipsis={{ tooltip: record.scope }}>
            {record.templateName || record.scope}
          </Typography.Text>
        </div>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 100,
      render: (status: DeliveryReceiptStatus) => deliveryStatusTag(status),
    },
    {
      title: '节点接收',
      key: 'receipt',
      render: (_: unknown, record: DeliveryRecord) => (
        <div className={styles.deliveryCounts}>
          <span>
            已应用 <strong>{record.appliedCount}</strong>
          </span>
          <span>
            已接收 <strong>{record.receivedCount}</strong>
          </span>
          <span className={record.failedCount ? styles.deliveryFailed : undefined}>
            失败 <strong>{record.failedCount}</strong>
          </span>
          <span>
            待接收 <strong>{record.pendingCount}</strong>
          </span>
        </div>
      ),
    },
    {
      title: '操作人',
      dataIndex: 'requestedBy',
      width: 90,
      render: (value: number) => (value ? `用户 ${value}` : '系统'),
    },
  ]

  const renderDeliveryTargets = (delivery: DeliveryRecord) => (
    <div className={styles.deliveryTargets}>
      <div className={styles.deliveryTargetHeader}>
        <span>节点</span>
        <span>接收状态</span>
        <span>探针回执</span>
      </div>
      {delivery.targets.map((target) => (
        <div className={styles.deliveryTargetRow} key={target.nodeId}>
          <div className={styles.deliveryTargetNode}>
            <Typography.Text strong>{target.nodeName || target.nodeId}</Typography.Text>
            <span>{target.nodeId}</span>
            <span>{target.clusterId || '-'}</span>
          </div>
          <div>{deliveryStatusTag(target.status)}</div>
          <div className={styles.deliveryProbes}>
            {target.probes.map((probe) => (
              <div className={styles.deliveryProbe} key={probe.probeType}>
                <div>
                  <Typography.Text strong>{probe.probeType}</Typography.Text>
                  <span>目标 generation {probe.targetGeneration}</span>
                </div>
                <div className={styles.deliveryGenerations}>
                  <span>接收 {probe.desiredGeneration || '-'}</span>
                  <span>应用 {probe.appliedGeneration || '-'}</span>
                </div>
                <div>{deliveryStatusTag(probe.status)}</div>
                <div className={styles.deliveryProbeReport}>
                  <span>{formatDeliveryTime(probe.reportedAt)}</span>
                  {(probe.reason || probe.agentStatus) && (
                    <span title={probe.reason || probe.agentStatus}>
                      {probe.reason || probe.agentStatus}
                    </span>
                  )}
                </div>
              </div>
            ))}
          </div>
        </div>
      ))}
    </div>
  )

  return (
    <div className={styles.managementPage}>
      <div className={styles.sectionHeader}>
        <div>
          <Typography.Title level={4}>探针管理</Typography.Title>
          <Typography.Text type="secondary">按集群查看节点运行状态</Typography.Text>
        </div>
        <Space className={styles.headerActions} wrap>
          <Button
            icon={<Network size={15} />}
            onClick={() => navigate('/probe-management/clusters')}
          >
            集群管理
          </Button>
          <Button
            type="primary"
            icon={<Plus size={15} />}
            onClick={() => navigate('/probe-management/settings')}
          >
            接入新集群
          </Button>
          <Button icon={<History size={15} />} onClick={openDeliveryHistory}>
            下发记录
          </Button>
          <Button
            icon={<Settings size={15} />}
            onClick={() => navigate('/probe-management/config?scope=global')}
          >
            默认配置
          </Button>
          <Button
            icon={<SlidersHorizontal size={15} />}
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
          scroll={{ x: 1040 }}
          rowClassName={(record) =>
            record.kind === 'cluster' ? styles.clusterRow : styles.nodeRow
          }
          locale={{
            emptyText: (
              <span className={styles.resourceEmpty}>
                <Activity size={15} />
                暂无接入集群或已注册节点
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
        open={deliveryDrawerOpen}
        title="下发记录"
        width="min(1080px, 94vw)"
        placement="right"
        extra={
          <Tooltip title="刷新下发记录">
            <Button
              type="text"
              aria-label="刷新下发记录"
              icon={<RefreshCw size={15} />}
              loading={deliveryLoading}
              onClick={loadDeliveries}
            />
          </Tooltip>
        }
        onClose={() => setDeliveryDrawerOpen(false)}
      >
        <Table
          rowKey="id"
          columns={deliveryColumns}
          dataSource={deliveries}
          loading={deliveryLoading}
          size="small"
          scroll={{ x: 820 }}
          pagination={{ pageSize: 10, showSizeChanger: false }}
          expandable={{
            rowExpandable: (record) => record.targets.length > 0,
            expandedRowRender: renderDeliveryTargets,
          }}
          locale={{ emptyText: '暂无下发记录' }}
        />
      </Drawer>
      <Drawer
        open={logNodes.length > 0}
        title="实时日志"
        width="min(960px, 92vw)"
        placement="right"
        onClose={() => {
          setLogNodes([])
          setLogFiles([])
          setSelectedLogFile(undefined)
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
          <Space>
            <Switch checked={autoRefresh} onChange={setAutoRefresh} />
            <Typography.Text>自动刷新</Typography.Text>
          </Space>
          <Button icon={<RefreshCw size={15} />} loading={logLoading} onClick={refreshLogFiles}>
            刷新日志
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
                key={`${file.probeType}:${file.path}`}
                className={
                  selectedLogFile?.path === file.path &&
                  selectedLogFile.probeType === file.probeType
                    ? styles.logFileActive
                    : ''
                }
                onClick={() => setSelectedLogFile(file)}
              >
                <strong>{file.name}</strong>
                <span>
                  {file.probeType} · {file.path}
                </span>
              </button>
            ))}
            {!logFiles.length && !logLoading && (
              <Typography.Text type="secondary">没有可查看的日志文件</Typography.Text>
            )}
          </div>
          <div className={styles.logPreview}>
            <div className={styles.logPreviewHeader}>
              <Space size={8}>
                <Typography.Text strong>
                  {logPreview?.file?.name || selectedLogFile?.name || '选择日志文件'}
                </Typography.Text>
                {selectedLogFile && (
                  <Typography.Text type="secondary">{selectedLogFile.probeType}</Typography.Text>
                )}
              </Space>
              <Space>
                <Button
                  icon={<Download size={15} />}
                  disabled={!selectedLogFile}
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
      <Modal
        open={Boolean(displayNameCluster)}
        title="修改集群展示名"
        okText="保存"
        cancelText="取消"
        confirmLoading={displayNameSaving}
        onOk={saveDisplayName}
        onCancel={() => setDisplayNameCluster(undefined)}
      >
        <Typography.Text type="secondary">cluster_id: {displayNameCluster?.id}</Typography.Text>
        <Input
          className="mt-3"
          value={displayName}
          placeholder="留空时使用 cluster_id"
          onChange={(event) => setDisplayName(event.target.value)}
          onPressEnter={saveDisplayName}
        />
      </Modal>
      <Modal
        open={Boolean(installCluster)}
        title="安装配置"
        footer={null}
        centered
        width={800}
        styles={{ body: { maxHeight: '70vh', overflowY: 'auto', overflowX: 'hidden' } }}
        onCancel={() => setInstallCluster(undefined)}
      >
        {installCluster && (
          <InstallCmd
            clusterId={installCluster.id}
            clusterType={installCluster.clusterType}
            apoCollector={installCluster.apoCollector}
          />
        )}
      </Modal>
    </div>
  )
}
