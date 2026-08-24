import {
  Button,
  Card,
  Col,
  Form,
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
  Typography,
} from 'antd'
import { LeftOutlined, PlusOutlined } from '@ant-design/icons'
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
} from 'lucide-react'
import { useCallback, useEffect, useMemo, useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import {
  batchAssignAgentConfigApi,
  batchUpdateCircuitBreakerApi,
  batchUpdateLogSourcesApi,
  createAgentConfigTemplateApi,
  getAgentConfigNodesApi,
  getAgentConfigStatesApi,
  getAgentConfigTemplatesApi,
  getGlobalCircuitBreakerApi,
  getNodeCircuitBreakerApi,
  getNodeConfigHistoryApi,
  getNodeCurrentConfigsApi,
  getNodeLogSourcesApi,
  updateAgentConfigTemplateApi,
  updateGlobalCircuitBreakerApi,
  updateNodeCircuitBreakerApi,
  updateNodeCurrentConfigsApi,
  updateNodeLogSourcesApi,
} from 'src/core/api/integration'
import { notify } from 'src/core/utils/notify'
import styles from './probeConfigPage.module.scss'

type AgentNode = { id: string; name?: string; clusterId?: string }
type ProbeState = { nodeId: string; probeType: string }
type TemplateVariable = { name: string; required?: boolean; defaultValue?: string }
type ConfigFile = { path: string; content: string; contentSha256: string }
type AgentTemplate = {
  id: string
  probeType: string
  name: string
  content: string
  files?: ConfigFile[]
  format: string
  version: number
  variables?: TemplateVariable[]
}
type CircuitPolicy = {
  enabled: boolean
  maxCpuPercent: number
  maxMemoryBytes: number
  durationSeconds: number
  autoRecover: boolean
  recoveryInitialDelaySeconds: number
  recoveryMaxDelaySeconds: number
  probePolicies?: Record<string, CircuitPolicy>
}
type CurrentConfig = {
  probeType: string
  generation: number
  content: string
  files: ConfigFile[]
  contentSha256: string
  source: string
  createdAt?: string
}
type LogSource = {
  type: 'file' | 'stdout'
  path: string
  pattern: string
  maxBytes: number
}
type DeliveryRecord = {
  id: string
  kind: 'CONFIG' | 'CIRCUIT' | 'LOG'
  target: string
  succeeded: number
  total: number
  createdAt: string
}

const DEFAULT_CIRCUIT: CircuitPolicy = {
  enabled: true,
  maxCpuPercent: 80,
  maxMemoryBytes: 512 * 1024 * 1024,
  durationSeconds: 30,
  autoRecover: true,
  recoveryInitialDelaySeconds: 30,
  recoveryMaxDelaySeconds: 3600,
}

const toMemoryMiB = (bytes: number) => Math.round(bytes / 1024 / 1024)
const fromMemoryMiB = (mib: number) => mib * 1024 * 1024
const newLogSource = (): LogSource => ({ type: 'file', path: '', pattern: '', maxBytes: 0 })
const draftKey = (probeType: string, path: string) => `${probeType}\u0000${path}`
const concreteCircuitPolicy = (value?: Partial<CircuitPolicy>): CircuitPolicy => {
  const { probePolicies: _ignored, ...policy } = value || {}
  return { ...DEFAULT_CIRCUIT, ...policy }
}
const normalizeCircuitPolicy = (value?: Partial<CircuitPolicy>): CircuitPolicy => ({
  ...concreteCircuitPolicy(value),
  probePolicies: Object.fromEntries(
    Object.entries(value?.probePolicies || {}).map(([probeType, policy]) => [
      probeType,
      concreteCircuitPolicy(policy),
    ]),
  ),
})
const filesForConfig = (config: CurrentConfig): ConfigFile[] =>
  config.files?.length
    ? config.files
    : [{ path: 'config.yaml', content: config.content || '', contentSha256: config.contentSha256 }]
const filesForTemplate = (template?: AgentTemplate): ConfigFile[] => {
  if (!template) return []
  if (template.files?.length) return template.files
  return [{ path: '', content: template.content || '', contentSha256: '' }]
}
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
  value,
  onChange,
}: {
  value: CircuitPolicy
  onChange: (next: CircuitPolicy) => void
}) {
  const patch = (next: Partial<CircuitPolicy>) => onChange({ ...value, ...next })
  return (
    <div className={styles.tabContent}>
      <div className={styles.sectionHeading}>
        <Typography.Text strong>资源熔断策略</Typography.Text>
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
            <strong>CPU 使用率</strong>
            <small>10% - 100%</small>
          </span>
          <InputNumber
            aria-label="CPU 熔断阈值"
            min={10}
            max={100}
            value={value.maxCpuPercent}
            disabled={!value.enabled}
            addonAfter="%"
            onChange={(next) => patch({ maxCpuPercent: Number(next || 10) })}
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
      </div>
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
  const isGlobal = searchParams.get('scope') === 'global'
  const isNode = Boolean(nodeId)
  const mode = isGlobal ? 'global' : isNode ? 'node' : 'batch'
  const [nodes, setNodes] = useState<AgentNode[]>([])
  const [states, setStates] = useState<ProbeState[]>([])
  const [templates, setTemplates] = useState<AgentTemplate[]>([])
  const [selectedProbe, setSelectedProbe] = useState('')
  const [selectedTemplateId, setSelectedTemplateId] = useState('')
  const [selectedClusters, setSelectedClusters] = useState<string[]>(
    initialCluster ? [initialCluster] : [],
  )
  const [variables, setVariables] = useState<Record<string, string>>({})
  const [currentConfigs, setCurrentConfigs] = useState<CurrentConfig[]>([])
  const [configDrafts, setConfigDrafts] = useState<Record<string, string>>({})
  const [selectedFilePath, setSelectedFilePath] = useState('')
  const [historyConfigs, setHistoryConfigs] = useState<CurrentConfig[]>([])
  const [selectedHistoryGeneration, setSelectedHistoryGeneration] = useState<number>()
  const [templateDraftFiles, setTemplateDraftFiles] = useState<ConfigFile[]>([])
  const [selectedTemplateFileIndex, setSelectedTemplateFileIndex] = useState(0)
  const [circuit, setCircuit] = useState<CircuitPolicy>(DEFAULT_CIRCUIT)
  const [circuitSource, setCircuitSource] = useState('')
  const [circuitScope, setCircuitScope] = useState<'all' | 'probe'>('all')
  const [circuitProbe, setCircuitProbe] = useState('')
  const [logSources, setLogSources] = useState<LogSource[]>([])
  const [logSourcesConfigured, setLogSourcesConfigured] = useState(false)
  const [logSourcesLoading, setLogSourcesLoading] = useState(false)
  const [activeTab, setActiveTab] = useState('config')
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [records, setRecords] = useState<DeliveryRecord[]>([])
  const [createOpen, setCreateOpen] = useState(false)
  const [createForm] = Form.useForm()

  useEffect(() => {
    setCircuitScope('all')
  }, [mode, nodeId])

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
      setTemplates((templateResult as any)?.templates || [])
    } catch (error: any) {
      notify({ type: 'error', message: error.message || '加载探针配置失败' })
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    loadBase()
  }, [loadBase])

  const availableProbeTypes = useMemo(() => {
    if (isNode) return currentConfigs.map((config) => config.probeType).sort()
    const values = new Set<string>()
    states.forEach((state) => {
      values.add(state.probeType)
    })
    templates.forEach((template) => values.add(template.probeType))
    return [...values].sort()
  }, [currentConfigs, isNode, states, templates])

  const circuitProbeTypes = useMemo(() => {
    if (mode === 'global') return []
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
    setCircuitProbe((current) =>
      circuitProbeTypes.includes(current) ? current : circuitProbeTypes[0] || '',
    )
  }, [circuitProbeTypes])

  const matchingTemplates = templates.filter((template) => template.probeType === selectedProbe)
  const selectedTemplate = matchingTemplates.find((template) => template.id === selectedTemplateId)

  useEffect(() => {
    setSelectedTemplateId((current) =>
      matchingTemplates.some((template) => template.id === current)
        ? current
        : matchingTemplates[0]?.id || '',
    )
  }, [selectedProbe, templates])

  useEffect(() => {
    setTemplateDraftFiles(filesForTemplate(selectedTemplate))
    setSelectedTemplateFileIndex(0)
    const defaults: Record<string, string> = {}
    selectedTemplate?.variables?.forEach((variable) => {
      defaults[variable.name] = variable.defaultValue || ''
    })
    setVariables(defaults)
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
  const selectedFiles = selectedCurrentConfig ? filesForConfig(selectedCurrentConfig) : []

  useEffect(() => {
    setSelectedFilePath((current) =>
      selectedFiles.some((file) => file.path === current) ? current : selectedFiles[0]?.path || '',
    )
  }, [selectedProbe, selectedCurrentConfig?.generation])

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

  useEffect(() => {
    if (mode !== 'node' || !selectedProbe) {
      setLogSources([])
      setLogSourcesConfigured(false)
      return
    }
    setLogSourcesLoading(true)
    getNodeLogSourcesApi(nodeId, selectedProbe)
      .then((result: any) => {
        setLogSources((result?.logSources || []) as LogSource[])
        setLogSourcesConfigured(Boolean(result?.configured))
      })
      .catch((error: any) => {
        console.error('[ProbeConfigPage] failed to load log sources', error)
        notify({ type: 'error', message: error.message || '加载日志配置失败' })
        setLogSources([])
        setLogSourcesConfigured(false)
      })
      .finally(() => setLogSourcesLoading(false))
  }, [mode, nodeId, selectedProbe])

  const clusterOptions = useMemo(
    () => [...new Set(nodes.map((node) => node.clusterId).filter(Boolean) as string[])].sort(),
    [nodes],
  )
  const targetCount = nodes.filter((node) => selectedClusters.includes(node.clusterId || '')).length
  const nodeConfigChanged = currentConfigs.some((config) =>
    filesForConfig(config).some(
      (file) =>
        (configDrafts[draftKey(config.probeType, file.path)] ?? file.content) !== file.content,
    ),
  )

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
          files: filesForConfig(config).map((file) => ({
            path: file.path,
            content: configDrafts[draftKey(config.probeType, file.path)] ?? file.content,
          })),
          expectedGeneration: config.generation,
        })),
      })
      const configs = (result?.configs || []) as CurrentConfig[]
      setCurrentConfigs(configs)
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
      notify({ type: 'success', message: `该节点的 ${configs.length} 个探针配置已一次下发` })
    } catch (error: any) {
      notify({ type: 'error', message: error.message || '节点配置保存失败' })
    } finally {
      setSaving(false)
    }
  }

  const loadHistory = () => {
    const history = historyConfigs.find((config) => config.generation === selectedHistoryGeneration)
    if (!history || !selectedCurrentConfig) return
    const currentPaths = new Set(selectedFiles.map((file) => file.path))
    const historyFiles = filesForConfig(history)
    if (
      historyFiles.length !== selectedFiles.length ||
      historyFiles.some((file) => !currentPaths.has(file.path))
    ) {
      notify({ type: 'error', message: '该历史版本的文件列表与当前版本不一致，无法载入' })
      return
    }
    setConfigDrafts((current) => ({
      ...current,
      ...Object.fromEntries(
        historyFiles.map((file) => [draftKey(selectedProbe, file.path), file.content]),
      ),
    }))
    notify({ type: 'success', message: `已载入版本 ${history.generation}` })
  }

  const loadTemplateIntoFile = () => {
    if (!selectedTemplate || !selectedFilePath) return
    const templateFiles = filesForTemplate(selectedTemplate)
    if (templateFiles.length > 1) {
      const currentPaths = new Set(selectedFiles.map((file) => file.path))
      if (
        templateFiles.length !== selectedFiles.length ||
        templateFiles.some((file) => !currentPaths.has(file.path))
      ) {
        notify({ type: 'error', message: '模板文件列表与当前探针配置不一致，无法载入' })
        return
      }
      setConfigDrafts((current) => ({
        ...current,
        ...Object.fromEntries(
          templateFiles.map((file) => [draftKey(selectedProbe, file.path), file.content]),
        ),
      }))
      notify({
        type: 'success',
        message: `模板“${selectedTemplate.name}”已载入全部 ${templateFiles.length} 个文件`,
      })
      return
    }
    const templateFile = templateFiles[0]
    const targetPath = templateFile?.path || selectedFilePath
    if (!selectedFiles.some((file) => file.path === targetPath)) {
      notify({ type: 'error', message: `模板文件 ${targetPath} 不属于当前探针配置` })
      return
    }
    setConfigDrafts((current) => ({
      ...current,
      [draftKey(selectedProbe, targetPath)]: templateFile?.content || '',
    }))
    setSelectedFilePath(targetPath)
    notify({ type: 'success', message: `模板“${selectedTemplate.name}”已载入当前文件` })
  }

  const saveTemplate = async () => {
    if (!selectedTemplate) return
    const legacy = templateDraftFiles.length === 1 && !templateDraftFiles[0].path.trim()
    setSaving(true)
    try {
      await updateAgentConfigTemplateApi(selectedTemplate.id, {
        ...selectedTemplate,
        content: templateDraftFiles.length === 1 ? templateDraftFiles[0].content : '',
        files: legacy
          ? []
          : templateDraftFiles.map((file) => ({ path: file.path.trim(), content: file.content })),
      })
      await loadBase()
      notify({ type: 'success', message: '全局模板已保存' })
    } catch (error: any) {
      notify({ type: 'error', message: error.message || '模板保存失败' })
    } finally {
      setSaving(false)
    }
  }

  const dispatchTemplate = async () => {
    if (!selectedTemplate || !selectedClusters.length) return
    setSaving(true)
    try {
      const result: any = await batchAssignAgentConfigApi({
        clusterIds: selectedClusters,
        probeType: selectedProbe,
        templateId: selectedTemplate.id,
        variables,
      })
      addRecord('CONFIG', selectedClusters.join(', '), result)
      notify({
        type: result.failedNodeIds?.length ? 'warning' : 'success',
        message: result.failedNodeIds?.length
          ? `${result.succeededCount} 个节点成功，${result.failedNodeIds.length} 个失败`
          : `已覆盖下发到 ${result.succeededCount} 个节点`,
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
          circuitBreaker: circuit,
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

  const updateLogSource = (index: number, patch: Partial<LogSource>) => {
    setLogSources((current) =>
      current.map((source, sourceIndex) =>
        sourceIndex === index ? { ...source, ...patch } : source,
      ),
    )
  }

  const saveLogSources = async () => {
    if (!selectedProbe) return
    const normalized = logSources.map((source) =>
      source.type === 'stdout'
        ? {
            type: 'stdout' as const,
            path: '',
            pattern: '',
            maxBytes: source.maxBytes || 10 * 1024 * 1024,
          }
        : {
            type: 'file' as const,
            path: source.path.trim(),
            pattern: source.pattern.trim(),
            maxBytes: 0,
          },
    )
    if (normalized.some((source) => source.type === 'file' && !source.path.startsWith('/'))) {
      notify({ type: 'error', message: '文件日志路径必须是绝对路径' })
      return
    }
    setSaving(true)
    try {
      if (mode === 'node') {
        const result: any = await updateNodeLogSourcesApi(nodeId, selectedProbe, {
          logSources: normalized,
        })
        setLogSources((result?.logSources || normalized) as LogSource[])
        setLogSourcesConfigured(true)
        addRecord('LOG', nodeId, { succeededCount: 1, targetNodeCount: 1 })
        notify({ type: 'success', message: '日志配置已保存到当前节点' })
      } else {
        const result: any = await batchUpdateLogSourcesApi({
          clusterIds: selectedClusters,
          probeType: selectedProbe,
          logSources: normalized,
        })
        addRecord('LOG', selectedClusters.join(', '), result)
        notify({ type: 'success', message: `日志配置已下发到 ${result.succeededCount} 个节点` })
      }
    } catch (error: any) {
      notify({ type: 'error', message: error.message || '日志配置保存失败' })
    } finally {
      setSaving(false)
    }
  }

  const createTemplate = async () => {
    const values = await createForm.validateFields()
    setSaving(true)
    try {
      const { path, content, ...template } = values
      await createAgentConfigTemplateApi({
        ...template,
        content,
        files: [{ path, content }],
      })
      setCreateOpen(false)
      createForm.resetFields()
      await loadBase()
      notify({ type: 'success', message: '模板已创建' })
    } finally {
      setSaving(false)
    }
  }

  const selectedFile = selectedFiles.find((file) => file.path === selectedFilePath)
  const selectedDraft = selectedFile
    ? configDrafts[draftKey(selectedProbe, selectedFile.path)] ?? selectedFile.content
    : ''
  const selectedLanguage = languageForPath(selectedFilePath)
  const selectedTemplateFile = templateDraftFiles[selectedTemplateFileIndex]
  const originalTemplateFiles = filesForTemplate(selectedTemplate)
  const templateChanged =
    JSON.stringify(
      templateDraftFiles.map((file) => ({ path: file.path, content: file.content })),
    ) !==
    JSON.stringify(
      originalTemplateFiles.map((file) => ({ path: file.path, content: file.content })),
    )
  const updateTemplateFile = (patch: Partial<ConfigFile>) => {
    setTemplateDraftFiles((current) =>
      current.map((file, index) =>
        index === selectedTemplateFileIndex ? { ...file, ...patch } : file,
      ),
    )
  }
  const addTemplateFile = () => {
    setTemplateDraftFiles((current) => [
      ...current,
      { path: `config-${current.length + 1}.yaml`, content: '', contentSha256: '' },
    ])
    setSelectedTemplateFileIndex(templateDraftFiles.length)
  }
  const removeTemplateFile = () => {
    if (templateDraftFiles.length <= 1) return
    setTemplateDraftFiles((current) =>
      current.filter((_, index) => index !== selectedTemplateFileIndex),
    )
    setSelectedTemplateFileIndex((current) => Math.max(0, current - 1))
  }

  const hasProbeCircuitPolicy = Boolean(circuitProbe && circuit.probePolicies?.[circuitProbe])
  const selectedCircuitPolicy =
    circuitScope === 'probe' && circuitProbe
      ? concreteCircuitPolicy(circuit.probePolicies?.[circuitProbe] || circuit)
      : concreteCircuitPolicy(circuit)
  const updateSelectedCircuitPolicy = (next: CircuitPolicy) => {
    if (circuitScope === 'all' || !circuitProbe) {
      setCircuit((current) => ({
        ...concreteCircuitPolicy(next),
        probePolicies: current.probePolicies,
      }))
      return
    }
    setCircuit((current) => ({
      ...current,
      probePolicies: {
        ...current.probePolicies,
        [circuitProbe]: concreteCircuitPolicy(next),
      },
    }))
  }
  const removeProbeCircuitPolicy = () => {
    if (!circuitProbe) return
    setCircuit((current) => {
      const probePolicies = { ...current.probePolicies }
      delete probePolicies[circuitProbe]
      return { ...current, probePolicies }
    })
  }

  const title =
    mode === 'global' ? '全局配置' : mode === 'node' ? `节点配置 · ${nodeId}` : '多集群批量配置'
  const subtitle =
    mode === 'global'
      ? '管理批量下发模板和新节点使用的默认熔断策略'
      : mode === 'node'
        ? '修改当前节点配置'
        : '选择模板和目标集群，批量覆盖对应探针的原配置'

  const configTab =
    mode === 'batch' ? (
      <div className={styles.tabContent}>
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
          <label>
            <span>配置模板</span>
            <Select
              value={selectedTemplateId || undefined}
              onChange={setSelectedTemplateId}
              options={matchingTemplates.map((template) => ({
                value: template.id,
                label: `${template.name} · v${template.version}`,
              }))}
            />
          </label>
        </div>
        {!!selectedTemplate?.variables?.length && (
          <div className={styles.variableGrid}>
            {selectedTemplate.variables.map((variable) => (
              <label key={variable.name}>
                <span>
                  {variable.name}
                  {variable.required ? ' *' : ''}
                </span>
                <Input
                  value={variables[variable.name]}
                  onChange={(event) =>
                    setVariables((current) => ({ ...current, [variable.name]: event.target.value }))
                  }
                />
              </label>
            ))}
          </div>
        )}
        <Row className={styles.configWorkspace} wrap={false}>
          <Col flex="240px" className={styles.filePanel}>
            <div className={styles.filePanelTitle}>
              <FileText size={16} />
              <Typography.Text strong>模板文件</Typography.Text>
              <Tag>{templateDraftFiles.length}</Tag>
            </div>
            <div className={styles.fileList}>
              {templateDraftFiles.map((file, index) => (
                <button
                  type="button"
                  key={`${index}:${file.path}`}
                  className={`${styles.fileItem} ${selectedTemplateFileIndex === index ? styles.fileItemActive : ''}`}
                  onClick={() => setSelectedTemplateFileIndex(index)}
                >
                  <span>{file.path || '默认配置文件'}</span>
                </button>
              ))}
            </div>
          </Col>
          <Col flex="1" className={styles.templatePreviewPanel}>
            <pre className={styles.configPreview}>
              {selectedTemplateFile?.content || '请选择模板'}
            </pre>
          </Col>
        </Row>
        <div className={styles.editorFooter}>
          <Typography.Text type="secondary">
            {selectedClusters.length} 个集群 · {targetCount} 个节点
          </Typography.Text>
          <Button
            type="primary"
            icon={<Send size={15} />}
            loading={saving}
            disabled={!selectedTemplate || !targetCount}
            onClick={dispatchTemplate}
          >
            按模板批量覆盖下发
          </Button>
        </div>
      </div>
    ) : (
      <div className={styles.tabContent}>
        {mode === 'global' && (
          <div className={styles.templateBar}>
            <Select
              value={selectedTemplateId || undefined}
              onChange={setSelectedTemplateId}
              options={matchingTemplates.map((template) => ({
                value: template.id,
                label: `${template.name} · v${template.version}`,
              }))}
            />
            <Button
              icon={<PlusOutlined />}
              onClick={() => {
                createForm.setFieldsValue({ probeType: selectedProbe, format: 'yaml' })
                setCreateOpen(true)
              }}
            >
              新建模板
            </Button>
          </div>
        )}
        {mode === 'node' && (
          <div className={styles.sourceToolbar}>
            <label className={styles.sourceGroup}>
              <span>探针</span>
              <Select
                value={selectedProbe || undefined}
                onChange={setSelectedProbe}
                options={availableProbeTypes.map((value) => ({
                  value,
                  label: `${value} · 版本 ${currentConfigs.find((config) => config.probeType === value)?.generation ?? 0}`,
                }))}
              />
            </label>
            <label className={styles.sourceGroup}>
              <span>历史版本</span>
              <Space.Compact>
                <Select
                  value={selectedHistoryGeneration}
                  placeholder="暂无历史版本"
                  onChange={setSelectedHistoryGeneration}
                  options={historyConfigs.map((config) => ({
                    value: config.generation,
                    label: `版本 ${config.generation}${config.generation === selectedCurrentConfig?.generation ? ' · 当前' : ''}`,
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
              <span>配置模板</span>
              <Space.Compact>
                <Select
                  value={selectedTemplateId || undefined}
                  placeholder="暂无可用模板"
                  onChange={setSelectedTemplateId}
                  options={matchingTemplates.map((template) => ({
                    value: template.id,
                    label: `${template.name} · v${template.version}`,
                  }))}
                />
                <Button
                  disabled={!selectedTemplate || !selectedFilePath}
                  onClick={loadTemplateIntoFile}
                >
                  使用模板
                </Button>
              </Space.Compact>
            </label>
          </div>
        )}
        {mode === 'node' ? (
          <Row className={styles.configWorkspace} wrap={false}>
            <Col flex="240px" className={styles.filePanel}>
              <div className={styles.filePanelTitle}>
                <FileText size={16} />
                <Typography.Text strong>配置文件</Typography.Text>
                <Tag>{selectedFiles.length}</Tag>
              </div>
              <div className={styles.fileList}>
                {selectedFiles.map((file) => {
                  const changed =
                    (configDrafts[draftKey(selectedProbe, file.path)] ?? file.content) !==
                    file.content
                  return (
                    <button
                      type="button"
                      key={file.path}
                      className={`${styles.fileItem} ${selectedFilePath === file.path ? styles.fileItemActive : ''}`}
                      onClick={() => setSelectedFilePath(file.path)}
                    >
                      <span>{file.path}</span>
                      {changed && <Tag color="warning">已修改</Tag>}
                    </button>
                  )
                })}
              </div>
            </Col>
            <Col flex="1" className={styles.diffPanel}>
              <DiffEditor
                key={`${selectedProbe}:${selectedFilePath}`}
                original={selectedFile?.content || ''}
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
                <Typography.Text strong>模板文件</Typography.Text>
                <Tag>{templateDraftFiles.length}</Tag>
              </div>
              <div className={styles.fileList}>
                {templateDraftFiles.map((file, index) => (
                  <button
                    type="button"
                    key={`${index}:${file.path}`}
                    className={`${styles.fileItem} ${selectedTemplateFileIndex === index ? styles.fileItemActive : ''}`}
                    onClick={() => setSelectedTemplateFileIndex(index)}
                  >
                    <span>{file.path || '默认配置文件'}</span>
                  </button>
                ))}
              </div>
              <Space.Compact block>
                <Button icon={<PlusOutlined />} onClick={addTemplateFile}>
                  新增文件
                </Button>
                <Button
                  icon={<Trash2 size={15} />}
                  disabled={templateDraftFiles.length <= 1}
                  onClick={removeTemplateFile}
                />
              </Space.Compact>
            </Col>
            <Col flex="1" className={styles.templateEditorPanel}>
              <Input
                value={selectedTemplateFile?.path || ''}
                placeholder="配置文件相对路径"
                onChange={(event) => updateTemplateFile({ path: event.target.value })}
              />
              <div className={styles.templateEditor}>
                <Editor
                  value={selectedTemplateFile?.content || ''}
                  language={languageForPath(
                    selectedTemplateFile?.path || '',
                    selectedTemplate?.format || 'yaml',
                  )}
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
          <Button
            type="primary"
            icon={<Save size={15} />}
            loading={saving}
            disabled={mode === 'node' ? !nodeConfigChanged : !selectedTemplate || !templateChanged}
            onClick={mode === 'node' ? saveNodeConfig : saveTemplate}
          >
            {mode === 'node' ? '保存并下发到当前节点' : '保存全局模板'}
          </Button>
        </div>
      </div>
    )

  const logSourcesTab = mode !== 'global' && (
    <div className={styles.tabContent}>
      <div className={styles.logSourceToolbar}>
        {mode === 'node' ? (
          <label className={styles.sourceGroup}>
            <span>探针</span>
            <Select
              value={selectedProbe || undefined}
              onChange={setSelectedProbe}
              options={availableProbeTypes.map((value) => ({ value, label: value }))}
            />
          </label>
        ) : (
          <label className={styles.sourceGroup}>
            <span>目标集群</span>
            <Select
              mode="multiple"
              value={selectedClusters}
              onChange={setSelectedClusters}
              options={clusterOptions.map((value) => ({ value, label: value }))}
            />
          </label>
        )}
        {mode === 'node' && (
          <Tag color={logSourcesConfigured ? 'processing' : 'default'}>
            {logSourcesConfigured ? '节点配置' : '部署默认'}
          </Tag>
        )}
        <Button
          icon={<PlusOutlined />}
          onClick={() => setLogSources((current) => [...current, newLogSource()])}
        >
          添加日志源
        </Button>
      </div>
      <div className={styles.logSourceList}>
        {logSources.map((source, index) => (
          <div className={styles.logSourceRow} key={`${index}:${source.type}`}>
            <Select
              aria-label={`日志源 ${index + 1} 类型`}
              value={source.type}
              onChange={(type) =>
                updateLogSource(index, {
                  type,
                  path: '',
                  pattern: '',
                  maxBytes: type === 'stdout' ? 10 * 1024 * 1024 : 0,
                })
              }
              options={[
                { value: 'file', label: '文件' },
                { value: 'stdout', label: '标准输出' },
              ]}
            />
            {source.type === 'file' ? (
              <>
                <Input
                  aria-label={`日志源 ${index + 1} 路径`}
                  value={source.path}
                  placeholder="/var/log/pods"
                  onChange={(event) => updateLogSource(index, { path: event.target.value })}
                />
                <Input
                  aria-label={`日志源 ${index + 1} 文件匹配`}
                  value={source.pattern}
                  placeholder="文件名正则（可选）"
                  onChange={(event) => updateLogSource(index, { pattern: event.target.value })}
                />
              </>
            ) : (
              <InputNumber
                aria-label={`日志源 ${index + 1} 保留大小`}
                min={0.0625}
                max={1024}
                step={1}
                value={(source.maxBytes || 10 * 1024 * 1024) / 1024 / 1024}
                addonAfter="MiB"
                onChange={(value) =>
                  updateLogSource(index, {
                    maxBytes: Math.round(Number(value || 0.0625) * 1024 * 1024),
                  })
                }
              />
            )}
            <Button
              aria-label={`删除日志源 ${index + 1}`}
              icon={<Trash2 size={15} />}
              onClick={() =>
                setLogSources((current) =>
                  current.filter((_, sourceIndex) => sourceIndex !== index),
                )
              }
            />
          </div>
        ))}
        {!logSources.length && !logSourcesLoading && (
          <Typography.Text type="secondary">暂无日志源</Typography.Text>
        )}
      </div>
      <div className={styles.editorFooter}>
        <Button
          type="primary"
          icon={<Save size={15} />}
          loading={saving || logSourcesLoading}
          disabled={!selectedProbe || (mode === 'batch' && !targetCount)}
          onClick={saveLogSources}
        >
          {mode === 'node' ? '保存日志配置' : `下发到 ${targetCount} 个节点`}
        </Button>
      </div>
    </div>
  )

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
      <div className={styles.circuitScopeBar}>
        <label>
          <span>设置范围</span>
          <Segmented
            value={circuitScope}
            options={[
              { value: 'all', label: '全部探针' },
              {
                value: 'probe',
                label: '单个探针',
                disabled: mode === 'global' || !circuitProbeTypes.length,
              },
            ]}
            onChange={(value) => setCircuitScope(value as 'all' | 'probe')}
          />
        </label>
        {circuitScope === 'probe' && (
          <label>
            <span>探针</span>
            <Select
              value={circuitProbe || undefined}
              onChange={setCircuitProbe}
              options={circuitProbeTypes.map((value) => ({ value, label: value }))}
            />
          </label>
        )}
        {circuitScope === 'probe' && hasProbeCircuitPolicy && (
          <Button icon={<Trash2 size={15} />} onClick={removeProbeCircuitPolicy}>
            移除单独设置
          </Button>
        )}
      </div>
      <CircuitEditor value={selectedCircuitPolicy} onChange={updateSelectedCircuitPolicy} />
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
          {mode !== 'node' && (activeTab === 'config' || activeTab === 'logs') && (
            <Space>
              <Typography.Text strong>目标探针</Typography.Text>
              <Select
                className={styles.probeSelect}
                value={selectedProbe || undefined}
                onChange={setSelectedProbe}
                options={availableProbeTypes.map((value) => ({ value, label: value }))}
              />
            </Space>
          )}
        </div>
        <Tabs
          activeKey={activeTab}
          onChange={setActiveTab}
          items={[
            { key: 'config', label: '配置文件', children: configTab },
            ...(mode === 'global'
              ? []
              : [{ key: 'logs', label: '日志配置', children: logSourcesTab }]),
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
        title="新建全局模板"
        open={createOpen}
        confirmLoading={saving}
        onOk={createTemplate}
        onCancel={() => setCreateOpen(false)}
      >
        <Form form={createForm} layout="vertical">
          <Form.Item name="name" label="模板名称" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="probeType" label="探针类型" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="path" label="配置文件相对路径" rules={[{ required: true }]}>
            <Input placeholder="例如 config/local/collection.yaml" />
          </Form.Item>
          <Form.Item name="format" label="格式" rules={[{ required: true }]}>
            <Select
              options={['yaml', 'json', 'xml', 'hcl', 'text'].map((value) => ({
                value,
                label: value,
              }))}
            />
          </Form.Item>
          <Form.Item name="content" label="模板内容" rules={[{ required: true }]}>
            <Input.TextArea rows={10} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
