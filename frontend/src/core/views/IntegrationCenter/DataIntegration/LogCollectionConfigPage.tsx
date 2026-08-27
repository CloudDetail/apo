import {
  Button,
  Empty,
  Input,
  InputNumber,
  Collapse,
  Checkbox,
  Dropdown,
  Modal,
  Segmented,
  Select,
  Space,
  Spin,
  Switch,
  Tooltip,
  Typography,
} from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import {
  Container as ContainerIcon,
  ExternalLink,
  Eye,
  FileText,
  GitBranch,
  RotateCcw,
  Save,
  ScanSearch,
  Send,
  Trash2,
  X,
} from 'lucide-react'
import { useCallback, useEffect, useState } from 'react'
import { useSelector } from 'react-redux'
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter'
import { oneDark, oneLight } from 'react-syntax-highlighter/dist/esm/styles/prism'
import {
  createAgentCommandApi,
  getAgentCommandsApi,
  getContainerStdoutCollectionApi,
  getHostFileCollectionApi,
  getLogCollectionDefaultsApi,
  previewContainerStdoutCollectionApi,
  previewHostFileCollectionApi,
  updateContainerStdoutCollectionApi,
  updateHostFileCollectionApi,
  updateLogCollectionDefaultsApi,
} from 'src/core/api/integration'
import { notify } from 'src/core/utils/notify'
import styles from './logCollectionConfigPage.module.scss'

type CollectionMode = 'host' | 'container'

type LogCollectionConfigPanelProps = {
  mode: CollectionMode
  nodeId?: string
  clusterIds?: string[]
  clusterOptions?: string[]
  showK8sFilters?: boolean
  onClusterIdsChange?: (clusterIds: string[]) => void
  inheritedClusterId?: string
  inheritedClusterEnvironment?: string
  onEditInherited?: () => void
  enableJavaLogScan?: boolean
  javaLogScanProbeType?: string
}

type HostFileSource = {
  id: string
  name: string
  enabled: boolean
  path: string
  maxDirSearchDepth: number
  filePattern: string
  processingRule?: string
  outputRule?: string
}

type FilterMapEntry = {
  id: string
  key: string
  value: string
}

type ContainerStdoutSource = {
  id: string
  name: string
  enabled: boolean
  namespaceRegex: string
  podRegex: string
  containerRegex: string
  collectStdout: boolean
  collectStderr: boolean
  maxEventsPerSecond: number
  ignoreParseWarning: boolean
  keepingSourceWhenParseFail: boolean
  tailSizeKB: number
  flushTimeoutSecs: number
  allowingIncludedByMultiConfigs: boolean
  multilineMode: 'off' | 'custom' | 'JSON'
  multilineStartPattern: string
  multilineContinuePattern: string
  multilineEndPattern: string
  unmatchedContentTreatment: 'single_line' | 'discard'
  includeK8sLabel: string
  excludeK8sLabel: FilterMapEntry[]
  includeEnv: string
  excludeEnv: string
  includeContainerLabel: string
  excludeContainerLabel: string
  externalK8sLabelTag: string
  externalEnvTag: string
  tags: string
  activeFilters: ContainerFilterField[]
  processingRule?: string
  outputRule?: string
}

type PipelineDefaults = {
  processingRule: string
  outputRule: string
  processorCount?: number
  outputCount?: number
}

type LogCollectionDefaults = {
  version: number
  file: PipelineDefaults
  container: PipelineDefaults
}

const EMPTY_PIPELINE_DEFAULTS: PipelineDefaults = {
  processingRule: '',
  outputRule: '',
}

type PreviewFile = {
  path: string
  content: string
  inputCount?: number
  processorCount?: number
  flusherCount?: number
}

type JavaLogScanDirectory = { path: string; fileCount?: number }

type ContainerFilterField =
  | 'namespaceRegex'
  | 'podRegex'
  | 'containerRegex'
  | 'includeK8sLabel'
  | 'excludeK8sLabel'
  | 'includeEnv'
  | 'excludeEnv'
  | 'includeContainerLabel'
  | 'excludeContainerLabel'

type ContainerTextField = 'externalK8sLabelTag' | 'externalEnvTag' | 'tags'

type ContainerFilterDefinition = {
  field: ContainerFilterField
  label: string
  placeholder: string
  k8sOnly?: boolean
  multiline?: boolean
  tags?: boolean
  map?: boolean
}

const containerFilterDefinitions: ContainerFilterDefinition[] = [
  {
    field: 'namespaceRegex',
    label: 'Namespace 正则',
    placeholder: '例如 ^production$',
    k8sOnly: true,
  },
  { field: 'podRegex', label: 'Pod 正则', placeholder: '例如 ^api-.*', k8sOnly: true },
  {
    field: 'containerRegex',
    label: 'Container 正则',
    placeholder: '例如 ^app$',
    k8sOnly: true,
  },
  {
    field: 'includeK8sLabel',
    label: '包含 Pod 标签',
    placeholder: '每行填写 key=value',
    k8sOnly: true,
    multiline: true,
  },
  {
    field: 'excludeK8sLabel',
    label: '排除 Pod 标签',
    placeholder: '配置要排除的 Pod 标签',
    k8sOnly: true,
    map: true,
  },
  {
    field: 'includeEnv',
    label: '包含环境变量',
    placeholder: '每行填写 key=value',
    multiline: true,
  },
  {
    field: 'excludeEnv',
    label: '排除环境变量',
    placeholder: '每行填写 key=value',
    multiline: true,
  },
  {
    field: 'includeContainerLabel',
    label: '包含容器标签',
    placeholder: '每行填写 key=value',
    multiline: true,
  },
  {
    field: 'excludeContainerLabel',
    label: '排除容器标签',
    placeholder: '每行填写 key=value',
    multiline: true,
  },
]

const tagMappingFields: [ContainerTextField, string][] = [
  ['externalK8sLabelTag', 'Pod 标签附加为 Tag'],
  ['externalEnvTag', '环境变量附加为 Tag'],
  ['tags', 'Tag 重命名或删除'],
]

const createSourceId = () =>
  typeof crypto !== 'undefined' && crypto.randomUUID ? crypto.randomUUID() : `source-${Date.now()}`

const newFilterMapEntry = (key = '', value = ''): FilterMapEntry => ({
  id: createSourceId(),
  key,
  value,
})

const normalizeFilterMapEntries = (value: unknown): FilterMapEntry[] => {
  if (Array.isArray(value)) {
    return value.map((entry) => {
      if (typeof entry === 'string') {
        const [key, ...rest] = entry.split('=')
        return newFilterMapEntry(key.trim(), rest.join('=').trim())
      }
      const item = entry as Partial<FilterMapEntry>
      return {
        id: item.id || createSourceId(),
        key: String(item.key || ''),
        value: String(item.value || ''),
      }
    })
  }
  if (typeof value === 'string') {
    return value
      .split('\n')
      .map((line) => line.trim())
      .filter(Boolean)
      .map((line) => {
        const [key, ...rest] = line.split('=')
        return newFilterMapEntry(key.trim(), rest.join('=').trim())
      })
  }
  if (value && typeof value === 'object') {
    return Object.entries(value as Record<string, unknown>).map(([key, entryValue]) =>
      newFilterMapEntry(key, String(entryValue ?? '')),
    )
  }
  return []
}

const filterMapPayload = (entries: FilterMapEntry[]) =>
  Object.fromEntries(entries.map(({ key, value }) => [key.trim(), value.trim()]))

const filterHasValue = (value: unknown) => {
  if (Array.isArray(value)) return value.length > 0
  return typeof value === 'string' && value.trim().length > 0
}

const formatFilterValue = (value: string | string[] | FilterMapEntry[]) => {
  if (!Array.isArray(value)) return value
  if (!value.length) return ''
  if (typeof value[0] === 'string') return (value as string[]).join('、')
  return (value as FilterMapEntry[])
    .map(({ key, value: entryValue }) => `${key}${entryValue ? `=${entryValue}` : ''}`)
    .join('、')
}

const newHostFile = (path = '/var/log', filePattern = '*.log'): HostFileSource => {
  const id = createSourceId()
  return {
    id,
    name: `host-file-${id}`,
    enabled: true,
    path,
    maxDirSearchDepth: 0,
    filePattern,
  }
}

const newContainerStdout = (): ContainerStdoutSource => {
  const id = createSourceId()
  return {
    id,
    name: `container-stdout-${id}`,
    enabled: true,
    namespaceRegex: '',
    podRegex: '',
    containerRegex: '',
    collectStdout: true,
    collectStderr: true,
    maxEventsPerSecond: 0,
    ignoreParseWarning: false,
    keepingSourceWhenParseFail: true,
    tailSizeKB: 1024,
    flushTimeoutSecs: 5,
    allowingIncludedByMultiConfigs: false,
    multilineMode: 'off',
    multilineStartPattern: '',
    multilineContinuePattern: '',
    multilineEndPattern: '',
    unmatchedContentTreatment: 'single_line',
    includeK8sLabel: '',
    excludeK8sLabel: [],
    includeEnv: '',
    excludeEnv: '',
    includeContainerLabel: '',
    excludeContainerLabel: '',
    externalK8sLabelTag: '',
    externalEnvTag: '',
    tags: '',
    activeFilters: [],
  }
}

const hostPayload = (sources: HostFileSource[]) =>
  sources.map((source, index) => ({
    ...source,
    name: source.name.trim() || `host-file-${source.id || index + 1}`,
  }))

const normalizeContainerSources = (sources: Partial<ContainerStdoutSource>[]) =>
  sources.map((source, index) => {
    const normalized = {
      ...newContainerStdout(),
      ...source,
      name: source.name?.trim() || `container-stdout-${source.id || index + 1}`,
      excludeK8sLabel: normalizeFilterMapEntries(source.excludeK8sLabel),
    }
    return {
      ...normalized,
      activeFilters:
        source.activeFilters ||
        containerFilterDefinitions
          .map(({ field }) => field)
          .filter((field) => {
            return filterHasValue(normalized[field])
          }),
    }
  })

const containerPayload = (sources: ContainerStdoutSource[]) =>
  normalizeContainerSources(sources).map(
    ({ activeFilters: _activeFilters, excludeK8sLabel, ...source }) => ({
      ...source,
      excludeK8sLabel: filterMapPayload(excludeK8sLabel),
    }),
  )

const containerSourcesForEnvironment = (
  sources: Partial<ContainerStdoutSource>[],
  showK8sFilters: boolean,
) =>
  normalizeContainerSources(sources).map((source) => {
    if (showK8sFilters) return source
    return {
      ...source,
      namespaceRegex: '',
      podRegex: '',
      containerRegex: '',
      includeK8sLabel: '',
      excludeK8sLabel: [],
      activeFilters: source.activeFilters.filter(
        (field) =>
          !containerFilterDefinitions.find((definition) => definition.field === field)?.k8sOnly,
      ),
    }
  })

function RulePipelineEditors({
  source,
  defaults,
  onChange,
}: {
  source: Pick<HostFileSource, 'processingRule' | 'outputRule'>
  defaults: PipelineDefaults
  onChange: (
    patch: Pick<HostFileSource, 'processingRule'> | Pick<HostFileSource, 'outputRule'>,
  ) => void
}) {
  const processingCustomized = source.processingRule !== undefined
  const outputCustomized = source.outputRule !== undefined
  return (
    <Collapse
      className={styles.rulePipeline}
      ghost
      size="small"
      items={[
        {
          key: 'processing',
          label: (
            <span className={styles.pipelineRuleTitle}>
              <span>处理规则</span>
              <Typography.Text type="secondary">
                {processingCustomized ? '已自定义' : '使用默认处理规则'}
              </Typography.Text>
            </span>
          ),
          children: (
            <div className={styles.pipelineEditorSection}>
              <div className={styles.pipelineEditorHeader}>
                <Typography.Text type="secondary">
                  iLogtail 原生 YAML，顶层字段固定为 processors。
                </Typography.Text>
                <Button
                  type="text"
                  size="small"
                  icon={<RotateCcw size={14} />}
                  disabled={!processingCustomized}
                  onClick={() => onChange({ processingRule: undefined })}
                >
                  恢复默认
                </Button>
              </div>
              <Input.TextArea
                className={styles.pipelineEditor}
                aria-label="处理规则 YAML"
                value={source.processingRule ?? defaults.processingRule}
                placeholder={'processors:\n  - Type: processor_example'}
                autoSize={{ minRows: 6, maxRows: 18 }}
                spellCheck={false}
                onChange={(event) => onChange({ processingRule: event.target.value })}
              />
            </div>
          ),
        },
        {
          key: 'output',
          label: (
            <span className={styles.pipelineRuleTitle}>
              <span>输出规则</span>
              <Typography.Text type="secondary">
                {outputCustomized ? '已自定义' : '使用默认输出规则'}
              </Typography.Text>
            </span>
          ),
          children: (
            <div className={styles.pipelineEditorSection}>
              <div className={styles.pipelineEditorHeader}>
                <Typography.Text type="secondary">
                  iLogtail 原生 YAML，顶层字段固定为 flushers，至少保留一个输出。
                </Typography.Text>
                <Button
                  type="text"
                  size="small"
                  icon={<RotateCcw size={14} />}
                  disabled={!outputCustomized}
                  onClick={() => onChange({ outputRule: undefined })}
                >
                  恢复默认
                </Button>
              </div>
              <Input.TextArea
                className={styles.pipelineEditor}
                aria-label="输出规则 YAML"
                value={source.outputRule ?? defaults.outputRule}
                placeholder={'flushers:\n  - Type: flusher_http'}
                autoSize={{ minRows: 6, maxRows: 18 }}
                spellCheck={false}
                onChange={(event) => onChange({ outputRule: event.target.value })}
              />
            </div>
          ),
        },
      ]}
    />
  )
}

export function LogCollectionDefaultsPanel() {
  const [kind, setKind] = useState<'container' | 'file'>('container')
  const [value, setValue] = useState<LogCollectionDefaults>({
    version: 0,
    file: EMPTY_PIPELINE_DEFAULTS,
    container: EMPTY_PIPELINE_DEFAULTS,
  })
  const [original, setOriginal] = useState('')
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const changed = Boolean(original) && JSON.stringify(value) !== original
  const current = value[kind]

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const result: any = await getLogCollectionDefaultsApi()
      const next = {
        version: Number(result?.version || 0),
        file: { ...EMPTY_PIPELINE_DEFAULTS, ...(result?.file || {}) },
        container: { ...EMPTY_PIPELINE_DEFAULTS, ...(result?.container || {}) },
      }
      setValue(next)
      setOriginal(JSON.stringify(next))
    } catch (error: any) {
      console.error('[LogCollectionDefaultsPanel] failed to load defaults', error)
      notify({ type: 'error', message: error.message || '加载日志默认配置失败' })
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    load()
  }, [load])

  const updateCurrent = (patch: Partial<PipelineDefaults>) =>
    setValue((currentValue) => ({
      ...currentValue,
      [kind]: { ...currentValue[kind], ...patch },
    }))

  const save = async () => {
    setSaving(true)
    try {
      const result: any = await updateLogCollectionDefaultsApi({
        file: value.file,
        container: value.container,
      })
      const next = {
        version: Number(result?.version || value.version + 1),
        file: { ...EMPTY_PIPELINE_DEFAULTS, ...(result?.file || value.file) },
        container: { ...EMPTY_PIPELINE_DEFAULTS, ...(result?.container || value.container) },
      }
      setValue(next)
      setOriginal(JSON.stringify(next))
      notify({ type: 'success', message: '日志采集默认规则已保存' })
    } catch (error: any) {
      console.error('[LogCollectionDefaultsPanel] failed to save defaults', error)
      notify({ type: 'error', message: error.message || '保存日志默认配置失败' })
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className={styles.panel}>
      <div className={styles.pageHeading}>
        <div>
          <Typography.Text strong>日志采集默认规则</Typography.Text>
          <Typography.Text type="secondary">
            新增采集规则默认继承同类型的处理和输出规则，单条规则可独立覆盖
          </Typography.Text>
        </div>
        <Typography.Text type="secondary">
          {value.version ? `版本 ${value.version}` : '使用已上报的 iLogtail 配置作为初始值'}
        </Typography.Text>
      </div>
      <Spin spinning={loading}>
        <section className={styles.workbench}>
          <div className={styles.defaultsToolbar}>
            <Segmented
              value={kind}
              options={[
                { value: 'container', label: '容器标准输出' },
                { value: 'file', label: '文件采集' },
              ]}
              onChange={(next) => setKind(next as 'container' | 'file')}
            />
            <Typography.Text type="secondary">
              当前规则作为{kind === 'container' ? '容器标准输出' : '文件采集'}的默认流水线
            </Typography.Text>
          </div>
          <Collapse
            className={styles.pipelineRules}
            expandIconPosition="end"
            items={[
              {
                key: 'processing',
                label: (
                  <span className={styles.pipelineRuleTitle}>
                    <span>默认处理规则</span>
                    <Typography.Text type="secondary">
                      {current.processorCount ?? 0} 个处理器
                    </Typography.Text>
                  </span>
                ),
                children: (
                  <Input.TextArea
                    className={styles.pipelineEditor}
                    aria-label={`${kind === 'container' ? '容器' : '文件'}默认处理规则 YAML`}
                    value={current.processingRule}
                    placeholder={'processors:\n  - Type: processor_example'}
                    autoSize={{ minRows: 8, maxRows: 24 }}
                    spellCheck={false}
                    onChange={(event) => updateCurrent({ processingRule: event.target.value })}
                  />
                ),
              },
              {
                key: 'output',
                label: (
                  <span className={styles.pipelineRuleTitle}>
                    <span>默认输出规则</span>
                    <Typography.Text type="secondary">
                      {current.outputCount ?? 0} 个输出
                    </Typography.Text>
                  </span>
                ),
                children: (
                  <Input.TextArea
                    className={styles.pipelineEditor}
                    aria-label={`${kind === 'container' ? '容器' : '文件'}默认输出规则 YAML`}
                    value={current.outputRule}
                    placeholder={'flushers:\n  - Type: flusher_http'}
                    autoSize={{ minRows: 8, maxRows: 24 }}
                    spellCheck={false}
                    onChange={(event) => updateCurrent({ outputRule: event.target.value })}
                  />
                ),
              },
            ]}
          />
          <div className={styles.footer}>
            <Typography.Text type="secondary">
              {changed ? '有未保存修改' : '容器与文件默认规则分别管理'}
            </Typography.Text>
            <Button
              type="primary"
              icon={<Save size={15} />}
              loading={saving}
              disabled={!changed}
              onClick={save}
            >
              保存默认规则
            </Button>
          </div>
        </section>
      </Spin>
    </div>
  )
}

export function LogCollectionConfigPanel({
  mode,
  nodeId = '',
  clusterIds = [],
  clusterOptions = [],
  showK8sFilters = true,
  onClusterIdsChange,
  inheritedClusterId = '',
  inheritedClusterEnvironment = '',
  onEditInherited,
  enableJavaLogScan = false,
  javaLogScanProbeType = '',
}: LogCollectionConfigPanelProps) {
  const theme = useSelector((state: any) => state.settingReducer.theme)
  const selectedNode = nodeId
  const selectedCluster = clusterIds[0] || ''
  const [hostFiles, setHostFiles] = useState<HostFileSource[]>([])
  const [containerStdout, setContainerStdout] = useState<ContainerStdoutSource[]>([])
  const [inheritedContainerStdout, setInheritedContainerStdout] = useState<ContainerStdoutSource[]>(
    [],
  )
  const [inheritedHostFiles, setInheritedHostFiles] = useState<HostFileSource[]>([])
  const [hostVersion, setHostVersion] = useState(0)
  const [containerVersion, setContainerVersion] = useState(0)
  const [inheritedVersion, setInheritedVersion] = useState(0)
  const [originalHostFiles, setOriginalHostFiles] = useState('[]')
  const [originalContainerStdout, setOriginalContainerStdout] = useState('[]')
  const [disableClusterRules, setDisableClusterRules] = useState(false)
  const [originalDisableClusterRules, setOriginalDisableClusterRules] = useState(false)
  const [loading, setLoading] = useState(true)
  const [loadFailed, setLoadFailed] = useState(false)
  const [inheritedLoading, setInheritedLoading] = useState(false)
  const [inheritedLoadFailed, setInheritedLoadFailed] = useState(false)
  const [saving, setSaving] = useState(false)
  const [previewLoading, setPreviewLoading] = useState(false)
  const [previewOpen, setPreviewOpen] = useState(false)
  const [previewFiles, setPreviewFiles] = useState<PreviewFile[]>([])
  const [selectedPreviewPath, setSelectedPreviewPath] = useState('')
  const [scanningJavaLogs, setScanningJavaLogs] = useState(false)
  const [pipelineDefaults, setPipelineDefaults] = useState<LogCollectionDefaults>({
    version: 0,
    file: EMPTY_PIPELINE_DEFAULTS,
    container: EMPTY_PIPELINE_DEFAULTS,
  })

  const selectedPreviewFile = previewFiles.find((file) => file.path === selectedPreviewPath)
  const hostChanged = JSON.stringify(hostFiles) !== originalHostFiles
  const containerChanged = JSON.stringify(containerStdout) !== originalContainerStdout
  const nodeChanged =
    hostChanged || containerChanged || disableClusterRules !== originalDisableClusterRules
  const changed = mode === 'host' ? nodeChanged : hostChanged || containerChanged
  const currentScope = mode === 'host' ? selectedNode : selectedCluster
  const currentVersion = mode === 'host' ? hostVersion : containerVersion
  const sourceCount = containerStdout.length + hostFiles.length
  const inheritedSourceCount =
    inheritedContainerStdout.filter((source) => source.enabled).length +
    inheritedHostFiles.filter((source) => source.enabled).length
  const inheritedRuleCount = inheritedContainerStdout.length + inheritedHostFiles.length
  const visibleRuleCount =
    mode === 'host'
      ? inheritedContainerStdout.length + inheritedHostFiles.length + sourceCount
      : containerStdout.length + hostFiles.length
  const availableFilterDefinitions = containerFilterDefinitions.filter(
    (definition) => showK8sFilters || !definition.k8sOnly,
  )

  const loadHostFiles = useCallback(async () => {
    if (!selectedNode) {
      setHostFiles([])
      setContainerStdout([])
      setOriginalHostFiles('[]')
      setOriginalContainerStdout('[]')
      setDisableClusterRules(false)
      setOriginalDisableClusterRules(false)
      setLoading(false)
      return
    }
    setLoading(true)
    setLoadFailed(false)
    try {
      const result: any = await getHostFileCollectionApi(selectedNode)
      const values = hostPayload((result?.hostFiles || []) as HostFileSource[])
      setHostFiles(values)
      setOriginalHostFiles(JSON.stringify(values))
      const containerValues = containerSourcesForEnvironment(
        (result?.containerStdout || []) as ContainerStdoutSource[],
        showK8sFilters,
      )
      setContainerStdout(containerValues)
      setOriginalContainerStdout(JSON.stringify(containerValues))
      const inheritedDisabled = Boolean(result?.disableClusterRules)
      setDisableClusterRules(inheritedDisabled)
      setOriginalDisableClusterRules(inheritedDisabled)
      setHostVersion(Number(result?.version || 0))
    } catch (error: any) {
      console.error('[LogCollectionConfigPage] failed to load host file collection', error)
      setHostFiles([])
      setContainerStdout([])
      setOriginalHostFiles('[]')
      setOriginalContainerStdout('[]')
      setDisableClusterRules(false)
      setOriginalDisableClusterRules(false)
      setLoadFailed(true)
      notify({ type: 'error', message: error.message || '加载宿主机文件采集配置失败' })
    } finally {
      setLoading(false)
    }
  }, [selectedNode, showK8sFilters])

  const loadContainerStdout = useCallback(async () => {
    if (!selectedCluster) {
      setContainerStdout([])
      setHostFiles([])
      setOriginalContainerStdout('[]')
      setOriginalHostFiles('[]')
      setLoading(false)
      return
    }
    setLoading(true)
    setLoadFailed(false)
    try {
      const result: any = await getContainerStdoutCollectionApi(selectedCluster)
      const values = containerSourcesForEnvironment(
        (result?.containerStdout || []) as ContainerStdoutSource[],
        showK8sFilters,
      )
      setContainerStdout(values)
      const fileValues = hostPayload((result?.hostFiles || []) as HostFileSource[])
      setHostFiles(fileValues)
      setOriginalContainerStdout(JSON.stringify(values))
      setOriginalHostFiles(JSON.stringify(fileValues))
      setContainerVersion(Number(result?.version || 0))
    } catch (error: any) {
      console.error('[LogCollectionConfigPage] failed to load container stdout collection', error)
      setContainerStdout([])
      setHostFiles([])
      setOriginalContainerStdout('[]')
      setOriginalHostFiles('[]')
      setLoadFailed(true)
      notify({ type: 'error', message: error.message || '加载容器标准输出采集配置失败' })
    } finally {
      setLoading(false)
    }
  }, [selectedCluster, showK8sFilters])

  const loadInheritedContainerStdout = useCallback(async () => {
    if (mode !== 'host' || !inheritedClusterId) {
      setInheritedContainerStdout([])
      setInheritedHostFiles([])
      setInheritedVersion(0)
      setInheritedLoadFailed(false)
      return
    }
    setInheritedLoading(true)
    setInheritedLoadFailed(false)
    try {
      const result: any = await getContainerStdoutCollectionApi(inheritedClusterId)
      setInheritedContainerStdout(
        normalizeContainerSources((result?.containerStdout || []) as ContainerStdoutSource[]),
      )
      setInheritedHostFiles(hostPayload((result?.hostFiles || []) as HostFileSource[]))
      setInheritedVersion(Number(result?.version || 0))
    } catch (error) {
      console.error('[LogCollectionConfigPage] failed to load inherited cluster collection', error)
      setInheritedContainerStdout([])
      setInheritedHostFiles([])
      setInheritedVersion(0)
      setInheritedLoadFailed(true)
    } finally {
      setInheritedLoading(false)
    }
  }, [inheritedClusterId, mode])

  useEffect(() => {
    if (mode === 'host') loadHostFiles()
    else loadContainerStdout()
  }, [loadContainerStdout, loadHostFiles, mode])

  useEffect(() => {
    loadInheritedContainerStdout()
  }, [loadInheritedContainerStdout])

  useEffect(() => {
    getLogCollectionDefaultsApi()
      .then((result: any) =>
        setPipelineDefaults({
          version: Number(result?.version || 0),
          file: { ...EMPTY_PIPELINE_DEFAULTS, ...(result?.file || {}) },
          container: { ...EMPTY_PIPELINE_DEFAULTS, ...(result?.container || {}) },
        }),
      )
      .catch((error) => {
        console.error('[LogCollectionConfigPage] failed to load pipeline defaults', error)
      })
  }, [])

  const updateHostFile = (index: number, patch: Partial<HostFileSource>) => {
    setHostFiles((current) =>
      current.map((source, sourceIndex) =>
        sourceIndex === index ? { ...source, ...patch } : source,
      ),
    )
  }

  const updateContainerStdout = (index: number, patch: Partial<ContainerStdoutSource>) => {
    setContainerStdout((current) =>
      current.map((source, sourceIndex) =>
        sourceIndex === index ? { ...source, ...patch } : source,
      ),
    )
  }

  const updateContainerFilter = (
    index: number,
    field: ContainerFilterField,
    value: string | string[] | FilterMapEntry[],
  ) => {
    updateContainerStdout(index, { [field]: value } as Partial<ContainerStdoutSource>)
  }

  const updateFilterMapEntry = (
    sourceIndex: number,
    entryId: string,
    patch: Partial<FilterMapEntry>,
  ) => {
    const source = containerStdout[sourceIndex]
    updateContainerFilter(
      sourceIndex,
      'excludeK8sLabel',
      source.excludeK8sLabel.map((entry) =>
        entry.id === entryId ? { ...entry, ...patch } : entry,
      ),
    )
  }

  const validateContainerFilters = () => {
    const incomplete = containerStdout.some((source) =>
      source.activeFilters.some((field) => {
        const value = source[field]
        if (field === 'excludeK8sLabel') {
          const entries = value as FilterMapEntry[]
          const keys = entries.map(({ key }) => key.trim()).filter(Boolean)
          return (
            entries.length === 0 ||
            keys.length !== entries.length ||
            new Set(keys).size !== keys.length
          )
        }
        return !filterHasValue(value)
      }),
    )
    if (!incomplete) return true
    notify({ type: 'error', message: '请填写已添加过滤器的匹配条件，或移除该过滤器' })
    return false
  }

  const save = async () => {
    if (!currentScope || loadFailed) return
    if (!validateContainerFilters()) return
    setSaving(true)
    try {
      if (mode === 'host') {
        const result: any = await updateHostFileCollectionApi(selectedNode, {
          containerStdout: containerPayload(containerStdout),
          hostFiles: hostPayload(hostFiles),
          disableClusterRules,
        })
        const values = hostPayload((result?.hostFiles || []) as HostFileSource[])
        const inheritedDisabled =
          typeof result?.disableClusterRules === 'boolean'
            ? result.disableClusterRules
            : disableClusterRules
        setHostFiles(values)
        setOriginalHostFiles(JSON.stringify(values))
        const containerValues = containerSourcesForEnvironment(
          (result?.containerStdout || []) as ContainerStdoutSource[],
          showK8sFilters,
        )
        setContainerStdout(containerValues)
        setOriginalContainerStdout(JSON.stringify(containerValues))
        setDisableClusterRules(inheritedDisabled)
        setOriginalDisableClusterRules(inheritedDisabled)
        setHostVersion(Number(result?.version || hostVersion + 1))
        notify({ type: 'success', message: '节点日志采集配置已保存' })
      } else {
        const payload = {
          containerStdout: containerPayload(containerStdout),
          hostFiles: hostPayload(hostFiles),
        }
        const results = await Promise.all(
          clusterIds.map((clusterId) => updateContainerStdoutCollectionApi(clusterId, payload)),
        )
        const result: any = results[0]
        const values = containerSourcesForEnvironment(
          (result?.containerStdout || []) as ContainerStdoutSource[],
          showK8sFilters,
        )
        setContainerStdout(values)
        const fileValues = hostPayload((result?.hostFiles || []) as HostFileSource[])
        setHostFiles(fileValues)
        setOriginalContainerStdout(JSON.stringify(values))
        setOriginalHostFiles(JSON.stringify(fileValues))
        setContainerVersion(Number(result?.version || containerVersion + 1))
        notify({
          type: 'success',
          message: `日志采集配置已保存到 ${clusterIds.length} 个集群`,
        })
      }
    } catch (error: any) {
      console.error('[LogCollectionConfigPage] failed to save collection config', error)
      notify({ type: 'error', message: error.message || '日志采集配置保存失败' })
    } finally {
      setSaving(false)
    }
  }

  const preview = async () => {
    if (
      !currentScope ||
      loadFailed ||
      (mode === 'host' && !disableClusterRules && inheritedLoadFailed)
    )
      return
    if (!validateContainerFilters()) return
    setPreviewLoading(true)
    try {
      let files: PreviewFile[]
      if (mode === 'host') {
        const [nodeResult, clusterResult]: any[] = await Promise.all([
          previewHostFileCollectionApi(selectedNode, {
            containerStdout: containerPayload(containerStdout),
            hostFiles: hostPayload(hostFiles),
            disableClusterRules,
          }),
          inheritedClusterId && !disableClusterRules
            ? previewContainerStdoutCollectionApi(inheritedClusterId, {
                containerStdout: containerPayload(inheritedContainerStdout),
                hostFiles: hostPayload(inheritedHostFiles),
              })
            : Promise.resolve({ files: [] }),
        ])
        files = [...(clusterResult?.files || []), ...(nodeResult?.files || [])]
      } else {
        const result: any = await previewContainerStdoutCollectionApi(selectedCluster, {
          containerStdout: containerPayload(containerStdout),
          hostFiles: hostPayload(hostFiles),
        })
        files = (result?.files || []) as PreviewFile[]
      }
      setPreviewFiles(files)
      setSelectedPreviewPath(files[0]?.path || '')
      setPreviewOpen(true)
    } catch (error: any) {
      console.error('[LogCollectionConfigPage] failed to preview collection config', error)
      notify({ type: 'error', message: error.message || '生成配置预览失败' })
    } finally {
      setPreviewLoading(false)
    }
  }

  const addHostFile = () => setHostFiles((current) => [...current, newHostFile()])
  const addContainerStdout = () =>
    setContainerStdout((current) => [...current, newContainerStdout()])

  const scanJavaLogs = async () => {
    if (!selectedNode || !javaLogScanProbeType) return
    setScanningJavaLogs(true)
    try {
      const created: any = await createAgentCommandApi(
        selectedNode,
        javaLogScanProbeType,
        'SCAN_JAVA_LOG_FILES',
      )
      let command: any
      for (let attempt = 0; attempt < 75; attempt += 1) {
        const response: any = await getAgentCommandsApi(selectedNode, javaLogScanProbeType)
        command = (response?.commands || []).find((item: any) => item.id === created.id)
        if (command?.status === 'SUCCEEDED' || command?.status === 'FAILED') break
        await new Promise((resolve) => window.setTimeout(resolve, 1000))
      }
      if (!command || (command.status !== 'SUCCEEDED' && command.status !== 'FAILED')) {
        throw new Error('等待节点返回扫描结果超时')
      }
      let result: any = {}
      try {
        result = command.result ? JSON.parse(command.result) : {}
      } catch (error) {
        console.error('[LogCollectionConfigPage] failed to parse Java log scan result', error)
      }
      if (command.status === 'FAILED') {
        throw new Error(result?.error?.message || command.result || 'Java 日志扫描失败')
      }
      const directories = ((result.directories || []) as JavaLogScanDirectory[])
        .map((directory) => directory.path?.trim())
        .filter(Boolean) as string[]
      if (!directories.length) {
        notify({
          type: 'warning',
          message: result.javaProcesses
            ? `发现 ${result.javaProcesses} 个 Java 进程，但未找到日志文件`
            : '未发现正在运行的 Java 服务',
        })
        return
      }
      const existingPaths = new Set(hostFiles.map((source) => source.path.trim()))
      const additions = directories
        .filter((path) => !existingPaths.has(path))
        .map((path) => newHostFile(path, '*.log*'))
      if (additions.length) {
        setHostFiles((current) => [...current, ...additions])
      }
      notify({
        type: additions.length ? 'success' : 'info',
        message: additions.length
          ? `扫描到 ${directories.length} 个日志目录，已新增 ${additions.length} 条采集规则`
          : `扫描到的 ${directories.length} 个日志目录均已配置`,
      })
    } catch (error: any) {
      console.error('[LogCollectionConfigPage] failed to scan Java logs', error)
      notify({ type: 'error', message: error.message || 'Java 日志扫描失败' })
    } finally {
      setScanningJavaLogs(false)
    }
  }

  const statusText = loading
    ? '正在加载'
    : loadFailed
      ? '配置加载失败'
      : changed
        ? '有未保存修改'
        : '配置已同步'

  return (
    <div className={styles.panel}>
      <div className={styles.pageHeading}>
        <div>
          <Typography.Text strong>日志采集配置</Typography.Text>
          <Typography.Text type="secondary">
            {mode === 'host'
              ? '查看集群继承规则，并管理当前节点追加的容器或文件采集规则'
              : '管理集群统一生效的容器标准输出和文件采集规则'}
          </Typography.Text>
        </div>
        <div className={`${styles.syncStatus} ${loadFailed ? styles.syncStatusError : ''}`}>
          <span />
          {statusText}
        </div>
      </div>

      <Spin spinning={loading}>
        <section className={styles.workbench}>
          {mode === 'host' && (
            <div
              className={`${styles.inheritedSection} ${
                disableClusterRules ? styles.inheritedSectionDisabled : ''
              }`}
            >
              <div className={styles.inheritedHeader}>
                <div>
                  <Typography.Text strong>集群继承 · {inheritedClusterId}</Typography.Text>
                  <Typography.Text type="secondary">
                    {disableClusterRules
                      ? '当前节点已停用集群日志采集规则'
                      : inheritedClusterEnvironment
                        ? `${inheritedClusterEnvironment} · 集群配置自动应用到当前节点`
                        : '集群配置自动应用到当前节点'}
                  </Typography.Text>
                </div>
                <div className={styles.inheritanceActions}>
                  {!!inheritedVersion && (
                    <Typography.Text type="secondary">版本 {inheritedVersion}</Typography.Text>
                  )}
                  <label className={styles.inheritanceToggle}>
                    <span>应用集群规则</span>
                    <Switch
                      size="small"
                      checked={!disableClusterRules}
                      aria-label="应用集群日志采集规则"
                      onChange={(checked) => setDisableClusterRules(!checked)}
                    />
                  </label>
                  <Button
                    icon={<ExternalLink size={14} />}
                    disabled={!inheritedClusterId || !onEditInherited}
                    onClick={onEditInherited}
                  >
                    修改集群日志采集
                  </Button>
                </div>
              </div>
              <Spin spinning={inheritedLoading}>
                {inheritedLoadFailed && !disableClusterRules ? (
                  <Typography.Text type="danger">
                    集群日志采集配置加载失败，无法生成完整预览
                  </Typography.Text>
                ) : (
                  <Typography.Text type="secondary">
                    {disableClusterRules
                      ? inheritedRuleCount
                        ? `${inheritedRuleCount} 项集群规则已停用，不会进入当前节点最终配置`
                        : '集群规则已停用'
                      : inheritedRuleCount
                        ? `已继承 ${inheritedRuleCount} 项集群采集规则，详情见下方采集规则`
                        : '集群尚未配置采集规则'}
                  </Typography.Text>
                )}
              </Spin>
            </div>
          )}

          <div className={styles.scopeToolbar}>
            {mode === 'container' && (
              <label className={styles.field}>
                <span>目标集群</span>
                <Select
                  mode="multiple"
                  showSearch
                  value={clusterIds}
                  optionFilterProp="label"
                  placeholder="选择一个或多个集群"
                  onChange={(values) => onClusterIdsChange?.(values)}
                  options={clusterOptions.map((clusterId) => ({
                    value: clusterId,
                    label: clusterId,
                  }))}
                />
              </label>
            )}

            {mode === 'host' && (
              <div className={styles.scopeIdentity}>
                <Typography.Text strong>当前节点配置</Typography.Text>
                <Typography.Text type="secondary">
                  可追加当前节点独立需要的容器标准输出或文件采集规则
                </Typography.Text>
              </div>
            )}

            <Typography.Text className={styles.version} type="secondary">
              {mode === 'container' && clusterIds.length > 1
                ? `以 ${selectedCluster} 的配置为编辑起点 · 共 ${clusterIds.length} 个集群`
                : currentVersion
                  ? `版本 ${currentVersion}`
                  : '尚未配置'}
            </Typography.Text>
          </div>

          <div className={styles.rulesHeader}>
            <div className={styles.rulesTitle}>
              <Typography.Text strong>采集规则</Typography.Text>
              <Typography.Text type="secondary">{visibleRuleCount} 项</Typography.Text>
            </div>
            <Space>
              {mode === 'host' && enableJavaLogScan && (
                <Tooltip title={javaLogScanProbeType ? '' : '当前节点没有可用的探针命令通道'}>
                  <Button
                    icon={<ScanSearch size={15} />}
                    loading={scanningJavaLogs}
                    disabled={!currentScope || !javaLogScanProbeType || loadFailed}
                    onClick={scanJavaLogs}
                  >
                    自动扫描 Java 日志
                  </Button>
                </Tooltip>
              )}
              <Dropdown
                trigger={['click']}
                menu={{
                  items: [
                    {
                      key: 'container',
                      icon: <ContainerIcon size={14} />,
                      label: '容器标准输出采集',
                    },
                    { key: 'file', icon: <FileText size={14} />, label: '文件采集' },
                  ],
                  onClick: ({ key }) =>
                    key === 'container' ? addContainerStdout() : addHostFile(),
                }}
              >
                <Button icon={<PlusOutlined />} disabled={!currentScope || loadFailed}>
                  新增采集规则
                </Button>
              </Dropdown>
            </Space>
          </div>

          <div className={styles.sourceList}>
            {mode === 'host' &&
              inheritedHostFiles.map((source, index) => (
                <div
                  className={`${styles.inheritedSourceRow} ${
                    disableClusterRules ? styles.inheritedSourceRowDisabled : ''
                  }`}
                  key={`inherited-file-${source.id}`}
                >
                  <span
                    className={
                      source.enabled && !disableClusterRules
                        ? styles.enabledDot
                        : styles.disabledDot
                    }
                  />
                  <div className={styles.ruleOrigin}>
                    <span className={styles.inheritedBadge}>
                      {disableClusterRules ? '集群继承 · 已停用' : '集群继承'}
                    </span>
                    <small>文件采集</small>
                  </div>
                  <div className={styles.inheritedRuleContent}>
                    <Typography.Text strong>文件采集 {index + 1}</Typography.Text>
                    <Typography.Text type="secondary">
                      {source.path}/{source.filePattern} · 递归深度 {source.maxDirSearchDepth}
                    </Typography.Text>
                  </div>
                </div>
              ))}

            {mode === 'host' &&
              inheritedContainerStdout.map((source, index) => (
                <div
                  className={`${styles.inheritedSourceRow} ${
                    disableClusterRules ? styles.inheritedSourceRowDisabled : ''
                  }`}
                  key={`inherited-${source.id}`}
                >
                  <span
                    className={
                      source.enabled && !disableClusterRules
                        ? styles.enabledDot
                        : styles.disabledDot
                    }
                  />
                  <div className={styles.ruleOrigin}>
                    <span className={styles.inheritedBadge}>
                      {disableClusterRules ? '集群继承 · 已停用' : '集群继承'}
                    </span>
                    <small>容器标准输出</small>
                  </div>
                  <div className={styles.inheritedRuleContent}>
                    <Typography.Text strong>容器标准输出采集 {index + 1}</Typography.Text>
                    {source.activeFilters.length ? (
                      <div className={styles.inheritedFilters}>
                        {source.activeFilters.map((field) => (
                          <span key={field}>
                            <small>
                              {
                                containerFilterDefinitions.find(
                                  (definition) => definition.field === field,
                                )?.label
                              }
                            </small>
                            <code>{formatFilterValue(source[field])}</code>
                          </span>
                        ))}
                      </div>
                    ) : (
                      <Typography.Text type="secondary">未使用过滤器</Typography.Text>
                    )}
                  </div>
                  <Typography.Text type="secondary">
                    {[source.collectStdout && 'stdout', source.collectStderr && 'stderr']
                      .filter(Boolean)
                      .join(' + ')}
                  </Typography.Text>
                </div>
              ))}

            {hostFiles.map((source, index) => (
              <div className={styles.hostSourceRow} key={source.id}>
                <Switch
                  size="small"
                  aria-label={`启用宿主机文件采集 ${index + 1}`}
                  checked={source.enabled}
                  onChange={(enabled) => updateHostFile(index, { enabled })}
                />
                <div className={styles.ruleOrigin}>
                  <span className={styles.nodeBadge}>
                    {mode === 'host' ? '当前节点' : '集群配置'}
                  </span>
                  <small>文件采集</small>
                </div>
                <label className={styles.field}>
                  <span>日志目录</span>
                  <Input
                    aria-label={`宿主机文件采集 ${index + 1} 路径`}
                    value={source.path}
                    placeholder="/var/log/application"
                    onChange={(event) => updateHostFile(index, { path: event.target.value })}
                  />
                </label>
                <label className={styles.field}>
                  <span>递归深度</span>
                  <InputNumber
                    aria-label={`宿主机文件采集 ${index + 1} 深度`}
                    min={0}
                    max={1000}
                    value={source.maxDirSearchDepth}
                    onChange={(value) =>
                      updateHostFile(index, { maxDirSearchDepth: Number(value || 0) })
                    }
                  />
                </label>
                <label className={styles.field}>
                  <span>文件名匹配</span>
                  <Input
                    aria-label={`宿主机文件采集 ${index + 1} 文件匹配`}
                    value={source.filePattern}
                    placeholder="*.log"
                    onChange={(event) => updateHostFile(index, { filePattern: event.target.value })}
                  />
                </label>
                <Tooltip title="删除">
                  <Button
                    aria-label={`删除宿主机文件采集 ${index + 1}`}
                    icon={<Trash2 size={15} />}
                    onClick={() =>
                      setHostFiles((current) =>
                        current.filter((_, itemIndex) => itemIndex !== index),
                      )
                    }
                  />
                </Tooltip>
                <RulePipelineEditors
                  source={source}
                  defaults={pipelineDefaults.file}
                  onChange={(patch) => updateHostFile(index, patch)}
                />
              </div>
            ))}

            {containerStdout.map((source, index) => (
              <div className={styles.containerSourceRow} key={source.id}>
                <Switch
                  size="small"
                  aria-label={`启用容器标准输出采集 ${index + 1}`}
                  checked={source.enabled}
                  onChange={(enabled) => updateContainerStdout(index, { enabled })}
                />
                <div className={styles.sourceIdentity}>
                  <Typography.Text strong>容器标准输出采集 {index + 1}</Typography.Text>
                  <Typography.Text type="secondary">
                    {source.activeFilters.length
                      ? `${source.activeFilters.length} 个过滤器`
                      : '未使用过滤器'}
                  </Typography.Text>
                </div>
                <div className={styles.filterEditor}>
                  <div className={styles.filterHeader}>
                    <Typography.Text strong>过滤器</Typography.Text>
                    <Select
                      value={undefined}
                      placeholder="添加过滤器"
                      disabled={source.activeFilters.length >= availableFilterDefinitions.length}
                      onChange={(field: ContainerFilterField) => {
                        const definition = containerFilterDefinitions.find(
                          (item) => item.field === field,
                        )
                        updateContainerStdout(index, {
                          ...(definition?.map ? { [field]: [newFilterMapEntry()] } : {}),
                          activeFilters: [...source.activeFilters, field],
                        })
                      }}
                      options={availableFilterDefinitions
                        .filter(({ field }) => !source.activeFilters.includes(field))
                        .map(({ field, label }) => ({ value: field, label }))}
                    />
                  </div>
                  {source.activeFilters.length ? (
                    <div className={styles.filterList}>
                      {source.activeFilters.map((field) => {
                        const definition = containerFilterDefinitions.find(
                          (item) => item.field === field,
                        )
                        if (!definition || (!showK8sFilters && definition.k8sOnly)) return null
                        return (
                          <div className={styles.filterRow} key={field}>
                            <Typography.Text>{definition.label}</Typography.Text>
                            {definition.tags ? (
                              <Select
                                mode="tags"
                                value={source[field] as string[]}
                                placeholder={definition.placeholder}
                                tokenSeparators={[',', ' ']}
                                onChange={(value) => updateContainerFilter(index, field, value)}
                              />
                            ) : definition.map ? (
                              <div className={styles.mapFilterEditor}>
                                {(source[field] as FilterMapEntry[]).map((entry) => (
                                  <div className={styles.mapFilterRow} key={entry.id}>
                                    <Input
                                      aria-label={`${definition.label}键`}
                                      value={entry.key}
                                      placeholder="标签键，例如 app.kubernetes.io/instance"
                                      onChange={(event) =>
                                        updateFilterMapEntry(index, entry.id, {
                                          key: event.target.value,
                                        })
                                      }
                                    />
                                    <Input
                                      aria-label={`${definition.label}值`}
                                      value={entry.value}
                                      placeholder="标签值或正则；留空表示仅匹配标签键"
                                      onChange={(event) =>
                                        updateFilterMapEntry(index, entry.id, {
                                          value: event.target.value,
                                        })
                                      }
                                    />
                                    <Tooltip title="删除标签条件">
                                      <Button
                                        aria-label="删除标签条件"
                                        icon={<Trash2 size={14} />}
                                        onClick={() =>
                                          updateContainerFilter(
                                            index,
                                            field,
                                            source.excludeK8sLabel.filter(
                                              (item) => item.id !== entry.id,
                                            ),
                                          )
                                        }
                                      />
                                    </Tooltip>
                                  </div>
                                ))}
                                <Button
                                  type="text"
                                  icon={<PlusOutlined />}
                                  onClick={() =>
                                    updateContainerFilter(index, field, [
                                      ...source.excludeK8sLabel,
                                      newFilterMapEntry(),
                                    ])
                                  }
                                >
                                  添加标签条件
                                </Button>
                              </div>
                            ) : definition.multiline ? (
                              <Input.TextArea
                                autoSize={{ minRows: 1, maxRows: 3 }}
                                value={source[field] as string}
                                placeholder={definition.placeholder}
                                onChange={(event) =>
                                  updateContainerFilter(index, field, event.target.value)
                                }
                              />
                            ) : (
                              <Input
                                value={source[field] as string}
                                placeholder={definition.placeholder}
                                onChange={(event) =>
                                  updateContainerFilter(index, field, event.target.value)
                                }
                              />
                            )}
                            <Tooltip title="移除过滤器">
                              <Button
                                aria-label={`移除${definition.label}`}
                                icon={<X size={14} />}
                                onClick={() => {
                                  updateContainerStdout(index, {
                                    [field]: definition.tags || definition.map ? [] : '',
                                    activeFilters: source.activeFilters.filter(
                                      (activeField) => activeField !== field,
                                    ),
                                  } as Partial<ContainerStdoutSource>)
                                }}
                              />
                            </Tooltip>
                          </div>
                        )
                      })}
                    </div>
                  ) : (
                    <Typography.Text type="secondary">
                      未使用过滤器，将采集当前集群中的全部容器
                    </Typography.Text>
                  )}
                </div>
                <Collapse
                  className={styles.advancedCollapse}
                  ghost
                  size="small"
                  items={[
                    {
                      key: 'advanced',
                      label: '高级配置',
                      children: (
                        <div className={styles.advancedContent}>
                          <div className={styles.advancedSection}>
                            <Typography.Text strong>采集行为</Typography.Text>
                            <div className={styles.advancedGrid}>
                              <div className={styles.field}>
                                <span>采集输出流</span>
                                <Space size={12}>
                                  <Checkbox
                                    checked={source.collectStdout}
                                    onChange={(event) =>
                                      updateContainerStdout(index, {
                                        collectStdout: event.target.checked,
                                      })
                                    }
                                  >
                                    stdout
                                  </Checkbox>
                                  <Checkbox
                                    checked={source.collectStderr}
                                    onChange={(event) =>
                                      updateContainerStdout(index, {
                                        collectStderr: event.target.checked,
                                      })
                                    }
                                  >
                                    stderr
                                  </Checkbox>
                                </Space>
                              </div>
                              <label className={styles.field}>
                                <span>最大事件数（条/秒）</span>
                                <InputNumber
                                  min={0}
                                  max={1000000000}
                                  precision={0}
                                  placeholder="0 表示不限制"
                                  value={source.maxEventsPerSecond}
                                  onChange={(value) =>
                                    updateContainerStdout(index, {
                                      maxEventsPerSecond: Number(value || 0),
                                    })
                                  }
                                />
                              </label>
                              <label className={styles.field}>
                                <span>首次采集末尾大小（KB）</span>
                                <InputNumber
                                  min={0}
                                  max={10485760}
                                  value={source.tailSizeKB}
                                  onChange={(value) =>
                                    updateContainerStdout(index, {
                                      tailSizeKB: Number(value ?? 1024),
                                    })
                                  }
                                />
                              </label>
                              <label className={styles.field}>
                                <span>完整行等待时间（秒）</span>
                                <InputNumber
                                  min={1}
                                  value={source.flushTimeoutSecs}
                                  onChange={(value) =>
                                    updateContainerStdout(index, {
                                      flushTimeoutSecs: Number(value ?? 5),
                                    })
                                  }
                                />
                              </label>
                              <label className={styles.toggleField}>
                                <span>解析失败保留原文</span>
                                <Switch
                                  size="small"
                                  checked={source.keepingSourceWhenParseFail}
                                  onChange={(value) =>
                                    updateContainerStdout(index, {
                                      keepingSourceWhenParseFail: value,
                                    })
                                  }
                                />
                              </label>
                              <label className={styles.toggleField}>
                                <span>忽略解析告警</span>
                                <Switch
                                  size="small"
                                  checked={source.ignoreParseWarning}
                                  onChange={(value) =>
                                    updateContainerStdout(index, { ignoreParseWarning: value })
                                  }
                                />
                              </label>
                              <label className={styles.toggleField}>
                                <span>允许被多配置采集</span>
                                <Switch
                                  size="small"
                                  checked={source.allowingIncludedByMultiConfigs}
                                  onChange={(value) =>
                                    updateContainerStdout(index, {
                                      allowingIncludedByMultiConfigs: value,
                                    })
                                  }
                                />
                              </label>
                            </div>
                          </div>

                          <div className={styles.advancedSection}>
                            <Typography.Text strong>多行聚合</Typography.Text>
                            <div className={styles.advancedGrid}>
                              <label className={styles.field}>
                                <span>聚合模式</span>
                                <Select
                                  value={source.multilineMode}
                                  onChange={(value) =>
                                    updateContainerStdout(index, { multilineMode: value })
                                  }
                                  options={[
                                    { value: 'off', label: '不开启' },
                                    { value: 'custom', label: '自定义正则' },
                                    { value: 'JSON', label: 'JSON' },
                                  ]}
                                />
                              </label>
                              {source.multilineMode === 'custom' && (
                                <>
                                  <label className={styles.field}>
                                    <span>行首正则</span>
                                    <Input
                                      value={source.multilineStartPattern}
                                      placeholder="至少填写一个匹配条件"
                                      onChange={(event) =>
                                        updateContainerStdout(index, {
                                          multilineStartPattern: event.target.value,
                                        })
                                      }
                                    />
                                  </label>
                                  <label className={styles.field}>
                                    <span>行继续正则</span>
                                    <Input
                                      value={source.multilineContinuePattern}
                                      onChange={(event) =>
                                        updateContainerStdout(index, {
                                          multilineContinuePattern: event.target.value,
                                        })
                                      }
                                    />
                                  </label>
                                  <label className={styles.field}>
                                    <span>行尾正则</span>
                                    <Input
                                      value={source.multilineEndPattern}
                                      onChange={(event) =>
                                        updateContainerStdout(index, {
                                          multilineEndPattern: event.target.value,
                                        })
                                      }
                                    />
                                  </label>
                                  <label className={styles.field}>
                                    <span>未匹配内容</span>
                                    <Select
                                      value={source.unmatchedContentTreatment}
                                      onChange={(value) =>
                                        updateContainerStdout(index, {
                                          unmatchedContentTreatment: value,
                                        })
                                      }
                                      options={[
                                        { value: 'single_line', label: '按单行采集' },
                                        { value: 'discard', label: '丢弃' },
                                      ]}
                                    />
                                  </label>
                                </>
                              )}
                            </div>
                          </div>

                          <div className={styles.advancedSection}>
                            <Typography.Text strong>Tag 映射</Typography.Text>
                            <div className={styles.advancedGrid}>
                              {tagMappingFields
                                .filter(
                                  ([field]) => showK8sFilters || field !== 'externalK8sLabelTag',
                                )
                                .map(([field, label]) => (
                                  <label className={styles.field} key={field}>
                                    <span>{label}</span>
                                    <Input.TextArea
                                      autoSize={{ minRows: 1, maxRows: 3 }}
                                      value={source[field]}
                                      placeholder="每行填写 key=value"
                                      onChange={(event) =>
                                        updateContainerStdout(index, {
                                          [field]: event.target.value,
                                        })
                                      }
                                    />
                                  </label>
                                ))}
                            </div>
                          </div>
                        </div>
                      ),
                    },
                  ]}
                />
                <RulePipelineEditors
                  source={source}
                  defaults={pipelineDefaults.container}
                  onChange={(patch) => updateContainerStdout(index, patch)}
                />
                <Tooltip title="删除">
                  <Button
                    aria-label={`删除容器标准输出采集 ${index + 1}`}
                    icon={<Trash2 size={15} />}
                    onClick={() =>
                      setContainerStdout((current) =>
                        current.filter((_, itemIndex) => itemIndex !== index),
                      )
                    }
                  />
                </Tooltip>
              </div>
            ))}

            {!visibleRuleCount && !loading && (
              <Empty
                image={Empty.PRESENTED_IMAGE_SIMPLE}
                description={
                  loadFailed
                    ? '配置加载失败，请重新选择目标后重试'
                    : mode === 'host'
                      ? '当前节点暂无独立采集规则'
                      : '当前集群暂无采集规则'
                }
              />
            )}
          </div>

          <div className={styles.footer}>
            <Typography.Text type="secondary">
              {mode === 'host'
                ? disableClusterRules
                  ? `未应用集群规则 · 当前节点 ${sourceCount} 项采集规则`
                  : `继承 ${inheritedSourceCount} 项集群规则 · 当前节点追加 ${sourceCount} 项`
                : `${sourceCount} 项采集规则`}
              {changed ? '待保存' : ''}
            </Typography.Text>
            <Space>
              <Button
                icon={<Eye size={15} />}
                loading={previewLoading}
                disabled={
                  !currentScope ||
                  loadFailed ||
                  (mode === 'host' && !disableClusterRules && inheritedLoadFailed)
                }
                onClick={preview}
              >
                最终配置预览
              </Button>
              <Button
                type="primary"
                icon={<Save size={15} />}
                loading={saving}
                disabled={!currentScope || !changed || loadFailed}
                onClick={save}
              >
                {mode === 'host' ? '下发节点日志采集' : '下发配置'}
              </Button>
            </Space>
          </div>
        </section>
      </Spin>

      <Modal
        open={previewOpen}
        width={1060}
        title={`最终配置预览 · ${currentScope}`}
        footer={<Button onClick={() => setPreviewOpen(false)}>关闭</Button>}
        onCancel={() => setPreviewOpen(false)}
      >
        {previewFiles.length ? (
          <div className={styles.previewWorkspace}>
            <div className={styles.previewFiles}>
              <div className={styles.previewFilesHeader}>
                <span>配置文件</span>
                <Typography.Text type="secondary">{previewFiles.length} 个</Typography.Text>
              </div>
              {previewFiles.map((file) => {
                const hasPipelineStats =
                  file.inputCount !== undefined ||
                  file.processorCount !== undefined ||
                  file.flusherCount !== undefined
                return (
                  <button
                    type="button"
                    key={file.path}
                    className={file.path === selectedPreviewPath ? styles.previewFileActive : ''}
                    onClick={() => setSelectedPreviewPath(file.path)}
                  >
                    <span className={styles.previewFileTitle}>
                      <FileText size={15} />
                      {file.path.split('/').pop()}
                    </span>
                    <span className={styles.previewFilePath}>
                      {file.path.includes('/')
                        ? file.path.slice(0, file.path.lastIndexOf('/'))
                        : '配置根目录'}
                    </span>
                    {hasPipelineStats && (
                      <span className={styles.previewFileStats}>
                        <span>
                          <FileText size={12} /> 输入 {file.inputCount ?? 0}
                        </span>
                        <span>
                          <GitBranch size={12} /> 处理器 {file.processorCount ?? 0}
                        </span>
                        <span>
                          <Send size={12} /> 输出 {file.flusherCount ?? 0}
                        </span>
                      </span>
                    )}
                  </button>
                )
              })}
            </div>
            <div className={styles.previewCodePane}>
              <div className={styles.previewCodeHeader}>
                <span>{selectedPreviewFile?.path.split('/').pop()}</span>
                <Typography.Text type="secondary">YAML · 最终下发内容</Typography.Text>
              </div>
              <SyntaxHighlighter
                language="yaml"
                style={theme === 'dark' ? oneDark : oneLight}
                showLineNumbers
                wrapLongLines
                customStyle={{
                  height: '100%',
                  margin: 0,
                  overflow: 'auto',
                  padding: '16px 18px',
                  background: 'transparent',
                  fontSize: 12,
                  lineHeight: 1.65,
                }}
                lineNumberStyle={{ minWidth: '2.5em', color: 'var(--ant-color-text-quaternary)' }}
              >
                {selectedPreviewFile?.content || ''}
              </SyntaxHighlighter>
            </div>
          </div>
        ) : (
          <Empty
            image={Empty.PRESENTED_IMAGE_SIMPLE}
            description="没有已开启的采集配置，不会生成配置文件"
          />
        )}
      </Modal>
    </div>
  )
}
