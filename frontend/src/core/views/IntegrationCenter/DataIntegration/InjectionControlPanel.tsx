import {
  Alert,
  Button,
  Checkbox,
  Empty,
  Input,
  Select,
  Spin,
  Tag,
  Tooltip,
  Tree,
  Typography,
} from 'antd'
import { Braces, Info, Plus, RefreshCw, Save, Trash2 } from 'lucide-react'
import { type Key, useCallback, useEffect, useMemo, useState } from 'react'
import {
  getClusterInjectionControlApi,
  updateClusterInjectionControlApi,
} from 'src/core/api/integration'
import { notify } from 'src/core/utils/notify'
import styles from './probeConfigPage.module.scss'

type NamespaceMode = 'disabled' | 'enabled' | 'enabledFuture'
type WorkloadKind = 'deployment' | 'statefulset' | 'daemonset'
type WorkloadRef = { namespace: string; kind: WorkloadKind; name: string }
type WorkloadRule = WorkloadRef & { enabled: boolean }
type NamespaceRule = { namespace: string; mode: NamespaceMode }
type ServiceNameMapping = {
  namespacePattern: string
  workloadPattern: string
  serviceName: string
}
type ServiceNameVariable = {
  name: string
  source: 'workloadName' | 'workflowName' | 'deployName' | 'containerName'
  regex: string
  replacement: string
}
type ServiceNamePolicy = {
  defaultFormat: string
  mappings: ServiceNameMapping[]
  variables: ServiceNameVariable[]
}
type NamespaceInjection = {
  instrumentAll: boolean
  instrumentNS: string[]
  instrumentDisabledNS: string[]
  namespaces: NamespaceRule[]
  workloads: WorkloadRule[]
  serviceName: ServiceNamePolicy
}
type WorkloadInventory = { namespaces: string[]; workloads: WorkloadRef[] }
type InjectionControl = {
  namespaceInjection?: Partial<NamespaceInjection>
  inventory?: Partial<WorkloadInventory>
}

const emptyServiceNamePolicy = (): ServiceNamePolicy => ({
  defaultFormat: '',
  mappings: [],
  variables: [],
})
const namespaceKey = (namespace: string) => `namespace:${namespace}`
const workloadKey = (workload: WorkloadRef) =>
  `workload:${workload.namespace}:${workload.kind}:${workload.name}`
const workloadLabel: Record<WorkloadKind, string> = {
  deployment: 'Deployment',
  statefulset: 'StatefulSet',
  daemonset: 'DaemonSet',
}
const modeLabel: Record<NamespaceMode, string> = {
  disabled: '暂不注入',
  enabled: '仅注入当前',
  enabledFuture: '注入当前及新增',
}
const nextMode: Record<NamespaceMode, NamespaceMode> = {
  disabled: 'enabled',
  enabled: 'enabledFuture',
  enabledFuture: 'disabled',
}
const previewContext: Record<string, string> = {
  namespaceName: 'production',
  workloadName: 'checkout-api',
  workflowName: 'checkout-api',
  deployName: 'checkout-api-v42',
  containerName: 'app',
}

const previewServiceName = (policy: ServiceNamePolicy) => {
  const values = { ...previewContext }
  for (const variable of policy.variables) {
    const source = values[variable.source] || ''
    if (!variable.name) continue
    try {
      values[variable.name] = variable.regex
        ? source.replace(new RegExp(variable.regex, 'g'), variable.replacement)
        : source
    } catch {
      console.error('[InjectionControl] service.name 变量正则无效', variable.name)
      values[variable.name] = source
    }
  }
  const format = policy.defaultFormat.trim() || '${workloadName}-${containerName}'
  return format.replace(/\$\{([^}]+)\}/g, (_, key) => values[key] || `\u0024{${key}}`)
}

const invalidRegex = (value: string) => {
  if (!value) return false
  try {
    new RegExp(value)
    return false
  } catch {
    return true
  }
}

const normalizeControl = (value?: InjectionControl) => {
  const raw = value?.namespaceInjection || {}
  return {
    config: {
      instrumentAll: Boolean(raw.instrumentAll),
      instrumentNS: raw.instrumentNS || [],
      instrumentDisabledNS: raw.instrumentDisabledNS || [],
      namespaces: raw.namespaces || [],
      workloads: raw.workloads || [],
      serviceName: {
        ...emptyServiceNamePolicy(),
        ...(raw.serviceName || {}),
        mappings: raw.serviceName?.mappings || [],
        variables: raw.serviceName?.variables || [],
      },
    } satisfies NamespaceInjection,
    inventory: {
      namespaces: value?.inventory?.namespaces || [],
      workloads: value?.inventory?.workloads || [],
    } satisfies WorkloadInventory,
  }
}

const modeForNamespace = (config: NamespaceInjection, namespace: string): NamespaceMode => {
  const rule = config.namespaces.find((item) => item.namespace === namespace)
  if (rule) return rule.mode
  if (config.instrumentDisabledNS.includes(namespace)) return 'disabled'
  if (config.instrumentNS.includes(namespace)) return 'enabledFuture'
  return config.instrumentAll ? 'enabledFuture' : 'disabled'
}

const stateFromControl = (value?: InjectionControl) => {
  const normalized = normalizeControl(value)
  const modes: Record<string, NamespaceMode> = {}
  const checked = new Set<string>()
  const workloadRules = new Map(
    normalized.config.workloads.map((rule) => [workloadKey(rule), rule.enabled]),
  )
  for (const namespace of normalized.inventory.namespaces) {
    const mode = modeForNamespace(normalized.config, namespace)
    modes[namespace] = mode
    for (const workload of normalized.inventory.workloads.filter(
      (item) => item.namespace === namespace,
    )) {
      const key = workloadKey(workload)
      const selected = workloadRules.has(key) ? workloadRules.get(key) : mode !== 'disabled'
      if (selected) checked.add(key)
    }
  }
  return { ...normalized, modes, checked: [...checked] }
}

const configFromState = (
  applyAll: boolean,
  modes: Record<string, NamespaceMode>,
  checkedWorkloads: string[],
  inventory: WorkloadInventory,
  serviceName: ServiceNamePolicy,
): NamespaceInjection => {
  const defaultMode: NamespaceMode = applyAll ? 'enabledFuture' : 'disabled'
  const namespaces = inventory.namespaces
    .filter((namespace) => (modes[namespace] || defaultMode) !== defaultMode)
    .map((namespace) => ({ namespace, mode: modes[namespace] || defaultMode }))
  const checked = new Set(checkedWorkloads)
  const workloads: WorkloadRule[] = []
  for (const workload of inventory.workloads) {
    const mode = modes[workload.namespace] || defaultMode
    const selected = checked.has(workloadKey(workload))
    if (selected !== (mode !== 'disabled')) workloads.push({ ...workload, enabled: selected })
  }
  return {
    instrumentAll: applyAll,
    instrumentNS: [],
    instrumentDisabledNS: [],
    namespaces,
    workloads,
    serviceName,
  }
}

export function InjectionControlPanel({ clusterId }: { clusterId: string }) {
  const [inventory, setInventory] = useState<WorkloadInventory>({ namespaces: [], workloads: [] })
  const [applyAll, setApplyAll] = useState(false)
  const [namespaceModes, setNamespaceModes] = useState<Record<string, NamespaceMode>>({})
  const [checkedWorkloads, setCheckedWorkloads] = useState<string[]>([])
  const [serviceName, setServiceName] = useState<ServiceNamePolicy>(emptyServiceNamePolicy)
  const [savedSnapshot, setSavedSnapshot] = useState('')
  const [search, setSearch] = useState('')
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)

  const currentConfig = useMemo(
    () => configFromState(applyAll, namespaceModes, checkedWorkloads, inventory, serviceName),
    [applyAll, checkedWorkloads, inventory, namespaceModes, serviceName],
  )
  const snapshot = useMemo(() => JSON.stringify(currentConfig), [currentConfig])
  const changed = Boolean(savedSnapshot) && snapshot !== savedSnapshot
  const invalidServiceName =
    serviceName.mappings.some(
      (item) =>
        !item.namespacePattern ||
        !item.workloadPattern ||
        !item.serviceName ||
        invalidRegex(item.namespacePattern) ||
        invalidRegex(item.workloadPattern),
    ) ||
    serviceName.variables.some((item) => !item.name || !item.source || invalidRegex(item.regex))
  const generatedPreview = useMemo(() => previewServiceName(serviceName), [serviceName])

  const applyControl = useCallback((value: InjectionControl) => {
    const state = stateFromControl(value)
    setInventory(state.inventory)
    setApplyAll(state.config.instrumentAll)
    setNamespaceModes(state.modes)
    setCheckedWorkloads(state.checked)
    setServiceName(state.config.serviceName)
    setSavedSnapshot(
      JSON.stringify(
        configFromState(
          state.config.instrumentAll,
          state.modes,
          state.checked,
          state.inventory,
          state.config.serviceName,
        ),
      ),
    )
  }, [])

  const load = useCallback(async () => {
    setLoading(true)
    try {
      applyControl((await getClusterInjectionControlApi(clusterId)) as InjectionControl)
    } catch (error: any) {
      notify({ type: 'error', message: error.message || '注入配置加载失败' })
    } finally {
      setLoading(false)
    }
  }, [applyControl, clusterId])

  useEffect(() => {
    load()
  }, [load])

  const save = async () => {
    setSaving(true)
    try {
      const result = (await updateClusterInjectionControlApi(clusterId, {
        namespaceInjection: currentConfig,
      })) as InjectionControl
      applyControl(result)
      notify({ type: 'success', message: '注入配置已保存' })
    } catch (error: any) {
      notify({ type: 'error', message: error.message || '注入配置保存失败' })
    } finally {
      setSaving(false)
    }
  }

  const workloadsByNamespace = useMemo(() => {
    const result = new Map<string, WorkloadRef[]>()
    for (const namespace of inventory.namespaces) result.set(namespace, [])
    for (const workload of inventory.workloads) {
      const items = result.get(workload.namespace) || []
      items.push(workload)
      result.set(workload.namespace, items)
    }
    return result
  }, [inventory])

  const treeData = useMemo(
    () =>
      [...workloadsByNamespace.entries()].map(([namespace, workloads]) => ({
        key: namespaceKey(namespace),
        title: (
          <span className={styles.workloadTreeLabel}>
            <span className={styles.treeNodeIdentity}>
              <span className={styles.treeNodeName}>{namespace}</span>
              <span className={styles.treeNodeType}>Namespace</span>
            </span>
            <Typography.Text type="secondary">
              {modeLabel[namespaceModes[namespace] || (applyAll ? 'enabledFuture' : 'disabled')]}
            </Typography.Text>
          </span>
        ),
        children: workloads.map((workload) => ({
          key: workloadKey(workload),
          title: (
            <span className={styles.workloadTreeLabel}>
              <span className={styles.treeNodeIdentity}>
                <span className={styles.treeNodeName}>{workload.name}</span>
                <span className={styles.treeNodeType}>工作负载</span>
              </span>
              <Typography.Text type="secondary">{workloadLabel[workload.kind]}</Typography.Text>
            </span>
          ),
        })),
      })),
    [applyAll, namespaceModes, workloadsByNamespace],
  )

  const checkedTreeKeys = useMemo(() => {
    const checked = [...checkedWorkloads]
    const halfChecked: string[] = []
    for (const namespace of workloadsByNamespace.keys()) {
      const mode = namespaceModes[namespace] || (applyAll ? 'enabledFuture' : 'disabled')
      if (mode === 'enabledFuture') checked.push(namespaceKey(namespace))
      if (mode === 'enabled') {
        halfChecked.push(namespaceKey(namespace))
      }
    }
    return { checked, halfChecked }
  }, [applyAll, checkedWorkloads, namespaceModes, workloadsByNamespace])

  const handleTreeCheck = (_keys: unknown, info: { node: { key: Key } }) => {
    const key = String(info.node.key)
    if (key.startsWith('namespace:')) {
      const namespace = key.slice('namespace:'.length)
      const mode = namespaceModes[namespace] || (applyAll ? 'enabledFuture' : 'disabled')
      const next = nextMode[mode]
      const childKeys = (workloadsByNamespace.get(namespace) || []).map(workloadKey)
      setNamespaceModes((current) => ({ ...current, [namespace]: next }))
      setCheckedWorkloads((current) => {
        const result = new Set(current)
        for (const childKey of childKeys) {
          if (next === 'disabled') result.delete(childKey)
          else result.add(childKey)
        }
        return [...result]
      })
      return
    }
    setCheckedWorkloads((current) => {
      const result = new Set(current)
      if (result.has(key)) result.delete(key)
      else result.add(key)
      return [...result]
    })
  }

  const changeApplyAll = (next: boolean) => {
    const previousDefault: NamespaceMode = applyAll ? 'enabledFuture' : 'disabled'
    const nextDefault: NamespaceMode = next ? 'enabledFuture' : 'disabled'
    const changedNamespaces = inventory.namespaces.filter(
      (namespace) => (namespaceModes[namespace] || previousDefault) === previousDefault,
    )
    setApplyAll(next)
    setNamespaceModes((current) => {
      const result = { ...current }
      for (const namespace of changedNamespaces) result[namespace] = nextDefault
      return result
    })
    setCheckedWorkloads((current) => {
      const result = new Set(current)
      for (const workload of inventory.workloads) {
        if (!changedNamespaces.includes(workload.namespace)) continue
        if (next) result.add(workloadKey(workload))
        else result.delete(workloadKey(workload))
      }
      return [...result]
    })
  }

  const patchVariable = (index: number, patch: Partial<ServiceNameVariable>) =>
    setServiceName((current) => ({
      ...current,
      variables: current.variables.map((variable, itemIndex) =>
        itemIndex === index ? { ...variable, ...patch } : variable,
      ),
    }))
  const patchMapping = (index: number, patch: Partial<ServiceNameMapping>) =>
    setServiceName((current) => ({
      ...current,
      mappings: current.mappings.map((mapping, itemIndex) =>
        itemIndex === index ? { ...mapping, ...patch } : mapping,
      ),
    }))

  return (
    <Spin spinning={loading}>
      <div className={styles.injectionControlPanel}>
        <section className={styles.injectionSection}>
          <div className={styles.injectionSectionHeading}>
            <div>
              <Typography.Text strong>注入对象</Typography.Text>
              <Checkbox
                checked={applyAll}
                onChange={(event) => changeApplyAll(event.target.checked)}
              >
                应用于所有 Namespace
              </Checkbox>
            </div>
            <Button aria-label="刷新工作负载" icon={<RefreshCw size={15} />} onClick={load} />
          </div>
          <Input.Search
            allowClear
            value={search}
            placeholder="搜索 Namespace 或工作负载"
            onChange={(event) => setSearch(event.target.value)}
          />
          {treeData.length ? (
            <div className={styles.workloadTree}>
              <Tree
                checkable
                checkStrictly
                blockNode
                defaultExpandAll
                height={420}
                checkedKeys={checkedTreeKeys}
                treeData={treeData}
                filterTreeNode={(node) =>
                  Boolean(search) && String(node.key).toLowerCase().includes(search.toLowerCase())
                }
                onCheck={handleTreeCheck}
              />
            </div>
          ) : (
            <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="尚未收到工作负载清单" />
          )}
        </section>

        <section className={`${styles.injectionSection} ${styles.serviceNameSection}`}>
          <div className={styles.serviceSectionHeader}>
            <div className={styles.serviceSectionTitle}>
              <span className={styles.serviceSectionIcon}>
                <Braces size={17} />
              </span>
              <div>
                <Typography.Text strong>service.name 命名</Typography.Text>
                <Typography.Text type="secondary">
                  按工作负载匹配映射，未匹配时使用默认规则
                </Typography.Text>
              </div>
            </div>
            <Tag color={serviceName.mappings.length ? 'blue' : 'default'}>
              {serviceName.mappings.length} 条映射
            </Tag>
          </div>
          <Alert
            className={styles.serviceNameAlert}
            type="info"
            showIcon
            icon={<Info size={15} />}
            message="映射优先级高于默认规则"
            description="先匹配 Namespace 和工作负载正则；同一工作负载命中多条时，按列表从上到下取第一条。"
          />
          <div className={styles.serviceVariableHeader}>
            <div className={styles.subsectionTitle}>
              <Typography.Text strong>名称映射</Typography.Text>
              <Typography.Text type="secondary">将指定工作负载固定到业务服务名</Typography.Text>
            </div>
            <Button
              type="dashed"
              icon={<Plus size={15} />}
              onClick={() =>
                setServiceName((current) => ({
                  ...current,
                  mappings: [
                    ...current.mappings,
                    { namespacePattern: '', workloadPattern: '', serviceName: '' },
                  ],
                }))
              }
            >
              添加映射
            </Button>
          </div>
          {serviceName.mappings.length ? (
            <div className={styles.serviceRuleList}>
              <div className={`${styles.serviceRuleColumns} ${styles.serviceRuleColumnsMapping}`}>
                <span />
                <span>Namespace 正则</span>
                <span>工作负载正则</span>
                <span>目标 service.name</span>
                <span />
              </div>
              <div className={styles.serviceVariableList}>
                {serviceName.mappings.map((mapping, index) => (
                  <div className={styles.serviceMappingRow} key={index}>
                    <span className={styles.ruleIndex}>{index + 1}</span>
                    <Input
                      aria-label="namespace 正则"
                      value={mapping.namespacePattern}
                      placeholder="例如：^prod$"
                      status={
                        !mapping.namespacePattern || invalidRegex(mapping.namespacePattern)
                          ? 'error'
                          : undefined
                      }
                      onChange={(event) =>
                        patchMapping(index, { namespacePattern: event.target.value })
                      }
                    />
                    <Input
                      aria-label="workload 正则"
                      value={mapping.workloadPattern}
                      placeholder="例如：^checkout-.*"
                      status={
                        !mapping.workloadPattern || invalidRegex(mapping.workloadPattern)
                          ? 'error'
                          : undefined
                      }
                      onChange={(event) =>
                        patchMapping(index, { workloadPattern: event.target.value })
                      }
                    />
                    <Input
                      aria-label="service.name"
                      value={mapping.serviceName}
                      placeholder="例如：checkout"
                      status={!mapping.serviceName ? 'error' : undefined}
                      onChange={(event) => patchMapping(index, { serviceName: event.target.value })}
                    />
                    <Tooltip title="删除映射">
                      <Button
                        danger
                        type="text"
                        aria-label="删除映射"
                        icon={<Trash2 size={15} />}
                        onClick={() =>
                          setServiceName((current) => ({
                            ...current,
                            mappings: current.mappings.filter(
                              (_, itemIndex) => itemIndex !== index,
                            ),
                          }))
                        }
                      />
                    </Tooltip>
                  </div>
                ))}
              </div>
            </div>
          ) : (
            <div className={styles.serviceEmptyState}>
              <Braces size={20} />
              <span>暂无名称映射</span>
              <Typography.Text type="secondary">
                所有未命中的工作负载将按默认规则生成名称
              </Typography.Text>
            </div>
          )}
        </section>

        <section className={`${styles.injectionSection} ${styles.serviceNameSection}`}>
          <div className={styles.serviceVariableHeader}>
            <div className={styles.subsectionTitle}>
              <Typography.Text strong>默认生成规则</Typography.Text>
              <Typography.Text type="secondary">
                为未命中映射的工作负载生成稳定的 service.name
              </Typography.Text>
            </div>
            <div className={styles.previewBadge}>
              <span />
              实时预览
            </div>
          </div>
          <div className={styles.serviceNameRuleEditor}>
            <label>
              <span>名称格式</span>
              <Input
                value={serviceName.defaultFormat}
                placeholder="${workloadName}-${containerName}"
                onChange={(event) =>
                  setServiceName((current) => ({ ...current, defaultFormat: event.target.value }))
                }
              />
              <Typography.Text type="secondary">
                使用 <code>${'{workloadName}'}</code>、<code>${'{containerName}'}</code>{' '}
                或派生变量组合
              </Typography.Text>
            </label>
            <div className={styles.servicePreview}>
              <span>示例结果</span>
              <code>{generatedPreview}</code>
              <Typography.Text type="secondary">production / checkout-api / app</Typography.Text>
            </div>
          </div>
          <div className={styles.serviceVariableHeader}>
            <div className={styles.subsectionTitle}>
              <Typography.Text strong>派生变量</Typography.Text>
              <Typography.Text type="secondary">先处理原始名称，再在格式中引用变量</Typography.Text>
            </div>
            <Button
              type="dashed"
              icon={<Plus size={15} />}
              onClick={() =>
                setServiceName((current) => ({
                  ...current,
                  variables: [
                    ...current.variables,
                    { name: '', source: 'workloadName', regex: '', replacement: '' },
                  ],
                }))
              }
            >
              添加变量
            </Button>
          </div>
          {serviceName.variables.length ? (
            <div className={styles.serviceRuleList}>
              <div className={styles.serviceRuleColumns}>
                <span>变量名</span>
                <span>来源</span>
                <span>匹配正则</span>
                <span>替换为</span>
                <span />
              </div>
              <div className={styles.serviceVariableList}>
                {serviceName.variables.map((variable, index) => (
                  <div className={styles.serviceVariableRow} key={`${index}-${variable.name}`}>
                    <Input
                      aria-label="变量名"
                      value={variable.name}
                      placeholder="例如：appName"
                      status={!variable.name ? 'error' : undefined}
                      onChange={(event) => patchVariable(index, { name: event.target.value })}
                    />
                    <Select
                      aria-label="变量来源"
                      value={variable.source}
                      options={[
                        { value: 'workloadName', label: 'workloadName' },
                        { value: 'workflowName', label: 'workflowName' },
                        { value: 'deployName', label: 'deployName' },
                        { value: 'containerName', label: 'containerName' },
                      ]}
                      onChange={(source) => patchVariable(index, { source })}
                    />
                    <Input
                      aria-label="正则表达式"
                      value={variable.regex}
                      placeholder="例如：-v\\d+$"
                      status={invalidRegex(variable.regex) ? 'error' : undefined}
                      onChange={(event) => patchVariable(index, { regex: event.target.value })}
                    />
                    <Input
                      aria-label="替换内容"
                      value={variable.replacement}
                      placeholder="例如："
                      onChange={(event) =>
                        patchVariable(index, { replacement: event.target.value })
                      }
                    />
                    <Tooltip title="删除变量">
                      <Button
                        danger
                        type="text"
                        aria-label="删除变量"
                        icon={<Trash2 size={15} />}
                        onClick={() =>
                          setServiceName((current) => ({
                            ...current,
                            variables: current.variables.filter(
                              (_, itemIndex) => itemIndex !== index,
                            ),
                          }))
                        }
                      />
                    </Tooltip>
                  </div>
                ))}
              </div>
            </div>
          ) : (
            <div className={styles.serviceEmptyStateCompact}>
              未配置派生变量，可直接使用工作负载和容器名称
            </div>
          )}
        </section>

        <div className={styles.injectionControlFooter}>
          <Button
            type="primary"
            icon={<Save size={15} />}
            loading={saving}
            disabled={!changed || invalidServiceName}
            onClick={save}
          >
            下发配置
          </Button>
        </div>
      </div>
    </Spin>
  )
}
