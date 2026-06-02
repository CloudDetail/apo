/**
 * Copyright 2026 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { chartSeriesLimitForWidget, normalizeChartSeriesLimit, queryItemsForQuery } from './utils.mjs'

const executableQueryItemsForQuery = (query = {}) => (
  queryItemsForQuery(query).filter((item) => item.hidden !== true)
)

export const timeRangeOptions = [
  { labelKey: 'last15Minutes', value: 'now-15m', durationMs: 15 * 60 * 1000 },
  { labelKey: 'last1Hour', value: 'now-1h', durationMs: 60 * 60 * 1000 },
  { labelKey: 'last6Hours', value: 'now-6h', durationMs: 6 * 60 * 60 * 1000 },
  { labelKey: 'last24Hours', value: 'now-24h', durationMs: 24 * 60 * 60 * 1000 },
]

export const fixedDashboardColumns = 4
export const defaultDashboardWidgetWidth = 2
const legacyDashboardColumns = 12
const legacyTwoColumnCount = 2

const safeDashboardColumns = (columns = fixedDashboardColumns) => (
  Math.max(1, Math.min(Number(columns) || fixedDashboardColumns, fixedDashboardColumns))
)

const defaultWidgetWidthForColumns = (columns = fixedDashboardColumns) => (
  Math.max(1, Math.min(defaultDashboardWidgetWidth, safeDashboardColumns(columns)))
)

const legacyTwoColumnX = (x = 0, columns = fixedDashboardColumns) => {
  const safeColumns = safeDashboardColumns(columns)
  const legacyColumn = Math.max(0, Math.min(Math.round(Number(x) || 0), legacyTwoColumnCount - 1))
  return Math.min(safeColumns - 1, legacyColumn * defaultWidgetWidthForColumns(safeColumns))
}

const scaleLegacyDashboardLayout = (layout = {}, columns = fixedDashboardColumns) => {
  const safeColumns = safeDashboardColumns(columns)
  const sourceX = Number(layout.x)
  const sourceY = Number(layout.y)
  const sourceW = Number(layout.w)
  const sourceH = Number(layout.h)
  const width = Number.isFinite(sourceW) && sourceW > 0
    ? Math.max(1, Math.round(sourceW * safeColumns / legacyDashboardColumns))
    : defaultWidgetWidthForColumns(safeColumns)
  const w = Math.min(width, safeColumns)
  const x = Math.max(0, Math.min(
    Math.floor((Number.isFinite(sourceX) ? sourceX : 0) * safeColumns / legacyDashboardColumns),
    safeColumns - w,
  ))
  const rowHeight = Number.isFinite(sourceH) && sourceH > 0 ? sourceH : 4
  return {
    x,
    y: Math.floor(Math.max(0, Number.isFinite(sourceY) ? sourceY : 0) / rowHeight),
    w,
    h: 1,
  }
}

export const semanticVariableTypes = [
  { name: 'cluster', label: 'Cluster', aliases: ['cluster', 'cluster_id', 'clusterId', 'clusterID', 'cluster_name'], snippetLabel: 'cluster_id' },
  { name: 'namespace', label: 'Namespace', aliases: ['namespace', 'namespace_name', 'namespaceName', 'src_namespace', 'dst_namespace'], snippetLabel: 'namespace' },
  { name: 'service', label: 'Service', aliases: ['svc', 'svc_name', 'service', 'service_name'], snippetLabel: 'svc_name' },
  { name: 'endpoint', label: 'Endpoint', aliases: ['endpoint', 'endpoint_name', 'content_key', 'contentKey'], snippetLabel: 'endpoint' },
  { name: 'pod', label: 'Pod', aliases: ['pod', 'pod_name', 'src_pod', 'dst_pod'], snippetLabel: 'pod' },
  { name: 'instance', label: 'Instance', aliases: ['instance', 'instance_name', 'node_name', 'nodeName', 'host_name', 'hostName'], snippetLabel: 'instance_name' },
]

const localizedDefaultSemanticLabels = {
  cluster: ['\u96c6\u7fa4'],
  namespace: ['\u547d\u540d\u7a7a\u95f4'],
  service: ['\u670d\u52a1', '\u670d\u52a1\u540d'],
  endpoint: ['\u670d\u52a1\u7aef\u70b9'],
  instance: ['\u5b9e\u4f8b', '\u5b9e\u4f8b\u540d'],
}

const localizedDefaultVariableLabels = {
  service_name: {
    labelKey: 'application',
    labels: ['\u5e94\u7528\u540d', 'Application', 'Application Name'],
  },
  service_type: {
    labelKey: 'queryMode',
    labels: ['\u67e5\u8be2\u65b9\u5f0f', 'Query mode', 'Query Mode'],
  },
  related_nodes: {
    labelKey: 'relatedNodes',
    labels: ['\u76f8\u5173\u8282\u70b9', 'Related Nodes'],
  },
  resolution: {
    labelKey: 'resolution',
    labels: ['\u7cbe\u5ea6', 'Resolution'],
  },
  jvm_instance: {
    labelKey: 'jvmInstance',
    labels: ['JVM\u5b9e\u4f8b', 'JVM Instance'],
  },
  job_name: {
    labelKey: 'hostMetricsJob',
    labels: ['\u4e3b\u673a\u6307\u6807 Job', 'Host metrics Job', 'Host Metrics Job', 'Node Job'],
  },
  cadvisor_job_name: {
    labelKey: 'containerMetricsJob',
    labels: ['\u5bb9\u5668\u6307\u6807 Job', 'Container metrics Job', 'Container Metrics Job'],
  },
}

const localizedDefaultVariableOptionLabels = {
  service_type: [
    {
      labelKey: 'queryByK8s',
      labels: ['\u6309K8s\u67e5\u8be2', 'Query by K8s'],
      values: ['.*'],
    },
    {
      labelKey: 'queryByVirtualMachine',
      labels: ['\u6309\u865a\u673a\u67e5\u8be2', 'Query by VM'],
      values: [''],
    },
  ],
}

const isDefaultSemanticLabel = (semantic = {}, label = '') => {
  if (!label) return true
  return label === semantic.label
    || label === semantic.name
    || semantic.aliases?.includes(label)
    || localizedDefaultSemanticLabels[semantic.name]?.includes(label)
}

const localizedDefaultVariableLabelKey = (variable = {}) => {
  const config = localizedDefaultVariableLabels[variable.name]
  if (!config) return ''
  if (!variable.label || config.labels.includes(variable.label)) return config.labelKey
  return ''
}

const localizedDefaultVariableOptionLabelKey = (definition = {}, candidate = {}) => {
  const configs = localizedDefaultVariableOptionLabels[definition.name]
  if (!configs) return ''
  const label = String(candidate.label ?? '')
  const value = String(candidate.value ?? '')
  return configs.find((config) => (
    config.labels.includes(label) || (!label && config.values.includes(value))
  ))?.labelKey || ''
}

const globalPageContextVariableAliases = new Set([
  'dataGroup',
  'data_group',
  'data_group_id',
  'group_id',
  'groupId',
  'timeRange',
  'time_range',
])

export const isGlobalPageContextVariable = (definition = {}) => (
  globalPageContextVariableAliases.has(definition.name)
  || globalPageContextVariableAliases.has(definition.type)
  || (definition.labels || []).some((label) => globalPageContextVariableAliases.has(label))
)

const variableReadinessDefinition = (variable = {}) => ({
  name: variable.name || '',
  type: variable.type || '',
  label: variable.label || '',
  labels: variable.labels || [],
  query: variable.query || '',
  includeAll: Boolean(variable.includeAll),
  allValue: variable.allValue || '',
  multi: variable.multi !== false,
  options: (variable.options || []).map((option) => ({
    label: option?.label ?? option,
    value: option?.value ?? option?.label ?? option,
  })),
})

export const dashboardVariableReadinessScopeKey = ({
  dashboardId,
  dataGroupId,
  globalRefreshKey,
  queryText = '',
  variables = [],
  selectedValues = {},
} = {}) => JSON.stringify({
  dashboardId: dashboardId || '',
  dataGroupId: dataGroupId ?? null,
  globalRefreshKey: globalRefreshKey || '',
  queryText,
  variables: variables.map(variableReadinessDefinition),
  selectedValues,
})

export const semanticVariableByName = (name = '') => (
  semanticVariableTypes.find((item) => item.name === name)
)

export const semanticVariableForLabel = (label = '') => (
  semanticVariableTypes.find((item) => item.aliases.includes(label))
)

const standardResourceVariableNames = new Set(['cluster', 'namespace', 'service', 'endpoint', 'pod', 'instance'])
const customLabelVariableNames = new Set([
  'related_nodes',
  'related_pid',
  'jvm_instance',
  'job_name',
  'cadvisor_job_name',
])

const standardResourceVariableType = (definition = {}) => {
  if (customLabelVariableNames.has(String(definition.name || ''))) return ''
  const values = [
    definition.type,
    definition.name,
    ...(definition.labels || []),
  ].filter(Boolean).map(String)
  const exactType = values.find((value) => standardResourceVariableNames.has(value))
  if (exactType) return exactType
  return values.map((value) => semanticVariableForLabel(value)?.name).find(Boolean) || ''
}

export const labelsForVariable = (definition = {}) => {
  if (Array.isArray(definition.labels) && definition.labels.length) return definition.labels
  return semanticVariableByName(definition.name || definition.type)?.aliases || []
}

const resourceVariableSnippet = (definition = {}) => `${definition.snippetLabel || definition.aliases?.[0] || definition.name}=~"$${definition.name}"`

const resourceVariableHint = (definition = {}, labels = []) => ({
  name: definition.name,
  label: definition.label,
  type: definition.name,
  ...(labels.length ? { labels } : {}),
  snippet: resourceVariableSnippet(definition),
})

export const dashboardActions = (dashboard, canCreate = false) => ({
  view: Boolean(dashboard?.id),
  edit: Boolean(dashboard?.canEdit),
  copy: Boolean(dashboard?.canCopy && canCreate),
  delete: Boolean(dashboard?.canDelete),
})

export const createRequestForTemplate = (templateID, template, t) => {
  const request = {
    templateId: templateID || 'blank',
    blank: !templateID || templateID === 'blank',
  }
  if (template && t) {
    request.name = templateDisplayName(t, template)
    request.description = templateDisplayDescription(t, template)
  }
  return request
}

export const needsUnsavedPrompt = (dirty, target) => Boolean(dirty && target)

export const nextPreviewMode = (current) => !current

export const editorPanelMode = (selectedWidget) => (selectedWidget ? 'widget' : 'dashboard')

export const addWidgetPlaceholderMode = (widgets = [], previewing = false) => {
  if (previewing) return 'hidden'
  return widgets.length > 0 ? 'append' : 'empty'
}

export const addWidgetPlaceholderCopy = (mode = 'empty', t) => ({
  title: t(mode === 'empty' ? 'emptyCanvasTitle' : 'addWidgetPlaceholderTitle'),
  description: t(mode === 'empty' ? 'emptyCanvasHint' : 'addWidgetPlaceholderHint'),
  action: t('addWidget'),
})

export const addWidgetPlaceholderGridStyle = (widgets = [], columns = fixedDashboardColumns, previewing = false) => {
  const mode = addWidgetPlaceholderMode(widgets, previewing)
  if (mode === 'hidden') return {}
  if (mode === 'empty') {
    return {
      gridColumn: '1 / -1',
      gridRow: '1 / span 1',
    }
  }
  const safeColumns = safeDashboardColumns(columns)
  const width = defaultWidgetWidthForColumns(safeColumns)
  return gridPlacement({
    x: (widgets.length % legacyTwoColumnCount) * width,
    y: Math.floor(widgets.length / legacyTwoColumnCount),
    w: width,
  }, safeColumns)
}

export const draftWidgetGridStyle = (widgets = [], columns = fixedDashboardColumns) => {
  const safeColumns = safeDashboardColumns(columns)
  const width = defaultWidgetWidthForColumns(safeColumns)
  return gridPlacement({
    x: (widgets.length % legacyTwoColumnCount) * width,
    y: Math.floor(widgets.length / legacyTwoColumnCount),
    w: width,
  }, safeColumns)
}

export const copyWidgetTitle = (title, t) => `${title} - ${t('copySuffix')}`

export const gridPlacement = (layout = {}, columns = fixedDashboardColumns) => {
  const { x, y, w, h } = normalizeGridLayout(layout, columns)
  return {
    gridColumn: `${x + 1} / span ${w}`,
    gridRow: `${y + 1} / span ${h}`,
  }
}

export const normalizeGridLayout = (layout = {}, columns = fixedDashboardColumns) => {
  const safeColumns = safeDashboardColumns(columns)
  let x = Number(layout.x)
  let y = Number(layout.y)
  let width = Number(layout.w)
  let height = Number(layout.h)
  if (!Number.isFinite(x)) x = 0
  if (!Number.isFinite(y)) y = 0
  const legacyLayout = width > safeColumns || x >= safeColumns
  if (legacyLayout) {
    return scaleLegacyDashboardLayout(layout, safeColumns)
  }
  if (!Number.isFinite(width) || width <= 0) {
    x = legacyTwoColumnX(x, safeColumns)
    width = defaultWidgetWidthForColumns(safeColumns)
  }
  const normalizedX = Math.max(0, Math.min(x, safeColumns - 1))
  const maxWidth = safeColumns - normalizedX
  const w = Number.isFinite(width) && width > 0
    ? Math.max(1, Math.min(Math.round(width), maxWidth))
    : 1
  const h = Number.isFinite(height) && height > 0
    ? Math.max(1, Math.round(height))
    : 1
  return {
    x: normalizedX,
    y: Math.max(0, y),
    w,
    h,
  }
}

export const gridLayoutsOverlap = (first, second) => (
  first.x < second.x + second.w
  && first.x + first.w > second.x
  && first.y < second.y + second.h
  && first.y + first.h > second.y
)

const hasGridCollision = (layout, placed) => placed.some((item) => gridLayoutsOverlap(layout, item.layout))

export const compactWidgetLayouts = (widgets = [], columns = fixedDashboardColumns) => {
  const safeColumns = safeDashboardColumns(columns)
  const placed = []
  return widgets
    .map((widget, index) => ({ widget, index, layout: normalizeGridLayout(widget.layout, safeColumns) }))
    .sort((a, b) => a.layout.y - b.layout.y || a.layout.x - b.layout.x || a.index - b.index)
    .map(({ widget, layout }) => {
      let nextLayout = null
      for (let y = 0; nextLayout === null && y < 1000; y += 1) {
        for (let x = 0; x <= safeColumns - layout.w; x += 1) {
          const candidate = { ...layout, x, y }
          if (!hasGridCollision(candidate, placed)) {
            nextLayout = candidate
            break
          }
        }
      }
      const item = { ...widget, layout: nextLayout || layout }
      placed.push({ id: widget.id, layout: item.layout })
      return item
    })
}

export const findOpenGridLayout = (layout = {}, placed = [], columns = fixedDashboardColumns) => {
  const safeColumns = safeDashboardColumns(columns)
  const desired = normalizeGridLayout(layout, safeColumns)
  if (!hasGridCollision(desired, placed)) return desired

  const maxPlacedY = placed.reduce((max, item) => Math.max(max, item.layout.y), desired.y)
  const maxY = maxPlacedY + 100
  for (let y = desired.y; y <= maxY; y += 1) {
    for (let x = 0; x <= safeColumns - desired.w; x += 1) {
      const candidate = { ...desired, x, y }
      if (!hasGridCollision(candidate, placed)) return candidate
    }
  }
  return { ...desired, x: 0, y: maxPlacedY }
}

export const arrangeWidgetLayouts = (widgets = [], columns = fixedDashboardColumns) => {
  return Object.fromEntries(widgets.map((item) => [item.id, normalizeGridLayout(item.layout, columns)]))
}

export const updateWidgetGridLayout = (widgets = [], widgetID, patch = {}, columns = fixedDashboardColumns) => {
  const target = widgets.find((widget) => widget.id === widgetID)
  if (!target) return widgets

  const others = widgets
    .filter((widget) => widget.id !== widgetID)
    .map((widget, index) => ({ widget, index, layout: normalizeGridLayout(widget.layout, columns) }))
    .sort((a, b) => a.layout.y - b.layout.y || a.layout.x - b.layout.x || a.index - b.index)

  const targetLayout = normalizeGridLayout({ ...target.layout, ...patch }, columns)
  if (hasGridCollision(targetLayout, others)) return widgets

  return widgets.map((widget) => (
    widget.id === widgetID ? { ...widget, layout: targetLayout } : widget
  ))
}

export const moveWidgetToGridPosition = (widgets = [], widgetID, x, y, columns = fixedDashboardColumns) => (
  updateWidgetGridLayout(widgets, widgetID, { x, y }, columns)
)

export const resizeWidgetGridLayout = (widgets = [], widgetID, patch = {}, columns = fixedDashboardColumns) => (
  updateWidgetGridLayout(widgets, widgetID, patch, columns)
)

export const addWidgetToOpenCell = (widgets = [], widget, columns = fixedDashboardColumns) => {
  const placed = widgets.map((item, index) => ({ id: item.id, layout: normalizeGridLayout(item.layout, columns), index }))
  const layout = findOpenGridLayout(widget.layout, placed, columns)
  return [...widgets, { ...widget, layout }]
}

export const cloneWidgetDraft = (widget = null) => {
  if (!widget) return null
  return JSON.parse(JSON.stringify(widget))
}

export const widgetDraftMode = (widgets = [], draftWidget = null) => {
  if (!draftWidget?.id) return ''
  return widgets.some((widget) => widget.id === draftWidget.id) ? 'edit' : 'new'
}

export const shouldPromptWidgetDraftDiscard = (draftWidget = null, draftDirty = false) => (
  Boolean(draftWidget?.id && draftDirty)
)

export const applyWidgetDraftToDashboard = (widgets = [], draftWidget = null, columns = fixedDashboardColumns) => {
  if (!draftWidget?.id) return widgets
  const widget = cloneWidgetDraft(draftWidget)
  if (widgetDraftMode(widgets, widget) === 'edit') {
    return widgets.map((item) => (item.id === widget.id ? widget : item))
  }
  return addWidgetToOpenCell(widgets, widget, columns)
}

const resolveWidgetGridCollisions = (widgets = [], activeID = '', columns = fixedDashboardColumns) => {
  if (!activeID) return widgets
  const items = widgets.map((widget, index) => ({
    widget,
    index,
    id: widget.id,
    layout: normalizeGridLayout(widget.layout, columns),
  }))
  const itemByID = new Map(items.map((item) => [item.id, item]))
  if (!itemByID.has(activeID)) return widgets

  const queue = [activeID]
  while (queue.length) {
    const current = itemByID.get(queue.shift())
    if (!current) continue
    items
      .filter((item) => item.id !== current.id)
      .sort((a, b) => a.layout.y - b.layout.y || a.layout.x - b.layout.x || a.index - b.index)
      .forEach((item) => {
        if (!gridLayoutsOverlap(current.layout, item.layout)) return
        const nextY = current.layout.y + current.layout.h
        if (item.layout.y >= nextY) return
        item.layout = { ...item.layout, y: nextY }
        queue.push(item.id)
      })
  }

  return items.map(({ widget, layout }) => ({ ...widget, layout }))
}

const compactEmptyGridRows = (widgets = [], columns = fixedDashboardColumns) => {
  const items = widgets.map((widget) => ({ ...widget, layout: normalizeGridLayout(widget.layout, columns) }))
  const occupiedRows = new Set()
  items.forEach(({ layout }) => {
    for (let y = layout.y; y < layout.y + layout.h; y += 1) {
      occupiedRows.add(y)
    }
  })
  const maxBottom = items.reduce((max, item) => Math.max(max, item.layout.y + item.layout.h), 0)
  const emptyRows = []
  for (let y = 0; y < maxBottom; y += 1) {
    if (!occupiedRows.has(y)) emptyRows.push(y)
  }
  if (!emptyRows.length) return items

  return items.map((item) => {
    const shift = emptyRows.filter((row) => row < item.layout.y).length
    return shift ? { ...item, layout: { ...item.layout, y: item.layout.y - shift } } : item
  })
}

export const mergeWidgetGridLayouts = (widgets = [], layouts = [], columns = fixedDashboardColumns, activeLayout = null) => {
  const layoutByID = new Map(layouts.map((layout) => [layout.i, layout]))
  const activeID = activeLayout?.i
  const updateIDs = activeID ? new Set([activeID]) : new Set(layoutByID.keys())
  const nextWidgets = widgets.map((widget) => {
    if (!updateIDs.has(widget.id)) return widget
    const layout = activeID === widget.id ? activeLayout : layoutByID.get(widget.id)
    return layout ? { ...widget, layout: normalizeGridLayout(layout, columns) } : widget
  })
  return activeID ? compactEmptyGridRows(resolveWidgetGridCollisions(nextWidgets, activeID, columns), columns) : nextWidgets
}

export const responsiveGridColumns = () => fixedDashboardColumns

export const visibilityLabel = (t, value) => t(value || 'personal')

const translatedCatalogText = (t, key, fallback = '') => {
  const value = t(key)
  return value && value !== key && value !== `t:${key}` ? value : fallback
}

export const templateDisplayName = (t, template = {}) => (
  translatedCatalogText(t, `templates.${template.id}.name`, template.name || '')
)

export const templateDisplayDescription = (t, template = {}) => (
  translatedCatalogText(t, `templates.${template.id}.description`, template.description || '')
)

export const presetDisplayName = (t, preset = {}) => (
  translatedCatalogText(t, `presetQueries.${preset.id}.name`, preset.name || '')
)

export const widgetDisplayTitle = (t, widget = {}, presets = []) => {
  const preset = presets.find((item) => item.id === widget.query?.presetId)
  if (!preset || widget.title !== preset.name) return widget.title || ''
  return presetDisplayName(t, preset)
}

export const presetParameterLabel = (t, param = {}) => (
  translatedCatalogText(t, `presetParameters.${param.name}.label`, param.label || param.name || '')
)

export const presetParameterDescription = (t, param = {}) => (
  translatedCatalogText(t, `presetParameters.${param.name}.description`, param.description || '')
)

const padDatePart = (value) => String(value).padStart(2, '0')

export const formatDashboardTimestamp = (timestamp) => {
  const value = Number(timestamp)
  if (!Number.isFinite(value) || value <= 0) return '-'
  const date = new Date(value)
  return [
    date.getFullYear(),
    padDatePart(date.getMonth() + 1),
    padDatePart(date.getDate()),
  ].join('-') + ' ' + [
    padDatePart(date.getHours()),
    padDatePart(date.getMinutes()),
    padDatePart(date.getSeconds()),
  ].join(':')
}

export const saveFailureMessage = (t, error) => {
  const data = error?.response?.data
  if (data?.code === 'A0006') return t('noEditPermission')
  return data?.message || data?.msg || error?.message || t('saveFailed')
}

export const runDashboardMutationWithPermissionRefresh = async (
  mutation,
  refreshPermission,
  onRefreshFailure,
) => {
  const result = await mutation()
  try {
    await refreshPermission?.()
  } catch (error) {
    onRefreshFailure?.(error)
  }
  return result
}

export const dashboardLoadFailureMessage = (t, error) => {
  const data = error?.response?.data
  if (data?.code === 'A0006') return t('noViewPermission')
  return data?.message || data?.msg || error?.message || t('dashboardNotFound')
}

export const previewStateKind = (response) => {
  const status = response?.status || 'empty'
  if (status === 'success') return response?.series?.length ? 'success' : 'no_data'
  if (['syntax_error', 'timeout', 'too_large', 'no_permission', 'metric_missing', 'no_data'].includes(status)) {
    return status
  }
  return 'query_failure'
}

const previewSuggestionKeys = {
  syntax_error: 'syntaxErrorSuggestion',
  timeout: 'timeoutSuggestion',
  too_large: 'tooLargeSuggestion',
  no_permission: 'noPermissionSuggestion',
  metric_missing: 'metricMissingSuggestion',
  no_data: 'noDataSuggestion',
  error: 'queryFailureSuggestion',
  query_failure: 'queryFailureSuggestion',
}

const previewStatusLabelKeys = {
  syntax_error: 'statusSyntaxError',
  timeout: 'statusTimeout',
  too_large: 'statusTooLarge',
  no_permission: 'statusNoPermission',
  metric_missing: 'statusMetricMissing',
  no_data: 'statusNoData',
  error: 'statusQueryFailure',
  query_failure: 'statusQueryFailure',
}

export const previewStatusLabel = (t, status = '') => {
  const key = previewStatusLabelKeys[status]
  return key ? t(key) : status
}

export const previewStatusDetails = (status = '') => ({
  color: status === 'no_data' ? 'default' : 'error',
  suggestionKey: previewSuggestionKeys[status] || previewSuggestionKeys.query_failure,
})

export const variableCandidateErrorMessage = (t, error) => {
  const data = error?.response?.data
  if (!error) return t('noAvailableValues')
  if (data?.code === 'A0006') return t('noVariablePermission')
  return data?.message || error?.message || t('noAvailableValues')
}

export const variableCandidateLabel = (t, item = {}, definition = {}) => {
  const defaultOptionLabelKey = localizedDefaultVariableOptionLabelKey(definition, item)
  if (defaultOptionLabelKey) return t(defaultOptionLabelKey)
  return item.type === 'dataGroup' && String(item.value) === '0' && item.label === 'All Data'
    ? t('allData')
    : item.label
}

export const variableDisplayLabel = (t, variable = {}) => {
  const defaultVariableLabelKey = localizedDefaultVariableLabelKey(variable)
  if (defaultVariableLabelKey) return t(defaultVariableLabelKey)

  const semantic = semanticVariableByName(variable.type)
    || semanticVariableByName(variable.name)
    || semanticVariableForLabel(variable.name)
  if (!semantic) return variable.label || t(variable.name)
  if (isDefaultSemanticLabel(semantic, variable.label)) {
    return t(semantic.name)
  }
  return variable.label
}

export const translatedTimeRangeOptions = (t) => timeRangeOptions.map((item) => ({
  label: t(item.labelKey),
  value: item.value,
}))

export const presetOptionsForChart = (presets = [], chartType, t = (key) => key) => {
  const groups = []
  const byCategory = new Map()
  presets
    .filter((preset) => !chartType || preset.compatibleCharts?.includes(chartType))
    .forEach((preset) => {
      const category = preset.category || 'other'
      if (!byCategory.has(category)) {
        const group = { label: t(`presetCategory.${category}`), options: [] }
        byCategory.set(category, group)
        groups.push(group)
      }
      byCategory.get(category).options.push({ label: presetDisplayName(t, preset), value: preset.id })
    })
  return groups
}

export const mergePresetDefaults = (preset, params = {}) => {
  const next = { ...params }
  ;(preset?.parameters || []).forEach((param) => {
    if (next[param.name] === undefined || next[param.name] === '') {
      next[param.name] = param.default || ''
    }
  })
  return next
}

export const dashboardQueryText = (widgets = [], presets = []) => widgets
  .map((widget) => {
    if (widget.query?.mode === 'preset') {
      return presets.find((preset) => preset.id === widget.query?.presetId)?.expression || ''
    }
    return executableQueryItemsForQuery(widget.query).map((query) => query.expression).join('\n')
  })
  .filter(Boolean)
  .join('\n')

export const widgetQueryText = (widget = {}, presets = []) => {
  const expression = widget.query?.mode === 'preset'
    ? presets.find((preset) => preset.id === widget.query?.presetId)?.expression || ''
    : executableQueryItemsForQuery(widget.query).map((query) => query.expression).join('\n')
  return [
    expression,
    ...Object.values(widget.query?.params || {}),
  ].filter(Boolean).join('\n')
}

export const extractMetricLabels = (queryText = '') => {
  const labels = new Set()
  const selectorPattern = /\{((?:\$\{[^}]*\}|[^}])*)\}/g
  let match
  while ((match = selectorPattern.exec(queryText)) !== null) {
    match[1].split(',').forEach((part) => {
      const key = part.trim().split(/[=!~]/)[0]?.trim()
      if (key) labels.add(key)
    })
  }
  return Array.from(labels).sort()
}

const extractMetricMatchers = (queryText = '') => {
  const matchers = []
  const selectorPattern = /\{((?:\$\{[^}]*\}|[^}])*)\}/g
  let selectorMatch
  while ((selectorMatch = selectorPattern.exec(queryText)) !== null) {
    selectorMatch[1].split(',').forEach((part) => {
      const matcher = part.trim().match(/^([A-Za-z_][A-Za-z0-9_]*)\s*(!=|=~|!~|=)\s*"((?:\\.|[^"\\])*)"$/)
      if (!matcher) return
      matchers.push({
        label: matcher[1],
        operator: matcher[2],
        value: matcher[3],
      })
    })
  }
  return matchers
}

const fixedRegexFilterHints = (queryText = '') => {
  const seen = new Set()
  return extractMetricMatchers(queryText)
    .filter((matcher) => ['=~', '!~'].includes(matcher.operator))
    .filter((matcher) => matcher.value && matcher.value !== '.*' && !matcher.value.includes('$'))
    .filter((matcher) => {
      const key = `${matcher.label}:${matcher.operator}:${matcher.value}`
      if (seen.has(key)) return false
      seen.add(key)
      return true
    })
}

export const dashboardLabelOptions = (widgets = [], presets = []) => (
  (() => {
    const labels = extractMetricLabels(dashboardQueryText(widgets, presets))
      .filter((label) => !globalPageContextVariableAliases.has(label))
    const semanticOptions = semanticVariableTypes
      .map((variable) => ({
        variable,
        labels: labels.filter((label) => variable.aliases.includes(label)),
      }))
      .filter(({ labels }) => labels.length)
      .map(({ variable, labels }) => ({ label: variable.label, value: variable.name, type: variable.name, labels }))
    const semanticAliases = new Set(semanticVariableTypes.flatMap((variable) => variable.aliases))
    const rawOptions = labels
      .filter((label) => !semanticAliases.has(label))
      .map((label) => ({ label, value: label, type: 'label', labels: [label] }))
    return [...semanticOptions, ...rawOptions]
  })()
)

const builtinResourceVariableOptions = () => semanticVariableTypes.map((variable) => ({
  label: variable.label,
  value: variable.name,
  type: variable.name,
  labels: variable.aliases,
}))

const enabledVariableOptionKeys = (variables = []) => {
  const keys = new Set()
  variables.forEach((variable) => {
    const resourceType = standardResourceVariableType(variable)
    if (resourceType) {
      keys.add(resourceType)
      return
    }
    if (variable.name) keys.add(variable.name)
    ;(variable.labels || []).forEach((label) => keys.add(label))
  })
  return keys
}

export const dashboardVariableOptionGroups = (widgets = [], presets = [], variables = []) => {
  const enabledKeys = enabledVariableOptionKeys(variables)
  const builtinOptions = builtinResourceVariableOptions()
    .filter((option) => !enabledKeys.has(option.value))
  const queryLabelOptions = dashboardLabelOptions(widgets, presets)
    .filter((option) => option.type === 'label')
    .filter((option) => !enabledKeys.has(option.value))
    .filter((option) => !(option.labels || []).some((label) => enabledKeys.has(label)))
  return [
    ...(builtinOptions.length ? [{ labelKey: 'builtinVariableGroup', options: builtinOptions }] : []),
    ...(queryLabelOptions.length ? [{ labelKey: 'queryLabelVariableGroup', options: queryLabelOptions }] : []),
  ]
}

const defaultAllResourceVariableTypes = standardResourceVariableNames

export const dashboardVariableDefinition = (name = '', option = {}, t = (key) => key) => {
  const type = option?.type || name
  const resourceType = standardResourceVariableType({ ...option, name, type })
  const includeAll = option?.includeAll === true || defaultAllResourceVariableTypes.has(resourceType)
  const allValue = option?.allValue || (includeAll ? '.*' : undefined)
  return {
    name,
    label: option?.type === 'label' ? option.label : t(name),
    type,
    labels: option?.labels,
    multi: true,
    ...(includeAll ? { includeAll: true, allValue } : {}),
  }
}

export const promQLResourceVariableHints = (queryText = '') => {
  const labels = extractMetricLabels(queryText)
    .filter((label) => !globalPageContextVariableAliases.has(label))
  const semanticAliases = new Set(semanticVariableTypes.flatMap((variable) => variable.aliases))
  const fixed = fixedRegexFilterHints(queryText)
  const fixedLabels = new Set(fixed.map((item) => item.label))
  const supported = semanticVariableTypes.map((variable) => ({
    ...resourceVariableHint(variable),
    aliases: variable.aliases,
  }))
  const recognized = semanticVariableTypes
    .map((variable) => ({
      variable,
      labels: labels.filter((label) => variable.aliases.includes(label)),
    }))
    .filter(({ labels }) => labels.length)
    .map(({ variable, labels }) => resourceVariableHint(variable, labels))
  const unrecognized = labels
    .filter((label) => !semanticAliases.has(label))
    .filter((label) => !fixedLabels.has(label))
    .map((label) => ({ label }))
  return { supported, recognized, fixed, unrecognized }
}

export const insertPromQLMatcherSnippet = (expression = '', snippet = '') => {
  const text = String(expression || '')
  if (!snippet) return text
  if (!text.trim()) return `up{${snippet}}`
  const selectorPattern = /\{((?:\$\{[^}]*\}|[^}])*)\}/
  const selectorMatch = selectorPattern.exec(text)
  if (!selectorMatch) return `${text}{${snippet}}`
  const selector = selectorMatch[0]
  const body = selectorMatch[1].trim()
  const nextSelector = `{${body ? `${body},${snippet}` : snippet}}`
  return `${text.slice(0, selectorMatch.index)}${nextSelector}${text.slice(selectorMatch.index + selector.length)}`
}

export const widgetUsesVariable = (widget = {}, definition = {}, presets = []) => {
  const text = [widget.title, widget.description, widgetQueryText(widget, presets)].filter(Boolean).join('\n')
  const labels = extractMetricLabels(text)
  const aliases = labelsForVariable(definition)
  if (labels.some((label) => aliases.includes(label))) return true
  const token = definition.name || definition.type
  if (!token) return false
  return textUsesVariableToken(text, token)
}

const escapeVariableToken = (value = '') => String(value).replace(/[.*+?^${}()|[\]\\]/g, '\\$&')

export const textUsesVariableToken = (text = '', token = '') => {
  if (!token) return false
  return new RegExp(`(\\$\\{${escapeVariableToken(token)}(?::[A-Za-z_][A-Za-z0-9_]*)?\\}|\\$${escapeVariableToken(token)}\\b|\\{\\{\\s*${escapeVariableToken(token)}\\s*\\}\\})`).test(text)
}

export const variableDeletionState = (variable = {}, variables = [], widgets = [], presets = []) => {
  const token = variable.name || variable.type
  const widgetUsageCount = widgets.filter((widget) => widgetUsesVariable(widget, variable, presets)).length
  const variableQueryUsageCount = variables.filter((item) => (
    (item.name || item.type) !== token && textUsesVariableToken(item.query || '', token)
  )).length
  const usageCount = widgetUsageCount + variableQueryUsageCount
  return {
    disabled: usageCount > 0,
    usageCount,
  }
}

const grafanaIntervalMacroPattern = /(\$__rate_interval|\$\{__rate_interval\}|\$__interval|\$\{__interval\})/

const variableControlsGrafanaInterval = (variable = {}, text = '') => (
  (variable.name === 'resolution' || variable.type === 'resolution') && grafanaIntervalMacroPattern.test(text)
)

export const enabledVariablesForDashboard = (variables = [], widgets = [], presets = []) => (
  (() => {
    const widgetText = dashboardQueryText(widgets, presets)
    const variableQueryText = variables.map((variable) => variable.query || '').filter(Boolean).join('\n')
    const combinedText = [widgetText, variableQueryText].filter(Boolean).join('\n')
    return variables.filter((variable) => (
      !isGlobalPageContextVariable(variable) && (
        isStandardResourceVariable(variable) ||
        textUsesVariableToken(variableQueryText, variable.name || variable.type) ||
        variableControlsGrafanaInterval(variable, combinedText) ||
        widgets.some((widget) => widgetUsesVariable(widget, variable, presets))
      )
    ))
  })()
)

const variableAllValue = (definition = {}) => String(definition.allValue || '.*')

const asVariableValues = (value) => {
  if (Array.isArray(value)) return value.map((item) => String(item))
  if (value === undefined || value === null || value === '') return []
  return [String(value)]
}

const isAllCandidate = (definition = {}, candidate = {}) => {
  if (candidate.meta?.all === true || candidate.meta?.all === 'true') return true
  const value = String(candidate.value ?? '')
  const label = String(candidate.label ?? '').toLowerCase()
  if (definition.includeAll && value === variableAllValue(definition)) return true
  if ((definition.type === 'dataGroup' || definition.name === 'dataGroup') && value === '0') return true
  return label === 'all' || label === 'all data'
}

const variableAllValues = (definition = {}, candidates = []) => {
  const values = new Set()
  if (definition.includeAll) values.add(variableAllValue(definition))
  candidates.forEach((candidate) => {
    if (isAllCandidate(definition, candidate)) values.add(String(candidate.value))
  })
  return values
}

const isVariableAllSelected = (definition = {}, values = []) => (
  Boolean(definition.includeAll)
  && values.some((value) => String(value) === variableAllValue(definition))
)

const normalizeVariableCandidate = (definition = {}, option = {}) => {
  const value = typeof option === 'string' ? option : (option.value ?? option.label)
  const label = typeof option === 'string' ? option : (option.label ?? option.value)
  if (value === undefined || value === null || value === '') return null
  return {
    label: String(label ?? value),
    value: String(value),
    type: option.type || definition.type || 'custom',
    ...(option.meta ? { meta: option.meta } : {}),
  }
}

export const variableCandidatesWithAll = (definition = {}, candidates = []) => {
  const normalized = candidates
    .map((candidate) => normalizeVariableCandidate(definition, candidate))
    .filter(Boolean)
  if (!definition.includeAll) return normalized
  const allValue = variableAllValue(definition)
  return [
    { label: definition.allLabel || 'All', value: allValue, type: definition.type || 'custom', meta: { all: 'true' } },
    ...normalized.filter((candidate) => candidate.value !== allValue),
  ]
}

export const variableCandidatesForDefinition = (definition = {}) => (
  Array.isArray(definition.options)
    ? variableCandidatesWithAll(definition, definition.options)
    : []
)

export const variableSelectMode = (definition = {}) => (
  definition.multi !== false ? 'multiple' : undefined
)

export const normalizeVariableSelection = (definition = {}, value, previousValue, candidates = []) => {
  const allValues = variableAllValues(definition, candidates)
  const allValue = allValues.values().next().value
  const values = asVariableValues(value)
  if (definition.multi === false) {
    if (allValues.size > 0 && values.length === 0) return allValue
    return value ?? undefined
  }
  if (allValues.size === 0) return values
  if (values.length === 0) return [allValue]

  const selectedAllValues = values.filter((item) => allValues.has(String(item)))
  if (selectedAllValues.length === 0) return values

  const previousAllSelected = asVariableValues(previousValue).some((item) => allValues.has(String(item)))
  if (previousAllSelected && values.length > selectedAllValues.length) {
    return values.filter((item) => !allValues.has(String(item)))
  }
  return [selectedAllValues[selectedAllValues.length - 1]]
}

export const syncVariableSelectionWithCandidates = (definition = {}, value, candidates = []) => {
  const availableValues = new Set(candidates.map((candidate) => String(candidate.value)))
  const defaultValue = candidates.length ? String(candidates[0].value) : undefined
  const normalized = normalizeVariableSelection(definition, value, undefined, candidates)

  if (Array.isArray(normalized)) {
    const next = normalized.filter((item) => availableValues.has(String(item)))
    if (next.length || defaultValue === undefined) return next
    return [defaultValue]
  }
  if (normalized === undefined || normalized === null || normalized === '') return defaultValue
  return availableValues.has(String(normalized)) ? normalized : defaultValue
}

export const variableSelectOptions = (t, definition = {}, candidates = []) => {
  const allOptions = []
  const valueOptions = []
  candidates.forEach((item) => {
    const option = { label: variableCandidateLabel(t, item, definition), value: item.value }
    if (isAllCandidate(definition, item)) {
      allOptions.push(option)
    } else {
      valueOptions.push(option)
    }
  })
  if (allOptions.length === 0) return valueOptions
  return [
    { label: t('allOptionGroup'), options: allOptions },
    ...(valueOptions.length ? [{ label: t('availableOptionGroup'), options: valueOptions }] : []),
  ]
}

export const variableRequestQuery = (definition = {}, dashboardQuery = '') => (
  definition.query || dashboardQuery
)

export const variableValuesForQuery = (variables = {}, definitions = []) => {
  const result = {}
  definitions.forEach((definition) => {
    const value = variables[definition.name]
    if (Array.isArray(value)) {
      result[definition.name] = value.map((item) => String(item))
    } else if (value !== undefined && value !== null) {
      result[definition.name] = [String(value)]
    }
  })
  return result
}

const escapeRegExp = (value = '') => value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')

const escapePromQLString = (value = '') => String(value)
  .replace(/\\/g, '\\\\')
  .replace(/"/g, '\\"')
  .replace(/\n/g, '\\n')

const escapePromQLRegexValue = (value = '') => escapePromQLString(escapeRegExp(value))

export const applyLabelVariableMatchers = (expression = '', variables = {}, definitions = []) => {
  let next = expression
  definitions.forEach((definition) => {
    const value = variables[definition.name]
    const values = Array.isArray(value) ? value : String(value || '').split('|')
    const selectedValues = values.filter(Boolean)
    const matcher = isVariableAllSelected(definition, selectedValues)
      ? variableAllValue(definition)
      : selectedValues.map(escapePromQLRegexValue).join('|')
    const labels = labelsForVariable(definition)
    if (!matcher || labels.length === 0) return
    const labelMatchers = labels.map((label) => ({
      label,
      pattern: new RegExp(`^${escapeRegExp(label)}\\s*(=|!=|=~|!~)`),
    }))
    const selectorPattern = /\{((?:\$\{[^}]*\}|[^}])*)\}/g
    const variableTokenPattern = /(\$\{[A-Za-z_][A-Za-z0-9_]*(?::[A-Za-z_][A-Za-z0-9_]*)?\}|\$[A-Za-z_][A-Za-z0-9_]*|\{\{\s*[A-Za-z_][A-Za-z0-9_]*\s*\}\})/
    next = next.replace(selectorPattern, (_, selector) => {
      const parts = selector.split(',').map((part) => part.trim()).filter(Boolean)
      const matchedLabels = labelMatchers
        .filter(({ pattern }) => parts.some((part) => pattern.test(part) && !variableTokenPattern.test(part)))
        .map(({ label }) => label)
      if (matchedLabels.length === 0) {
        return `{${parts.join(',')}}`
      }
      const withoutMatchedLabels = parts.filter((part) => (
        !labelMatchers.some(({ label, pattern }) => (
          matchedLabels.includes(label) && pattern.test(part) && !variableTokenPattern.test(part)
        ))
      ))
      return `{${[...matchedLabels.map((label) => `${label}=~"${matcher}"`), ...withoutMatchedLabels].join(',')}}`
    })
  })
  return next
}

export const displayText = (t, message, position) => {
  if (position === undefined || position === null || position < 0) {
    return message
  }
  return `${message} (${t('position')} ${position})`
}

export const rangeToQueryWindow = (rangeValue, now = Date.now()) => {
  const option = timeRangeOptions.find((item) => item.value === rangeValue) || timeRangeOptions[1]
  return {
    startTime: now - option.durationMs,
    endTime: now,
    step: option.durationMs > 6 * 60 * 60 * 1000 ? 300 : 60,
  }
}

export const globalTimeToQueryWindow = (timeRange = {}, now = Date.now()) => {
  const start = Number(timeRange.startTime)
  const end = Number(timeRange.endTime)
  if (Number.isFinite(start) && Number.isFinite(end) && end > start) {
    const startTime = start > 9999999999999 ? Math.floor(start / 1000) : start
    const endTime = end > 9999999999999 ? Math.floor(end / 1000) : end
    const duration = endTime - startTime
    return {
      startTime,
      endTime,
      step: duration > 6 * 60 * 60 * 1000 ? 300 : 60,
    }
  }
  return rangeToQueryWindow('now-15m', now)
}

export const timeRangeRefreshDependency = (timeRange = {}) => [
  timeRange.startTime ?? '',
  timeRange.endTime ?? '',
  timeRange.refreshTimestamp ?? '',
].join(':')

export const WIDGET_PREVIEW_CONCURRENCY_LIMIT = 2

export const WIDGET_PREVIEW_PRIORITY = Object.freeze({
  fullscreen: 0,
  retry: 0,
  visible: 1,
  near: 2,
  fallback: 3,
})

export const widgetPreviewEligibilityForVisibility = ({
  visibility = '',
  observationAvailable = true,
  priority = '',
} = {}) => {
  if (priority) return priority
  if (!observationAvailable) return 'fallback'
  if (visibility === 'visible' || visibility === 'near') return visibility
  return ''
}

export const widgetPreviewVisibilityFromRects = ({
  rootTop,
  rootBottom,
  itemTop,
  itemBottom,
  preloadMargin = 0,
} = {}) => {
  const values = [rootTop, rootBottom, itemTop, itemBottom].map(Number)
  if (values.some((value) => !Number.isFinite(value))) return 'offscreen'
  const [safeRootTop, safeRootBottom, safeItemTop, safeItemBottom] = values
  const margin = Math.max(0, Number(preloadMargin) || 0)
  if (safeItemBottom > safeRootTop && safeItemTop < safeRootBottom) return 'visible'
  if (safeItemBottom > safeRootTop - margin && safeItemTop < safeRootBottom + margin) return 'near'
  return 'offscreen'
}

export const shouldRefreshAllWidgetPreviews = ({ previews = {}, changedVariables = [] } = {}) => (
  !Object.keys(previews).some((key) => !String(key).startsWith('draft:'))
  || changedVariables.length === 0
)

export const shouldRefreshWidgetVisibilityOnScroll = ({ observationAvailable = true } = {}) => (
  !observationAvailable
)

export const shouldRenderWidgetChart = ({
  draft = false,
  alreadyRendered = false,
  visibility = '',
  scrolling = false,
} = {}) => Boolean(
  draft
  || alreadyRendered
  || (visibility === 'visible' && !scrolling)
)

const renderedChartTimestamp = (value, fallback = 0) => {
  const timestamp = Number(value)
  return Number.isFinite(timestamp) ? timestamp : fallback
}

export const reconcileRenderedChartWidgets = ({
  widgetIds = [],
  rendered = {},
  visibilities = {},
  distances = {},
  pinnedWidgetIds = [],
  now = Date.now(),
  maxActiveCharts = 12,
  minHoldMs = 2500,
} = {}) => {
  const safeNow = renderedChartTimestamp(now, Date.now())
  const safeMaxActiveCharts = Math.max(1, Number(maxActiveCharts) || 1)
  const safeMinHoldMs = Math.max(0, Number(minHoldMs) || 0)
  const widgetIdSet = new Set(widgetIds.filter(Boolean))
  const pinnedWidgetIdSet = new Set(pinnedWidgetIds.filter(Boolean))
  const candidateIds = new Set()

  widgetIdSet.forEach((widgetId) => {
    const visibility = visibilities[widgetId]
    if (rendered[widgetId] || visibility === 'visible' || visibility === 'near') {
      candidateIds.add(widgetId)
    }
  })

  const candidates = [...candidateIds].map((widgetId) => {
    const previous = rendered[widgetId]
    const previousMeta = previous && typeof previous === 'object' ? previous : {}
    const visibility = visibilities[widgetId] || 'offscreen'
    const renderedAt = renderedChartTimestamp(previousMeta.renderedAt, safeNow)
    const previousLastVisibleAt = renderedChartTimestamp(previousMeta.lastVisibleAt, renderedAt)
    const lastVisibleAt = visibility === 'visible' ? safeNow : previousLastVisibleAt
    const distance = renderedChartTimestamp(distances[widgetId], Number.POSITIVE_INFINITY)
    const retainedAt = Math.max(renderedAt, previousLastVisibleAt)
    const age = Math.max(0, safeNow - retainedAt)

    return {
      widgetId,
      visibility,
      distance,
      pinned: pinnedWidgetIdSet.has(widgetId),
      recent: age < safeMinHoldMs,
      meta: {
        renderedAt,
        lastVisibleAt,
      },
    }
  })

  const keptIds = new Set(candidates.map((candidate) => candidate.widgetId))
  if (keptIds.size > safeMaxActiveCharts) {
    candidates
      .filter((candidate) => (
        candidate.visibility !== 'visible'
        && !candidate.pinned
        && !candidate.recent
      ))
      .sort((left, right) => right.distance - left.distance)
      .some((candidate) => {
        if (keptIds.size <= safeMaxActiveCharts) return true
        keptIds.delete(candidate.widgetId)
        return false
      })
  }

  return candidates.reduce((next, candidate) => {
    if (keptIds.has(candidate.widgetId)) {
      next[candidate.widgetId] = candidate.meta
    }
    return next
  }, {})
}

const completedWidgetPreviewStatuses = new Set(['success', 'no_data'])
const activeWidgetPreviewStatuses = new Set(['queued', 'loading', 'retrying'])

export const shouldScheduleWidgetPreview = (state = {}, requestKey = '', { retry = false } = {}) => {
  if (!requestKey) return false
  if (retry) return true
  if (!state || !state.status) return true
  if (state.requestKey !== requestKey) return true
  if (state.stale) return true
  if (activeWidgetPreviewStatuses.has(state.status)) return false
  if (completedWidgetPreviewStatuses.has(state.status)) return false
  if (state.status === 'failed') return false
  return true
}

const schedulerAbortController = () => {
  if (typeof AbortController !== 'undefined') return new AbortController()
  const signal = { aborted: false }
  return {
    signal,
    abort: () => {
      signal.aborted = true
    },
  }
}

export const createDashboardPreviewScheduler = ({ limit = WIDGET_PREVIEW_CONCURRENCY_LIMIT } = {}) => {
  const maxConcurrent = Math.max(1, Number(limit) || WIDGET_PREVIEW_CONCURRENCY_LIMIT)
  const queued = new Map()
  const active = new Map()
  let order = 0
  let runID = 0
  let generation = 0

  const nextQueuedJob = () => (
    [...queued.values()].sort((a, b) => a.priority - b.priority || a.order - b.order)[0]
  )

  const drain = () => {
    while (active.size < maxConcurrent && queued.size > 0) {
      const job = nextQueuedJob()
      if (!job) return
      queued.delete(job.widgetId)
      const controller = schedulerAbortController()
      const entry = {
        controller,
        generation,
      }
      const id = ++runID
      active.set(id, entry)
      Promise.resolve()
        .then(() => job.run({
          signal: controller.signal,
          isCancelled: () => active.get(id) !== entry || generation !== entry.generation,
        }))
        .catch(() => {})
        .finally(() => {
          if (active.get(id) !== entry) return
          active.delete(id)
          drain()
        })
    }
  }

  return {
    schedule: (job = {}) => {
      if (!job.widgetId || !job.requestKey || typeof job.run !== 'function') return
      queued.set(job.widgetId, {
        ...job,
        priority: Number.isFinite(Number(job.priority)) ? Number(job.priority) : WIDGET_PREVIEW_PRIORITY.near,
        order: ++order,
      })
      drain()
    },
    clear: () => {
      generation += 1
      queued.clear()
      active.forEach((entry) => entry.controller?.abort?.())
      active.clear()
    },
    activeCount: () => active.size,
    queuedCount: () => queued.size,
  }
}

export const createVariableCandidateScheduler = ({ limit = VARIABLE_CANDIDATE_CONCURRENCY_LIMIT } = {}) => {
  const maxConcurrent = Math.max(1, Number(limit) || VARIABLE_CANDIDATE_CONCURRENCY_LIMIT)
  const queued = new Map()
  const active = new Map()
  const idleWaiters = []
  let order = 0
  let runID = 0
  let generation = 0

  const notifyIdle = () => {
    if (queued.size > 0 || active.size > 0) return
    idleWaiters.splice(0).forEach((resolve) => resolve())
  }

  const drain = () => {
    while (active.size < maxConcurrent && queued.size > 0) {
      const job = [...queued.values()].sort((a, b) => a.order - b.order)[0]
      if (!job) return
      queued.delete(job.variableName)
      const controller = schedulerAbortController()
      const entry = {
        controller,
        generation,
      }
      const id = ++runID
      active.set(id, entry)
      Promise.resolve()
        .then(() => job.run({
          signal: controller.signal,
          isCancelled: () => active.get(id) !== entry || generation !== entry.generation,
        }))
        .catch(() => {})
        .finally(() => {
          if (active.get(id) !== entry) return
          active.delete(id)
          drain()
          notifyIdle()
        })
    }
    notifyIdle()
  }

  return {
    schedule: (job = {}) => {
      if (!job.variableName || !job.requestKey || typeof job.run !== 'function') return
      queued.set(job.variableName, {
        ...job,
        order: ++order,
      })
      drain()
    },
    clear: () => {
      generation += 1
      queued.clear()
      active.forEach((entry) => entry.controller?.abort?.())
      active.clear()
      notifyIdle()
    },
    idle: () => (
      queued.size === 0 && active.size === 0
        ? Promise.resolve()
        : new Promise((resolve) => idleWaiters.push(resolve))
    ),
    activeCount: () => active.size,
    queuedCount: () => queued.size,
  }
}

export const isStandardResourceVariable = (definition = {}) => (
  (() => {
    const type = standardResourceVariableType(definition)
    if (!type) return false
    const semantic = semanticVariableByName(type)
    if (!semantic) return false
    const aliases = new Set([semantic.name, ...semantic.aliases])
    return [definition.type, definition.name, ...(definition.labels || [])]
      .some((value) => aliases.has(value || ''))
  })()
)

const resourceRequestExtra = (definition = {}) => {
  const type = standardResourceVariableType(definition) || definition.type || definition.name
  if (type === 'endpoint') return 'endpoint'
  if (type === 'pod' || type === 'instance') return 'instance'
  return undefined
}

export const buildResourceVariableCandidateRequest = ({ variable = {}, groupId, timeRange = {}, now = Date.now() }) => {
  const start = Number(timeRange.startTime)
  const end = Number(timeRange.endTime)
  const window = Number.isFinite(start) && Number.isFinite(end) && end > start
    ? { startTime: start, endTime: end }
    : { startTime: (now - 15 * 60 * 1000) * 1000, endTime: now * 1000 }
  const extra = resourceRequestExtra(variable)
  return {
    ...(groupId !== undefined && groupId !== null ? { groupId } : {}),
    category: 'apm',
    ...(extra ? { extra } : {}),
    startTime: window.startTime,
    endTime: window.endTime,
  }
}

export const buildResourceVariableCandidatePlan = ({
  variables = [],
  selectedValues = {},
  groupId,
  timeRange = {},
  now = Date.now(),
} = {}) => {
  const requestByKey = new Map()
  const variablePlans = variables
    .filter(isStandardResourceVariable)
    .map((variable) => {
      const request = buildResourceVariableCandidateRequest({ variable, groupId, timeRange, now })
      const treeRequestKey = JSON.stringify(request)
      if (!requestByKey.has(treeRequestKey)) {
        requestByKey.set(treeRequestKey, { requestKey: treeRequestKey, request })
      }
      return {
        variable,
        request,
        treeRequestKey,
        candidateRequestKey: JSON.stringify({ request, selectedValues, variableName: variable.name }),
      }
    })
  return {
    treeRequests: Array.from(requestByKey.values()),
    variablePlans,
  }
}

export const buildVariableCandidateRequest = ({
  variable = {},
  queryText = '',
  variables = {},
  enabledVariables = [],
  groupId,
  timeRange = {},
}) => {
  const label = labelsForVariable(variable)[0] || variable.type
  const query = variableRequestQuery(variable, queryText)
  const start = Number(timeRange.startTime)
  const end = Number(timeRange.endTime)
  return {
    type: variable.type === 'label' ? 'label' : variable.type,
    label,
    variableName: variable.name,
    query,
    variables: variableCandidateValuesForQuery({
      variable,
      queryText: query,
      variables,
      enabledVariables,
    }),
    ...(groupId !== undefined && groupId !== null ? { groupId } : {}),
    ...(Number.isFinite(start) && Number.isFinite(end) && end > start ? { startTime: start, endTime: end } : {}),
  }
}

const variableTokenNamesFromText = (text = '') => {
  const names = new Set()
  const tokenPattern = /\$\{([A-Za-z_][A-Za-z0-9_]*)(?::[A-Za-z_][A-Za-z0-9_]*)?\}|\$([A-Za-z_][A-Za-z0-9_]*)\b|\{\{\s*([A-Za-z_][A-Za-z0-9_]*)\s*\}\}/g
  let match
  while ((match = tokenPattern.exec(text)) !== null) {
    const name = match[1] || match[2] || match[3]
    if (name) names.add(name)
  }
  return names
}

const variableCandidateValuesForQuery = ({
  variable = {},
  queryText = '',
  variables = {},
  enabledVariables = [],
} = {}) => {
  const tokenNames = variableTokenNamesFromText(queryText)
  const definitions = enabledVariables.filter((definition) => {
    if (!definition.name || definition.name === variable.name) return false
    if (isStandardResourceVariable(definition)) return true
    return tokenNames.has(definition.name)
  })
  return variableValuesForQuery(variables, definitions)
}

export const VARIABLE_CANDIDATE_CONCURRENCY_LIMIT = 4

export const runLimitedVariableCandidateRequests = async (
  jobs = [],
  { limit = VARIABLE_CANDIDATE_CONCURRENCY_LIMIT } = {},
) => {
  const runnableJobs = jobs
    .map((job) => (typeof job === 'function' ? job : job?.run))
    .filter((job) => typeof job === 'function')
  const maxConcurrent = Math.max(1, Number(limit) || VARIABLE_CANDIDATE_CONCURRENCY_LIMIT)
  let index = 0
  const workers = Array.from({ length: Math.min(maxConcurrent, runnableJobs.length) }, async () => {
    for (;;) {
      const currentIndex = index
      index += 1
      if (currentIndex >= runnableJobs.length) return
      await runnableJobs[currentIndex]()
    }
  })
  await Promise.all(workers)
}

const nodeIdentity = (node = {}) => String(node.id ?? node.name ?? '')

const nodeValue = (node = {}, type = '') => {
  switch (type) {
    case 'cluster':
      return node.clusterId || node.id || node.name
    case 'namespace':
      return node.namespace || node.name
    case 'service':
      return node.service || node.name
    case 'endpoint':
      return node.endpoint || node.name
    case 'pod':
      return node.pod || (node.type === 'pod' ? node.name : '')
    case 'instance':
      return node.node || node.instance || node.instanceName || node.host_name || node.hostName || node.name || node.id
    default:
      return node.name || node.id
  }
}

const selectionMatches = (value, values = []) => {
  if (!values.length || values.some((item) => String(item) === '.*' || String(item) === '0')) return true
  return value !== undefined && value !== null && values.map(String).includes(String(value))
}

const resourceSelectionAliases = {
  cluster: ['cluster', 'cluster_id', 'clusterId', 'clusterID', 'cluster_name'],
  namespace: ['namespace', 'namespace_name', 'namespaceName', 'src_namespace', 'dst_namespace'],
  service: ['service', 'svc', 'svc_name', 'service_name'],
  endpoint: ['endpoint', 'endpoint_name', 'content_key', 'contentKey'],
  pod: ['pod', 'pod_name', 'src_pod', 'dst_pod'],
  instance: ['instance', 'instance_name', 'node_name', 'nodeName', 'host_name', 'hostName'],
}

const resourceUpstreamTypes = {
  cluster: [],
  namespace: ['cluster'],
  service: ['cluster', 'namespace'],
  endpoint: ['cluster', 'namespace', 'service'],
  pod: ['cluster', 'namespace', 'service'],
  instance: ['cluster', 'namespace', 'service'],
}

const selectedResourceValues = (selected = {}, type = '') => (
  (resourceSelectionAliases[type] || [type])
    .flatMap((key) => selected[key] || [])
)

const resourcePathValue = (path = {}, type = '') => {
  switch (type) {
    case 'cluster':
      return path.clusterId
    case 'namespace':
      return path.namespace
    case 'service':
      return path.service
    case 'endpoint':
      return path.endpoint
    case 'pod':
      return path.pod
    case 'instance':
      return path.instance
    default:
      return undefined
  }
}

const resourcePathMatchesSelected = (path = {}, selected = {}, type = '') => (
  (resourceUpstreamTypes[type] || [])
    .every((upstreamType) => selectionMatches(
      resourcePathValue(path, upstreamType),
      selectedResourceValues(selected, upstreamType),
    ))
)

const resourceCandidateMeta = (path = {}, type = '') => {
  const meta = {}
  if (type !== 'cluster' && path.clusterId) meta.clusterId = path.clusterId
  if (!['cluster', 'namespace'].includes(type) && path.namespace) meta.namespace = path.namespace
  if (!['cluster', 'namespace', 'service'].includes(type) && path.service) meta.service = path.service
  return Object.keys(meta).length ? meta : undefined
}

export const resourceVariableCandidatesFromTree = (root = {}, definition = {}, selected = {}) => {
  const type = standardResourceVariableType(definition) || definition.type || definition.name
  const seen = new Map()
  const addCandidate = (node, path) => {
    if (!resourcePathMatchesSelected(path, selected, type)) return
    const value = nodeValue(node, type)
    if (value === undefined || value === null || value === '') return
    const candidate = {
      label: type === 'cluster' ? (node.clusterName || node.name || value) : String(value),
      value: String(value),
      type,
    }
    const meta = resourceCandidateMeta(path, type)
    if (meta) candidate.meta = meta
    seen.set(`${type}:${candidate.value}`, candidate)
  }
  const walk = (node, path = {}) => {
    if (!node) return
    const nextPath = {
      ...path,
      clusterId: node.clusterId || path.clusterId,
      namespace: node.namespace || path.namespace,
      service: node.service || path.service,
    }
    if (node.type === 'endpoint') nextPath.endpoint = nodeValue(node, 'endpoint')
    if (node.type === 'pod') nextPath.pod = nodeValue(node, 'pod')
    if (node.type === 'instance') {
      nextPath.instance = nodeValue(node, 'instance')
      nextPath.pod = node.pod || nextPath.pod
      nextPath.service = node.service || nextPath.service
    }

    if (
      node.type === type
      || (type === 'pod' && node.type === 'instance' && node.pod)
    ) {
      addCandidate(node, nextPath)
    }
    ;[...(node.children || []), ...(node.extraChildren || [])].forEach((child) => walk(child, nextPath))
  }
  walk(root)
  return Array.from(seen.values()).sort((a, b) => a.label.localeCompare(b.label))
}

export const buildPreviewRequest = ({ widget, dashboardId, variables = {}, variableDefinitions = [], timeRange = {}, groupId, seriesLimit, now = Date.now() }) => {
  const mode = widget.query?.mode || 'custom'
  const queries = mode === 'custom'
    ? executableQueryItemsForQuery(widget.query).map((query) => ({
      ...query,
      expression: applyLabelVariableMatchers(query.expression, variables, variableDefinitions),
    }))
    : []
  const firstQuery = queries[0] || {}
  return {
    ...(dashboardId ? { dashboardId } : {}),
    widgetId: widget.id,
    mode,
    presetId: widget.query?.presetId,
    expr: mode === 'custom'
      ? firstQuery.expression || ''
      : widget.query?.expression,
    params: widget.query?.params || {},
    queries,
    variables,
    ...(groupId !== undefined && groupId !== null ? { groupId } : {}),
    chartType: widget.type,
    seriesLimit: normalizeChartSeriesLimit(seriesLimit, chartSeriesLimitForWidget(widget)),
    ...globalTimeToQueryWindow(timeRange, now),
  }
}

export const previewRequestSignature = (request = {}) => JSON.stringify(request)
