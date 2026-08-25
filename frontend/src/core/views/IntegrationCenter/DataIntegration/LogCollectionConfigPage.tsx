import {
  Button,
  Empty,
  Input,
  InputNumber,
  Collapse,
  Checkbox,
  Modal,
  Select,
  Space,
  Spin,
  Switch,
  Tooltip,
  Typography,
} from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import { ExternalLink, Eye, Save, Trash2 } from 'lucide-react'
import { useCallback, useEffect, useState } from 'react'
import {
  getContainerStdoutCollectionApi,
  getHostFileCollectionApi,
  previewContainerStdoutCollectionApi,
  previewHostFileCollectionApi,
  updateContainerStdoutCollectionApi,
  updateHostFileCollectionApi,
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
}

type HostFileSource = {
  id: string
  name: string
  enabled: boolean
  path: string
  maxDirSearchDepth: number
  filePattern: string
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
  excludeK8sLabel: string
  includeEnv: string
  excludeEnv: string
  includeContainerLabel: string
  excludeContainerLabel: string
  externalK8sLabelTag: string
  externalEnvTag: string
  tags: string
  activeFilters: ContainerFilterField[]
}

type PreviewFile = { path: string; content: string }

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
    placeholder: '每行填写 key=value',
    k8sOnly: true,
    multiline: true,
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

const newHostFile = (): HostFileSource => {
  const id = createSourceId()
  return {
    id,
    name: `host-file-${id}`,
    enabled: true,
    path: '/var/log',
    maxDirSearchDepth: 0,
    filePattern: '*.log',
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
    excludeK8sLabel: '',
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
    }
    return {
      ...normalized,
      activeFilters:
        source.activeFilters ||
        containerFilterDefinitions
          .map(({ field }) => field)
          .filter((field) => normalized[field].trim()),
    }
  })

const containerPayload = (sources: ContainerStdoutSource[]) =>
  normalizeContainerSources(sources).map(({ activeFilters: _activeFilters, ...source }) => source)

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
      excludeK8sLabel: '',
      activeFilters: source.activeFilters.filter(
        (field) =>
          !containerFilterDefinitions.find((definition) => definition.field === field)?.k8sOnly,
      ),
    }
  })

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
}: LogCollectionConfigPanelProps) {
  const selectedNode = nodeId
  const selectedCluster = clusterIds[0] || ''
  const [hostFiles, setHostFiles] = useState<HostFileSource[]>([])
  const [containerStdout, setContainerStdout] = useState<ContainerStdoutSource[]>([])
  const [inheritedContainerStdout, setInheritedContainerStdout] = useState<ContainerStdoutSource[]>(
    [],
  )
  const [hostVersion, setHostVersion] = useState(0)
  const [containerVersion, setContainerVersion] = useState(0)
  const [inheritedVersion, setInheritedVersion] = useState(0)
  const [originalHostFiles, setOriginalHostFiles] = useState('[]')
  const [originalContainerStdout, setOriginalContainerStdout] = useState('[]')
  const [loading, setLoading] = useState(true)
  const [loadFailed, setLoadFailed] = useState(false)
  const [inheritedLoading, setInheritedLoading] = useState(false)
  const [inheritedLoadFailed, setInheritedLoadFailed] = useState(false)
  const [saving, setSaving] = useState(false)
  const [previewLoading, setPreviewLoading] = useState(false)
  const [previewOpen, setPreviewOpen] = useState(false)
  const [previewFiles, setPreviewFiles] = useState<PreviewFile[]>([])
  const [selectedPreviewPath, setSelectedPreviewPath] = useState('')

  const selectedPreviewFile = previewFiles.find((file) => file.path === selectedPreviewPath)
  const hostChanged = JSON.stringify(hostFiles) !== originalHostFiles
  const containerChanged = JSON.stringify(containerStdout) !== originalContainerStdout
  const changed = mode === 'host' ? hostChanged : containerChanged
  const currentScope = mode === 'host' ? selectedNode : selectedCluster
  const currentVersion = mode === 'host' ? hostVersion : containerVersion
  const sourceCount = mode === 'host' ? hostFiles.length : containerStdout.length
  const inheritedSourceCount = inheritedContainerStdout.filter((source) => source.enabled).length
  const visibleRuleCount =
    mode === 'host' ? inheritedContainerStdout.length + hostFiles.length : containerStdout.length
  const availableFilterDefinitions = containerFilterDefinitions.filter(
    (definition) => showK8sFilters || !definition.k8sOnly,
  )

  const loadHostFiles = useCallback(async () => {
    if (!selectedNode) {
      setHostFiles([])
      setOriginalHostFiles('[]')
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
      setHostVersion(Number(result?.version || 0))
    } catch (error: any) {
      console.error('[LogCollectionConfigPage] failed to load host file collection', error)
      setHostFiles([])
      setOriginalHostFiles('[]')
      setLoadFailed(true)
      notify({ type: 'error', message: error.message || '加载宿主机文件采集配置失败' })
    } finally {
      setLoading(false)
    }
  }, [selectedNode])

  const loadContainerStdout = useCallback(async () => {
    if (!selectedCluster) {
      setContainerStdout([])
      setOriginalContainerStdout('[]')
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
      setOriginalContainerStdout(JSON.stringify(values))
      setContainerVersion(Number(result?.version || 0))
    } catch (error: any) {
      console.error('[LogCollectionConfigPage] failed to load container stdout collection', error)
      setContainerStdout([])
      setOriginalContainerStdout('[]')
      setLoadFailed(true)
      notify({ type: 'error', message: error.message || '加载容器标准输出采集配置失败' })
    } finally {
      setLoading(false)
    }
  }, [selectedCluster, showK8sFilters])

  const loadInheritedContainerStdout = useCallback(async () => {
    if (mode !== 'host' || !inheritedClusterId) {
      setInheritedContainerStdout([])
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
      setInheritedVersion(Number(result?.version || 0))
    } catch (error) {
      console.error('[LogCollectionConfigPage] failed to load inherited cluster collection', error)
      setInheritedContainerStdout([])
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

  const validateContainerFilters = () => {
    const incomplete = containerStdout.some((source) =>
      source.activeFilters.some((field) => !source[field].trim()),
    )
    if (!incomplete) return true
    notify({ type: 'error', message: '请填写已添加过滤器的匹配条件，或移除该过滤器' })
    return false
  }

  const save = async () => {
    if (!currentScope || loadFailed) return
    if (mode === 'container' && !validateContainerFilters()) return
    setSaving(true)
    try {
      if (mode === 'host') {
        const result: any = await updateHostFileCollectionApi(selectedNode, {
          hostFiles: hostPayload(hostFiles),
        })
        const values = hostPayload((result?.hostFiles || []) as HostFileSource[])
        setHostFiles(values)
        setOriginalHostFiles(JSON.stringify(values))
        setHostVersion(Number(result?.version || hostVersion + 1))
        notify({ type: 'success', message: '节点追加的宿主机文件采集规则已保存' })
      } else {
        const payload = { containerStdout: containerPayload(containerStdout) }
        const results = await Promise.all(
          clusterIds.map((clusterId) => updateContainerStdoutCollectionApi(clusterId, payload)),
        )
        const result: any = results[0]
        const values = containerSourcesForEnvironment(
          (result?.containerStdout || []) as ContainerStdoutSource[],
          showK8sFilters,
        )
        setContainerStdout(values)
        setOriginalContainerStdout(JSON.stringify(values))
        setContainerVersion(Number(result?.version || containerVersion + 1))
        notify({
          type: 'success',
          message: `容器标准输出采集配置已保存到 ${clusterIds.length} 个集群`,
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
    if (!currentScope || loadFailed || (mode === 'host' && inheritedLoadFailed)) return
    if (mode === 'container' && !validateContainerFilters()) return
    setPreviewLoading(true)
    try {
      let files: PreviewFile[]
      if (mode === 'host') {
        const [nodeResult, clusterResult]: any[] = await Promise.all([
          previewHostFileCollectionApi(selectedNode, { hostFiles: hostPayload(hostFiles) }),
          inheritedClusterId
            ? previewContainerStdoutCollectionApi(inheritedClusterId, {
                containerStdout: containerPayload(inheritedContainerStdout),
              })
            : Promise.resolve({ files: [] }),
        ])
        files = [...(clusterResult?.files || []), ...(nodeResult?.files || [])]
      } else {
        const result: any = await previewContainerStdoutCollectionApi(selectedCluster, {
          containerStdout: containerPayload(containerStdout),
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

  const addSource = () => {
    if (mode === 'host') setHostFiles((current) => [...current, newHostFile()])
    else setContainerStdout((current) => [...current, newContainerStdout()])
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
          <Typography.Text strong>
            {mode === 'host' ? '日志采集规则' : '容器标准输出采集'}
          </Typography.Text>
          <Typography.Text type="secondary">
            {mode === 'host'
              ? '查看集群继承规则，并管理当前节点追加的宿主机日志路径'
              : '按需添加过滤器，选择需要采集标准输出的容器'}
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
            <div className={styles.inheritedSection}>
              <div className={styles.inheritedHeader}>
                <div>
                  <Typography.Text strong>集群继承 · {inheritedClusterId}</Typography.Text>
                  <Typography.Text type="secondary">
                    {inheritedClusterEnvironment
                      ? `${inheritedClusterEnvironment} · 集群配置自动应用到当前节点`
                      : '集群配置自动应用到当前节点'}
                  </Typography.Text>
                </div>
                <Space>
                  {!!inheritedVersion && (
                    <Typography.Text type="secondary">版本 {inheritedVersion}</Typography.Text>
                  )}
                  <Button
                    icon={<ExternalLink size={14} />}
                    disabled={!inheritedClusterId || !onEditInherited}
                    onClick={onEditInherited}
                  >
                    修改集群日志采集
                  </Button>
                </Space>
              </div>
              <Spin spinning={inheritedLoading}>
                {inheritedLoadFailed ? (
                  <Typography.Text type="danger">
                    集群日志采集配置加载失败，无法生成完整预览
                  </Typography.Text>
                ) : (
                  <Typography.Text type="secondary">
                    {inheritedContainerStdout.length
                      ? `已继承 ${inheritedContainerStdout.length} 项容器标准输出规则，详情见下方生效规则`
                      : '集群尚未配置容器标准输出规则'}
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
                  仅追加当前节点需要采集的宿主机文件
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

            <Button
              icon={<PlusOutlined />}
              disabled={!currentScope || loadFailed}
              onClick={addSource}
            >
              {mode === 'host' ? '添加文件路径' : '添加容器选择'}
            </Button>
          </div>

          <div className={styles.rulesHeader}>
            <Typography.Text strong>{mode === 'host' ? '生效规则' : '容器规则'}</Typography.Text>
            <Typography.Text type="secondary">{visibleRuleCount} 项</Typography.Text>
          </div>

          <div className={styles.sourceList}>
            {mode === 'host' &&
              inheritedContainerStdout.map((source, index) => (
                <div className={styles.inheritedSourceRow} key={`inherited-${source.id}`}>
                  <span className={source.enabled ? styles.enabledDot : styles.disabledDot} />
                  <div className={styles.ruleOrigin}>
                    <span className={styles.inheritedBadge}>集群继承</span>
                    <small>容器标准输出</small>
                  </div>
                  <div className={styles.inheritedRuleContent}>
                    <Typography.Text strong>容器规则 {index + 1}</Typography.Text>
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
                            <code>{source[field]}</code>
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

            {mode === 'host' &&
              hostFiles.map((source, index) => (
                <div className={styles.hostSourceRow} key={source.id}>
                  <Switch
                    size="small"
                    aria-label={`启用宿主机文件采集 ${index + 1}`}
                    checked={source.enabled}
                    onChange={(enabled) => updateHostFile(index, { enabled })}
                  />
                  <div className={styles.ruleOrigin}>
                    <span className={styles.nodeBadge}>当前节点</span>
                    <small>宿主机文件</small>
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
                      onChange={(event) =>
                        updateHostFile(index, { filePattern: event.target.value })
                      }
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
                </div>
              ))}

            {mode === 'container' &&
              containerStdout.map((source, index) => (
                <div className={styles.containerSourceRow} key={source.id}>
                  <Switch
                    size="small"
                    aria-label={`启用容器标准输出采集 ${index + 1}`}
                    checked={source.enabled}
                    onChange={(enabled) => updateContainerStdout(index, { enabled })}
                  />
                  <div className={styles.sourceIdentity}>
                    <Typography.Text strong>容器规则 {index + 1}</Typography.Text>
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
                        onChange={(field: ContainerFilterField) =>
                          updateContainerStdout(index, {
                            activeFilters: [...source.activeFilters, field],
                          })
                        }
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
                              {definition.multiline ? (
                                <Input.TextArea
                                  autoSize={{ minRows: 1, maxRows: 3 }}
                                  value={source[field]}
                                  placeholder={definition.placeholder}
                                  onChange={(event) =>
                                    updateContainerStdout(index, {
                                      [field]: event.target.value,
                                    })
                                  }
                                />
                              ) : (
                                <Input
                                  value={source[field]}
                                  placeholder={definition.placeholder}
                                  onChange={(event) =>
                                    updateContainerStdout(index, {
                                      [field]: event.target.value,
                                    })
                                  }
                                />
                              )}
                              <Tooltip title="移除过滤器">
                                <Button
                                  aria-label={`移除${definition.label}`}
                                  icon={<Trash2 size={14} />}
                                  onClick={() =>
                                    updateContainerStdout(index, {
                                      [field]: '',
                                      activeFilters: source.activeFilters.filter(
                                        (activeField) => activeField !== field,
                                      ),
                                    })
                                  }
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
                      ? '当前节点暂无宿主机文件采集配置'
                      : '当前集群暂无容器标准输出采集配置'
                }
              />
            )}
          </div>

          <div className={styles.footer}>
            <Typography.Text type="secondary">
              {mode === 'host'
                ? `继承 ${inheritedSourceCount} 项集群规则 · 追加 ${sourceCount} 个文件路径`
                : `${sourceCount} 项容器规则`}
              {changed ? '待保存' : ''}
            </Typography.Text>
            <Space>
              <Button
                icon={<Eye size={15} />}
                loading={previewLoading}
                disabled={!currentScope || loadFailed || (mode === 'host' && inheritedLoadFailed)}
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
                {mode === 'host' ? '保存节点追加规则' : '保存配置'}
              </Button>
            </Space>
          </div>
        </section>
      </Spin>

      <Modal
        open={previewOpen}
        width={920}
        title={`最终配置预览 · ${currentScope}`}
        footer={<Button onClick={() => setPreviewOpen(false)}>关闭</Button>}
        onCancel={() => setPreviewOpen(false)}
      >
        {previewFiles.length ? (
          <div className={styles.previewWorkspace}>
            <div className={styles.previewFiles}>
              {previewFiles.map((file) => (
                <button
                  type="button"
                  key={file.path}
                  className={file.path === selectedPreviewPath ? styles.previewFileActive : ''}
                  onClick={() => setSelectedPreviewPath(file.path)}
                >
                  {file.path}
                </button>
              ))}
            </div>
            <pre>{selectedPreviewFile?.content || ''}</pre>
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
