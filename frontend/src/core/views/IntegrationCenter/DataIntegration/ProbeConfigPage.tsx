import {
  Button,
  Card,
  Col,
  Collapse,
  Input,
  InputNumber,
  Modal,
  Row,
  Segmented,
  Select,
  Space,
  Switch,
  Table,
  Tabs,
  Tag,
  Tree,
  Typography,
} from 'antd'
import { LeftOutlined } from '@ant-design/icons'
import Editor, { DiffEditor } from '@monaco-editor/react'
import {
  Cpu,
  FileText,
  History,
  MemoryStick,
  Network,
  RotateCcw,
  Save,
  Send,
  Server,
  Settings,
  ShieldAlert,
  Trash2,
  Upload,
} from 'lucide-react'
import { type ChangeEvent, useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import {
  batchAssignAgentConfigApi,
  batchUpdateCircuitBreakerApi,
  createAgentConfigTemplateApi,
  getAgentConfigNodesApi,
  getAgentConfigStatesApi,
  getAgentConfigTemplatesApi,
  getGlobalCircuitBreakerApi,
  getNodeCircuitBreakerApi,
  getNodeConfigHistoryApi,
  getNodeCurrentConfigsApi,
  importAgentConfigTemplateApi,
  updateAgentConfigTemplateApi,
  updateGlobalCircuitBreakerApi,
  updateNodeCircuitBreakerApi,
  updateNodeCurrentConfigsApi,
} from 'src/core/api/integration'
import { notify } from 'src/core/utils/notify'
import { LogCollectionConfigPanel } from './LogCollectionConfigPage'
import styles from './probeConfigPage.module.scss'

type AgentNode = {
  id: string
  name?: string
  clusterId?: string
  environment?: string
  labels?: Record<string, string>
  reportedVariables?: Record<string, string>
}
type ProbeState = { nodeId: string; probeType: string }
type ClusterType = 'kubernetes' | 'vm'
type TemplateVariable = { name: string; required?: boolean; defaultValue?: string }
type ConfigFile = { path: string; content: string; contentSha256: string }
type TemplateProbeConfig = { probeType: string; files: ConfigFile[] }
type AgentTemplate = {
  id: string
  name: string
  description?: string
  clusterType?: string
  configs: TemplateProbeConfig[]
  version: number
  variables?: TemplateVariable[]
}
type CircuitPolicy = {
  enabled: boolean
  maxCpuCores: number
  maxMemoryBytes: number
  durationSeconds: number
  autoRecover: boolean
  recoveryInitialDelaySeconds: number
  recoveryMaxDelaySeconds: number
  systemPressure?: SystemPressurePolicy
  componentPolicies?: Record<string, CircuitPolicy>
}
type SystemPressurePolicy = {
  enabled: boolean
  maxCpuPercent: number
  maxMemoryPercent: number
}
type CurrentConfig = {
  probeType: string
  generation: number
  version: number
  content: string
  files: ConfigFile[]
  contentSha256: string
  source: string
  createdAt?: string
}
type DeliveryRecord = {
  id: string
  kind: 'CONFIG' | 'CIRCUIT'
  target: string
  succeeded: number
  total: number
  createdAt: string
}
type BatchPreviewFile = ConfigFile & {
  key: string
  probeType: string
}

const DEFAULT_CIRCUIT: CircuitPolicy = {
  enabled: true,
  maxCpuCores: 1,
  maxMemoryBytes: 2 * 1024 * 1024 * 1024,
  durationSeconds: 30,
  autoRecover: true,
  recoveryInitialDelaySeconds: 30,
  recoveryMaxDelaySeconds: 3600,
  systemPressure: {
    enabled: false,
    maxCpuPercent: 90,
    maxMemoryPercent: 90,
  },
}

const toMemoryMiB = (bytes: number) => Math.round(bytes / 1024 / 1024)
const fromMemoryMiB = (mib: number) => mib * 1024 * 1024
const isK8sEnvironment = (value?: string) =>
  ['', 'k8s', 'kubernetes', 'helm'].includes(
    String(value || '')
      .trim()
      .toLowerCase(),
  )
const normalizeClusterType = (value?: string): ClusterType =>
  ['vm', 'binary'].includes(
    String(value || '')
      .trim()
      .toLowerCase(),
  )
    ? 'vm'
    : 'kubernetes'
const templateClusterType = (template?: AgentTemplate) =>
  normalizeClusterType(template?.clusterType)
const clusterTypeLabel = (value?: string) =>
  normalizeClusterType(value) === 'vm' ? 'VM' : 'Kubernetes'
const templateLabel = (template: AgentTemplate) =>
  `${clusterTypeLabel(template.clusterType)} · ${template.name} · v${template.version}`
const draftKey = (probeType: string, path: string) => `${probeType}\u0000${path}`
const concreteCircuitPolicy = (value?: Partial<CircuitPolicy>): CircuitPolicy => {
  const { componentPolicies: _ignored, ...policy } = value || {}
  return {
    ...DEFAULT_CIRCUIT,
    ...policy,
    systemPressure: {
      ...DEFAULT_CIRCUIT.systemPressure!,
      ...value?.systemPressure,
    },
  }
}
const concreteComponentCircuitPolicy = (value?: Partial<CircuitPolicy>): CircuitPolicy => {
  const { systemPressure: _systemPressure, ...policy } = concreteCircuitPolicy(value)
  return policy
}
const normalizeCircuitPolicy = (value?: Partial<CircuitPolicy>): CircuitPolicy => ({
  ...concreteCircuitPolicy(value),
  componentPolicies: Object.fromEntries(
    Object.entries(value?.componentPolicies || {}).map(([componentType, policy]) => [
      componentType,
      concreteComponentCircuitPolicy(policy),
    ]),
  ),
})
const clusterCircuitPolicy = (value: CircuitPolicy): CircuitPolicy => {
  const normalized = normalizeCircuitPolicy(value)
  return {
    ...normalized,
    maxCpuCores: 0,
    maxMemoryBytes: 0,
    systemPressure: {
      ...normalized.systemPressure!,
      enabled: normalized.enabled,
    },
  }
}
const filesForConfig = (config: CurrentConfig): ConfigFile[] =>
  config.files?.length
    ? config.files
    : [{ path: 'config.yaml', content: config.content || '', contentSha256: config.contentSha256 }]
const configsForTemplate = (template?: AgentTemplate): TemplateProbeConfig[] =>
  template?.configs || []
const languageForPath = (path: string, fallback = 'yaml') =>
  path.endsWith('.xml')
    ? 'xml'
    : path.endsWith('.json')
      ? 'json'
      : path.endsWith('.hcl')
        ? 'hcl'
        : path.endsWith('.yaml') || path.endsWith('.yml')
          ? 'yaml'
          : fallback

function CircuitEditor({
  title,
  value,
  onChange,
  showSystemPressure = false,
  systemOnly = false,
}: {
  title: string
  value: CircuitPolicy
  onChange: (next: CircuitPolicy) => void
  showSystemPressure?: boolean
  systemOnly?: boolean
}) {
  const patch = (next: Partial<CircuitPolicy>) => onChange({ ...value, ...next })
  const systemPressure = value.systemPressure || DEFAULT_CIRCUIT.systemPressure!
  const patchSystemPressure = (next: Partial<SystemPressurePolicy>) =>
    patch({ systemPressure: { ...systemPressure, ...next } })
  const durationField = (
    <label className={styles.thresholdItem}>
      <span className={`${styles.thresholdIcon} ${styles.durationIcon}`}>
        <ShieldAlert size={17} />
      </span>
      <span className={styles.thresholdCopy}>
        <strong>持续时间</strong>
        <small>5 秒 - 10 分钟</small>
      </span>
      <InputNumber
        aria-label="持续超限时间"
        min={5}
        max={600}
        value={value.durationSeconds}
        disabled={!value.enabled}
        addonAfter="秒"
        onChange={(next) => patch({ durationSeconds: Number(next || 5) })}
      />
    </label>
  )
  const systemThresholdFields = (
    <>
      <label className={styles.thresholdItem}>
        <span className={`${styles.thresholdIcon} ${styles.cpuIcon}`}>
          <Cpu size={17} />
        </span>
        <span className={styles.thresholdCopy}>
          <strong>系统 CPU 使用率</strong>
          <small>1% - 100%</small>
        </span>
        <InputNumber
          aria-label="系统 CPU 熔断阈值"
          min={1}
          max={100}
          value={systemPressure.maxCpuPercent}
          disabled={!value.enabled || (!systemOnly && !systemPressure.enabled)}
          addonAfter="%"
          onChange={(next) => patchSystemPressure({ maxCpuPercent: Number(next || 1) })}
        />
      </label>
      <label className={styles.thresholdItem}>
        <span className={`${styles.thresholdIcon} ${styles.memoryIcon}`}>
          <MemoryStick size={17} />
        </span>
        <span className={styles.thresholdCopy}>
          <strong>系统内存使用率</strong>
          <small>1% - 100%</small>
        </span>
        <InputNumber
          aria-label="系统内存熔断阈值"
          min={1}
          max={100}
          value={systemPressure.maxMemoryPercent}
          disabled={!value.enabled || (!systemOnly && !systemPressure.enabled)}
          addonAfter="%"
          onChange={(next) => patchSystemPressure({ maxMemoryPercent: Number(next || 1) })}
        />
      </label>
    </>
  )
  return (
    <div className={styles.tabContent}>
      <div className={styles.sectionHeading}>
        <Typography.Text strong>{title}</Typography.Text>
        <Switch
          checked={value.enabled}
          checkedChildren="启用"
          unCheckedChildren="停用"
          onChange={(enabled) =>
            patch({
              enabled,
              ...(systemOnly
                ? {
                    maxCpuCores: 0,
                    maxMemoryBytes: 0,
                    systemPressure: { ...systemPressure, enabled },
                  }
                : {}),
            })
          }
        />
      </div>
      {systemOnly ? (
        <div className={styles.thresholdGrid}>
          {systemThresholdFields}
          {durationField}
        </div>
      ) : (
        <div className={styles.thresholdGrid}>
          <label className={styles.thresholdItem}>
            <span className={`${styles.thresholdIcon} ${styles.cpuIcon}`}>
              <Cpu size={17} />
            </span>
            <span className={styles.thresholdCopy}>
              <strong>CPU 配额</strong>
              <small>0.1 - 256 Core</small>
            </span>
            <InputNumber
              aria-label="CPU 熔断阈值"
              min={0.1}
              max={256}
              step={0.1}
              precision={1}
              value={value.maxCpuCores}
              disabled={!value.enabled}
              addonAfter="Core"
              onChange={(next) => patch({ maxCpuCores: Number(next || 0.1) })}
            />
          </label>
          <label className={styles.thresholdItem}>
            <span className={`${styles.thresholdIcon} ${styles.memoryIcon}`}>
              <MemoryStick size={17} />
            </span>
            <span className={styles.thresholdCopy}>
              <strong>内存占用</strong>
              <small>64 MiB - 64 GiB</small>
            </span>
            <InputNumber
              aria-label="内存熔断阈值"
              min={64}
              max={65536}
              step={64}
              value={toMemoryMiB(value.maxMemoryBytes)}
              disabled={!value.enabled}
              addonAfter="MiB"
              onChange={(next) => patch({ maxMemoryBytes: fromMemoryMiB(Number(next || 64)) })}
            />
          </label>
          {durationField}
        </div>
      )}
      {showSystemPressure && !systemOnly && (
        <div className={styles.systemPressureSection}>
          <div className={styles.sectionHeading}>
            <div>
              <span className={`${styles.thresholdIcon} ${styles.systemIcon}`}>
                <Server size={17} />
              </span>
              <Typography.Text strong>系统资源保护</Typography.Text>
            </div>
            <Switch
              checked={systemPressure.enabled}
              checkedChildren="启用"
              unCheckedChildren="停用"
              disabled={!value.enabled}
              onChange={(enabled) => patchSystemPressure({ enabled })}
            />
          </div>
          <div className={`${styles.thresholdGrid} ${styles.systemThresholdGrid}`}>
            {systemThresholdFields}
          </div>
        </div>
      )}
      <div className={styles.recoveryRow}>
        <div className={styles.recoveryToggle}>
          <span className={`${styles.thresholdIcon} ${styles.recoveryIcon}`}>
            <RotateCcw size={17} />
          </span>
          <Typography.Text strong>自动恢复</Typography.Text>
          <Switch
            checked={value.autoRecover}
            disabled={!value.enabled}
            onChange={(autoRecover) => patch({ autoRecover })}
          />
        </div>
        <label>
          <span>首次重试间隔</span>
          <InputNumber
            aria-label="首次恢复重试间隔"
            min={5}
            max={600}
            value={value.recoveryInitialDelaySeconds}
            disabled={!value.enabled || !value.autoRecover}
            addonAfter="秒"
            onChange={(next) => {
              const recoveryInitialDelaySeconds = Number(next || 5)
              patch({
                recoveryInitialDelaySeconds,
                recoveryMaxDelaySeconds: Math.max(
                  recoveryInitialDelaySeconds,
                  value.recoveryMaxDelaySeconds,
                ),
              })
            }}
          />
        </label>
        <label>
          <span>最长重试间隔</span>
          <InputNumber
            aria-label="最长恢复重试间隔"
            min={value.recoveryInitialDelaySeconds}
            max={3600}
            value={value.recoveryMaxDelaySeconds}
            disabled={!value.enabled || !value.autoRecover}
            addonAfter="秒"
            onChange={(next) => patch({ recoveryMaxDelaySeconds: Number(next || 3600) })}
          />
        </label>
      </div>
    </div>
  )
}

export default function ProbeConfigPage() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const nodeId = searchParams.get('nodeId') || ''
  const initialCluster = searchParams.get('clusterId') || ''
  const initialClusterEnvironment = searchParams.get('clusterType') || ''
  const isGlobal = searchParams.get('scope') === 'global'
  const isNode = Boolean(nodeId)
  const mode = isGlobal ? 'global' : isNode ? 'node' : 'batch'
  const [nodes, setNodes] = useState<AgentNode[]>([])
  const [states, setStates] = useState<ProbeState[]>([])
  const [templates, setTemplates] = useState<AgentTemplate[]>([])
  const [globalClusterType, setGlobalClusterType] = useState<ClusterType>(
    normalizeClusterType(initialClusterEnvironment),
  )
  const [selectedProbe, setSelectedProbe] = useState('')
  const [selectedTemplateId, setSelectedTemplateId] = useState('')
  const [selectedClusters, setSelectedClusters] = useState<string[]>(
    initialCluster ? [initialCluster] : [],
  )
  const [selectedBatchTemplateId, setSelectedBatchTemplateId] = useState('')
  const [selectedBatchFileKey, setSelectedBatchFileKey] = useState('')
  const [batchDraftConfigs, setBatchDraftConfigs] = useState<TemplateProbeConfig[]>([])
  const [saveBatchTemplateOpen, setSaveBatchTemplateOpen] = useState(false)
  const [batchTemplateCopyName, setBatchTemplateCopyName] = useState('')
  const [batchNodeConfigs, setBatchNodeConfigs] = useState<Record<string, CurrentConfig[]>>({})
  const [batchConfigsLoading, setBatchConfigsLoading] = useState(false)
  const [currentConfigs, setCurrentConfigs] = useState<CurrentConfig[]>([])
  const [configDrafts, setConfigDrafts] = useState<Record<string, string>>({})
  const [configFileDrafts, setConfigFileDrafts] = useState<Record<string, ConfigFile[]>>({})
  const [selectedFilePath, setSelectedFilePath] = useState('')
  const [historyConfigs, setHistoryConfigs] = useState<CurrentConfig[]>([])
  const [selectedHistoryGeneration, setSelectedHistoryGeneration] = useState<number>()
  const [templateDraftConfigs, setTemplateDraftConfigs] = useState<TemplateProbeConfig[]>([])
  const [selectedTemplateProbe, setSelectedTemplateProbe] = useState('')
  const [selectedTemplateFileIndex, setSelectedTemplateFileIndex] = useState(0)
  const [circuit, setCircuit] = useState<CircuitPolicy>(DEFAULT_CIRCUIT)
  const [circuitSource, setCircuitSource] = useState('')
  const [circuitComponent, setCircuitComponent] = useState('')
  const [activeTab, setActiveTab] = useState(
    searchParams.get('section') === 'logs' ? 'logs' : 'config',
  )
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [records, setRecords] = useState<DeliveryRecord[]>([])
  const importTemplateInputRef = useRef<HTMLInputElement>(null)

  const loadBase = useCallback(async () => {
    setLoading(true)
    try {
      const [nodeResult, stateResult, templateResult] = await Promise.all([
        getAgentConfigNodesApi(),
        getAgentConfigStatesApi(),
        getAgentConfigTemplatesApi(),
      ])
      setNodes((nodeResult as any)?.nodes || [])
      setStates((stateResult as any)?.states || [])
      setTemplates(
        (((templateResult as any)?.templates || []) as AgentTemplate[]).map((template) => ({
          ...template,
          clusterType: normalizeClusterType(template.clusterType),
        })),
      )
    } catch (error: any) {
      notify({ type: 'error', message: error.message || '加载节点配置失败' })
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    loadBase()
  }, [loadBase])

  useEffect(() => {
    if (mode === 'global' && activeTab === 'logs') setActiveTab('config')
  }, [activeTab, mode])

  useEffect(() => {
    if (mode !== 'global' && searchParams.get('section') === 'logs') setActiveTab('logs')
  }, [mode, searchParams])

  useEffect(() => {
    if (mode !== 'batch') return
    setSelectedClusters(initialCluster ? [initialCluster] : [])
  }, [initialCluster, mode])

  const availableProbeTypes = useMemo(() => {
    if (isNode) return currentConfigs.map((config) => config.probeType).sort()
    const values = new Set<string>()
    states.forEach((state) => {
      values.add(state.probeType)
    })
    return [...values].sort()
  }, [currentConfigs, isNode, states, templates])

  const circuitComponentTypes = useMemo(() => {
    const values = new Set<string>()
    states.forEach((state) => {
      if (mode !== 'node' || state.nodeId === nodeId) values.add(state.probeType)
    })
    if (mode === 'node') {
      currentConfigs.forEach((config) => values.add(config.probeType))
    }
    return [...values].sort()
  }, [currentConfigs, mode, nodeId, states])

  useEffect(() => {
    setSelectedProbe((current) =>
      availableProbeTypes.includes(current) ? current : availableProbeTypes[0] || '',
    )
  }, [availableProbeTypes])

  useEffect(() => {
    setCircuitComponent((current) =>
      circuitComponentTypes.includes(current) ? current : circuitComponentTypes[0] || '',
    )
  }, [circuitComponentTypes])

  const nodeClusterType = normalizeClusterType(
    nodes.find((node) => node.id === nodeId)?.environment || initialClusterEnvironment,
  )
  const selectableTemplates = useMemo(
    () =>
      mode === 'global'
        ? templates.filter((template) => templateClusterType(template) === globalClusterType)
        : mode === 'node'
          ? templates.filter((template) => templateClusterType(template) === nodeClusterType)
          : templates,
    [globalClusterType, mode, nodeClusterType, templates],
  )
  const selectedTemplate = selectableTemplates.find(
    (template) => template.id === selectedTemplateId,
  )

  useEffect(() => {
    setSelectedTemplateId((current) =>
      selectableTemplates.some((template) => template.id === current)
        ? current
        : selectableTemplates[0]?.id || '',
    )
  }, [selectableTemplates])

  useEffect(() => {
    const configs = configsForTemplate(selectedTemplate).map((config) => ({
      ...config,
      files: config.files.map((file) => ({ ...file })),
    }))
    setTemplateDraftConfigs(configs)
    setSelectedTemplateProbe(configs[0]?.probeType || '')
    setSelectedTemplateFileIndex(0)
  }, [selectedTemplateId, selectedTemplate?.version])

  useEffect(() => {
    const loadDetail = async () => {
      try {
        if (mode === 'node') {
          const [configResult, circuitResult] = await Promise.all([
            getNodeCurrentConfigsApi(nodeId),
            getNodeCircuitBreakerApi(nodeId),
          ])
          const configs = ((configResult as any)?.configs || []) as CurrentConfig[]
          setCurrentConfigs(configs)
          setConfigFileDrafts({})
          setConfigDrafts(
            Object.fromEntries(
              configs.flatMap((config) =>
                filesForConfig(config).map((file) => [
                  draftKey(config.probeType, file.path),
                  file.content,
                ]),
              ),
            ),
          )
          setCircuit(normalizeCircuitPolicy((circuitResult as any)?.circuitBreaker))
          setCircuitSource((circuitResult as any)?.source || '')
        } else {
          setCircuit(normalizeCircuitPolicy((await getGlobalCircuitBreakerApi()) as any))
          setCircuitSource('GLOBAL_DEFAULT')
        }
      } catch (error: any) {
        notify({ type: 'error', message: error.message || '加载当前配置失败' })
      }
    }
    loadDetail()
  }, [mode, nodeId])

  const selectedCurrentConfig = currentConfigs.find((config) => config.probeType === selectedProbe)
  const filesForDraft = (config: CurrentConfig) =>
    configFileDrafts[config.probeType] || filesForConfig(config)
  const selectedFiles = selectedCurrentConfig ? filesForDraft(selectedCurrentConfig) : []
  const selectedFileSetKey = selectedFiles.map((file) => file.path).join('\u0000')
  const nodeFileCount = currentConfigs.reduce(
    (count, config) => count + filesForDraft(config).length,
    0,
  )
  const currentNode = nodes.find((node) => node.id === nodeId)
  const inheritedClusterId = currentNode?.clusterId || initialCluster
  const inheritedClusterEnvironment =
    currentNode?.environment || initialClusterEnvironment || 'kubernetes'
  const nodeProbeTypes = [
    ...new Set([
      ...states.filter((state) => state.nodeId === nodeId).map((state) => state.probeType),
      ...currentConfigs.map((config) => config.probeType),
    ]),
  ]
  const javaLogScanProbeType =
    nodeProbeTypes.find((probeType) =>
      ['ilogtail', 'loongcollector'].some((name) => probeType.toLowerCase().includes(name)),
    ) || nodeProbeTypes[0]
  const reportedVariables = currentNode?.reportedVariables || currentNode?.labels || {}
  const reportedVariableEntries = Object.entries(reportedVariables).sort(([left], [right]) =>
    left.localeCompare(right),
  )

  useEffect(() => {
    setSelectedFilePath((current) =>
      selectedFiles.some((file) => file.path === current) ? current : selectedFiles[0]?.path || '',
    )
  }, [selectedProbe, selectedCurrentConfig?.generation, selectedFileSetKey])

  useEffect(() => {
    if (mode !== 'node' || !selectedProbe) {
      setHistoryConfigs([])
      setSelectedHistoryGeneration(undefined)
      return
    }
    getNodeConfigHistoryApi(nodeId, selectedProbe)
      .then((result: any) => {
        const configs = (result?.configs || []) as CurrentConfig[]
        setHistoryConfigs(configs)
        setSelectedHistoryGeneration(configs[0]?.generation)
      })
      .catch((error: any) => {
        console.error('[ProbeConfigPage] failed to load configuration history', error)
        setHistoryConfigs([])
      })
  }, [mode, nodeId, selectedProbe, selectedCurrentConfig?.generation])

  const clusterOptions = useMemo(
    () => [...new Set(nodes.map((node) => node.clusterId).filter(Boolean) as string[])].sort(),
    [nodes],
  )
  const selectedClusterTypes = useMemo(() => {
    const values = new Set<ClusterType>()
    selectedClusters.forEach((clusterId) => {
      const clusterNodes = nodes.filter((node) => node.clusterId === clusterId)
      if (clusterNodes.length) {
        clusterNodes.forEach((node) => values.add(normalizeClusterType(node.environment)))
      } else if (clusterId === initialCluster) {
        values.add(normalizeClusterType(initialClusterEnvironment))
      }
    })
    return [...values]
  }, [initialCluster, initialClusterEnvironment, nodes, selectedClusters])
  const selectedClusterType =
    selectedClusterTypes.length === 1 ? selectedClusterTypes[0] : undefined
  const selectedClusterEnvironment = selectedClusterType || ''
  const hasMixedClusterTypes = selectedClusterTypes.length > 1
  const batchTemplates = useMemo(
    () =>
      selectedClusterType
        ? templates.filter((template) => templateClusterType(template) === selectedClusterType)
        : [],
    [selectedClusterType, templates],
  )
  const targetNodes = useMemo(
    () => nodes.filter((node) => selectedClusters.includes(node.clusterId || '')),
    [nodes, selectedClusters],
  )
  const targetCount = targetNodes.length
  useEffect(() => {
    if (mode !== 'batch' || !targetNodes.length) {
      setBatchNodeConfigs({})
      setBatchConfigsLoading(false)
      return
    }
    let cancelled = false
    setBatchConfigsLoading(true)
    Promise.all(
      targetNodes.map(async (node) => {
        const result = (await getNodeCurrentConfigsApi(node.id)) as any
        return [node.id, (result?.configs || []) as CurrentConfig[]] as const
      }),
    )
      .then((entries) => {
        if (!cancelled) setBatchNodeConfigs(Object.fromEntries(entries))
      })
      .catch((error: any) => {
        console.error('[ProbeConfigPage] failed to load batch node configurations', error)
        if (!cancelled) {
          setBatchNodeConfigs({})
          notify({ type: 'error', message: error.message || '读取节点当前配置失败' })
        }
      })
      .finally(() => {
        if (!cancelled) setBatchConfigsLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [mode, targetNodes])

  const batchProbeTypes = useMemo(() => {
    if (mode !== 'batch' || !targetNodes.length) return []
    return [
      ...new Set(
        targetNodes.flatMap((node) =>
          (batchNodeConfigs[node.id] || []).map((config) => config.probeType),
        ),
      ),
    ].sort()
  }, [batchNodeConfigs, mode, targetNodes])
  const selectedBatchTemplate = batchTemplates.find(
    (template) => template.id === selectedBatchTemplateId,
  )
  const batchFiles = useMemo<BatchPreviewFile[]>(
    () =>
      batchDraftConfigs.flatMap((config) =>
        config.files.map((file, index) => ({
          ...file,
          key: `${config.probeType}\u0000${file.path || index}`,
          probeType: config.probeType,
        })),
      ),
    [batchDraftConfigs],
  )
  const selectedBatchFile = batchFiles.find((file) => file.key === selectedBatchFileKey)
  const batchTemplateProbeTypes = batchDraftConfigs.map((config) => config.probeType)
  const batchTemplateChanged =
    JSON.stringify(
      batchDraftConfigs.map((config) => ({
        probeType: config.probeType,
        files: config.files.map((file) => ({ path: file.path, content: file.content })),
      })),
    ) !==
    JSON.stringify(
      configsForTemplate(selectedBatchTemplate).map((config) => ({
        probeType: config.probeType,
        files: config.files.map((file) => ({ path: file.path, content: file.content })),
      })),
    )
  const batchTemplateComplete = Boolean(
    selectedBatchTemplate &&
      selectedClusterType === templateClusterType(selectedBatchTemplate) &&
      targetNodes.length &&
      targetNodes.every((node) => {
        const current = batchNodeConfigs[node.id] || []
        const templateConfigs = batchDraftConfigs
        if (current.length !== templateConfigs.length) return false
        const currentByProbe = new Map(current.map((config) => [config.probeType, config]))
        return templateConfigs.every((config) => {
          const registered = currentByProbe.get(config.probeType)
          if (!registered) return false
          const paths = new Set(filesForConfig(registered).map((file) => file.path))
          return (
            config.files.length === paths.size && config.files.every((file) => paths.has(file.path))
          )
        })
      }),
  )

  useEffect(() => {
    if (mode !== 'batch') return
    setSelectedBatchTemplateId((current) =>
      batchTemplates.some((template) => template.id === current)
        ? current
        : batchTemplates[0]?.id || '',
    )
  }, [batchTemplates, mode])

  useEffect(() => {
    if (mode !== 'batch') return
    setBatchDraftConfigs(
      configsForTemplate(selectedBatchTemplate).map((config) => ({
        ...config,
        files: config.files.map((file) => ({ ...file })),
      })),
    )
  }, [mode, selectedBatchTemplateId, selectedBatchTemplate?.version])

  useEffect(() => {
    setSelectedBatchFileKey((current) =>
      batchFiles.some((file) => file.key === current) ? current : batchFiles[0]?.key || '',
    )
  }, [batchFiles])

  const nodeConfigChanged = currentConfigs.some((config) => {
    const original = filesForConfig(config)
    const draft = filesForDraft(config)
    const originalByPath = new Map(original.map((file) => [file.path, file.content]))
    return (
      original.length !== draft.length ||
      draft.some(
        (file) =>
          !originalByPath.has(file.path) ||
          (configDrafts[draftKey(config.probeType, file.path)] ?? file.content) !==
            originalByPath.get(file.path),
      )
    )
  })

  const addRecord = (kind: DeliveryRecord['kind'], target: string, result: any) => {
    setRecords((current) => [
      {
        id: `${kind}-${Date.now()}`,
        kind,
        target,
        succeeded: Number(result?.succeededCount || 1),
        total: Number(result?.targetNodeCount || 1),
        createdAt: new Date().toLocaleString(),
      },
      ...current,
    ])
  }

  const saveNodeConfig = async () => {
    if (!currentConfigs.length) return
    setSaving(true)
    try {
      const result: any = await updateNodeCurrentConfigsApi(nodeId, {
        configs: currentConfigs.map((config) => ({
          probeType: config.probeType,
          files: filesForDraft(config).map((file) => ({
            path: file.path,
            content: configDrafts[draftKey(config.probeType, file.path)] ?? file.content,
          })),
          expectedGeneration: config.generation,
        })),
      })
      const configs = (result?.configs || []) as CurrentConfig[]
      setCurrentConfigs(configs)
      setConfigFileDrafts({})
      setConfigDrafts(
        Object.fromEntries(
          configs.flatMap((config) =>
            filesForConfig(config).map((file) => [
              draftKey(config.probeType, file.path),
              file.content,
            ]),
          ),
        ),
      )
      addRecord('CONFIG', nodeId, { succeededCount: 1, targetNodeCount: 1 })
      notify({ type: 'success', message: '节点配置已保存并下发' })
    } catch (error: any) {
      notify({ type: 'error', message: error.message || '节点配置保存失败' })
    } finally {
      setSaving(false)
    }
  }

  const loadHistory = () => {
    const history = historyConfigs.find((config) => config.generation === selectedHistoryGeneration)
    if (!history || !selectedCurrentConfig) return
    const historyFiles = filesForConfig(history)
    setConfigFileDrafts((current) => ({ ...current, [selectedProbe]: historyFiles }))
    setConfigDrafts((current) => ({
      ...current,
      ...Object.fromEntries(
        historyFiles.map((file) => [draftKey(selectedProbe, file.path), file.content]),
      ),
    }))
    notify({ type: 'success', message: `已载入版本 ${history.version}` })
  }

  const loadTemplateIntoNode = () => {
    if (!selectedTemplate) return
    const templateConfigs = configsForTemplate(selectedTemplate)
    const currentByProbe = new Map(currentConfigs.map((config) => [config.probeType, config]))
    const complete =
      templateConfigs.length === currentConfigs.length &&
      templateConfigs.every((config) => {
        const current = currentByProbe.get(config.probeType)
        if (!current) return false
        const registeredPaths = new Set(filesForConfig(current).map((file) => file.path))
        return (
          config.files.length === registeredPaths.size &&
          config.files.every((file) => registeredPaths.has(file.path))
        )
      })
    if (!complete) {
      notify({ type: 'error', message: '配置模版与当前节点注册的配置文件不一致，无法载入' })
      return
    }
    setConfigFileDrafts(
      Object.fromEntries(templateConfigs.map((config) => [config.probeType, config.files])),
    )
    setConfigDrafts((current) => ({
      ...current,
      ...Object.fromEntries(
        templateConfigs.flatMap((config) =>
          config.files.map((file) => [draftKey(config.probeType, file.path), file.content]),
        ),
      ),
    }))
    notify({ type: 'success', message: `配置模版“${selectedTemplate.name}”已载入当前节点` })
  }

  const saveTemplate = async () => {
    if (!selectedTemplate) return
    setSaving(true)
    try {
      await updateAgentConfigTemplateApi(selectedTemplate.id, {
        name: selectedTemplate.name,
        description: selectedTemplate.description,
        clusterType: templateClusterType(selectedTemplate),
        variables: selectedTemplate.variables,
        configs: templateDraftConfigs.map((config) => ({
          probeType: config.probeType,
          files: config.files.map((file) => ({ path: file.path.trim(), content: file.content })),
        })),
      })
      await loadBase()
      notify({ type: 'success', message: '配置模版已保存' })
    } catch (error: any) {
      notify({ type: 'error', message: error.message || '配置模版保存失败' })
    } finally {
      setSaving(false)
    }
  }

  const updateBatchFileContent = (content: string) => {
    if (!selectedBatchFile) return
    setBatchDraftConfigs((current) =>
      current.map((config) =>
        config.probeType === selectedBatchFile.probeType
          ? {
              ...config,
              files: config.files.map((file) =>
                file.path === selectedBatchFile.path ? { ...file, content } : file,
              ),
            }
          : config,
      ),
    )
  }

  const openSaveBatchTemplate = () => {
    if (!selectedBatchTemplate || !batchTemplateChanged) return
    setBatchTemplateCopyName(`${selectedBatchTemplate.name}-副本`)
    setSaveBatchTemplateOpen(true)
  }

  const saveBatchTemplateCopy = async () => {
    const name = batchTemplateCopyName.trim()
    if (!selectedBatchTemplate || !batchTemplateChanged || !name) return
    setSaving(true)
    try {
      const created = (await createAgentConfigTemplateApi({
        name,
        description: selectedBatchTemplate.description,
        clusterType: templateClusterType(selectedBatchTemplate),
        variables: selectedBatchTemplate.variables,
        configs: batchDraftConfigs.map((config) => ({
          probeType: config.probeType,
          files: config.files.map((file) => ({ path: file.path, content: file.content })),
        })),
      })) as any
      await loadBase()
      setSelectedBatchTemplateId(created.id)
      setSaveBatchTemplateOpen(false)
      notify({ type: 'success', message: `已保存为新模版“${created.name || name}”` })
    } catch (error: any) {
      notify({ type: 'error', message: error.message || '新模版保存失败' })
    } finally {
      setSaving(false)
    }
  }

  const dispatchTemplate = async () => {
    if (batchTemplateChanged) {
      notify({ type: 'warning', message: '请先将修改保存为新模版' })
      return
    }
    if (!batchTemplateComplete || !selectedClusters.length) return
    setSaving(true)
    try {
      const result: any = await batchAssignAgentConfigApi({
        clusterIds: selectedClusters,
        templateId: selectedBatchTemplateId,
      })
      addRecord('CONFIG', selectedClusters.join(', '), result)
      notify({
        type: result.failedNodeIds?.length ? 'warning' : 'success',
        message: result.failedNodeIds?.length
          ? `${result.succeededCount} 个节点成功，${result.failedNodeIds.length} 个失败`
          : `完整配置已覆盖下发到 ${result.succeededCount} 个节点`,
      })
    } catch (error: any) {
      notify({ type: 'error', message: error.message || '批量配置下发失败' })
    } finally {
      setSaving(false)
    }
  }

  const saveCircuit = async () => {
    setSaving(true)
    try {
      let result: any
      let target = '全局默认'
      if (mode === 'global') {
        result = await updateGlobalCircuitBreakerApi(circuit)
      } else if (mode === 'node') {
        result = await updateNodeCircuitBreakerApi(nodeId, circuit)
        target = nodeId
        setCircuitSource('NODE')
      } else {
        result = await batchUpdateCircuitBreakerApi({
          clusterIds: selectedClusters,
          circuitBreaker: clusterCircuitPolicy(circuit),
        })
        target = selectedClusters.join(', ')
      }
      addRecord('CIRCUIT', target, result)
      notify({
        type: 'success',
        message: mode === 'global' ? '全局默认熔断策略已保存' : '熔断策略已独立下发',
      })
    } catch (error: any) {
      notify({ type: 'error', message: error.message || '熔断策略保存失败' })
    } finally {
      setSaving(false)
    }
  }

  const importTemplate = async (event: ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0]
    if (!file) return
    setSaving(true)
    try {
      const imported = (await importAgentConfigTemplateApi({ yaml: await file.text() })) as any
      await loadBase()
      setGlobalClusterType(normalizeClusterType(imported?.clusterType))
      setSelectedTemplateId(imported?.id || '')
      notify({ type: 'success', message: `配置模版“${imported?.name || file.name}”已导入` })
    } catch (error: any) {
      notify({ type: 'error', message: error.message || '配置模版导入失败' })
    } finally {
      event.target.value = ''
      setSaving(false)
    }
  }

  const selectedFile = selectedFiles.find((file) => file.path === selectedFilePath)
  const selectedOriginalFile = selectedCurrentConfig
    ? filesForConfig(selectedCurrentConfig).find((file) => file.path === selectedFilePath)
    : undefined
  const selectedDraft = selectedFile
    ? configDrafts[draftKey(selectedProbe, selectedFile.path)] ?? selectedFile.content
    : ''
  const selectedLanguage = languageForPath(selectedFilePath)
  const selectedTemplateConfig = templateDraftConfigs.find(
    (config) => config.probeType === selectedTemplateProbe,
  )
  const selectedTemplateFile = selectedTemplateConfig?.files[selectedTemplateFileIndex]
  const templateFileCount = templateDraftConfigs.reduce(
    (total, config) => total + config.files.length,
    0,
  )
  const templateChanged =
    JSON.stringify(
      templateDraftConfigs.map((config) => ({
        probeType: config.probeType,
        files: config.files.map((file) => ({ path: file.path, content: file.content })),
      })),
    ) !==
    JSON.stringify(
      configsForTemplate(selectedTemplate).map((config) => ({
        probeType: config.probeType,
        files: config.files.map((file) => ({ path: file.path, content: file.content })),
      })),
    )
  const updateTemplateFile = (patch: Partial<ConfigFile>) => {
    setTemplateDraftConfigs((current) =>
      current.map((config) =>
        config.probeType === selectedTemplateProbe
          ? {
              ...config,
              files: config.files.map((file, index) =>
                index === selectedTemplateFileIndex ? { ...file, ...patch } : file,
              ),
            }
          : config,
      ),
    )
  }
  const hasComponentCircuitPolicy = Boolean(
    circuitComponent && circuit.componentPolicies?.[circuitComponent],
  )
  const selectedComponentPolicy = concreteCircuitPolicy(
    circuit.componentPolicies?.[circuitComponent] || circuit,
  )
  const updateNodeCircuitPolicy = (next: CircuitPolicy) => {
    setCircuit((current) => ({
      ...concreteCircuitPolicy(next),
      componentPolicies: current.componentPolicies,
    }))
  }
  const updateComponentCircuitPolicy = (next: CircuitPolicy) => {
    if (!circuitComponent) return
    const { systemPressure: _systemPressure, ...componentPolicy } = concreteCircuitPolicy(next)
    setCircuit((current) => ({
      ...current,
      componentPolicies: {
        ...current.componentPolicies,
        [circuitComponent]: componentPolicy,
      },
    }))
  }
  const removeComponentCircuitPolicy = () => {
    if (!circuitComponent) return
    setCircuit((current) => {
      const componentPolicies = { ...current.componentPolicies }
      delete componentPolicies[circuitComponent]
      return { ...current, componentPolicies }
    })
  }

  const isClusterConfig = mode === 'batch' && Boolean(initialCluster)
  const title =
    mode === 'global'
      ? '默认配置'
      : mode === 'node'
        ? `节点配置 · ${nodeId}`
        : isClusterConfig
          ? `集群配置 · ${initialCluster}`
          : '批量配置'
  const subtitle =
    mode === 'global'
      ? '管理配置模版和新节点默认使用的熔断策略'
      : mode === 'node'
        ? '修改当前节点配置'
        : isClusterConfig
          ? '管理当前集群的配置文件、日志采集和熔断策略'
          : '选择目标集群和配置方案，将完整配置统一下发至集群节点'

  const configTab =
    mode === 'batch' ? (
      <div className={styles.tabContent}>
        <div className={styles.batchTargetRow}>
          <label className={styles.sourceGroup}>
            <span>目标集群</span>
            <Select
              mode="multiple"
              value={selectedClusters}
              onChange={setSelectedClusters}
              options={clusterOptions.map((value) => ({ value, label: value }))}
            />
          </label>
        </div>
        <div className={styles.batchTemplateSection}>
          <div className={styles.sectionHeading}>
            <div>
              <Typography.Text strong>配置方案</Typography.Text>
              <Typography.Text type="secondary">
                {targetCount ? `${targetCount} 个节点的完整配置` : '选择集群后生成完整配置方案'}
              </Typography.Text>
            </div>
            <Tag>{batchTemplateProbeTypes.length} 个组件</Tag>
          </div>
          {hasMixedClusterTypes && (
            <Typography.Text type="danger">
              Kubernetes 与 VM 集群使用不同模版，请按同一类别分别下发。
            </Typography.Text>
          )}
          <Select
            style={{ width: '100%' }}
            value={selectedBatchTemplateId || undefined}
            placeholder={
              hasMixedClusterTypes
                ? '不能为混合类别集群选择模版'
                : selectedClusterType
                  ? `选择 ${clusterTypeLabel(selectedClusterType)} 配置模版`
                  : '先选择目标集群'
            }
            onChange={setSelectedBatchTemplateId}
            disabled={!selectedClusterType}
            options={batchTemplates.map((template) => ({
              value: template.id,
              label: templateLabel(template),
            }))}
          />
        </div>
        <Row className={styles.configWorkspace} wrap={false}>
          <Col flex="280px" className={styles.filePanel}>
            <div className={styles.filePanelTitle}>
              <FileText size={16} />
              <Typography.Text strong>配置文件</Typography.Text>
              <Tag>{batchFiles.length}</Tag>
            </div>
            <Tree
              key={batchProbeTypes.join(':')}
              className={styles.probeFileTree}
              blockNode
              defaultExpandAll
              showLine={{ showLeafIcon: false }}
              selectedKeys={selectedBatchFileKey ? [selectedBatchFileKey] : []}
              treeData={batchTemplateProbeTypes.map((probeType) => {
                const files = batchFiles.filter((file) => file.probeType === probeType)
                return {
                  key: `probe:${probeType}`,
                  selectable: false,
                  title: (
                    <span className={styles.probeTreeTitle}>
                      <strong>{probeType}</strong>
                      <small>{files.length} 个文件</small>
                    </span>
                  ),
                  children: files.map((file) => ({
                    key: file.key,
                    title: (
                      <span className={styles.fileTreeTitle}>
                        <span>{file.path || '默认配置文件'}</span>
                      </span>
                    ),
                  })),
                }
              })}
              onSelect={(keys) => setSelectedBatchFileKey(String(keys[0] || ''))}
            />
          </Col>
          <Col flex="1" className={styles.templateEditorPanel}>
            <Input value={selectedBatchFile?.path || ''} placeholder="配置文件相对路径" readOnly />
            <div className={styles.templateEditor}>
              <Editor
                key={selectedBatchFile?.key || 'batch-empty'}
                value={selectedBatchFile?.content || ''}
                language={languageForPath(selectedBatchFile?.path || '')}
                theme="vs"
                onChange={(value) => updateBatchFileContent(value || '')}
                options={{
                  readOnly: !selectedBatchFile,
                  minimap: { enabled: false },
                  automaticLayout: true,
                  wordWrap: 'on',
                  scrollBeyondLastLine: false,
                }}
              />
            </div>
          </Col>
        </Row>
        <div className={styles.editorFooter}>
          <Typography.Text type="secondary">
            {selectedClusters.length} 个集群 · {targetCount} 个节点 ·{' '}
            {batchTemplateProbeTypes.length} 项组件配置
          </Typography.Text>
          <Space>
            {batchTemplateChanged && <Tag color="warning">未保存修改</Tag>}
            <Button
              icon={<Save size={15} />}
              disabled={!selectedBatchTemplate || !batchTemplateChanged}
              onClick={openSaveBatchTemplate}
            >
              另存为新模版
            </Button>
            <Button
              type="primary"
              icon={<Send size={15} />}
              loading={saving}
              disabled={!targetCount || !batchTemplateComplete || batchTemplateChanged}
              onClick={dispatchTemplate}
            >
              覆盖下发到所选集群
            </Button>
          </Space>
        </div>
      </div>
    ) : (
      <div className={styles.tabContent}>
        {mode === 'global' && (
          <div className={styles.templateBar}>
            <Segmented
              value={globalClusterType}
              options={[
                { label: 'Kubernetes', value: 'kubernetes' },
                { label: 'VM', value: 'vm' },
              ]}
              onChange={(value) => setGlobalClusterType(value as ClusterType)}
            />
            <Select
              value={selectedTemplateId || undefined}
              placeholder={`暂无 ${clusterTypeLabel(globalClusterType)} 配置模版`}
              onChange={setSelectedTemplateId}
              options={selectableTemplates.map((template) => ({
                value: template.id,
                label: templateLabel(template),
              }))}
            />
            <input
              ref={importTemplateInputRef}
              type="file"
              accept=".yaml,.yml,application/yaml,text/yaml"
              hidden
              onChange={importTemplate}
            />
            <Button
              icon={<Upload size={15} />}
              loading={saving}
              onClick={() => importTemplateInputRef.current?.click()}
            >
              导入配置模版
            </Button>
          </div>
        )}
        {mode === 'node' && (
          <>
            <div className={styles.nodeVariablesSection}>
              <div className={styles.sectionHeading}>
                <div>
                  <Typography.Text strong>节点变量</Typography.Text>
                  <Typography.Text type="secondary">探针上报</Typography.Text>
                </div>
                <Tag>{reportedVariableEntries.length}</Tag>
              </div>
              {reportedVariableEntries.length ? (
                <div className={styles.nodeVariableList}>
                  {reportedVariableEntries.map(([name, value]) => (
                    <div className={styles.nodeVariableItem} key={name}>
                      <Typography.Text>{name}</Typography.Text>
                      <Typography.Text code>{value}</Typography.Text>
                    </div>
                  ))}
                </div>
              ) : (
                <Typography.Text type="secondary">暂无探针上报变量</Typography.Text>
              )}
            </div>
            <div className={styles.sourceToolbar}>
              <label className={styles.sourceGroup}>
                <span>当前组件历史版本</span>
                <Space.Compact>
                  <Select
                    value={selectedHistoryGeneration}
                    placeholder="暂无历史版本"
                    onChange={setSelectedHistoryGeneration}
                    options={historyConfigs.map((config) => ({
                      value: config.generation,
                      label: `版本 ${config.version}${config.generation === selectedCurrentConfig?.generation ? ' · 当前' : ''}`,
                    }))}
                  />
                  <Button
                    icon={<History size={15} />}
                    disabled={selectedHistoryGeneration === undefined}
                    onClick={loadHistory}
                  >
                    载入
                  </Button>
                </Space.Compact>
              </label>
              <label className={styles.sourceGroup}>
                <span>配置模版</span>
                <Space.Compact>
                  <Select
                    value={selectedTemplateId || undefined}
                    placeholder="暂无可用配置模版"
                    onChange={setSelectedTemplateId}
                    options={selectableTemplates.map((template) => ({
                      value: template.id,
                      label: templateLabel(template),
                    }))}
                  />
                  <Button disabled={!selectedTemplate} onClick={loadTemplateIntoNode}>
                    使用模版
                  </Button>
                </Space.Compact>
              </label>
            </div>
          </>
        )}
        {mode === 'node' ? (
          <Row className={styles.configWorkspace} wrap={false}>
            <Col flex="280px" className={styles.filePanel}>
              <div className={`${styles.filePanelTitle} ${styles.nodeFilePanelTitle}`}>
                <FileText size={16} />
                <div>
                  <Typography.Text strong>配置文件</Typography.Text>
                  <Typography.Text type="secondary">
                    {currentConfigs.length} 个组件 · {nodeFileCount} 个文件
                  </Typography.Text>
                </div>
              </div>
              <Tree
                key={currentConfigs
                  .flatMap((config) =>
                    filesForDraft(config).map((file) => `${config.probeType}:${file.path}`),
                  )
                  .join('|')}
                className={styles.probeFileTree}
                blockNode
                defaultExpandAll
                showLine={{ showLeafIcon: false }}
                selectedKeys={
                  selectedProbe && selectedFilePath
                    ? [draftKey(selectedProbe, selectedFilePath)]
                    : []
                }
                treeData={currentConfigs.map((config) => ({
                  key: `probe:${config.probeType}`,
                  selectable: false,
                  title: (
                    <span className={styles.probeTreeTitle}>
                      <strong>{config.probeType}</strong>
                    </span>
                  ),
                  children: filesForDraft(config).map((file) => {
                    const original = filesForConfig(config).find(
                      (currentFile) => currentFile.path === file.path,
                    )
                    const changed =
                      !original ||
                      (configDrafts[draftKey(config.probeType, file.path)] ?? file.content) !==
                        original.content
                    return {
                      key: draftKey(config.probeType, file.path),
                      probeType: config.probeType,
                      filePath: file.path,
                      title: (
                        <span className={styles.fileTreeTitle}>
                          <span>{file.path}</span>
                          {changed && <Tag color="warning">已修改</Tag>}
                        </span>
                      ),
                    }
                  }),
                }))}
                onSelect={(_, info) => {
                  const node = info.node as any
                  if (!node.probeType || !node.filePath) return
                  setSelectedProbe(node.probeType)
                  setSelectedFilePath(node.filePath)
                }}
              />
            </Col>
            <Col flex="1" className={styles.diffPanel}>
              <DiffEditor
                key={`${selectedProbe}:${selectedFilePath}`}
                original={selectedOriginalFile?.content || ''}
                modified={selectedDraft}
                language={selectedLanguage}
                theme="vs"
                onMount={(editor) => {
                  const modifiedEditor = editor.getModifiedEditor()
                  const listener = modifiedEditor.onDidChangeModelContent(() => {
                    setConfigDrafts((current) => ({
                      ...current,
                      [draftKey(selectedProbe, selectedFilePath)]: modifiedEditor.getValue(),
                    }))
                  })
                  editor.onDidDispose(() => listener.dispose())
                }}
                options={{
                  renderSideBySide: false,
                  originalEditable: false,
                  minimap: { enabled: false },
                  automaticLayout: true,
                  wordWrap: 'on',
                  scrollBeyondLastLine: false,
                }}
              />
            </Col>
          </Row>
        ) : (
          <Row className={styles.configWorkspace} wrap={false}>
            <Col flex="240px" className={styles.filePanel}>
              <div className={styles.filePanelTitle}>
                <FileText size={16} />
                <Typography.Text strong>组件配置</Typography.Text>
                <Tag>{templateFileCount}</Tag>
              </div>
              <Tree
                className={styles.probeFileTree}
                blockNode
                defaultExpandAll
                showLine={{ showLeafIcon: false }}
                selectedKeys={
                  selectedTemplateFile
                    ? [draftKey(selectedTemplateProbe, selectedTemplateFile.path)]
                    : []
                }
                treeData={templateDraftConfigs.map((config) => ({
                  key: `template-probe:${config.probeType}`,
                  selectable: false,
                  title: (
                    <span className={styles.probeTreeTitle}>
                      <strong>{config.probeType}</strong>
                      <small>{config.files.length} 个文件</small>
                    </span>
                  ),
                  children: config.files.map((file, fileIndex) => ({
                    key: draftKey(config.probeType, file.path),
                    probeType: config.probeType,
                    fileIndex,
                    title: <span className={styles.fileTreeTitle}>{file.path}</span>,
                  })),
                }))}
                onSelect={(_, info) => {
                  const node = info.node as any
                  if (!node.probeType || node.fileIndex === undefined) return
                  setSelectedTemplateProbe(node.probeType)
                  setSelectedTemplateFileIndex(node.fileIndex)
                }}
              />
            </Col>
            <Col flex="1" className={styles.templateEditorPanel}>
              <Input
                value={selectedTemplateFile?.path || ''}
                placeholder="配置文件相对路径"
                readOnly
              />
              <div className={styles.templateEditor}>
                <Editor
                  value={selectedTemplateFile?.content || ''}
                  language={languageForPath(selectedTemplateFile?.path || '')}
                  theme="vs"
                  onChange={(value) => updateTemplateFile({ content: value || '' })}
                  options={{
                    minimap: { enabled: false },
                    automaticLayout: true,
                    wordWrap: 'on',
                    scrollBeyondLastLine: false,
                  }}
                />
              </div>
            </Col>
          </Row>
        )}
        <div className={styles.editorFooter}>
          {mode === 'node' && (
            <Typography.Text type="secondary">
              {currentConfigs.length} 个组件 · {nodeFileCount} 个配置文件将一次下发
            </Typography.Text>
          )}
          <Button
            type="primary"
            icon={<Save size={15} />}
            loading={saving}
            disabled={mode === 'node' ? !nodeConfigChanged : !selectedTemplate || !templateChanged}
            onClick={mode === 'node' ? saveNodeConfig : saveTemplate}
          >
            {mode === 'node' ? '保存并下发到当前节点' : '保存配置模版'}
          </Button>
        </div>
      </div>
    )

  const logCollectionTab =
    mode === 'node' ? (
      <LogCollectionConfigPanel
        mode="host"
        nodeId={nodeId}
        inheritedClusterId={inheritedClusterId}
        inheritedClusterEnvironment={inheritedClusterEnvironment}
        enableJavaLogScan={!isK8sEnvironment(inheritedClusterEnvironment)}
        javaLogScanProbeType={javaLogScanProbeType}
        onEditInherited={() =>
          navigate(
            `/probe-management/config?clusterId=${encodeURIComponent(inheritedClusterId)}&clusterType=${encodeURIComponent(inheritedClusterEnvironment)}&section=logs`,
          )
        }
      />
    ) : mode === 'batch' ? (
      <LogCollectionConfigPanel
        mode="container"
        clusterIds={selectedClusters}
        clusterOptions={clusterOptions}
        showK8sFilters={isK8sEnvironment(selectedClusterEnvironment)}
        onClusterIdsChange={setSelectedClusters}
      />
    ) : null

  const circuitTab = (
    <div className={styles.tabContent}>
      {mode === 'node' && (
        <Tag color={circuitSource === 'NODE' ? 'processing' : 'default'}>
          {circuitSource === 'NODE' ? '节点最后下发策略' : '当前使用全局默认'}
        </Tag>
      )}
      {mode === 'batch' && (
        <div className={styles.batchGrid}>
          <label>
            <span>目标集群</span>
            <Select
              mode="multiple"
              value={selectedClusters}
              onChange={setSelectedClusters}
              options={clusterOptions.map((value) => ({ value, label: value }))}
            />
          </label>
        </div>
      )}
      <CircuitEditor
        title={mode === 'batch' ? '集群系统资源熔断' : '探针消耗资源熔断'}
        value={concreteCircuitPolicy(circuit)}
        onChange={updateNodeCircuitPolicy}
        showSystemPressure={mode !== 'batch'}
        systemOnly={mode === 'batch'}
      />
      {!!circuitComponentTypes.length && (
        <Collapse
          ghost
          className={styles.advancedCircuit}
          items={[
            {
              key: 'component-quotas',
              label: '高级配置',
              children: (
                <div>
                  <div className={styles.circuitScopeBar}>
                    <label>
                      <span>组件</span>
                      <Select
                        value={circuitComponent || undefined}
                        onChange={setCircuitComponent}
                        options={circuitComponentTypes.map((value) => ({ value, label: value }))}
                      />
                    </label>
                    {hasComponentCircuitPolicy && (
                      <Button icon={<Trash2 size={15} />} onClick={removeComponentCircuitPolicy}>
                        移除组件配额
                      </Button>
                    )}
                  </div>
                  {hasComponentCircuitPolicy ? (
                    <CircuitEditor
                      title={`组件资源配额 · ${circuitComponent}`}
                      value={selectedComponentPolicy}
                      onChange={updateComponentCircuitPolicy}
                    />
                  ) : (
                    <Button
                      type="primary"
                      icon={<Settings size={15} />}
                      disabled={!circuitComponent}
                      onClick={() => updateComponentCircuitPolicy(selectedComponentPolicy)}
                    >
                      设置组件配额
                    </Button>
                  )}
                </div>
              ),
            },
          ]}
        />
      )}
      <div className={styles.editorFooter}>
        <Button
          type="primary"
          icon={<ShieldAlert size={15} />}
          loading={saving}
          disabled={mode === 'batch' && !targetCount}
          onClick={saveCircuit}
        >
          {mode === 'global'
            ? '保存全局默认'
            : mode === 'node'
              ? '下发到当前节点'
              : `下发到 ${targetCount} 个节点`}
        </Button>
      </div>
    </div>
  )

  return (
    <div className={styles.page}>
      <Button
        type="link"
        icon={<LeftOutlined />}
        onClick={() => navigate('/probe-management?tab=management')}
      >
        探针管理
      </Button>
      <Card className={styles.card} loading={loading}>
        <div className={styles.header}>
          <div className={styles.titleBlock}>
            <span
              className={
                mode === 'global'
                  ? styles.globalScopeIcon
                  : mode === 'node'
                    ? styles.nodeScopeIcon
                    : styles.clusterScopeIcon
              }
            >
              {mode === 'global' ? (
                <Settings size={18} />
              ) : mode === 'node' ? (
                <Server size={18} />
              ) : (
                <Network size={18} />
              )}
            </span>
            <div>
              <Typography.Title level={3}>{title}</Typography.Title>
              <Typography.Text type="secondary">{subtitle}</Typography.Text>
            </div>
          </div>
        </div>
        <Tabs
          activeKey={activeTab}
          onChange={setActiveTab}
          items={[
            { key: 'config', label: '配置文件', children: configTab },
            ...(mode === 'global'
              ? []
              : [{ key: 'logs', label: '日志采集', children: logCollectionTab }]),
            { key: 'circuit', label: '熔断策略', children: circuitTab },
          ]}
        />
        {!!records.length && (
          <Table
            size="small"
            pagination={false}
            rowKey="id"
            dataSource={records}
            columns={[
              {
                title: '类型',
                dataIndex: 'kind',
                render: (kind) => (
                  <Tag>
                    {kind === 'CONFIG' ? '配置文件' : kind === 'LOG' ? '日志配置' : '熔断策略'}
                  </Tag>
                ),
              },
              { title: '目标', dataIndex: 'target' },
              { title: '结果', render: (_, record) => `${record.succeeded} / ${record.total}` },
              { title: '时间', dataIndex: 'createdAt' },
            ]}
          />
        )}
      </Card>
      <Modal
        title="另存为新模版"
        open={saveBatchTemplateOpen}
        okText="保存新模版"
        cancelText="取消"
        confirmLoading={saving}
        okButtonProps={{ disabled: !batchTemplateCopyName.trim() }}
        onOk={saveBatchTemplateCopy}
        onCancel={() => setSaveBatchTemplateOpen(false)}
      >
        <Input
          autoFocus
          maxLength={255}
          value={batchTemplateCopyName}
          placeholder="新模版名称"
          onChange={(event) => setBatchTemplateCopyName(event.target.value)}
          onPressEnter={saveBatchTemplateCopy}
        />
      </Modal>
    </div>
  )
}
