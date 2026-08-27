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
  ArrowRight,
  Cpu,
  FileText,
  History,
  MemoryStick,
  Network,
  Pencil,
  Plus,
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
  batchUpdateNodeInjectionControlApi,
  batchUpdateCircuitBreakerApi,
  createAgentConfigTemplateApi,
  getAgentConfigNodesApi,
  getAgentConfigStatesApi,
  getAgentConfigTemplatesApi,
  getGlobalCircuitBreakerApi,
  getNodeCircuitBreakerApi,
  getNodeConfigHistoryApi,
  getNodeCurrentConfigsApi,
  getNodeInjectionControlApi,
  getNodeVariablesApi,
  importAgentConfigTemplateApi,
  updateAgentConfigTemplateApi,
  updateGlobalCircuitBreakerApi,
  updateNodeCircuitBreakerApi,
  updateNodeInjectionControlApi,
  updateNodeProbeConfigsApi,
  updateNodeVariablesApi,
} from 'src/core/api/integration'
import { notify } from 'src/core/utils/notify'
import { InjectionControlPanel } from './InjectionControlPanel'
import { LogCollectionConfigPanel, LogCollectionDefaultsPanel } from './LogCollectionConfigPage'
import { PreloadConfigEditor } from './PreloadConfigEditor'
import styles from './probeConfigPage.module.scss'

type AgentNode = {
  id: string
  name?: string
  clusterId?: string
  environment?: string
  labels?: Record<string, string>
  reportedVariables?: Record<string, string>
}
type NodeVariables = {
  nodeId: string
  initialClusterId?: string
  reported?: Record<string, string>
  overrides?: Record<string, string>
  effective?: Record<string, string>
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
type NodeInjectionControl = {
  nodeId: string
  available: boolean
  generation: number
  filePath?: string
  content?: string
  contentSha256?: string
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
const filesForConfig = (config: CurrentConfig): ConfigFile[] =>
  config.files?.length
    ? config.files
    : [{ path: 'config.yaml', content: config.content || '', contentSha256: config.contentSha256 }]
const allConfigsForTemplate = (template?: AgentTemplate): TemplateProbeConfig[] =>
  template?.configs || []
const isPreloadProbe = (probeType: string) => probeType.toLowerCase() === 'preload'
const configsForTemplate = (template?: AgentTemplate): TemplateProbeConfig[] =>
  allConfigsForTemplate(template).filter((config) => !isPreloadProbe(config.probeType))
const preloadFileForConfig = (config?: { probeType: string; files: ConfigFile[] }) =>
  config && isPreloadProbe(config.probeType)
    ? config.files.find((file) => file.path.split('/').at(-1) === 'libapoinstrument.conf')
    : undefined
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
  description,
  value,
  onChange,
  showSystemPressure = false,
}: {
  title: string
  description?: string
  value: CircuitPolicy
  onChange: (next: CircuitPolicy) => void
  showSystemPressure?: boolean
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
          disabled={!value.enabled || !systemPressure.enabled}
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
          disabled={!value.enabled || !systemPressure.enabled}
          addonAfter="%"
          onChange={(next) => patchSystemPressure({ maxMemoryPercent: Number(next || 1) })}
        />
      </label>
    </>
  )
  return (
    <div className={styles.tabContent}>
      <div className={styles.sectionHeading}>
        <div>
          <Typography.Text strong>{title}</Typography.Text>
          {description && <Typography.Text type="secondary">{description}</Typography.Text>}
        </div>
        <Switch
          checked={value.enabled}
          checkedChildren="启用"
          unCheckedChildren="停用"
          onChange={(enabled) => patch({ enabled })}
        />
      </div>
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
      {showSystemPressure && (
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
  const isClusterConfig = mode === 'batch' && Boolean(initialCluster)
  const isKubernetesClusterConfig =
    isClusterConfig && normalizeClusterType(initialClusterEnvironment) === 'kubernetes'
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
  const [batchPreloadDraft, setBatchPreloadDraft] = useState('')
  const [currentConfigs, setCurrentConfigs] = useState<CurrentConfig[]>([])
  const [configDrafts, setConfigDrafts] = useState<Record<string, string>>({})
  const [configFileDrafts, setConfigFileDrafts] = useState<Record<string, ConfigFile[]>>({})
  const [preloadControl, setPreloadControl] = useState<NodeInjectionControl>()
  const [preloadDraft, setPreloadDraft] = useState('')
  const [selectedFilePath, setSelectedFilePath] = useState('')
  const [historyConfigs, setHistoryConfigs] = useState<CurrentConfig[]>([])
  const [selectedHistoryGeneration, setSelectedHistoryGeneration] = useState<number>()
  const [templateDraftConfigs, setTemplateDraftConfigs] = useState<TemplateProbeConfig[]>([])
  const [selectedTemplateProbe, setSelectedTemplateProbe] = useState('')
  const [selectedTemplateFileIndex, setSelectedTemplateFileIndex] = useState(0)
  const [circuit, setCircuit] = useState<CircuitPolicy>(DEFAULT_CIRCUIT)
  const [circuitSource, setCircuitSource] = useState('')
  const [circuitComponent, setCircuitComponent] = useState('')
  const [componentPolicyPickerOpen, setComponentPolicyPickerOpen] = useState(false)
  const [componentPolicyDraft, setComponentPolicyDraft] = useState<CircuitPolicy>()
  const [nodeVariables, setNodeVariables] = useState<NodeVariables>()
  const [nodeVariableDraft, setNodeVariableDraft] = useState<Record<string, string>>({})
  const [nodeVariablesEditing, setNodeVariablesEditing] = useState(false)
  const [nodeVariablesSaving, setNodeVariablesSaving] = useState(false)
  const [activeTab, setActiveTab] = useState(() => {
    const section = searchParams.get('section')
    return ['logs', 'injection'].includes(section || '') ? section! : 'config'
  })
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
    const section = searchParams.get('section')
    if (section === 'logs' || (mode !== 'global' && section === 'injection')) {
      setActiveTab(section)
    }
  }, [mode, searchParams])

  useEffect(() => {
    if (mode !== 'batch') return
    setSelectedClusters(initialCluster ? [initialCluster] : [])
  }, [initialCluster, mode])

  const availableProbeTypes = useMemo(() => {
    if (isNode)
      return currentConfigs
        .filter((config) => !isPreloadProbe(config.probeType))
        .map((config) => config.probeType)
        .sort()
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
      setCircuitComponent('')
      setComponentPolicyPickerOpen(false)
      setComponentPolicyDraft(undefined)
      setNodeVariablesEditing(false)
      try {
        if (mode === 'node') {
          const [configResult, circuitResult, injectionResult, variablesResult] = await Promise.all(
            [
              getNodeCurrentConfigsApi(nodeId),
              getNodeCircuitBreakerApi(nodeId),
              getNodeInjectionControlApi(nodeId),
              getNodeVariablesApi(nodeId),
            ],
          )
          const configs = ((configResult as any)?.configs || []) as CurrentConfig[]
          const injectionControl = injectionResult as NodeInjectionControl
          setCurrentConfigs(configs)
          setPreloadControl(injectionControl)
          setPreloadDraft(injectionControl?.content || '')
          setConfigFileDrafts({})
          setConfigDrafts(
            Object.fromEntries(
              configs
                .filter((config) => !isPreloadProbe(config.probeType))
                .flatMap((config) =>
                  filesForConfig(config).map((file) => [
                    draftKey(config.probeType, file.path),
                    file.content,
                  ]),
                ),
            ),
          )
          setCircuit(normalizeCircuitPolicy((circuitResult as any)?.circuitBreaker))
          setCircuitSource((circuitResult as any)?.source || '')
          const variables = (variablesResult || {}) as NodeVariables
          setNodeVariables(variables)
          setNodeVariableDraft({ ...(variables.overrides || {}) })
        } else {
          setNodeVariables(undefined)
          setNodeVariableDraft({})
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
  const configPageCurrentConfigs = currentConfigs.filter(
    (config) => !isPreloadProbe(config.probeType),
  )
  const selectedFiles = selectedCurrentConfig ? filesForDraft(selectedCurrentConfig) : []
  const selectedFileSetKey = selectedFiles.map((file) => file.path).join('\u0000')
  const nodeFileCount = configPageCurrentConfigs.reduce(
    (count, config) => count + filesForDraft(config).length,
    0,
  )
  const currentNode = nodes.find((node) => node.id === nodeId)
  const inheritedClusterId = currentNode?.clusterId || initialCluster
  const inheritedClusterEnvironment =
    currentNode?.environment || initialClusterEnvironment || 'kubernetes'
  const isKubernetesNodeConfig = mode === 'node' && isK8sEnvironment(inheritedClusterEnvironment)
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
  const reportedVariables =
    nodeVariables?.reported || currentNode?.reportedVariables || currentNode?.labels || {}
  const variableNames = [
    ...new Set([
      ...Object.keys(reportedVariables),
      ...Object.keys(nodeVariables?.overrides || {}),
      ...Object.keys(nodeVariables?.effective || {}),
    ]),
  ].sort()
  const reportedVariableEntries = variableNames.map(
    (name) => [name, reportedVariables[name] || ''] as const,
  )
  const startEditingNodeVariables = () => {
    setNodeVariableDraft({ ...(nodeVariables?.overrides || {}) })
    setNodeVariablesEditing(true)
  }
  const cancelEditingNodeVariables = () => {
    setNodeVariableDraft({ ...(nodeVariables?.overrides || {}) })
    setNodeVariablesEditing(false)
  }
  const saveNodeVariables = async () => {
    setNodeVariablesSaving(true)
    try {
      const editableOverrides = Object.fromEntries(
        Object.entries(nodeVariableDraft).filter(
          ([name]) => !['CLUSTER_IP', 'CLUSTER_ID', 'INSTANCE_ID'].includes(name),
        ),
      )
      const saved = (await updateNodeVariablesApi(nodeId, editableOverrides)) as NodeVariables
      setNodeVariables(saved)
      setNodeVariableDraft({ ...(saved.overrides || {}) })
      setNodeVariablesEditing(false)
      notify({ type: 'success', message: '节点变量覆盖已保存' })
    } catch (error: any) {
      notify({ type: 'error', message: error.message || '节点变量覆盖保存失败' })
    } finally {
      setNodeVariablesSaving(false)
    }
  }
  const protectedNodeVariable = (name: string) =>
    ['CLUSTER_IP', 'CLUSTER_ID', 'INSTANCE_ID'].includes(name)
  const effectiveVariable = (name: string) =>
    nodeVariables?.effective?.[name] || nodeVariableDraft[name] || reportedVariables[name] || ''
  /* Keep reported values sorted for a stable, scannable node panel. */
  const sortedReportedVariableEntries = reportedVariableEntries.sort(([left], [right]) =>
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
  const batchPreloadControls = targetNodes.map((node) => {
    const config = (batchNodeConfigs[node.id] || []).find((item) => isPreloadProbe(item.probeType))
    return { node, config, file: preloadFileForConfig(config) }
  })
  const batchPreloadContents = new Set(
    batchPreloadControls
      .map(({ file }) => file?.content)
      .filter((content): content is string => content !== undefined),
  )
  const batchPreloadContentConsistent = batchPreloadContents.size <= 1
  const batchPreloadReady = Boolean(
    targetNodes.length &&
      !batchConfigsLoading &&
      batchPreloadControls.every(({ config, file }) => config && file),
  )
  const batchPreloadChanged = Boolean(
    batchPreloadReady &&
      batchPreloadDraft &&
      batchPreloadControls.some(({ file }) => file?.content !== batchPreloadDraft),
  )
  const showBatchPreloadTab = mode === 'batch' && selectedClusterType === 'vm'
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
        const current = (batchNodeConfigs[node.id] || []).filter(
          (config) => !isPreloadProbe(config.probeType),
        )
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
    if (!showBatchPreloadTab) {
      setBatchPreloadDraft('')
      return
    }
    const contents = new Set(
      batchPreloadControls
        .map(({ file }) => file?.content)
        .filter((content) => content !== undefined),
    )
    setBatchPreloadDraft([...contents][0] || '')
  }, [batchNodeConfigs, showBatchPreloadTab])

  useEffect(() => {
    setSelectedBatchFileKey((current) =>
      batchFiles.some((file) => file.key === current) ? current : batchFiles[0]?.key || '',
    )
  }, [batchFiles])

  const nodeFileConfigChanged = configPageCurrentConfigs.some((config) => {
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

  const saveNodeConfigGroup = async (
    configsToSave: CurrentConfig[],
    recordTarget: string,
    successMessage: string,
  ) => {
    if (!configsToSave.length) return
    setSaving(true)
    try {
      const result: any = await updateNodeProbeConfigsApi(nodeId, {
        configs: configsToSave.map((config) => ({
          probeType: config.probeType,
          files: filesForDraft(config).map((file) => ({
            path: file.path,
            content: configDrafts[draftKey(config.probeType, file.path)] ?? file.content,
          })),
          expectedGeneration: config.generation,
        })),
      })
      const configs = (result?.configs || []) as CurrentConfig[]
      const savedProbes = new Set(configsToSave.map((config) => config.probeType.toLowerCase()))
      setCurrentConfigs(configs)
      setConfigFileDrafts((current) =>
        Object.fromEntries(
          Object.entries(current).filter(
            ([probeType]) => !savedProbes.has(probeType.toLowerCase()),
          ),
        ),
      )
      setConfigDrafts((current) => ({
        ...current,
        ...Object.fromEntries(
          configs
            .filter((config) => savedProbes.has(config.probeType.toLowerCase()))
            .flatMap((config) =>
              filesForConfig(config).map((file) => [
                draftKey(config.probeType, file.path),
                file.content,
              ]),
            ),
        ),
      }))
      addRecord('CONFIG', recordTarget, { succeededCount: 1, targetNodeCount: 1 })
      notify({ type: 'success', message: successMessage })
    } catch (error: any) {
      notify({ type: 'error', message: error.message || '配置保存下发失败' })
    } finally {
      setSaving(false)
    }
  }

  const saveNodeConfig = () =>
    saveNodeConfigGroup(configPageCurrentConfigs, `${nodeId} · 配置文件`, '配置文件已保存并下发')

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
    const templateConfigs = configsForTemplate(selectedTemplate).filter(
      (config) => !isPreloadProbe(config.probeType),
    )
    const currentByProbe = new Map(
      configPageCurrentConfigs.map((config) => [config.probeType, config]),
    )
    const complete =
      templateConfigs.length === configPageCurrentConfigs.length &&
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
          circuitBreaker: normalizeCircuitPolicy(circuit),
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
  const preloadConfigChanged = Boolean(
    preloadControl?.available && preloadDraft !== preloadControl.content,
  )
  const hasPreloadControl =
    mode === 'node' &&
    !isK8sEnvironment(inheritedClusterEnvironment) &&
    Boolean(preloadControl?.available)

  useEffect(() => {
    if (
      activeTab === 'injection' &&
      !hasPreloadControl &&
      !showBatchPreloadTab &&
      !isKubernetesClusterConfig &&
      !isKubernetesNodeConfig
    ) {
      setActiveTab('config')
    }
  }, [
    activeTab,
    hasPreloadControl,
    isKubernetesClusterConfig,
    isKubernetesNodeConfig,
    showBatchPreloadTab,
  ])

  const updatePreloadConfigContent = (content: string) => {
    if (!preloadControl?.available) return
    setPreloadDraft(content)
  }
  const savePreloadConfig = async () => {
    if (!preloadControl?.available) return
    setSaving(true)
    try {
      const saved = (await updateNodeInjectionControlApi(nodeId, {
        content: preloadDraft,
        expectedGeneration: preloadControl.generation,
      })) as NodeInjectionControl
      setPreloadControl(saved)
      setPreloadDraft(saved.content || '')
      addRecord('CONFIG', `${nodeId} · 注入控制`, {
        succeededCount: 1,
        targetNodeCount: 1,
      })
      notify({ type: 'success', message: '注入控制已保存并下发' })
    } catch (error: any) {
      notify({ type: 'error', message: error.message || '注入控制保存下发失败' })
    } finally {
      setSaving(false)
    }
  }
  const saveBatchPreloadConfig = async () => {
    if (!batchPreloadReady || !batchPreloadChanged) return
    setSaving(true)
    try {
      const result: any = await batchUpdateNodeInjectionControlApi({
        clusterIds: selectedClusters,
        content: batchPreloadDraft,
        expectedGenerations: Object.fromEntries(
          batchPreloadControls.map(({ node, config }) => [node.id, config!.generation]),
        ),
      })
      setBatchNodeConfigs((current) =>
        Object.fromEntries(
          Object.entries(current).map(([targetNodeID, configs]) => [
            targetNodeID,
            configs.map((config) => {
              if (!isPreloadProbe(config.probeType)) return config
              const changed = filesForConfig(config).some(
                (file) =>
                  file.path.split('/').at(-1) === 'libapoinstrument.conf' &&
                  file.content !== batchPreloadDraft,
              )
              return {
                ...config,
                generation: config.generation + (changed ? 1 : 0),
                files: filesForConfig(config).map((file) =>
                  file.path.split('/').at(-1) === 'libapoinstrument.conf'
                    ? { ...file, content: batchPreloadDraft }
                    : file,
                ),
              }
            }),
          ]),
        ),
      )
      addRecord('CONFIG', `${selectedClusters.join(', ')} · 注入控制`, result)
      notify({
        type: 'success',
        message: `注入控制已下发到 ${result.succeededCount} 个节点`,
      })
    } catch (error: any) {
      notify({ type: 'error', message: error.message || '批量注入控制下发失败' })
    } finally {
      setSaving(false)
    }
  }
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
  const configuredComponentTypes = Object.keys(circuit.componentPolicies || {}).sort()
  const availableComponentTypes = circuitComponentTypes.filter(
    (componentType) => !circuit.componentPolicies?.[componentType],
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
    const nextComponent = configuredComponentTypes.find(
      (componentType) => componentType !== circuitComponent,
    )
    setCircuit((current) => {
      const componentPolicies = { ...current.componentPolicies }
      delete componentPolicies[circuitComponent]
      return { ...current, componentPolicies }
    })
    setCircuitComponent(nextComponent || '')
  }
  const startAddingComponentPolicy = () => {
    setCircuitComponent('')
    setComponentPolicyDraft(undefined)
    setComponentPolicyPickerOpen(true)
  }
  const selectComponentPolicyDraft = (componentType: string) => {
    setCircuitComponent(componentType)
    setComponentPolicyDraft(concreteComponentCircuitPolicy(circuit))
  }
  const cancelAddingComponentPolicy = () => {
    setComponentPolicyPickerOpen(false)
    setComponentPolicyDraft(undefined)
    setCircuitComponent(configuredComponentTypes[0] || '')
  }
  const addComponentCircuitPolicy = () => {
    if (!circuitComponent || !componentPolicyDraft) return
    updateComponentCircuitPolicy(componentPolicyDraft)
    setComponentPolicyPickerOpen(false)
    setComponentPolicyDraft(undefined)
  }

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
      ? '管理配置模版、日志采集默认规则和新节点默认使用的熔断策略'
      : mode === 'node'
        ? '修改当前节点配置'
        : isClusterConfig
          ? '批量管理当前集群各节点的配置文件、注入控制、日志采集和熔断策略'
          : '选择目标集群，分别下发配置文件和各项独立控制'

  const configTab =
    mode === 'batch' ? (
      <div className={styles.tabContent}>
        <div className={styles.batchSetupGrid}>
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
                  <Typography.Text type="secondary">
                    {nodeVariablesEditing ? '编辑变量覆盖' : '探针上报与当前生效值'}
                  </Typography.Text>
                </div>
                {!!sortedReportedVariableEntries.length &&
                  (nodeVariablesEditing ? (
                    <span className={styles.nodeVariableActions}>
                      <Button
                        size="small"
                        disabled={nodeVariablesSaving}
                        onClick={cancelEditingNodeVariables}
                      >
                        取消
                      </Button>
                      <Button
                        type="primary"
                        size="small"
                        icon={<Save size={14} />}
                        loading={nodeVariablesSaving}
                        onClick={saveNodeVariables}
                      >
                        保存变量
                      </Button>
                    </span>
                  ) : (
                    <Button
                      size="small"
                      icon={<Pencil size={14} />}
                      onClick={startEditingNodeVariables}
                    >
                      编辑变量
                    </Button>
                  ))}
              </div>
              {sortedReportedVariableEntries.length ? (
                <div className={styles.nodeVariableList}>
                  {sortedReportedVariableEntries.map(([name, value]) => (
                    <div
                      className={`${styles.nodeVariableItem} ${
                        nodeVariablesEditing ? styles.nodeVariableItemEditing : ''
                      }`}
                      key={name}
                    >
                      <Typography.Text code>{name}</Typography.Text>
                      <Typography.Text type="secondary">上报: {value || '-'}</Typography.Text>
                      {nodeVariablesEditing &&
                        (protectedNodeVariable(name) ? (
                          <Typography.Text type="secondary">由集群管理维护</Typography.Text>
                        ) : (
                          <Input
                            size="small"
                            value={nodeVariableDraft[name] || ''}
                            placeholder="可选覆盖值"
                            onChange={(event) =>
                              setNodeVariableDraft((current) => ({
                                ...current,
                                [name]: event.target.value,
                              }))
                            }
                          />
                        ))}
                      <Typography.Text code>生效: {effectiveVariable(name) || '-'}</Typography.Text>
                      {name === 'CLUSTER_IP' && (
                        <Button
                          type="link"
                          size="small"
                          icon={<ArrowRight size={14} />}
                          onClick={() =>
                            navigate(
                              `/probe-management?clusterId=${encodeURIComponent(inheritedClusterId)}`,
                            )
                          }
                        >
                          前往集群管理
                        </Button>
                      )}
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
                    {configPageCurrentConfigs.length} 个组件 · {nodeFileCount} 个文件
                  </Typography.Text>
                </div>
              </div>
              <Tree
                key={configPageCurrentConfigs
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
                treeData={configPageCurrentConfigs.map((config) => ({
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
              {configPageCurrentConfigs.length} 个组件 · {nodeFileCount} 个配置文件作为本页独立
              Release 组下发
            </Typography.Text>
          )}
          <Button
            type="primary"
            icon={<Save size={15} />}
            loading={saving}
            disabled={
              mode === 'node' ? !nodeFileConfigChanged : !selectedTemplate || !templateChanged
            }
            onClick={mode === 'node' ? saveNodeConfig : saveTemplate}
          >
            {mode === 'node' ? '下发配置' : '保存配置模版'}
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
        showK8sFilters={isK8sEnvironment(inheritedClusterEnvironment)}
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
    ) : (
      <LogCollectionDefaultsPanel />
    )

  const injectionTab = isKubernetesClusterConfig ? (
    <InjectionControlPanel clusterId={initialCluster} />
  ) : isKubernetesNodeConfig ? (
    <div className={styles.clusterInjectionGuide}>
      <span className={styles.clusterInjectionGuideIcon}>
        <Network size={22} />
      </span>
      <div className={styles.clusterInjectionGuideContent}>
        <Typography.Title level={4}>该节点的注入策略由集群统一管理</Typography.Title>
        <Typography.Paragraph type="secondary">
          当前节点自动应用所属集群“{inheritedClusterId}”中配置的 Namespace
          和工作负载注入范围，暂不支持按节点单独设置。
        </Typography.Paragraph>
        <Button
          type="primary"
          icon={<ArrowRight size={15} />}
          iconPosition="end"
          disabled={!inheritedClusterId}
          onClick={() =>
            navigate(
              `/probe-management/config?clusterId=${encodeURIComponent(inheritedClusterId)}&clusterType=kubernetes&section=injection`,
            )
          }
        >
          前往集群注入控制
        </Button>
      </div>
    </div>
  ) : showBatchPreloadTab ? (
    <div className={styles.tabContent}>
      <div className={styles.injectionTargetRow}>
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
      <div className={styles.injectionWorkspace}>
        <PreloadConfigEditor
          content={batchPreloadDraft}
          onChange={setBatchPreloadDraft}
          readOnly={!batchPreloadReady}
        />
      </div>
      <div className={`${styles.editorFooter} ${styles.injectionFooter}`}>
        <Typography.Text type={batchPreloadReady ? 'secondary' : 'danger'}>
          {batchConfigsLoading
            ? '正在读取目标节点注入配置'
            : !targetCount
              ? '请选择至少一个 VM 集群'
              : !batchPreloadReady
                ? '部分目标节点未注册 preload 注入配置，无法批量下发'
                : !batchPreloadContentConsistent
                  ? '目标节点当前注入配置不一致，当前以第一个目标节点为草稿，保存后将统一覆盖'
                  : `${targetCount} 个节点使用独立注入控制接口原子下发，不生成可回滚历史`}
        </Typography.Text>
        <Button
          type="primary"
          icon={<Send size={15} />}
          loading={saving}
          disabled={!batchPreloadChanged}
          onClick={saveBatchPreloadConfig}
        >
          下发到所选集群
        </Button>
      </div>
    </div>
  ) : hasPreloadControl ? (
    <div className={styles.tabContent}>
      <div className={styles.injectionWorkspace}>
        <PreloadConfigEditor content={preloadDraft} onChange={updatePreloadConfigContent} />
      </div>
      <div className={`${styles.editorFooter} ${styles.injectionFooter}`}>
        <Typography.Text type="secondary">
          使用独立注入控制接口下发，不生成可回滚的配置历史
        </Typography.Text>
        <Button
          type="primary"
          icon={<Save size={15} />}
          loading={saving}
          disabled={!preloadConfigChanged}
          onClick={savePreloadConfig}
        >
          下发注入配置
        </Button>
      </div>
    </div>
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
        title={mode === 'batch' ? '批量节点熔断策略' : '探针消耗资源熔断'}
        description={
          mode === 'batch'
            ? '同一策略批量设置到所选集群的每个节点，由各节点独立判断和熔断'
            : undefined
        }
        value={concreteCircuitPolicy(circuit)}
        onChange={updateNodeCircuitPolicy}
        showSystemPressure
      />
      {!!circuitComponentTypes.length && (
        <Collapse
          ghost
          className={styles.advancedCircuit}
          items={[
            {
              key: 'component-quotas',
              label: (
                <div className={styles.advancedCircuitLabel}>
                  <span>
                    <Typography.Text strong>高级配置</Typography.Text>
                    <Tag>{configuredComponentTypes.length} 项</Tag>
                  </span>
                  <Typography.Text type="secondary">按组件单独设置熔断策略</Typography.Text>
                </div>
              ),
              children: (
                <div className={styles.componentPolicySection}>
                  <div className={styles.componentPolicyHeader}>
                    <div>
                      <Typography.Text strong>组件专属策略</Typography.Text>
                      <Typography.Text type="secondary">
                        未单独配置的组件继续使用上方通用策略
                      </Typography.Text>
                    </div>
                    <Button
                      icon={<Plus size={15} />}
                      disabled={!availableComponentTypes.length}
                      onClick={startAddingComponentPolicy}
                    >
                      添加组件单独熔断策略
                    </Button>
                  </div>
                  {!!configuredComponentTypes.length && (
                    <div className={styles.componentPolicyList}>
                      {configuredComponentTypes.map((componentType) => {
                        const policy = concreteCircuitPolicy(
                          circuit.componentPolicies?.[componentType],
                        )
                        return (
                          <button
                            key={componentType}
                            type="button"
                            className={
                              componentType === circuitComponent && !componentPolicyPickerOpen
                                ? styles.componentPolicyRowActive
                                : styles.componentPolicyRow
                            }
                            onClick={() => {
                              setCircuitComponent(componentType)
                              setComponentPolicyPickerOpen(false)
                              setComponentPolicyDraft(undefined)
                            }}
                          >
                            <span className={styles.componentPolicyIdentity}>
                              <strong>{componentType}</strong>
                              <Tag color={policy.enabled ? 'success' : 'default'}>
                                {policy.enabled ? '已启用' : '已停用'}
                              </Tag>
                            </span>
                            <span className={styles.componentPolicySummary}>
                              CPU {policy.maxCpuCores} Core · 内存{' '}
                              {toMemoryMiB(policy.maxMemoryBytes)} MiB · 持续{' '}
                              {policy.durationSeconds} 秒
                            </span>
                            <span className={styles.componentPolicyEdit}>编辑</span>
                          </button>
                        )
                      })}
                    </div>
                  )}
                  {!configuredComponentTypes.length && !componentPolicyPickerOpen && (
                    <div className={styles.componentPolicyEmpty}>
                      <Settings size={18} />
                      <span>
                        <Typography.Text>暂无组件专属策略</Typography.Text>
                        <Typography.Text type="secondary">
                          当前所有组件都使用通用熔断策略
                        </Typography.Text>
                      </span>
                    </div>
                  )}
                  {componentPolicyPickerOpen && (
                    <div className={styles.componentPolicyComposer}>
                      <div className={styles.componentPolicyPicker}>
                        <label>
                          <span>选择组件</span>
                          <Select
                            value={circuitComponent || undefined}
                            placeholder="请选择需要单独配置的组件"
                            onChange={selectComponentPolicyDraft}
                            options={availableComponentTypes.map((value) => ({
                              value,
                              label: value,
                            }))}
                          />
                        </label>
                      </div>
                      {componentPolicyDraft && (
                        <CircuitEditor
                          title="设置熔断策略"
                          value={componentPolicyDraft}
                          onChange={setComponentPolicyDraft}
                        />
                      )}
                      <div className={styles.componentPolicyDraftActions}>
                        <Button onClick={cancelAddingComponentPolicy}>取消</Button>
                        <Button
                          type="primary"
                          disabled={!componentPolicyDraft}
                          onClick={addComponentCircuitPolicy}
                        >
                          添加
                        </Button>
                      </div>
                    </div>
                  )}
                  {hasComponentCircuitPolicy && !componentPolicyPickerOpen && (
                    <div className={styles.componentPolicyEditor}>
                      <div className={styles.componentPolicyEditorHeader}>
                        <div>
                          <Typography.Text strong>{circuitComponent}</Typography.Text>
                          <Typography.Text type="secondary">
                            仅覆盖当前组件；移除后恢复使用通用策略
                          </Typography.Text>
                        </div>
                        <Button
                          danger
                          type="text"
                          icon={<Trash2 size={15} />}
                          onClick={removeComponentCircuitPolicy}
                        >
                          移除专属策略
                        </Button>
                      </div>
                      <CircuitEditor
                        title="资源阈值"
                        value={selectedComponentPolicy}
                        onChange={updateComponentCircuitPolicy}
                      />
                    </div>
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
      <Button type="link" icon={<LeftOutlined />} onClick={() => navigate('/probe-management')}>
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
            ...(isKubernetesClusterConfig ||
            isKubernetesNodeConfig ||
            showBatchPreloadTab ||
            hasPreloadControl
              ? [
                  {
                    key: 'injection',
                    label: '注入控制',
                    children: injectionTab,
                  },
                ]
              : []),
            { key: 'logs', label: '日志采集', children: logCollectionTab },
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
