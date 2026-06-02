/**
 * Copyright 2026 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import test from 'node:test'
import assert from 'node:assert/strict'
import { existsSync, readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import {
  arrangeWidgetLayouts,
  addWidgetToOpenCell,
  addWidgetPlaceholderCopy,
  addWidgetPlaceholderGridStyle,
  addWidgetPlaceholderMode,
  applyWidgetDraftToDashboard,
  applyLabelVariableMatchers,
  buildPreviewRequest,
  buildResourceVariableCandidateRequest,
  buildResourceVariableCandidatePlan,
  dashboardVariableReadinessScopeKey,
  buildVariableCandidateRequest,
  copyWidgetTitle,
  createDashboardPreviewScheduler,
  createVariableCandidateScheduler,
  createRequestForTemplate,
  cloneWidgetDraft,
  dashboardActions,
  dashboardVariableDefinition,
  dashboardLoadFailureMessage,
  dashboardLabelOptions,
  dashboardVariableOptionGroups,
  draftWidgetGridStyle,
  displayText,
  editorPanelMode,
  enabledVariablesForDashboard,
  extractMetricLabels,
  formatDashboardTimestamp,
  gridPlacement,
  gridLayoutsOverlap,
  globalTimeToQueryWindow,
  insertPromQLMatcherSnippet,
  mergeWidgetGridLayouts,
  mergePresetDefaults,
  moveWidgetToGridPosition,
  normalizeGridLayout,
  needsUnsavedPrompt,
  nextPreviewMode,
  presetOptionsForChart,
  previewStatusLabel,
  previewStatusDetails,
  previewRequestSignature,
  previewStateKind,
  promQLResourceVariableHints,
  reconcileRenderedChartWidgets,
  resourceVariableCandidatesFromTree,
  responsiveGridColumns,
  resizeWidgetGridLayout,
  saveFailureMessage,
  shouldPromptWidgetDraftDiscard,
  presetParameterDescription,
  presetParameterLabel,
  timeRangeRefreshDependency,
  templateDisplayDescription,
  templateDisplayName,
  translatedTimeRangeOptions,
  updateWidgetGridLayout,
  variableCandidatesForDefinition,
  variableCandidateErrorMessage,
  variableCandidateLabel,
  variableCandidatesWithAll,
  variableDeletionState,
  variableDisplayLabel,
  runLimitedVariableCandidateRequests,
  isGlobalPageContextVariable,
  isStandardResourceVariable,
  variableSelectOptions,
  variableRequestQuery,
  variableSelectMode,
  normalizeVariableSelection,
  syncVariableSelectionWithCandidates,
  variableValuesForQuery,
  visibilityLabel,
  widgetDisplayTitle,
  widgetDraftMode,
  widgetPreviewEligibilityForVisibility,
  widgetPreviewVisibilityFromRects,
  VARIABLE_CANDIDATE_CONCURRENCY_LIMIT,
  WIDGET_PREVIEW_CONCURRENCY_LIMIT,
  WIDGET_PREVIEW_PRIORITY,
  runDashboardMutationWithPermissionRefresh,
  shouldRefreshAllWidgetPreviews,
  shouldRefreshWidgetVisibilityOnScroll,
  shouldScheduleWidgetPreview,
  widgetUsesVariable,
} from './behavior.mjs'
import {
  buildChartOption,
  barGaugeProgressPercent,
  chartLegendModes,
  chartLegendRows,
  chartSeriesLimitControlInfo,
  chartSupportsLegendMode,
  chartSupportsTimeBrush,
  chartTypes,
  chartTimeRangeFromBrush,
  formatAxisValue,
  formatWidgetValue,
  normalizeChartSeriesLimit,
  normalizeWidgetQuery,
  normalizeDashboard,
  queryItemsForQuery,
  serializeDashboard,
  seriesDisplayName,
  showNearestLineTooltip,
  tableColumnsForPreview,
  tableRowsForPreview,
} from './utils.mjs'

const currentDir = dirname(fileURLToPath(import.meta.url))
const flushMicrotasks = () => new Promise((resolve) => setTimeout(resolve, 0))

test('custom dashboard root menu has a sidebar icon mapping', () => {
  const navSource = readFileSync(resolve(currentDir, '../../../_nav.js'), 'utf8')

  assert.match(navSource, /customDashboard\s*:/)
})

test('custom dashboard viewer observes global refresh timestamp changes', () => {
  assert.notEqual(
    timeRangeRefreshDependency({ startTime: 1000, endTime: 2000, refreshTimestamp: 3000 }),
    timeRangeRefreshDependency({ startTime: 1000, endTime: 2000, refreshTimestamp: 4000 }),
  )

  const reducerSource = readFileSync(resolve(currentDir, '../../../core/store/reducers/timeRangeReducer.js'), 'utf8')
  assert.match(reducerSource, /refreshTimestamp:\s*timeRangeState\.refreshTimestamp/)

  const viewerSource = readFileSync(resolve(currentDir, 'Viewer.jsx'), 'utf8')
  assert.match(viewerSource, /timeRangeRefreshDependency\(globalTimeRange\)/)
})

test('dashboard viewer converts load failures into user-visible states', () => {
  const t = (key) => `t:${key}`

  assert.equal(
    dashboardLoadFailureMessage(t, { response: { data: { code: 'A0006', message: 'hidden detail' } } }),
    't:noViewPermission',
  )
  assert.equal(
    dashboardLoadFailureMessage(t, { response: { data: { message: 'dashboard is not visible to current user' } } }),
    'dashboard is not visible to current user',
  )
  assert.equal(dashboardLoadFailureMessage(t, null), 't:dashboardNotFound')

  const viewerSource = readFileSync(resolve(currentDir, 'Viewer.jsx'), 'utf8')
  assert.match(viewerSource, /dashboardLoadFailureMessage\(t,\s*error\)/)
  assert.match(viewerSource, /setLoadError/)
  assert.match(viewerSource, /loadError\s*\|\|\s*t\('dashboardNotFound'\)/)
})

test('dashboard viewer omits widget header actions in read-only mode', () => {
  const viewerSource = readFileSync(resolve(currentDir, 'Viewer.jsx'), 'utf8')

  assert.doesNotMatch(viewerSource, /custom-dashboard-widget-action-trigger/)
  assert.doesNotMatch(viewerSource, /LuEllipsisVertical|LuExpand|LuTrash2/)
  assert.doesNotMatch(viewerSource, /key:\s*'fullscreen'|key:\s*'edit'|key:\s*'delete'/)
  assert.doesNotMatch(viewerSource, /edit\?widgetId=/)
  assert.doesNotMatch(viewerSource, /deleteCustomDashboard\(dashboard\.id\)/)
  assert.doesNotMatch(viewerSource, /openFullscreen/)
})

test('editable dashboards remain on the viewer route and expose an edit action', () => {
  const viewerSource = readFileSync(resolve(currentDir, 'Viewer.jsx'), 'utf8')

  assert.doesNotMatch(viewerSource, /if \(data\.canEdit\) \{[\s\S]*navigate\(`\/custom-dashboard\/\$\{data\.id\}\/edit`, \{ replace: true \}\)[\s\S]*return[\s\S]*\}/)
  assert.match(viewerSource, /dashboard\.canEdit[\s\S]*navigate\(`\/custom-dashboard\/\$\{dashboard\.id\}\/edit`\)/)
})

test('dashboard editor converts load failures into user-visible states', () => {
  const editorSource = readFileSync(resolve(currentDir, 'Editor.jsx'), 'utf8')

  assert.match(editorSource, /dashboardLoadFailureMessage\(t,\s*error\)/)
  assert.match(editorSource, /setLoadError/)
  assert.match(editorSource, /loadError\s*\|\|\s*t\('dashboardNotFound'\)/)
})

test('dashboard list actions follow row permissions and create permission', () => {
  const ownerRow = { id: 'dash-1', canEdit: true, canCopy: true, canDelete: true }
  assert.deepEqual(dashboardActions(ownerRow, true), {
    view: true,
    edit: true,
    copy: true,
    delete: true,
  })

  const sharedRow = { id: 'dash-2', canEdit: false, canCopy: true, canDelete: false }
  assert.deepEqual(dashboardActions(sharedRow, false), {
    view: true,
    edit: false,
    copy: false,
    delete: false,
  })
})

test('dashboard list exposes update timestamp but hides last visit timestamp', () => {
  const timestamp = new Date(2026, 0, 2, 3, 4, 5).getTime()
  assert.equal(formatDashboardTimestamp(timestamp), '2026-01-02 03:04:05')
  assert.equal(formatDashboardTimestamp(0), '-')

  const listSource = readFileSync(resolve(currentDir, 'List.jsx'), 'utf8')
  assert.match(listSource, /dataIndex:\s*'updatedAt'/)
  assert.doesNotMatch(listSource, /dataIndex:\s*'lastVisitedAt'/)
  assert.doesNotMatch(listSource, /t\('lastVisit'\)/)
})

test('custom dashboard list route hides global data group and time refresh controls', () => {
  const routesSource = readFileSync(resolve(currentDir, '../../routes.js'), 'utf8')
  const routeBlock = (path) => {
    const start = routesSource.indexOf(`path: '${path}'`)
    assert.notEqual(start, -1)
    const end = routesSource.indexOf('\n  },', start)
    assert.notEqual(end, -1)
    return routesSource.slice(start, end)
  }

  const listBlock = routeBlock('/custom-dashboard')
  assert.match(listBlock, /hideSystemTimeRangePicker:\s*true/)
  assert.doesNotMatch(listBlock, /showDataGroup/)

  for (const path of ['/custom-dashboard/:id', '/custom-dashboard/:id/edit']) {
    const block = routeBlock(path)
    assert.match(block, /showDataGroup:\s*true/)
    assert.doesNotMatch(block, /hideSystemTimeRangePicker/)
  }
})

test('dashboard list relies on breadcrumb instead of repeating the page title', () => {
  const listSource = readFileSync(resolve(currentDir, 'List.jsx'), 'utf8')

  assert.doesNotMatch(listSource, /Typography\.Title/)
  assert.doesNotMatch(listSource, /t\('title'\)/)
})

test('create request distinguishes blank and template creation', () => {
  assert.deepEqual(createRequestForTemplate('blank'), { templateId: 'blank', blank: true })
  assert.deepEqual(createRequestForTemplate('service-health'), { templateId: 'service-health', blank: false })

  const t = (key) => ({
    'templates.service-health.name': '服务健康',
    'templates.service-health.description': '服务健康描述',
  }[key] || key)
  assert.deepEqual(
    createRequestForTemplate('service-health', { id: 'service-health', name: 'Service Health', description: 'Service health dashboard' }, t),
    {
      templateId: 'service-health',
      blank: false,
      name: '服务健康',
      description: '服务健康描述',
    },
  )

  const listSource = readFileSync(resolve(currentDir, 'List.jsx'), 'utf8')
  assert.match(listSource, /createCustomDashboard\(createRequestForTemplate\('blank'\)\)/)
  assert.doesNotMatch(listSource, /getCustomDashboardTemplates/)
  assert.doesNotMatch(listSource, /selectedTemplate/)
  assert.doesNotMatch(listSource, /<Select/)
})

test('dashboard menu mutations refresh user permission menu config', () => {
  const listSource = readFileSync(resolve(currentDir, 'List.jsx'), 'utf8')
  const editorSource = readFileSync(resolve(currentDir, 'Editor.jsx'), 'utf8')
  const userContextSource = readFileSync(resolve(currentDir, '../../../core/contexts/UserContext.js'), 'utf8')

  assert.match(userContextSource, /return getUserPermissionApi\(\{ userId: user\.userId, language: i18n\.language \}\)/)

  assert.match(listSource, /useUserContext/)
  assert.match(listSource, /const \{ getUserPermission \} = useUserContext\(\)/)
  assert.match(listSource, /runDashboardMutationWithPermissionRefresh\([\s\S]*\(\) => createCustomDashboard[\s\S]*getUserPermission[\s\S]*warnPermissionRefreshFailure[\s\S]*navigate\(`\/custom-dashboard\/\$\{dashboard\.id\}\/edit`\)/)
  assert.match(listSource, /runDashboardMutationWithPermissionRefresh\([\s\S]*\(\) => copyCustomDashboard\(id\)[\s\S]*getUserPermission[\s\S]*warnPermissionRefreshFailure[\s\S]*navigate\(`\/custom-dashboard\/\$\{dashboard\.id\}\/edit`\)/)
  assert.match(listSource, /runDashboardMutationWithPermissionRefresh\([\s\S]*\(\) => deleteCustomDashboard\(id\)[\s\S]*getUserPermission[\s\S]*warnPermissionRefreshFailure[\s\S]*await loadData\(\)/)

  assert.match(editorSource, /useUserContext/)
  assert.match(editorSource, /const \{ getUserPermission \} = useUserContext\(\)/)
  assert.match(editorSource, /runDashboardMutationWithPermissionRefresh\([\s\S]*\(\) => updateCustomDashboard[\s\S]*getUserPermission[\s\S]*warnPermissionRefreshFailure[\s\S]*message\.success\(t\('saved'\)\)/)
})

test('dashboard permission menu refresh failures do not fail completed dashboard mutations', async () => {
  const refreshError = new Error('menu refresh unavailable')
  let capturedError = null

  const result = await runDashboardMutationWithPermissionRefresh(
    async () => ({ id: 'dash-1' }),
    async () => { throw refreshError },
    (error) => { capturedError = error },
  )

  assert.deepEqual(result, { id: 'dash-1' })
  assert.equal(capturedError, refreshError)
})

test('editor prompts only when navigating away with unsaved changes', () => {
  assert.equal(needsUnsavedPrompt(true, '/custom-dashboard/dash-1'), true)
  assert.equal(needsUnsavedPrompt(false, '/custom-dashboard/dash-1'), false)
  assert.equal(needsUnsavedPrompt(true, ''), false)
  assert.equal(nextPreviewMode(false), true)
  assert.equal(nextPreviewMode(true), false)
})

test('existing widget edits stay in a draft until the widget is saved', () => {
  const widgets = [{
    id: 'w-1',
    title: 'CPU',
    layout: { x: 0, y: 0, w: 2, h: 1 },
    query: { mode: 'custom', queries: [] },
    display: { decimals: 2 },
  }]
  const draft = cloneWidgetDraft(widgets[0])
  draft.title = 'CPU usage'
  draft.display.decimals = 1

  assert.equal(widgetDraftMode(widgets, draft), 'edit')
  assert.equal(widgets[0].title, 'CPU')
  assert.equal(widgets[0].display.decimals, 2)

  const nextWidgets = applyWidgetDraftToDashboard(widgets, draft, 4)
  assert.equal(nextWidgets[0].title, 'CPU usage')
  assert.equal(nextWidgets[0].display.decimals, 1)
  assert.equal(widgets[0].title, 'CPU')
  assert.equal(widgets[0].display.decimals, 2)
})

test('new widget drafts are added only when the widget is saved', () => {
  const widgets = [{
    id: 'w-1',
    title: 'CPU',
    layout: { x: 0, y: 0, w: 2, h: 1 },
  }]
  const draft = {
    id: 'w-2',
    title: 'Memory',
    layout: { x: 0, y: 0, w: 2, h: 1 },
  }

  assert.equal(widgetDraftMode(widgets, draft), 'new')
  assert.equal(widgets.length, 1)

  const nextWidgets = applyWidgetDraftToDashboard(widgets, draft, 4)
  assert.equal(nextWidgets.length, 2)
  assert.equal(nextWidgets[1].id, 'w-2')
  assert.equal(widgets.length, 1)
})

test('returning from widget edit prompts only for dirty widget drafts', () => {
  assert.equal(shouldPromptWidgetDraftDiscard({ id: 'w-1' }, true), true)
  assert.equal(shouldPromptWidgetDraftDiscard({ id: 'w-1' }, false), false)
  assert.equal(shouldPromptWidgetDraftDiscard(null, true), false)
})

test('editor separates widget draft save from dashboard save', () => {
  const editorSource = readFileSync(resolve(currentDir, 'Editor.jsx'), 'utf8')

  assert.match(editorSource, /cloneWidgetDraft/)
  assert.match(editorSource, /applyWidgetDraftToDashboard/)
  assert.match(editorSource, /shouldPromptWidgetDraftDiscard/)
  assert.match(editorSource, /const \[widgetDraftDirty,\s*setWidgetDraftDirty\] = useState\(false\)/)
  assert.match(editorSource, /const startEditWidget = \(widget\) => \{[\s\S]*setDraftWidget\(cloneWidgetDraft\(widget\)\)/)
  assert.match(editorSource, /const saveWidgetDraft = \(\) => \{[\s\S]*applyWidgetDraftToDashboard\(prev\.widgets,\s*draftWidget,\s*gridColumns\)/)
  assert.match(editorSource, /const requestReturnToDashboard = \(\) => \{[\s\S]*shouldPromptWidgetDraftDiscard\(draftWidget,\s*widgetDraftDirty\)/)
  assert.match(editorSource, /widgetFocusMode \? null : \(/)
  assert.doesNotMatch(editorSource, /widgets: prev\.widgets\.map\(\(widget\) => widget\.id === selectedWidgetID \? \{ \.\.\.widget, \.\.\.patch \} : widget\)/)
})

test('widget edit labels distinguish dashboard and widget save actions', () => {
  const zh = JSON.parse(readFileSync(resolve(currentDir, '../../../../public/locales/zh/pro/customDashboard.json'), 'utf8'))
  const en = JSON.parse(readFileSync(resolve(currentDir, '../../../../public/locales/en/pro/customDashboard.json'), 'utf8'))

  assert.equal(zh.saveDashboard, '保存大盘')
  assert.equal(zh.saveWidget, '保存面板')
  assert.equal(zh.saveWidgetAndReturn, '保存面板并返回')
  assert.equal(en.saveDashboard, 'Save dashboard')
  assert.equal(en.saveWidget, 'Save widget')
  assert.equal(en.saveWidgetAndReturn, 'Save widget and return')
})

test('widget unsaved prompt keeps all footer actions aligned in one row', () => {
  const editorSource = readFileSync(resolve(currentDir, 'Editor.jsx'), 'utf8')
  const styleSource = readFileSync(resolve(currentDir, 'index.scss'), 'utf8')
  const promptBlock = editorSource.slice(
    editorSource.indexOf('const widgetLeavePrompt ='),
    editorSource.indexOf('const renderPromQLHints ='),
  )
  const leaveActionsRule = styleSource.match(/\.custom-dashboard-widget-leave-actions\s*\{[\s\S]*?\n\}/)?.[0] || ''
  const leaveActionButtonRule = styleSource.match(/\.custom-dashboard-widget-leave-actions \.ant-btn\s*\{[\s\S]*?\n\}/)?.[0] || ''

  assert.match(promptBlock, /footer=\{\(\s*<div className="custom-dashboard-widget-leave-actions">/)
  assert.match(leaveActionsRule, /display:\s*flex/)
  assert.match(leaveActionsRule, /align-items:\s*center/)
  assert.match(leaveActionsRule, /justify-content:\s*flex-end/)
  assert.match(leaveActionsRule, /gap:\s*8px/)
  assert.match(leaveActionsRule, /flex-wrap:\s*nowrap/)
  assert.match(leaveActionButtonRule, /flex:\s*0 0 auto/)
})

test('widget draft action row places cancel before save', () => {
  const editorSource = readFileSync(resolve(currentDir, 'Editor.jsx'), 'utf8')
  const settingsPanelStart = editorSource.indexOf('const widgetSettingsPanel =')
  const settingsPanelBlock = editorSource.slice(
    settingsPanelStart,
    editorSource.indexOf('\n  return (', settingsPanelStart),
  )

  assert.match(
    settingsPanelBlock,
    /<Button onClick=\{previewWidget\}>\{t\('preview'\)\}<\/Button>[\s\S]*<Button danger onClick=\{requestReturnToDashboard\}>\{t\('cancel'\)\}<\/Button>[\s\S]*<Button type="primary" onClick=\{confirmAddWidget\}>\{t\('saveWidget'\)\}<\/Button>/,
  )
})

test('widget settings actions stay on their own responsive row', () => {
  const editorSource = readFileSync(resolve(currentDir, 'Editor.jsx'), 'utf8')
  const styleSource = readFileSync(resolve(currentDir, 'index.scss'), 'utf8')
  const settingsPanelStart = editorSource.indexOf('const widgetSettingsPanel =')
  const settingsPanelBlock = editorSource.slice(
    settingsPanelStart,
    editorSource.indexOf('\n  return (', settingsPanelStart),
  )

  assert.match(settingsPanelBlock, /<Space className="custom-dashboard-widget-threshold-row" wrap>/)
  assert.match(settingsPanelBlock, /<Space className="custom-dashboard-widget-action-row" wrap>/)
  assert.match(styleSource, /\.custom-dashboard-widget-threshold-row,[\s\S]*\.custom-dashboard-widget-action-row\s*\{[\s\S]*display:\s*flex;[\s\S]*width:\s*100%;/)
  assert.match(styleSource, /\.custom-dashboard-widget-action-row\s*\{[\s\S]*margin-top:\s*4px;/)
})

test('editor toolbar omits the top-level cancel action', () => {
  const editorSource = readFileSync(resolve(currentDir, 'Editor.jsx'), 'utf8')

  assert.doesNotMatch(editorSource, /cancelEdit/)
  assert.doesNotMatch(editorSource, /onClick=\{cancelEdit\}/)
  assert.doesNotMatch(editorSource, /<Button icon=\{<LuArrowLeft \/>\} onClick=\{cancelEdit\}>/)
})

test('dashboard information modal scrolls inside the modal body', () => {
  const editorSource = readFileSync(resolve(currentDir, 'Editor.jsx'), 'utf8')
  const styleSource = readFileSync(resolve(currentDir, 'index.scss'), 'utf8')

  assert.match(editorSource, /className="custom-dashboard-settings-modal"/)
  assert.match(editorSource, /styles=\{\{[\s\S]*body: \{[\s\S]*maxHeight: 'calc\(100vh - 180px\)'[\s\S]*overflowY: 'auto'/)
  assert.match(styleSource, /\.custom-dashboard-settings-modal \{[\s\S]*top: 24px/)
  assert.match(styleSource, /\.custom-dashboard-settings-modal[\s\S]*\.ant-modal-content \{[\s\S]*max-height: calc\(100vh - 48px\)[\s\S]*display: flex[\s\S]*overflow: hidden/)
  assert.match(styleSource, /\.custom-dashboard-settings-modal[\s\S]*\.ant-modal-body \{[\s\S]*flex: 1 1 auto[\s\S]*min-height: 0[\s\S]*overflow-y: auto/)
})

test('dashboard information modal exposes an explicit save action', () => {
  const editorSource = readFileSync(resolve(currentDir, 'Editor.jsx'), 'utf8')

  assert.match(editorSource, /const saveDashboardPanel = async \(\) => \{[\s\S]*if \(await save\(\)\) setDashboardPanelOpen\(false\)[\s\S]*\}/)
  assert.match(editorSource, /footer=\{\s*<Button[\s\S]*type="primary"[\s\S]*loading=\{saving\}[\s\S]*onClick=\{saveDashboardPanel\}[\s\S]*\{t\('saveDashboard'\)\}[\s\S]*<\/Button>\s*\}/)
})

test('dashboard information modal renders grouped variable options', () => {
  const editorSource = readFileSync(resolve(currentDir, 'Editor.jsx'), 'utf8')
  const zh = JSON.parse(readFileSync(resolve(currentDir, '../../../../public/locales/zh/pro/customDashboard.json'), 'utf8'))
  const en = JSON.parse(readFileSync(resolve(currentDir, '../../../../public/locales/en/pro/customDashboard.json'), 'utf8'))

  assert.match(editorSource, /dashboardVariableOptionGroups/)
  assert.match(editorSource, /variableOptionGroups/)
  assert.match(editorSource, /flatVariableOptions/)
  assert.match(editorSource, /options=\{variableOptionGroups\}/)
  assert.equal(zh.builtinVariableGroup, '内置资源参数')
  assert.equal(zh.queryLabelVariableGroup, '从查询发现的 Label')
  assert.equal(en.builtinVariableGroup, 'Built-in resources')
  assert.equal(en.queryLabelVariableGroup, 'Labels from queries')
})

test('dashboard information modal lets users choose single or multi variable mode', () => {
  const editorSource = readFileSync(resolve(currentDir, 'Editor.jsx'), 'utf8')
  const zh = JSON.parse(readFileSync(resolve(currentDir, '../../../../public/locales/zh/pro/customDashboard.json'), 'utf8'))
  const en = JSON.parse(readFileSync(resolve(currentDir, '../../../../public/locales/en/pro/customDashboard.json'), 'utf8'))

  assert.match(editorSource, /updateVariableMulti/)
  assert.match(editorSource, /checked=\{variable\.multi !== false\}/)
  assert.match(editorSource, /onChange=\{\(multi\) => updateVariableMulti\(variable\.name,\s*multi\)\}/)
  assert.equal(zh.multiSelect, '多选')
  assert.equal(zh.singleSelect, '单选')
  assert.equal(en.multiSelect, 'Multi')
  assert.equal(en.singleSelect, 'Single')
})

test('custom dashboard toolbars omit redundant dashboard refresh actions', () => {
  const listSource = readFileSync(resolve(currentDir, 'List.jsx'), 'utf8')
  const editorSource = readFileSync(resolve(currentDir, 'Editor.jsx'), 'utf8')
  const viewerSource = readFileSync(resolve(currentDir, 'Viewer.jsx'), 'utf8')

  assert.doesNotMatch(listSource, /LuRefreshCw/)
  assert.doesNotMatch(listSource, /<Button icon=\{<LuRefreshCw \/>\}/)
  assert.doesNotMatch(editorSource, /LuRefreshCw/)
  assert.doesNotMatch(editorSource, /refreshDashboardWidgets/)
  assert.doesNotMatch(editorSource, /<Button icon=\{<LuRefreshCw \/>\}/)
  assert.doesNotMatch(viewerSource, /LuRefreshCw/)
  assert.doesNotMatch(viewerSource, /refreshAll/)
  assert.doesNotMatch(viewerSource, /<Button icon=\{<LuRefreshCw \/>\}/)
})

test('custom dashboard parameter changes debounce backend refreshes', () => {
  const editorSource = readFileSync(resolve(currentDir, 'Editor.jsx'), 'utf8')
  const viewerSource = readFileSync(resolve(currentDir, 'Viewer.jsx'), 'utf8')

  ;[editorSource, viewerSource].forEach((source) => {
    assert.match(source, /import \{ useDebounce \} from 'react-use'/)
    assert.match(source, /const \[debouncedVariables,\s*setDebouncedVariables\] = useState\(\{\}\)/)
    assert.match(source, /useDebounce\(\s*\(\) => \{[\s\S]*setDebouncedVariables\(variables\)[\s\S]*\},\s*300,\s*\[variables,\s*variableQueryPaused\]\s*\)/)
    assert.match(source, /setDebouncedVariables\(initialVars\)/)
    assert.match(source, /const widgetVars = useMemo\(\(\) => debouncedVariables,\s*\[debouncedVariables\]\)/)
    assert.match(source, /JSON\.stringify\(debouncedVariables\)/)
    assert.doesNotMatch(source, /\[dashboard\?\.id, queryText, JSON\.stringify\(variables\), enabledVariables, dataGroupId, globalRefreshKey\]/)
  })
})

test('multi-select dashboard parameter queries wait until the dropdown closes', () => {
  const viewerSource = readFileSync(resolve(currentDir, 'Viewer.jsx'), 'utf8')
  const editorSource = readFileSync(resolve(currentDir, 'Editor.jsx'), 'utf8')
  const resourceFilterSource = readFileSync(resolve(currentDir, 'CustomDashboardResourceFilter.jsx'), 'utf8')

  for (const source of [viewerSource, editorSource]) {
    assert.match(source, /openMultiVariableSelectors/)
    assert.match(source, /activeMultiVariableSelectorNames/)
    assert.match(source, /variableQueryPaused/)
    assert.match(source, /enabledVariables\.filter\(\(variable\) => variableSelectMode\(variable\) === 'multiple'\)/)
    assert.match(source, /Object\.keys\(openMultiVariableSelectors\)\.filter\(\(name\) => multiVariableNames\.has\(name\)\)/)
    assert.match(source, /activeMultiVariableSelectorNames\.length > 0/)
    assert.match(source, /useEffect\(\(\) => \{[\s\S]*setOpenMultiVariableSelectors/)
    assert.match(source, /const prevNames = Object\.keys\(prev\)/)
    assert.match(source, /activeMultiVariableSelectorNames\.length === prevNames\.length/)
    assert.match(source, /activeMultiVariableSelectorNames\.every\(\(name\) => prev\[name\]\)/)
    assert.match(source, /\}, \[activeMultiVariableSelectorNames\]\)/)
    assert.match(source, /if \(!variableQueryPaused\) \{[\s\S]*setDebouncedVariables\(variables\)[\s\S]*\}/)
    assert.match(source, /const updateVariableSelectorOpen = useCallback\(\(variable,\s*open\) => \{/)
    assert.match(source, /if \(variableSelectMode\(variable\) !== 'multiple'\) return/)
    assert.match(source, /onDropdownVisibleChange=\{updateVariableSelectorOpen\}/)
    assert.match(source, /onDropdownVisibleChange=\{\(open\) => updateVariableSelectorOpen\(variable,\s*open\)\}/)
  }
  assert.match(resourceFilterSource, /onDropdownVisibleChange/)
  assert.match(resourceFilterSource, /onDropdownVisibleChange=\{\(open\) => onDropdownVisibleChange\?\.\(variable,\s*open\)\}/)
})

test('editor schedules previews for existing widgets when entering edit mode', () => {
  const editorSource = readFileSync(resolve(currentDir, 'Editor.jsx'), 'utf8')
  const viewerSource = readFileSync(resolve(currentDir, 'Viewer.jsx'), 'utf8')
  const rendererSource = readFileSync(resolve(currentDir, 'WidgetRenderer.jsx'), 'utf8')

  for (const source of [editorSource, viewerSource]) {
    assert.match(source, /const shouldRefreshAll = shouldRefreshAllWidgetPreviews\(\{[\s\S]*previews:\s*previewsRef\.current,[\s\S]*changedVariables,[\s\S]*\}\)/)
    assert.doesNotMatch(source, /const shouldRefreshAll = Object\.keys\(widgetPreviewStatesRef\.current\)\.length === 0/)
  }
  assert.match(editorSource, /markWidgetsStale\(widgets\)/)
  assert.match(editorSource, /scheduleEligibleWidgets\(widgets\)/)
  assert.match(editorSource, /timeRangeRefreshDependency\(globalTimeRange\)/)
  assert.match(rendererSource, /if \(!preview\)/)
  assert.doesNotMatch(rendererSource, /if \(!preview\?\.series\?\.length\)\s*\{[\s\S]*<Empty/)
})

test('editor preview relies on the preview endpoint for custom query validation', () => {
  const editorSource = readFileSync(resolve(currentDir, 'Editor.jsx'), 'utf8')

  assert.doesNotMatch(editorSource, /validateCustomDashboardQuery/)
  assert.match(editorSource, /const widgetPreviewRequest = useCallback\(\(widget,\s*variableState = widgetVars\) => \{[\s\S]*const request = buildPreviewRequest\(/)
  assert.match(editorSource, /const previewDashboardWidget = async \(widget,\s*\{ previewKey = widget\.id \} = \{\}\) => \{[\s\S]*const \{ request, requestKey \} = widgetPreviewRequest\(widget,\s*variables\)/)
  assert.match(editorSource, /const response = await previewCustomDashboardQuery\(request\)/)
})

test('editor uses a single context panel instead of tabbed dashboard and widget panes', () => {
  assert.equal(editorPanelMode(null), 'dashboard')
  assert.equal(editorPanelMode({ id: 'w-1' }), 'widget')

  const editorSource = readFileSync(resolve(currentDir, 'Editor.jsx'), 'utf8')
  assert.doesNotMatch(editorSource, /<Tabs/)
  assert.match(editorSource, /editorPanelMode\(selectedWidget\)/)
  assert.match(editorSource, /const \[dashboardPanelOpen,\s*setDashboardPanelOpen\] = useState\(false\)/)
  assert.match(editorSource, /icon=\{<LuSettings \/>} onClick=\{openDashboardPanel\}>\{t\('settings'\)\}<\/Button>/)
  assert.match(editorSource, /setDashboardPanelOpen\(true\)/)
  assert.match(editorSource, /setDashboardPanelOpen\(false\)/)
  assert.match(editorSource, /setSelectedWidgetID\(''\)/)
  assert.match(editorSource, /t\('backToDashboard'\)/)
  assert.match(editorSource, /const dashboardSettingsModal = \(/)
  assert.match(editorSource, /<Modal[\s\S]*open=\{dashboardPanelOpen\}[\s\S]*title=\{t\('dashboard'\)\}/)
  assert.doesNotMatch(editorSource, /const dashboardPanel = !previewing && dashboardPanelOpen \? \(/)
  assert.doesNotMatch(editorSource, /\{dashboardPanel\}/)
  assert.doesNotMatch(editorSource, /<Button icon=\{<LuX \/>\} onClick=\{cancelEdit\}>\{t\('cancel'\)\}<\/Button>/)
  assert.doesNotMatch(editorSource, /<Button icon=\{<LuEye \/>\} onClick=\{\(\) => setPreviewing\(nextPreviewMode\)\}/)
})

test('dashboard variable controls render below the title row', () => {
  const viewerSource = readFileSync(resolve(currentDir, 'Viewer.jsx'), 'utf8')
  const editorSource = readFileSync(resolve(currentDir, 'Editor.jsx'), 'utf8')
  const styleSource = readFileSync(resolve(currentDir, 'index.scss'), 'utf8')

  assert.match(viewerSource, /custom-dashboard-toolbar-main/)
  assert.match(viewerSource, /custom-dashboard-toolbar-parameters/)
  assert.match(viewerSource, /custom-dashboard-toolbar-parameters[\s\S]*CustomDashboardResourceFilter[\s\S]*customLabelVariables\.map/)
  assert.doesNotMatch(viewerSource, /<Space wrap>\s*<Typography\.Title[\s\S]*enabledVariables\.map/)
  assert.match(editorSource, /custom-dashboard-toolbar-parameters/)
  assert.match(editorSource, /custom-dashboard-toolbar-parameters[\s\S]*CustomDashboardResourceFilter[\s\S]*customLabelVariables\.map/)
  assert.match(editorSource, /const \[variables,\s*setVariables\] = useState\(\{\}\)/)
  assert.match(editorSource, /variableValuesForQuery\(variableState,\s*enabledVariables\)/)
  assert.match(styleSource, /\.custom-dashboard-toolbar-main/)
  assert.match(styleSource, /\.custom-dashboard-toolbar-parameters/)
  assert.match(styleSource, /\.custom-dashboard-viewer \.custom-dashboard-toolbar,\s*\n\.custom-dashboard-editor \.custom-dashboard-toolbar[\s\S]*flex-direction:\s*column/)
})

test('dashboard toolbar parameter labels render to the left of filter selects', () => {
  const editorSource = readFileSync(resolve(currentDir, 'Editor.jsx'), 'utf8')
  const viewerSource = readFileSync(resolve(currentDir, 'Viewer.jsx'), 'utf8')
  const styleSource = readFileSync(resolve(currentDir, 'index.scss'), 'utf8')

  assert.match(editorSource, /className="custom-dashboard-toolbar-parameter"/)
  assert.match(editorSource, /custom-dashboard-toolbar-parameter-label[\s\S]*\{variableDisplayLabel\(t,\s*variable\)\}[\s\S]*<Select/)
  assert.match(editorSource, /placeholder=\{variable\.name\}/)
  assert.match(viewerSource, /className="custom-dashboard-toolbar-parameter"/)
  assert.match(viewerSource, /custom-dashboard-toolbar-parameter-label[\s\S]*\{variableDisplayLabel\(t,\s*variable\)\}[\s\S]*<Select/)
  assert.match(viewerSource, /placeholder=\{variable\.name\}/)
  assert.match(styleSource, /\.custom-dashboard-toolbar-parameter\s*\{[\s\S]*display:\s*inline-flex/)
  assert.match(styleSource, /\.custom-dashboard-toolbar-parameter-label/)
})

test('viewer widget edit deep-links into one widget without card-click editing', () => {
  const editorSource = readFileSync(resolve(currentDir, 'Editor.jsx'), 'utf8')

  assert.match(editorSource, /useSearchParams/)
  assert.match(editorSource, /const widgetIDFromSearch = searchParams\.get\('widgetId'\) \|\| ''/)
  assert.match(editorSource, /data\.widgets\?\.find\(\(widget\) => widget\.id === widgetIDFromSearch\)/)
  assert.match(editorSource, /setSelectedWidgetID\(widgetIDFromSearch\)/)
  assert.match(editorSource, /setDraftWidget\(cloneWidgetDraft\(widgetFromSearch\)\)/)
  assert.doesNotMatch(editorSource, /onClick=\{\(\) => \{\s*setDraftWidget\(null\)\s*setSelectedWidgetID\(widget\.id\)\s*\}\}/)
})

test('editor uses a split full-page focus layout while editing a widget', () => {
  const editorSource = readFileSync(resolve(currentDir, 'Editor.jsx'), 'utf8')
  const styleSource = readFileSync(resolve(currentDir, 'index.scss'), 'utf8')

  assert.match(editorSource, /import \{[\s\S]*Splitter[\s\S]*\} from 'antd'/)
  assert.match(editorSource, /const widgetFocusMode = !previewing && panelMode === 'widget'/)
  assert.match(editorSource, /const visibleWidgets = widgetFocusMode/)
  assert.match(editorSource, /\? \(draftWidget \? \[\] : dashboard\.widgets\.filter/)
  assert.match(editorSource, /dashboard\.widgets\.filter\(\(widget\) => widget\.id === selectedWidgetID\)/)
  assert.match(editorSource, /className=\{`custom-dashboard-editor-body \$\{widgetFocusMode \? 'widget-focus' : ''\}`\}/)
  assert.match(editorSource, /<Splitter[\s\S]*className="custom-dashboard-widget-focus-splitter"/)
  assert.match(editorSource, /const widgetFocusLayoutStorageKey = 'customDashboard\.widgetFocusLayout'/)
  assert.match(editorSource, /const \[widgetFocusLayout,\s*setWidgetFocusLayout\] = useState\(\(\) => readWidgetFocusLayout\(\)\)/)
  assert.match(editorSource, /writeWidgetFocusLayout\(nextLayout\)/)
  assert.match(editorSource, /onResize=\{\(sizes\) => updateWidgetFocusLayout\(\{ settingsWidth: Math\.round\(sizes\[1\]\) \}\)\}/)
  assert.match(editorSource, /onResize=\{\(sizes\) => updateWidgetFocusLayout\(\{ previewHeight: Math\.round\(sizes\[0\]\) \}\)\}/)
  assert.match(editorSource, /<Splitter\.Panel[\s\S]*className="custom-dashboard-widget-focus-main-panel"[\s\S]*min=\{420\}/)
  assert.match(editorSource, /<Splitter[\s\S]*layout="vertical"[\s\S]*className="custom-dashboard-widget-focus-main"/)
  assert.match(editorSource, /<Splitter\.Panel[\s\S]*className="custom-dashboard-widget-focus-preview"[\s\S]*min=\{280\}[\s\S]*size=\{widgetFocusLayout\.previewHeight\}/)
  assert.match(editorSource, /className="custom-dashboard-widget-focus-preview"[\s\S]*\{widgetGrid\}/)
  assert.match(editorSource, /<Splitter\.Panel[\s\S]*className="custom-dashboard-widget-query-panel"[\s\S]*min=\{200\}[\s\S]*<div className="custom-dashboard-widget-query-panel-inner">[\s\S]*\{widgetQueryPanel\}/)
  assert.match(editorSource, /<Splitter\.Panel[\s\S]*className="custom-dashboard-panel custom-dashboard-widget-settings-panel"[\s\S]*min=\{280\}[\s\S]*max=\{560\}[\s\S]*size=\{widgetFocusLayout\.settingsWidth\}/)
  assert.match(editorSource, /custom-dashboard-widget-focus-main/)
  assert.match(editorSource, /custom-dashboard-widget-focus-preview/)
  assert.match(editorSource, /custom-dashboard-widget-query-panel/)
  assert.match(editorSource, /custom-dashboard-widget-settings-panel/)
  assert.match(editorSource, /const widgetGrid = widgetFocusMode \? \(/)
  assert.match(editorSource, /custom-dashboard-focus-grid/)
  assert.match(editorSource, /<DashboardGridLayout/)
  assert.match(editorSource, /isDraggable=\{!previewing && !widgetFocusMode\}/)
  assert.match(editorSource, /isResizable=\{!previewing && !widgetFocusMode\}/)
  assert.match(editorSource, /layout=\{dashboardGridLayout\}/)
  assert.doesNotMatch(editorSource, /gridPlacement/)

  assert.match(styleSource, /\.custom-dashboard-editor-body\.widget-focus/)
  assert.match(styleSource, /\.custom-dashboard-editor-body\.widget-focus[\s\S]*display:\s*block/)
  assert.doesNotMatch(styleSource, /\.custom-dashboard-editor-body\.widget-focus[\s\S]*max-width:\s*960px/)
  assert.match(styleSource, /\.custom-dashboard-widget-focus-splitter[\s\S]*height:\s*100%/)
  assert.match(styleSource, /\.custom-dashboard-widget-focus-main[\s\S]*height:\s*100%/)
  assert.match(styleSource, /\.custom-dashboard-widget-query-panel[\s\S]*overflow:\s*auto/)
  assert.match(styleSource, /\.custom-dashboard-widget-settings-panel[\s\S]*overflow:\s*auto/)

  const previewPaneRule = styleSource.match(/\.custom-dashboard-widget-focus-preview\s*\{[\s\S]*?\n\}/)?.[0] || ''
  const queryPaneRule = styleSource.match(/\.custom-dashboard-widget-query-panel\s*\{[\s\S]*?\n\}/)?.[0] || ''
  const settingsPaneRule = styleSource.match(/\.custom-dashboard-panel\.custom-dashboard-widget-settings-panel\s*\{[\s\S]*?\n\}/)?.[0] || ''
  for (const paneRule of [previewPaneRule, settingsPaneRule]) {
    assert.match(paneRule, /padding:\s*12px/)
    assert.match(paneRule, /box-sizing:\s*border-box/)
  }
  assert.match(queryPaneRule, /padding:\s*12px/)
  assert.match(queryPaneRule, /box-sizing:\s*border-box/)
  assert.match(styleSource, /\.custom-dashboard-widget-query-panel-inner[\s\S]*padding:\s*14px 16px 16px/)

  const queryPanelBlock = editorSource.slice(
    editorSource.indexOf('const widgetQueryPanel ='),
    editorSource.indexOf('const widgetSettingsPanel ='),
  )
  const settingsPanelStart = editorSource.indexOf('const widgetSettingsPanel =')
  const settingsPanelBlock = editorSource.slice(
    settingsPanelStart,
    editorSource.indexOf('\n  return (', settingsPanelStart),
  )
  assert.doesNotMatch(queryPanelBlock, /<Button onClick=\{previewWidget\}>\{t\('preview'\)\}<\/Button>/)
  assert.match(settingsPanelBlock, /<Button onClick=\{previewWidget\}>\{t\('preview'\)\}<\/Button>/)
})

test('editor body uses a single dashboard area outside widget focus mode', () => {
  const styleSource = readFileSync(resolve(currentDir, 'index.scss'), 'utf8')
  const editorBodyRule = styleSource.match(/\.custom-dashboard-editor-body\s*\{[\s\S]*?\n\}/)?.[0] || ''

  assert.match(editorBodyRule, /display:\s*flex/)
  assert.match(editorBodyRule, /flex-direction:\s*column/)
  assert.doesNotMatch(editorBodyRule, /display:\s*grid/)
  assert.doesNotMatch(editorBodyRule, /grid-template-columns/)
  assert.doesNotMatch(styleSource, /@media \(max-width: 1100px\) \{[\s\S]*\.custom-dashboard-editor-body\s*\{[\s\S]*grid-template-columns/)
})

test('dashboard edit widget cards expose drag resize handles and action menu', () => {
  const editorSource = readFileSync(resolve(currentDir, 'Editor.jsx'), 'utf8')
  const styleSource = readFileSync(resolve(currentDir, 'index.scss'), 'utf8')
  const packageSource = readFileSync(resolve(currentDir, '../../../../package.json'), 'utf8')

  assert.match(packageSource, /"react-grid-layout":/)
  assert.match(editorSource, /from 'react-grid-layout\/legacy'/)
  assert.match(editorSource, /react-grid-layout\/css\/styles\.css/)
  assert.match(editorSource, /WidthProvider\(GridLayout\)/)
  assert.match(editorSource, /<DashboardGridLayout/)
  assert.match(editorSource, /draggableHandle="\.custom-dashboard-widget-drag-region"/)
  assert.match(editorSource, /onDragStop=\{syncWidgetGridLayouts\}/)
  assert.match(editorSource, /onResizeStop=\{syncWidgetGridLayouts\}/)
  assert.match(editorSource, /compactType=\{null\}/)
  assert.match(editorSource, /preventCollision=\{false\}/)
  assert.match(editorSource, /Dropdown/)
  assert.match(editorSource, /LuEllipsisVertical/)
  assert.match(editorSource, /LuGripVertical/)
  assert.match(editorSource, /menu=\{\{[\s\S]*key:\s*'edit'[\s\S]*key:\s*'delete'/)
  assert.match(editorSource, /setSelectedWidgetID\(widget\.id\)/)
  assert.match(editorSource, /deleteWidget\(widget\.id\)/)
  assert.doesNotMatch(editorSource, /startResizeWidget/)
  assert.doesNotMatch(editorSource, /draggable=\{!previewing && !widgetFocusMode\}/)

  assert.match(styleSource, /\.custom-dashboard-editor-body:not\(\.widget-focus\) \.custom-dashboard-widget-header[\s\S]*border-bottom:\s*0/)
  assert.match(styleSource, /\.custom-dashboard-widget-action-trigger/)
  assert.match(styleSource, /\.custom-dashboard-widget-drag-handle/)
  assert.match(styleSource, /\.custom-dashboard-grid-layout/)
  assert.match(styleSource, /\.react-grid-placeholder/)
  assert.match(styleSource, /\.react-resizable-handle/)
})

test('editor adds widgets from the canvas with an empty and append placeholder', () => {
  assert.equal(addWidgetPlaceholderMode([], false), 'empty')
  assert.equal(addWidgetPlaceholderMode([{ id: 'w-1' }], false), 'append')
  assert.equal(addWidgetPlaceholderMode([], true), 'hidden')
  assert.deepEqual(addWidgetPlaceholderGridStyle([], 4, false), {
    gridColumn: '1 / -1',
    gridRow: '1 / span 1',
  })
  assert.deepEqual(addWidgetPlaceholderGridStyle([{ id: 'w-1' }, { id: 'w-2' }, { id: 'w-3' }], 4, false), {
    gridColumn: '3 / span 2',
    gridRow: '2 / span 1',
  })

  const editorSource = readFileSync(resolve(currentDir, 'Editor.jsx'), 'utf8')
  assert.match(editorSource, /custom-dashboard-add-widget/)
  assert.match(editorSource, /startAddWidget\('line'\)/)
  assert.match(editorSource, /addWidgetGridItemID/)
  assert.match(editorSource, /findOpenGridLayout/)
  assert.match(editorSource, /const placeholderInGrid = placeholderMode === 'empty'/)
  assert.match(editorSource, /const placeholderAction = placeholderMode === 'hidden' \? null : \(/)
  assert.match(editorSource, /className=\{`custom-dashboard-add-widget \$\{placeholderMode\}`\}/)
  assert.doesNotMatch(editorSource, /\.\.\.\(!widgetFocusMode && placeholderLayout \?/)
  assert.match(editorSource, /confirmAddWidget/)
  assert.doesNotMatch(editorSource, /custom-dashboard-widget-palette/)
  const styleSource = readFileSync(resolve(currentDir, 'index.scss'), 'utf8')
  assert.match(styleSource, /\.custom-dashboard-add-widget\.append[\s\S]*min-height:\s*96px/)
  assert.match(styleSource, /\.custom-dashboard-grid-layout[\s\S]*flex:\s*0 0 auto/)
})

test('editor add-widget placeholder has state copy and a clear action', () => {
  const t = (key) => `t:${key}`

  assert.deepEqual(addWidgetPlaceholderCopy('empty', t), {
    title: 't:emptyCanvasTitle',
    description: 't:emptyCanvasHint',
    action: 't:addWidget',
  })
  assert.deepEqual(addWidgetPlaceholderCopy('append', t), {
    title: 't:addWidgetPlaceholderTitle',
    description: 't:addWidgetPlaceholderHint',
    action: 't:addWidget',
  })

  const editorSource = readFileSync(resolve(currentDir, 'Editor.jsx'), 'utf8')
  assert.match(editorSource, /addWidgetPlaceholderCopy\(placeholderMode,\s*t\)/)
  assert.match(editorSource, /custom-dashboard-add-widget-description/)
  assert.match(editorSource, /placeholderCopy\.action/)
})

test('draft widget appears on the canvas before it is added to dashboard widgets', () => {
  assert.deepEqual(draftWidgetGridStyle([], 4), {
    gridColumn: '1 / span 2',
    gridRow: '1 / span 1',
  })
  assert.deepEqual(draftWidgetGridStyle([{ id: 'w-1' }, { id: 'w-2' }, { id: 'w-3' }], 4), {
    gridColumn: '3 / span 2',
    gridRow: '2 / span 1',
  })

  const editorSource = readFileSync(resolve(currentDir, 'Editor.jsx'), 'utf8')
  const styleSource = readFileSync(resolve(currentDir, 'index.scss'), 'utf8')
  assert.match(editorSource, /draftWidget && renderWidgetCard\(draftWidget, \{ draft: true \}\)/)
  assert.match(editorSource, /selectedWidgetID === widget\.id \|\| draft \? 'selected' : ''/)
  assert.match(editorSource, /draft \? 'draft' : ''/)
  assert.match(editorSource, /<WidgetRenderer[\s\S]*widget=\{widget\}/)
  assert.doesNotMatch(styleSource, /\.custom-dashboard-widget\.draft[\s\S]*border-style:\s*dashed/)
})

test('draft widget missing preview does not stay in loading state', () => {
  const editorSource = readFileSync(resolve(currentDir, 'Editor.jsx'), 'utf8')
  const rendererSource = readFileSync(resolve(currentDir, 'WidgetRenderer.jsx'), 'utf8')

  assert.match(rendererSource, /loadingOnMissingPreview = true/)
  assert.match(rendererSource, /if \(!preview\)\s*\{[\s\S]*if \(loadingOnMissingPreview\)[\s\S]*<Spin/)
  assert.match(editorSource, /draftWidget && renderWidgetCard\(draftWidget, \{ draft: true \}\)/)
  assert.match(editorSource, /loadingOnMissingPreview=\{!draft\}/)
})

test('new widget keeps an empty preview placeholder after it is added to the canvas', () => {
  const editorSource = readFileSync(resolve(currentDir, 'Editor.jsx'), 'utf8')

  assert.match(editorSource, /emptyWidgetPreview/)
  assert.match(editorSource, /applyPreview\(\(prev\) => \(\{ \.\.\.prev, \[widget\.id\]: emptyWidgetPreview \}\)\)/)
})

test('copy widget title uses localized suffix', () => {
  assert.equal(copyWidgetTitle('Latency', (key) => key === 'copySuffix' ? '副本' : key), 'Latency - 副本')

  const editorSource = readFileSync(resolve(currentDir, 'Editor.jsx'), 'utf8')
  assert.match(editorSource, /copyWidgetTitle\(selectedWidget\.title,\s*t\)/)
})

test('query preview states preserve expected dashboard status categories', () => {
  assert.equal(previewStateKind({ status: 'success', series: [{ values: [] }] }), 'success')
  assert.equal(previewStateKind({ status: 'success', series: [] }), 'no_data')
  assert.equal(previewStateKind({ status: 'syntax_error' }), 'syntax_error')
  assert.equal(previewStateKind({ status: 'timeout' }), 'timeout')
  assert.equal(previewStateKind({ status: 'too_large' }), 'too_large')
  assert.equal(previewStateKind({ status: 'no_permission' }), 'no_permission')
  assert.equal(previewStateKind({ status: 'metric_missing' }), 'metric_missing')
  assert.equal(previewStateKind({ status: 'error' }), 'query_failure')
})

test('query preview status details include operator suggestions', () => {
  const t = (key) => `t:${key}`

  assert.equal(previewStatusLabel(t, 'syntax_error'), 't:statusSyntaxError')
  assert.equal(previewStatusLabel(t, 'no_data'), 't:statusNoData')
  assert.equal(previewStatusLabel(t, 'unknown'), 'unknown')

  assert.deepEqual(previewStatusDetails('syntax_error'), {
    color: 'error',
    suggestionKey: 'syntaxErrorSuggestion',
  })
  assert.deepEqual(previewStatusDetails('timeout'), {
    color: 'error',
    suggestionKey: 'timeoutSuggestion',
  })
  assert.deepEqual(previewStatusDetails('no_data'), {
    color: 'default',
    suggestionKey: 'noDataSuggestion',
  })

  const rendererSource = readFileSync(resolve(currentDir, 'WidgetRenderer.jsx'), 'utf8')
  assert.match(rendererSource, /previewStatusDetails\(status\)/)
  assert.match(rendererSource, /previewStatusLabel\(t,\s*status\)/)
  assert.match(rendererSource, /details\.suggestionKey/)
})

test('widget empty states use a plain Grafana-style panel state', () => {
  const rendererSource = readFileSync(resolve(currentDir, 'WidgetRenderer.jsx'), 'utf8')
  const styleSource = readFileSync(resolve(currentDir, 'index.scss'), 'utf8')

  assert.match(rendererSource, /renderWidgetState/)
  assert.match(rendererSource, /custom-dashboard-state-empty/)
  assert.match(rendererSource, /custom-dashboard-state-error/)
  assert.doesNotMatch(rendererSource, /<Empty/)
  assert.doesNotMatch(rendererSource, /Empty\.PRESENTED_IMAGE_SIMPLE/)
  assert.match(styleSource, /\.custom-dashboard-state\.custom-dashboard-state-empty/)
  assert.match(styleSource, /\.custom-dashboard-state-title/)
  assert.match(styleSource, /font-weight:\s*600/)
})

test('viewer time range controls drive query windows and payloads', () => {
  const now = Date.UTC(2026, 0, 1, 0, 0, 0)
  assert.deepEqual(globalTimeToQueryWindow({}, now), {
    startTime: now - 15 * 60 * 1000,
    endTime: now,
    step: 60,
  })
  assert.deepEqual(globalTimeToQueryWindow({ startTime: 1_000_000_000_000_000, endTime: 1_027_000_000_000_000 }), {
    startTime: 1_000_000_000_000,
    endTime: 1_027_000_000_000,
    step: 300,
  })

  const widget = {
    id: 'w-1',
    type: 'line',
    query: { mode: 'custom', expression: 'up', params: { service: '$service' } },
  }
	  assert.deepEqual(buildPreviewRequest({
	    widget,
	    dashboardId: 'dashboard-1',
	    variables: { service: 'checkout' },
	    groupId: 7,
	    timeRange: { startTime: now - 15 * 60 * 1000, endTime: now },
	    now,
	  }), {
	    dashboardId: 'dashboard-1',
	    widgetId: 'w-1',
	    mode: 'custom',
    presetId: undefined,
    expr: 'up',
    params: { service: '$service' },
    queries: [{ id: 'q-1', expression: 'up', legend: '' }],
    variables: { service: 'checkout' },
    groupId: 7,
    chartType: 'line',
    seriesLimit: 50,
    startTime: now - 15 * 60 * 1000,
    endTime: now,
    step: 60,
  })
})

test('preview request signatures distinguish selected data group context', () => {
  const widget = {
    id: 'w-1',
    type: 'line',
    query: { mode: 'custom', expression: 'up' },
  }

  const first = buildPreviewRequest({ widget, groupId: 7 })
  const second = buildPreviewRequest({ widget, groupId: 8 })

  assert.equal(first.groupId, 7)
  assert.notEqual(previewRequestSignature(first), previewRequestSignature(second))
})

test('preview requests ask the backend to cap chart series by default', () => {
  const request = buildPreviewRequest({
    widget: {
      id: 'latency',
      type: 'line',
      query: { mode: 'custom', expression: 'up' },
    },
  })

  assert.equal(request.seriesLimit, 50)
})

test('bar charts default to a lower series cap to reduce draw cost', () => {
  const barRequest = buildPreviewRequest({
    widget: {
      id: 'disk',
      type: 'bar',
      query: { mode: 'custom', expression: 'up' },
    },
  })
  const series = Array.from({ length: 60 }, (_, index) => ({
    metric: { instance: `node-${index + 1}` },
    values: [{ TimeStamp: 1000, Value: index + 1 }],
  }))
  const barOption = buildChartOption(
    { type: 'bar', display: {}, query: { legend: '{{instance}}' } },
    { series },
  )
  const lineOption = buildChartOption(
    { type: 'line', display: {}, query: { legend: '{{instance}}' } },
    { series },
  )

  assert.equal(barRequest.seriesLimit, 20)
  assert.equal(barOption.series.length, 20)
  assert.equal(lineOption.series.length, 50)
})

test('preview requests support per-widget chart series limit overrides', () => {
  const base = buildPreviewRequest({
    widget: {
      id: 'latency',
      type: 'line',
      query: { mode: 'custom', expression: 'up' },
    },
  })
  const expanded = buildPreviewRequest({
    widget: {
      id: 'latency',
      type: 'line',
      query: { mode: 'custom', expression: 'up' },
    },
    seriesLimit: 200,
  })

  assert.equal(expanded.seriesLimit, 200)
  assert.notEqual(previewRequestSignature(base), previewRequestSignature(expanded))
})

test('chart series limits are bounded to avoid full high-cardinality renders', () => {
  assert.equal(normalizeChartSeriesLimit(0), 50)
  assert.equal(normalizeChartSeriesLimit(200), 200)
  assert.equal(normalizeChartSeriesLimit(1000), 500)
})

test('chart series limit controls stay visible after expanding above the default cap', () => {
  const highCardinalityPreview = {
    status: 'success',
    series: Array.from({ length: 60 }, (_, index) => ({
      metric: { svc_name: `svc-${index + 1}` },
      values: [{ TimeStamp: 1000, Value: index + 1 }],
    })),
  }
  const expandedWidget = {
    type: 'line',
    display: { seriesLimit: 100 },
    query: { legend: '{{svc_name}}' },
  }

  assert.deepEqual(chartSeriesLimitControlInfo(expandedWidget, highCardinalityPreview), {
    limited: false,
    total: 60,
    shown: 60,
    limit: 100,
  })

  assert.equal(
    chartSeriesLimitControlInfo(expandedWidget, {
      status: 'success',
      series: highCardinalityPreview.series.slice(0, 40),
    }),
    null,
  )
})

test('variable candidate requests use POST body shape and only carry relevant upstream variables', () => {
  const variable = { name: 'service', type: 'service' }
  const customVariable = { name: 'job_name', type: 'label', labels: ['job_name'] }

  assert.equal(isStandardResourceVariable(variable), true)
  assert.equal(isStandardResourceVariable(customVariable), false)
  assert.equal(isStandardResourceVariable({ name: 'cluster_id', type: 'cluster', labels: ['cluster_id'] }), true)
  assert.deepEqual(
    buildResourceVariableCandidateRequest({
      variable,
      groupId: 7,
      timeRange: { startTime: 1000, endTime: 2000 },
    }),
    {
      groupId: 7,
      category: 'apm',
      startTime: 1000,
      endTime: 2000,
    },
  )
  assert.deepEqual(
    buildResourceVariableCandidateRequest({
      variable: { name: 'endpoint', type: 'endpoint' },
      groupId: 7,
      timeRange: { startTime: 1000, endTime: 2000 },
    }),
    {
      groupId: 7,
      category: 'apm',
      extra: 'endpoint',
      startTime: 1000,
      endTime: 2000,
    },
  )
  assert.deepEqual(
    buildVariableCandidateRequest({
      variable: customVariable,
      queryText: 'label_values(up{cluster_id=~"$cluster",namespace="$namespace"}, job_name)',
      variables: { cluster: 'dev', namespace: 'prod', resolution: '1m', job_name: 'old' },
      enabledVariables: [
        { name: 'cluster', type: 'cluster' },
        { name: 'namespace', type: 'namespace' },
        { name: 'resolution', type: 'custom' },
        customVariable,
      ],
      groupId: 7,
      timeRange: { startTime: 1000, endTime: 2000 },
    }),
    {
      type: 'label',
      label: 'job_name',
      variableName: 'job_name',
      query: 'label_values(up{cluster_id=~"$cluster",namespace="$namespace"}, job_name)',
      variables: { cluster: ['dev'], namespace: ['prod'] },
      groupId: 7,
      startTime: 1000,
      endTime: 2000,
    },
  )
})

test('imported semantic label variables use standard resource candidate flow', () => {
  const namespaceVariable = {
    name: 'namespace',
    type: 'label',
    labels: ['namespace'],
    query: 'query_result(group by (namespace) (container_memory_rss))',
  }
  const serviceVariable = {
    name: 'service_name',
    type: 'label',
    labels: ['service_name'],
    query: 'query_result(group by (svc_name) (kindling_span_trace_duration_nanoseconds_count))',
  }
  const podVariable = {
    name: 'pod',
    type: 'label',
    labels: ['pod'],
    query: 'query_result(group by (pod) (container_memory_rss))',
  }

  assert.equal(isStandardResourceVariable(namespaceVariable), true)
  assert.equal(isStandardResourceVariable(serviceVariable), true)
  assert.equal(isStandardResourceVariable(podVariable), true)
  assert.deepEqual(
    buildResourceVariableCandidateRequest({
      variable: namespaceVariable,
      groupId: 7,
      timeRange: { startTime: 1000, endTime: 2000 },
    }),
    {
      groupId: 7,
      category: 'apm',
      startTime: 1000,
      endTime: 2000,
    },
  )

  const plan = buildResourceVariableCandidatePlan({
    variables: [namespaceVariable, serviceVariable, podVariable],
    selectedValues: { namespace: ['prod'] },
    groupId: 7,
    timeRange: { startTime: 1000, endTime: 2000 },
  })
  assert.deepEqual(plan.treeRequests.map((item) => item.request), [
    { groupId: 7, category: 'apm', startTime: 1000, endTime: 2000 },
    { groupId: 7, category: 'apm', extra: 'instance', startTime: 1000, endTime: 2000 },
  ])

  const tree = {
    id: 'root',
    name: 'root',
    children: [
      {
        id: 'cluster-a',
        name: 'Production Cluster',
        type: 'cluster',
        clusterId: 'cluster-a',
        children: [
          {
            id: 'cluster-a#prod',
            name: 'prod',
            type: 'namespace',
            clusterId: 'cluster-a',
            namespace: 'prod',
            children: [
              {
                id: 'cluster-a#prod#checkout',
                name: 'checkout',
                type: 'service',
                clusterId: 'cluster-a',
                namespace: 'prod',
                service: 'checkout',
                extraChildren: [
                  { id: 'pod-1', name: 'checkout@node-a@123', type: 'instance', pod: 'checkout-0', node: 'node-a', pid: 123, service: 'checkout' },
                ],
              },
            ],
          },
        ],
      },
    ],
  }
  assert.deepEqual(resourceVariableCandidatesFromTree(tree, namespaceVariable), [
    { label: 'prod', value: 'prod', type: 'namespace', meta: { clusterId: 'cluster-a' } },
  ])
  assert.deepEqual(resourceVariableCandidatesFromTree(tree, serviceVariable, { namespace: ['prod'] }), [
    { label: 'checkout', value: 'checkout', type: 'service', meta: { clusterId: 'cluster-a', namespace: 'prod' } },
  ])
  assert.deepEqual(resourceVariableCandidatesFromTree(tree, podVariable, { service_name: ['checkout'] }), [
    { label: 'checkout-0', value: 'checkout-0', type: 'pod', meta: { clusterId: 'cluster-a', namespace: 'prod', service: 'checkout' } },
  ])
})

test('custom dashboard variable option groups expose builtin resources before query labels', () => {
  const groups = dashboardVariableOptionGroups([
    {
      query: {
        mode: 'custom',
        expression: 'up{namespace="prod",service_name="checkout",job_name="integrations/unix"}',
      },
    },
  ], [], [
    { name: 'namespace', type: 'label', labels: ['namespace'] },
  ])

  assert.deepEqual(groups.map((group) => group.labelKey), [
    'builtinVariableGroup',
    'queryLabelVariableGroup',
  ])
  assert.deepEqual(groups[0].options.map((option) => option.value), [
    'cluster',
    'service',
    'endpoint',
    'pod',
    'instance',
  ])
  assert.deepEqual(groups[1].options, [
    { label: 'job_name', value: 'job_name', type: 'label', labels: ['job_name'] },
  ])
})

test('custom dashboard builtin resource filter is dedicated to custom dashboards', () => {
  const filterPath = resolve(currentDir, 'CustomDashboardResourceFilter.jsx')
  const editorSource = readFileSync(resolve(currentDir, 'Editor.jsx'), 'utf8')
  const viewerSource = readFileSync(resolve(currentDir, 'Viewer.jsx'), 'utf8')

  assert.equal(existsSync(filterPath), true)
  const filterSource = readFileSync(filterPath, 'utf8')
  assert.match(filterSource, /CustomDashboardResourceFilter/)
  assert.match(filterSource, /maxTagCount=\{2\}/)
  assert.match(filterSource, /variableSelectOptions/)
  assert.doesNotMatch(filterSource, /DataSourceFilter/)
  assert.match(editorSource, /CustomDashboardResourceFilter/)
  assert.match(viewerSource, /CustomDashboardResourceFilter/)
})

test('empty dashboards still offer builtin resource variables', () => {
  const groups = dashboardVariableOptionGroups([], [], [])

  assert.deepEqual(groups, [
    {
      labelKey: 'builtinVariableGroup',
      options: [
        { label: 'Cluster', value: 'cluster', type: 'cluster', labels: ['cluster', 'cluster_id', 'clusterId', 'clusterID', 'cluster_name'] },
        { label: 'Namespace', value: 'namespace', type: 'namespace', labels: ['namespace', 'namespace_name', 'namespaceName', 'src_namespace', 'dst_namespace'] },
        { label: 'Service', value: 'service', type: 'service', labels: ['svc', 'svc_name', 'service', 'service_name'] },
        { label: 'Endpoint', value: 'endpoint', type: 'endpoint', labels: ['endpoint', 'endpoint_name', 'content_key', 'contentKey'] },
        { label: 'Pod', value: 'pod', type: 'pod', labels: ['pod', 'pod_name', 'src_pod', 'dst_pod'] },
        { label: 'Instance', value: 'instance', type: 'instance', labels: ['instance', 'instance_name', 'node_name', 'nodeName', 'host_name', 'hostName'] },
      ],
    },
  ])
})

test('custom label variables are not treated as builtin resources', () => {
  const customVariables = [
    { name: 'related_nodes', type: 'label', labels: ['node_name'] },
    { name: 'related_pid', type: 'label', labels: ['pid'] },
    { name: 'jvm_instance', type: 'label', labels: ['instance'] },
    { name: 'job_name', type: 'label', labels: ['job_name'] },
    { name: 'cadvisor_job_name', type: 'label', labels: ['job_name'] },
  ]

  customVariables.forEach((variable) => {
    assert.equal(isStandardResourceVariable(variable), false, variable.name)
  })

  assert.deepEqual(buildResourceVariableCandidatePlan({ variables: customVariables }).variablePlans, [])
  assert.deepEqual(
    buildVariableCandidateRequest({
      variable: customVariables[0],
      queryText: 'label_values(container_memory_rss{pod=~"$pod"}, node_name)',
      variables: { pod: ['checkout-0'], related_nodes: ['node-a'] },
      enabledVariables: [{ name: 'pod', type: 'pod' }, customVariables[0]],
    }),
    {
      type: 'label',
      label: 'node_name',
      variableName: 'related_nodes',
      query: 'label_values(container_memory_rss{pod=~"$pod"}, node_name)',
      variables: { pod: ['checkout-0'] },
    },
  )
})

test('query-discovered custom labels are deduplicated by enabled variable aliases', () => {
  const groups = dashboardVariableOptionGroups([
    {
      query: {
        mode: 'custom',
        expression: 'node_cpu_seconds_total{job_name="integrations/unix",environment="prod"}',
      },
    },
  ], [], [
    { name: 'cadvisor_job_name', type: 'label', labels: ['job_name'] },
  ])

  assert.deepEqual(groups[1].options, [
    { label: 'environment', value: 'environment', type: 'label', labels: ['environment'] },
  ])
})

test('builtin resources default to All while custom labels preserve explicit All only', () => {
  ;['cluster', 'namespace', 'service', 'endpoint', 'pod', 'instance'].forEach((name) => {
    const definition = dashboardVariableDefinition(name, { value: name, type: name })
    assert.equal(definition.includeAll, true, name)
    assert.equal(definition.allValue, '.*', name)
  })

  const custom = dashboardVariableDefinition('job_name', {
    label: 'job_name',
    value: 'job_name',
    type: 'label',
    labels: ['job_name'],
  })
  assert.equal(custom.multi, true)
  assert.equal(custom.includeAll, undefined)
  assert.equal(custom.allValue, undefined)

  const importedCustom = dashboardVariableDefinition('job_name', {
    label: 'job_name',
    value: 'job_name',
    type: 'label',
    labels: ['job_name'],
    includeAll: true,
    allValue: '$__all',
  })
  assert.equal(importedCustom.includeAll, true)
  assert.equal(importedCustom.allValue, '$__all')
})

test('custom variable candidate jobs are concurrency limited', async () => {
  let active = 0
  let maxActive = 0
  const completed = []
  const jobs = Array.from({ length: VARIABLE_CANDIDATE_CONCURRENCY_LIMIT + 3 }, (_, index) => async () => {
    active += 1
    maxActive = Math.max(maxActive, active)
    await new Promise((resolve) => {
      setTimeout(resolve, 1)
    })
    completed.push(index)
    active -= 1
  })

  await runLimitedVariableCandidateRequests(jobs, { limit: VARIABLE_CANDIDATE_CONCURRENCY_LIMIT })

  assert.equal(completed.length, jobs.length)
  assert.ok(maxActive <= VARIABLE_CANDIDATE_CONCURRENCY_LIMIT)
})

test('custom variable candidate scheduler limits concurrency across schedule batches', async () => {
  const scheduler = createVariableCandidateScheduler({ limit: VARIABLE_CANDIDATE_CONCURRENCY_LIMIT })
  let active = 0
  let maxActive = 0
  const completed = []
  const total = VARIABLE_CANDIDATE_CONCURRENCY_LIMIT * 2 + 3

  const makeJob = (index) => ({
    variableName: `var-${index}`,
    requestKey: `request-${index}`,
    run: async () => {
      active += 1
      maxActive = Math.max(maxActive, active)
      await new Promise((resolve) => {
        setTimeout(resolve, 2)
      })
      completed.push(index)
      active -= 1
    },
  })

  Array.from({ length: VARIABLE_CANDIDATE_CONCURRENCY_LIMIT + 1 }, (_, index) => makeJob(index))
    .forEach((job) => scheduler.schedule(job))
  Array.from({ length: total - VARIABLE_CANDIDATE_CONCURRENCY_LIMIT - 1 }, (_, offset) => makeJob(offset + VARIABLE_CANDIDATE_CONCURRENCY_LIMIT + 1))
    .forEach((job) => scheduler.schedule(job))

  await scheduler.idle()

  assert.equal(completed.length, total)
  assert.ok(maxActive <= VARIABLE_CANDIDATE_CONCURRENCY_LIMIT)
})

test('custom variable candidate scheduler clears queued work and aborts active work', async () => {
  const scheduler = createVariableCandidateScheduler({ limit: 1 })
  const signals = []
  let queuedRan = false

  scheduler.schedule({
    variableName: 'first',
    requestKey: 'first',
    run: ({ signal }) => new Promise(() => {
      signals.push(signal)
    }),
  })
  scheduler.schedule({
    variableName: 'second',
    requestKey: 'second',
    run: () => {
      queuedRan = true
    },
  })
  await flushMicrotasks()

  scheduler.clear()
  await flushMicrotasks()

  assert.equal(signals[0]?.aborted, true)
  assert.equal(queuedRan, false)
  assert.equal(scheduler.activeCount(), 0)
  assert.equal(scheduler.queuedCount(), 0)
})

test('standard resource variable candidate requests are grouped by resource tree request', () => {
  const variables = [
    { name: 'cluster', type: 'cluster' },
    { name: 'namespace', type: 'namespace' },
    { name: 'service', type: 'service' },
    { name: 'endpoint', type: 'endpoint' },
    { name: 'pod', type: 'pod' },
    { name: 'instance', type: 'instance' },
    { name: 'job_name', type: 'label' },
  ]
  const first = buildResourceVariableCandidatePlan({
    variables,
    selectedValues: { namespace: ['prod'] },
    groupId: 7,
    timeRange: { startTime: 1000, endTime: 2000 },
  })
  const second = buildResourceVariableCandidatePlan({
    variables,
    selectedValues: { namespace: ['staging'] },
    groupId: 7,
    timeRange: { startTime: 1000, endTime: 2000 },
  })

  assert.deepEqual(first.treeRequests.map((item) => item.request), [
    { groupId: 7, category: 'apm', startTime: 1000, endTime: 2000 },
    { groupId: 7, category: 'apm', extra: 'endpoint', startTime: 1000, endTime: 2000 },
    { groupId: 7, category: 'apm', extra: 'instance', startTime: 1000, endTime: 2000 },
  ])
  assert.deepEqual(first.variablePlans.map((item) => item.variable.name), [
    'cluster',
    'namespace',
    'service',
    'endpoint',
    'pod',
    'instance',
  ])
  assert.deepEqual(first.variablePlans.map((item) => item.treeRequestKey), [
    first.treeRequests[0].requestKey,
    first.treeRequests[0].requestKey,
    first.treeRequests[0].requestKey,
    first.treeRequests[1].requestKey,
    first.treeRequests[2].requestKey,
    first.treeRequests[2].requestKey,
  ])
  assert.deepEqual(
    first.treeRequests.map((item) => item.requestKey),
    second.treeRequests.map((item) => item.requestKey),
  )
  assert.notDeepEqual(
    first.variablePlans.map((item) => item.candidateRequestKey),
    second.variablePlans.map((item) => item.candidateRequestKey),
  )
})

test('resource filter tree adapts into cascaded dashboard variable candidates', () => {
  const tree = {
    type: '',
    children: [
      {
        id: 'cluster-scope',
        name: 'Production Cluster',
        type: 'cluster',
        clusterId: 'cluster-a',
        children: [
          {
            id: 'cluster-a#prod',
            name: 'prod',
            type: 'namespace',
            clusterId: 'cluster-a',
            namespace: 'prod',
            children: [
              {
                id: 'cluster-a#prod#checkout',
                name: 'checkout',
                type: 'service',
                clusterId: 'cluster-a',
                namespace: 'prod',
                service: 'checkout',
                extraChildren: [
                  { id: 'endpoint-1', name: '/pay', type: 'endpoint', endpoint: '/pay', service: 'checkout' },
                  { id: 'pod-1', name: 'checkout@node-a@123', type: 'instance', pod: 'checkout-0', node: 'node-a', pid: 123, service: 'checkout' },
                ],
              },
            ],
          },
        ],
      },
    ],
  }

  assert.deepEqual(resourceVariableCandidatesFromTree(tree, { name: 'cluster', type: 'cluster' }), [
    { label: 'Production Cluster', value: 'cluster-a', type: 'cluster' },
  ])
  assert.deepEqual(resourceVariableCandidatesFromTree(tree, { name: 'service', type: 'service' }, { namespace: ['prod'] }), [
    { label: 'checkout', value: 'checkout', type: 'service', meta: { clusterId: 'cluster-a', namespace: 'prod' } },
  ])
  assert.deepEqual(resourceVariableCandidatesFromTree(tree, { name: 'service', type: 'service' }, { namespace: ['staging'] }), [])
  assert.deepEqual(resourceVariableCandidatesFromTree(tree, { name: 'service', type: 'service' }, { namespace_name: ['prod'] }), [
    { label: 'checkout', value: 'checkout', type: 'service', meta: { clusterId: 'cluster-a', namespace: 'prod' } },
  ])
  assert.deepEqual(resourceVariableCandidatesFromTree(tree, { name: 'service', type: 'service' }, { dst_namespace: ['staging'] }), [])
  assert.deepEqual(resourceVariableCandidatesFromTree(tree, { name: 'namespace', type: 'namespace' }, { cluster_name: ['cluster-a'] }), [
    { label: 'prod', value: 'prod', type: 'namespace', meta: { clusterId: 'cluster-a' } },
  ])
  assert.deepEqual(resourceVariableCandidatesFromTree(tree, { name: 'namespace', type: 'namespace' }, { service: ['checkout'] }), [
    { label: 'prod', value: 'prod', type: 'namespace', meta: { clusterId: 'cluster-a' } },
  ])
  assert.deepEqual(resourceVariableCandidatesFromTree(tree, { name: 'endpoint', type: 'endpoint' }, { service: ['checkout'] }), [
    { label: '/pay', value: '/pay', type: 'endpoint', meta: { clusterId: 'cluster-a', namespace: 'prod', service: 'checkout' } },
  ])
  assert.deepEqual(resourceVariableCandidatesFromTree(tree, { name: 'endpoint', type: 'endpoint' }, { service_name: ['billing'] }), [])
  assert.deepEqual(resourceVariableCandidatesFromTree(tree, { name: 'pod', type: 'pod' }, { service: ['checkout'] }), [
    { label: 'checkout-0', value: 'checkout-0', type: 'pod', meta: { clusterId: 'cluster-a', namespace: 'prod', service: 'checkout' } },
  ])
  assert.deepEqual(resourceVariableCandidatesFromTree(tree, { name: 'instance', type: 'instance' }, { service_name: ['checkout'] }), [
    { label: 'node-a', value: 'node-a', type: 'instance', meta: { clusterId: 'cluster-a', namespace: 'prod', service: 'checkout' } },
  ])
})

test('preview request supports multiple custom query expressions', () => {
  const now = Date.UTC(2026, 0, 1, 0, 0, 0)
  const request = buildPreviewRequest({
    widget: {
      id: 'w-multi',
      type: 'line',
      query: {
        mode: 'custom',
        queries: [
          { id: 'q-a', expression: 'up{svc_name="old"}', legend: 'Availability' },
          { id: 'q-b', expression: 'rate(http_requests_total{svc="old"}[5m])', legend: 'Requests' },
        ],
      },
    },
    variables: { service: ['checkout'] },
    variableDefinitions: [{ name: 'service', type: 'service' }],
    timeRange: { startTime: now - 15 * 60 * 1000, endTime: now },
    now,
  })

  assert.deepEqual(request.queries.map((query) => query.expression), [
    'up{svc_name=~"checkout"}',
    'rate(http_requests_total{svc=~"checkout"}[5m])',
  ])
  assert.deepEqual(request.queries.map((query) => query.legend), ['Availability', 'Requests'])
  assert.equal(request.expr, 'up{svc_name=~"checkout"}')
})

test('hidden custom query items are preserved during frontend normalization', () => {
  const normalized = normalizeWidgetQuery({
    mode: 'custom',
    queries: [
      { id: 'q-a', expression: 'up', legend: 'Availability', hidden: true },
      { id: 'q-b', expression: 'rate(http_requests_total[5m])', legend: 'Requests' },
    ],
  })

  assert.equal(queryItemsForQuery(normalized)[0].hidden, true)
  assert.equal(queryItemsForQuery(normalized)[1].hidden, undefined)
  assert.equal(normalized.queries[0].hidden, true)
})

test('preview request ignores hidden custom query items', () => {
  const request = buildPreviewRequest({
    widget: {
      id: 'w-hidden-query',
      type: 'line',
      query: {
        mode: 'custom',
        queries: [
          { id: 'q-a', expression: 'up{job="api"}', legend: 'API', hidden: true },
          { id: 'q-b', expression: 'up{job="worker"}', legend: 'Worker' },
        ],
      },
    },
  })

  assert.deepEqual(request.queries, [
    { id: 'q-b', expression: 'up{job="worker"}', legend: 'Worker' },
  ])
  assert.equal(request.expr, 'up{job="worker"}')
})

test('preview request allows all custom query items to be hidden', () => {
  const request = buildPreviewRequest({
    widget: {
      id: 'w-all-hidden-query',
      type: 'line',
      query: {
        mode: 'custom',
        queries: [
          { id: 'q-a', expression: 'up{job="api"}', legend: 'API', hidden: true },
          { id: 'q-b', expression: 'up{job="worker"}', legend: 'Worker', hidden: true },
        ],
      },
    },
  })

  assert.deepEqual(request.queries, [])
  assert.equal(request.expr, '')

  const editorSource = readFileSync(resolve(currentDir, 'Editor.jsx'), 'utf8')
  assert.doesNotMatch(editorSource, /keepOneQueryVisible|hideDisabled|visibleCount\s*<=\s*1/)
})

test('grafana-style static and label variables feed viewer query requests', () => {
  const staticVariable = {
    name: 'resolution',
    label: 'Resolution',
    type: 'custom',
    options: [
      { label: '1m', value: '1m' },
      { label: '5m', value: '5m' },
    ],
  }
  const labelVariable = {
    name: 'job_name',
    label: 'Node Job',
    type: 'label',
    labels: ['job_name'],
    query: 'label_values(node_cpu_seconds_total{cluster_id=~"$cluster"},job_name)',
  }

  assert.deepEqual(variableCandidatesForDefinition(staticVariable), [
    { label: '1m', value: '1m', type: 'custom' },
    { label: '5m', value: '5m', type: 'custom' },
  ])
  assert.equal(variableSelectMode(staticVariable), 'multiple')
  assert.equal(variableSelectMode({ ...staticVariable, multi: false }), undefined)
  assert.equal(variableSelectMode(labelVariable), 'multiple')
  assert.equal(variableSelectMode({ ...labelVariable, multi: false }), undefined)
  assert.equal(variableSelectMode({ ...labelVariable, multi: true }), 'multiple')
  assert.equal(variableRequestQuery(labelVariable, 'dashboard query'), labelVariable.query)

  const widget = {
    id: 'w-grafana',
    type: 'line',
    query: {
      mode: 'custom',
      expression: 'sum(rate(node_cpu_seconds_total{job_name="$job_name",cluster_id=~"$cluster"}[$resolution]))',
    },
  }
  assert.deepEqual(
    enabledVariablesForDashboard([staticVariable, labelVariable], [widget]),
    [staticVariable, labelVariable],
  )
  assert.deepEqual(variableValuesForQuery({ resolution: '1m', job_name: 'integrations/unix' }, [staticVariable, labelVariable]), {
    resolution: ['1m'],
    job_name: ['integrations/unix'],
  })
})

test('grafana includeAll variables expose all candidates and keep all matcher raw', () => {
  const clusterVariable = {
    name: 'cluster_id',
    label: 'Cluster',
    type: 'label',
    labels: ['cluster_id'],
    includeAll: true,
    allValue: '.*',
    options: [
      { label: 'dev', value: 'dev' },
      { label: 'prod', value: 'prod' },
    ],
  }

  assert.deepEqual(variableCandidatesForDefinition(clusterVariable), [
    { label: 'All', value: '.*', type: 'label', meta: { all: 'true' } },
    { label: 'dev', value: 'dev', type: 'label' },
    { label: 'prod', value: 'prod', type: 'label' },
  ])
  assert.deepEqual(variableCandidatesWithAll({ ...clusterVariable, options: undefined }, [{ label: 'dev', value: 'dev', type: 'label' }]), [
    { label: 'All', value: '.*', type: 'label', meta: { all: 'true' } },
    { label: 'dev', value: 'dev', type: 'label' },
  ])

  const expression = applyLabelVariableMatchers(
    'up{cluster_id="dev",svc_name="checkout"}',
    { cluster_id: '.*' },
    [clusterVariable],
  )
  assert.equal(expression, 'up{cluster_id=~".*",svc_name="checkout"}')
})

test('include-all variable selections are mutually exclusive and rendered as separate option groups', () => {
  const t = (key) => `t:${key}`
  const variable = {
    name: 'cluster_id',
    type: 'label',
    includeAll: true,
    allValue: '.*',
    multi: true,
  }
  const candidates = [
    { label: 'All', value: '.*', type: 'label', meta: { all: 'true' } },
    { label: 'dev', value: 'dev', type: 'label' },
    { label: 'prod', value: 'prod', type: 'label' },
  ]

  assert.deepEqual(normalizeVariableSelection(variable, ['dev', '.*'], ['dev']), ['.*'])
  assert.deepEqual(normalizeVariableSelection(variable, ['.*', 'prod'], ['.*']), ['prod'])
  assert.deepEqual(normalizeVariableSelection(variable, ['dev', 'prod'], ['dev']), ['dev', 'prod'])
  assert.deepEqual(normalizeVariableSelection(variable, [], ['dev'], candidates), ['.*'])
  assert.equal(normalizeVariableSelection({ ...variable, multi: false }, undefined, 'dev', candidates), '.*')
  assert.deepEqual(syncVariableSelectionWithCandidates(variable, [], candidates), ['.*'])
  assert.deepEqual(variableSelectOptions(t, variable, candidates), [
    { label: 't:allOptionGroup', options: [{ label: 'All', value: '.*' }] },
    { label: 't:availableOptionGroup', options: [{ label: 'dev', value: 'dev' }, { label: 'prod', value: 'prod' }] },
  ])
})

test('builtin resource variables include All first and default to it', () => {
  const namespaceVariable = dashboardVariableDefinition(
    'namespace',
    { label: 'Namespace', value: 'namespace', type: 'namespace', labels: ['namespace'] },
    (key) => key,
  )
  const candidates = variableCandidatesWithAll(namespaceVariable, [
    { label: 'prod', value: 'prod', type: 'namespace' },
    { label: 'staging', value: 'staging', type: 'namespace' },
  ])

  assert.equal(namespaceVariable.includeAll, true)
  assert.equal(namespaceVariable.allValue, '.*')
  assert.deepEqual(candidates.map((candidate) => candidate.value), ['.*', 'prod', 'staging'])
  assert.deepEqual(syncVariableSelectionWithCandidates(namespaceVariable, [], candidates), ['.*'])

  const serviceVariable = dashboardVariableDefinition(
    'service',
    { label: 'Service', value: 'service', type: 'service', labels: ['svc_name'] },
    (key) => key,
  )
  assert.equal(serviceVariable.includeAll, true)
  assert.equal(serviceVariable.allValue, '.*')
  assert.equal(serviceVariable.multi, true)
})

test('variable selections are pruned when linked candidate options refresh', () => {
  const variable = { name: 'service', type: 'service', multi: true }
  const candidates = [
    { label: 'checkout', value: 'checkout', type: 'service' },
    { label: 'cart', value: 'cart', type: 'service' },
  ]

  assert.deepEqual(
    syncVariableSelectionWithCandidates(variable, ['checkout', 'payment'], candidates),
    ['checkout'],
  )
  assert.equal(
    syncVariableSelectionWithCandidates({ ...variable, multi: false }, 'payment', candidates),
    'checkout',
  )
  assert.deepEqual(syncVariableSelectionWithCandidates(variable, [], candidates), ['checkout'])
  assert.deepEqual(syncVariableSelectionWithCandidates(variable, ['payment'], candidates), ['checkout'])
  assert.equal(syncVariableSelectionWithCandidates({ ...variable, multi: false }, undefined, candidates), 'checkout')
})

test('resource defaults outside page group are pruned to All', () => {
  const namespaceVariable = { name: 'namespace', type: 'namespace', multi: false, includeAll: true, allValue: '.*' }
  const candidates = variableCandidatesWithAll(namespaceVariable, [
    { label: 'train-ticket', value: 'train-ticket', type: 'namespace' },
  ])

  assert.equal(syncVariableSelectionWithCandidates(namespaceVariable, 'apisix', candidates), '.*')
})

test('viewer and editor normalize variable selections and sync linked candidate lists', () => {
  const viewerSource = readFileSync(resolve(currentDir, 'Viewer.jsx'), 'utf8')
  const editorSource = readFileSync(resolve(currentDir, 'Editor.jsx'), 'utf8')

  for (const source of [viewerSource, editorSource]) {
    assert.match(source, /candidateRequestKeysRef = useRef\(\{\}\)/)
    assert.match(source, /normalizeVariableSelection\(variable,\s*value,\s*prev\[variable\.name\]/)
    assert.match(source, /syncVariableSelectionWithCandidates\(variable,\s*prev\[variable\.name\]/)
    assert.match(source, /variableSelectOptions\(t,\s*variable,\s*candidates\[variable\.name\]\s*\|\|\s*\[\]\)/)
  }
})

test('viewer and editor cache grouped standard resource filter tree requests', () => {
  const viewerSource = readFileSync(resolve(currentDir, 'Viewer.jsx'), 'utf8')
  const editorSource = readFileSync(resolve(currentDir, 'Editor.jsx'), 'utf8')

  for (const source of [viewerSource, editorSource]) {
    assert.match(source, /buildResourceVariableCandidatePlan/)
    assert.match(source, /resourceTreeCacheRef = useRef\(\{\}\)/)
    assert.match(source, /resourceTreeScopeKeyRef = useRef\(''\)/)
    assert.match(source, /const resourceVariables = \[\]/)
    assert.match(source, /Promise\.all\(treeRequests\.map/)
    assert.match(source, /resourceTreeCacheRef\.current\[requestKey\]/)
    assert.match(source, /getDatasourceFilterApiV2\(request\)/)
  }
})

test('viewer and editor defer previews until page group resource variables are ready', () => {
  const viewerSource = readFileSync(resolve(currentDir, 'Viewer.jsx'), 'utf8')
  const editorSource = readFileSync(resolve(currentDir, 'Editor.jsx'), 'utf8')

  for (const source of [viewerSource, editorSource]) {
    assert.match(source, /dashboardVariablesReady/)
    assert.match(source, /dashboardId:\s*dashboard\?\.id/)
    assert.match(source, /status:\s*'deferred'/)
    assert.match(source, /setDashboardVariablesReadyKey\(dashboardVariableScopeKey\)/)
  }
})

test('viewer and editor wait for all variable candidates before previewing', () => {
  assert.notEqual(
    dashboardVariableReadinessScopeKey({
      dashboardId: 'dash-1',
      dataGroupId: 7,
      globalRefreshKey: '1000:2000:1',
      queryText: 'up',
      variables: [{ name: 'service', type: 'service' }],
    }),
    dashboardVariableReadinessScopeKey({
      dashboardId: 'dash-1',
      dataGroupId: 7,
      globalRefreshKey: '1000:2000:1',
      queryText: 'up{job_name="$job_name"}',
      variables: [{ name: 'service', type: 'service' }, { name: 'job_name', type: 'label', query: 'label_values(up,job_name)' }],
    }),
  )

  const viewerSource = readFileSync(resolve(currentDir, 'Viewer.jsx'), 'utf8')
  const editorSource = readFileSync(resolve(currentDir, 'Editor.jsx'), 'utf8')

  for (const source of [viewerSource, editorSource]) {
    assert.match(source, /dashboardVariableReadinessScopeKey/)
    assert.match(source, /pendingCandidateRequestKeys/)
    assert.match(source, /markVariableCandidateReady/)
    assert.match(source, /applyVariableCandidateError/)
  }
  assert.match(editorSource, /const previewDashboardWidget = async \(widget,\s*\{ previewKey = widget\.id \} = \{\}\) => \{[\s\S]*if \(dataGroupId === null \|\| dataGroupId === undefined \|\| !dashboardVariablesReady\)/)
})

test('grafana braced regex variables feed preview requests without label matcher overwrite', () => {
  const request = buildPreviewRequest({
    widget: {
      id: 'w-regex',
      type: 'line',
      query: {
        mode: 'custom',
        expression: 'sum(rate(http_requests_total{cluster_id=~"${cluster_id:regex}",job_name="$job_name"}[$resolution]))',
      },
    },
    variables: {
      cluster_id: ['dev', 'prod'],
      job_name: ['integrations/unix'],
      resolution: ['1m'],
    },
    variableDefinitions: [
      { name: 'cluster_id', type: 'label', labels: ['cluster_id'], multi: true },
      { name: 'job_name', type: 'label', labels: ['job_name'] },
      { name: 'resolution', type: 'custom' },
    ],
  })

  assert.equal(
    request.expr,
    'sum(rate(http_requests_total{cluster_id=~"${cluster_id:regex}",job_name="$job_name"}[$resolution]))',
  )
})

test('explicit grafana variable tokens are not overwritten by variables sharing the same label', () => {
  const request = buildPreviewRequest({
    widget: {
      id: 'w-node',
      type: 'line',
      query: {
        mode: 'custom',
        expression: 'sum(rate(node_cpu_seconds_total{job_name="$job_name"}[$resolution]))',
      },
    },
    variables: {
      job_name: ['integrations/unix'],
      cadvisor_job_name: ['integrations/kubernetes/cadvisor'],
      resolution: ['1m'],
    },
    variableDefinitions: [
      { name: 'job_name', type: 'label', labels: ['job_name'] },
      { name: 'cadvisor_job_name', type: 'label', labels: ['job_name'] },
      { name: 'resolution', type: 'custom' },
    ],
  })

  assert.equal(
    request.expr,
    'sum(rate(node_cpu_seconds_total{job_name="$job_name"}[$resolution]))',
  )
})

test('preview request signatures distinguish selected pod variable values', () => {
  const widget = {
    id: 'grafana-6',
    type: 'area',
    query: {
      mode: 'custom',
      expression: 'sum(rate(container_cpu_usage_seconds_total{namespace="$namespace",pod=~"${pod:regex}"}[$resolution]))',
    },
  }
  const base = buildPreviewRequest({
    widget,
    variables: { namespace: ['apo'], pod: ['.*'], resolution: ['1m'] },
    timeRange: { startTime: 1778125183000, endTime: 1778126083000 },
  })
  const selected = buildPreviewRequest({
    widget,
    variables: { namespace: ['apo'], pod: ['apo-alertmanager-858d59b696-n67ns'], resolution: ['1m'] },
    timeRange: { startTime: 1778125183000, endTime: 1778126083000 },
  })

  assert.notEqual(previewRequestSignature(base), previewRequestSignature(selected))
})

test('viewer and editor ignore stale preview responses after variables change', () => {
  const viewerSource = readFileSync(resolve(currentDir, 'Viewer.jsx'), 'utf8')
  const editorSource = readFileSync(resolve(currentDir, 'Editor.jsx'), 'utf8')

  assert.match(viewerSource, /previewRequestKeysRef = useRef\(\{\}\)/)
  assert.match(viewerSource, /previewRequestSignature\(request\)/)
  assert.match(viewerSource, /previewRequestKeysRef\.current\[widget\.id\] !== requestKey/)
  assert.match(editorSource, /previewRequestKeysRef = useRef\(\{\}\)/)
  assert.match(editorSource, /previewRequestSignature\(request\)/)
  assert.match(editorSource, /previewRequestKeysRef\.current\[previewKey\] !== requestKey/)
})

test('viewer dispatches widget previews through viewport scheduling instead of eager refresh', () => {
  const viewerSource = readFileSync(resolve(currentDir, 'Viewer.jsx'), 'utf8')

  assert.match(viewerSource, /createDashboardPreviewScheduler/)
  assert.match(viewerSource, /IntersectionObserver/)
  assert.match(viewerSource, /widgetPreviewPreloadMarginPx = 320/)
  assert.match(viewerSource, /const root = gridRef\.current/)
  assert.match(viewerSource, /root\.getBoundingClientRect/)
  assert.match(viewerSource, /rootMargin:\s*`\$\{widgetPreviewPreloadMarginPx\}px 0px`/)
  assert.match(viewerSource, /root,\s*\n\s*rootMargin/)
  assert.match(viewerSource, /seedWidgetVisibility/)
  assert.match(viewerSource, /widgetPreviewVisibilityFromRects/)
  assert.match(viewerSource, /widgetVisibilityRef/)
  assert.match(viewerSource, /widgetPreviewEligibilityForVisibility/)
  assert.match(viewerSource, /WIDGET_PREVIEW_PRIORITY/)
  assert.doesNotMatch(viewerSource, /Promise\.all\(widgets\.map\(\(widget\) => refreshWidget\(widget\)\)\)/)
})

test('editor dispatches dashboard widget previews through viewport scheduling instead of eager refresh', () => {
  const editorSource = readFileSync(resolve(currentDir, 'Editor.jsx'), 'utf8')

  assert.match(editorSource, /createDashboardPreviewScheduler/)
  assert.match(editorSource, /IntersectionObserver/)
  assert.match(editorSource, /widgetPreviewPreloadMarginPx = 320/)
  assert.match(editorSource, /const root = editorBodyRef\.current/)
  assert.match(editorSource, /root\.getBoundingClientRect/)
  assert.match(editorSource, /root,\s*\n\s*rootMargin/)
  assert.match(editorSource, /registerWidgetElement/)
  assert.match(editorSource, /findWidgetElement/)
  assert.match(editorSource, /data-widget-id=\{draft \? undefined : widget\.id\}/)
  assert.match(editorSource, /widgetPreviewEligibilityForVisibility/)
  assert.match(editorSource, /markWidgetsStale/)
  assert.doesNotMatch(editorSource, /Promise\.all\(widgets\.map\(async \(widget\) => \[widget\.id, await previewDashboardWidget\(widget\)\]\)\)/)
})

test('viewer and editor keep widget element ref callbacks stable across renders', () => {
  const viewerSource = readFileSync(resolve(currentDir, 'Viewer.jsx'), 'utf8')
  const editorSource = readFileSync(resolve(currentDir, 'Editor.jsx'), 'utf8')

  for (const source of [viewerSource, editorSource]) {
    assert.match(source, /const widgetElementRefCallbacksRef = useRef\(\{\}\)/)
    assert.match(source, /if \(!widgetElementRefCallbacksRef\.current\[widgetID\]\)/)
    assert.match(source, /widgetElementRefCallbacksRef\.current\[widgetID\] = \(node\) =>/)
    assert.match(source, /return widgetElementRefCallbacksRef\.current\[widgetID\]/)
  }
})

test('editor defers expensive chart mounting while dashboard canvas scrolls', () => {
  const editorSource = readFileSync(resolve(currentDir, 'Editor.jsx'), 'utf8')
  const rendererSource = readFileSync(resolve(currentDir, 'WidgetRenderer.jsx'), 'utf8')

  assert.match(editorSource, /const \[widgetVisibilities,\s*setWidgetVisibilities\] = useState\(\{\}\)/)
  assert.match(editorSource, /const \[renderedChartWidgets,\s*setRenderedChartWidgets\] = useState\(\{\}\)/)
  assert.match(editorSource, /const \[dashboardScrolling,\s*setDashboardScrolling\] = useState\(false\)/)
  assert.match(editorSource, /dashboardScrollIdleTimerRef/)
  assert.match(editorSource, /onScroll=\{handleDashboardScroll\}/)
  assert.match(editorSource, /shouldRenderWidgetChart\(\{[\s\S]*alreadyRendered:\s*renderedChartWidgets\[widget\.id\],[\s\S]*visibility:\s*widgetVisibilities\[widget\.id\],[\s\S]*scrolling:\s*dashboardScrolling,[\s\S]*\}\)/)
  assert.doesNotMatch(editorSource, /renderChart=\{widgetVisibilities\[widget\.id\] === 'visible' && !dashboardScrolling\}/)
  assert.match(editorSource, /setRenderedChartWidgets/)

  assert.match(rendererSource, /renderChart = true/)
  assert.match(rendererSource, /!renderChart && !thumbnail/)
  assert.match(rendererSource, /t\('widgetDeferred'\)/)
})

test('chart render window releases distant offscreen ECharts after the hold period', () => {
  const next = reconcileRenderedChartWidgets({
    widgetIds: ['visible', 'near', 'recent', 'far-a', 'far-b'],
    rendered: {
      visible: { renderedAt: 1000, lastVisibleAt: 1000 },
      near: { renderedAt: 1000, lastVisibleAt: 1000 },
      recent: { renderedAt: 9000, lastVisibleAt: 9000 },
      'far-a': { renderedAt: 1000, lastVisibleAt: 1000 },
      'far-b': { renderedAt: 1000, lastVisibleAt: 1000 },
    },
    visibilities: {
      visible: 'visible',
      near: 'near',
      recent: 'offscreen',
      'far-a': 'offscreen',
      'far-b': 'offscreen',
    },
    distances: {
      visible: 0,
      near: 240,
      recent: 1200,
      'far-a': 2200,
      'far-b': 3200,
    },
    now: 10000,
    maxActiveCharts: 3,
    minHoldMs: 2500,
  })

  assert.deepEqual(Object.keys(next).sort(), ['near', 'recent', 'visible'])
  assert.equal(next.visible.lastVisibleAt, 10000)
})

test('chart render window holds recently visible widgets after they leave the viewport', () => {
  const next = reconcileRenderedChartWidgets({
    widgetIds: ['visible', 'recently-left', 'stale'],
    rendered: {
      visible: { renderedAt: 1000, lastVisibleAt: 1000 },
      'recently-left': { renderedAt: 1000, lastVisibleAt: 9000 },
      stale: { renderedAt: 1000, lastVisibleAt: 1000 },
    },
    visibilities: {
      visible: 'visible',
      'recently-left': 'offscreen',
      stale: 'offscreen',
    },
    distances: {
      visible: 0,
      'recently-left': 5000,
      stale: 100,
    },
    now: 10000,
    maxActiveCharts: 2,
    minHoldMs: 2500,
  })

  assert.deepEqual(Object.keys(next).sort(), ['recently-left', 'visible'])
})

test('chart render window keeps pinned widgets even when they are far offscreen', () => {
  const next = reconcileRenderedChartWidgets({
    widgetIds: ['visible', 'selected', 'far'],
    rendered: {
      visible: { renderedAt: 1000, lastVisibleAt: 1000 },
      selected: { renderedAt: 1000, lastVisibleAt: 1000 },
      far: { renderedAt: 1000, lastVisibleAt: 1000 },
    },
    visibilities: {
      visible: 'visible',
      selected: 'offscreen',
      far: 'offscreen',
    },
    distances: {
      visible: 0,
      selected: 5000,
      far: 1000,
    },
    pinnedWidgetIds: ['selected'],
    now: 10000,
    maxActiveCharts: 2,
    minHoldMs: 2500,
  })

  assert.deepEqual(Object.keys(next).sort(), ['selected', 'visible'])
})

test('editor passes raw near visibility to the chart render window', () => {
  const editorSource = readFileSync(resolve(currentDir, 'Editor.jsx'), 'utf8')

  assert.match(editorSource, /const chartWidgetVisibilities = chartWidgetIDs\.reduce/)
  assert.match(editorSource, /widgetVisibilityRef\.current\[widgetID\] \|\| widgetVisibilities\[widgetID\]/)
  assert.match(editorSource, /visibilities:\s*chartWidgetVisibilities/)
  assert.doesNotMatch(editorSource, /visibilities:\s*widgetVisibilities/)
})

test('viewer limits active ECharts with the chart render window', () => {
  const viewerSource = readFileSync(resolve(currentDir, 'Viewer.jsx'), 'utf8')

  assert.match(viewerSource, /reconcileRenderedChartWidgets/)
  assert.match(viewerSource, /shouldRenderWidgetChart/)
  assert.match(viewerSource, /const \[renderedChartWidgets,\s*setRenderedChartWidgets\] = useState\(\{\}\)/)
  assert.match(viewerSource, /const \[dashboardScrolling,\s*setDashboardScrolling\] = useState\(false\)/)
  assert.match(viewerSource, /onScroll=\{handleDashboardScroll\}/)
  assert.match(viewerSource, /const chartWidgetVisibilities = chartWidgetIDs\.reduce/)
  assert.match(viewerSource, /visibilities:\s*chartWidgetVisibilities/)
  assert.match(viewerSource, /renderChart=\{renderChart\}/)
})

test('viewer refreshes widget visibility once when dashboard scrolling becomes idle', () => {
  const viewerSource = readFileSync(resolve(currentDir, 'Viewer.jsx'), 'utf8')
  const idleTimerBlock = viewerSource.match(/dashboardScrollIdleTimerRef\.current = window\.setTimeout\(\(\) => \{[\s\S]*?\}, 180\)/)?.[0] || ''

  assert.match(viewerSource, /shouldRefreshWidgetVisibilityOnScroll/)
  assert.match(idleTimerBlock, /refreshWidgetVisibilities\(\)/)
  assert.match(idleTimerBlock, /setDashboardScrolling\(false\)/)
})

test('editor refreshes widget visibility classification while dashboard canvas scrolls', () => {
  const editorSource = readFileSync(resolve(currentDir, 'Editor.jsx'), 'utf8')

  assert.match(editorSource, /const dashboardScrollVisibilityFrameRef = useRef\(null\)/)
  assert.match(editorSource, /const refreshWidgetVisibilities = useCallback/)
  assert.match(editorSource, /const scheduleScrollVisibilityRefresh = useCallback/)
  assert.match(editorSource, /scheduleScrollVisibilityRefresh\(\)[\s\S]*if \(!dashboardScrollingRef\.current\)/)
  assert.match(editorSource, /window\.requestAnimationFrame\(\(\) => \{[\s\S]*refreshWidgetVisibilities\(\)/)
  assert.match(editorSource, /window\.cancelAnimationFrame\?\.\(dashboardScrollVisibilityFrameRef\.current\)/)
})

test('editor refreshes widget visibility after grid items mount', () => {
  const editorSource = readFileSync(resolve(currentDir, 'Editor.jsx'), 'utf8')
  const registerWidgetElementBlock = editorSource.match(/const registerWidgetElement = useCallback[\s\S]*?\n  }, \[[^\]]+\]\)/)?.[0] || ''

  assert.match(editorSource, /const refreshWidgetVisibilitiesRef = useRef\(null\)/)
  assert.match(editorSource, /const schedulePostLayoutVisibilityRefresh = useCallback/)
  assert.match(registerWidgetElementBlock, /widgetElementsRef\.current\[widgetID\] = node/)
  assert.match(registerWidgetElementBlock, /schedulePostLayoutVisibilityRefresh\(\)/)
  assert.match(editorSource, /refreshWidgetVisibilitiesRef\.current = refreshWidgetVisibilities/)
})

test('viewer exposes deferred, stale, and retry preview states per widget', () => {
  const viewerSource = readFileSync(resolve(currentDir, 'Viewer.jsx'), 'utf8')
  const rendererSource = readFileSync(resolve(currentDir, 'WidgetRenderer.jsx'), 'utf8')
  const styleSource = readFileSync(resolve(currentDir, 'index.scss'), 'utf8')
  const enMessages = readFileSync(resolve(currentDir, '../../../../public/locales/en/pro/customDashboard.json'), 'utf8')
  const zhMessages = readFileSync(resolve(currentDir, '../../../../public/locales/zh/pro/customDashboard.json'), 'utf8')

  assert.match(viewerSource, /const \[widgetPreviewStates,\s*setWidgetPreviewStates\] = useState\(\{\}\)/)
  assert.match(viewerSource, /shouldScheduleWidgetPreview/)
  assert.match(viewerSource, /previewStateKind\(response\)/)
  assert.match(viewerSource, /retryWidgetPreview/)
  assert.doesNotMatch(viewerSource, /priority:\s*'fullscreen'/)
  assert.match(viewerSource, /onRetry=\{\(\) => retryWidgetPreview\(widget\)\}/)
  assert.match(viewerSource, /custom-dashboard-widget-refreshing/)

  assert.match(rendererSource, /previewState/)
  assert.match(rendererSource, /t\('widgetDeferred'\)/)
  assert.match(rendererSource, /onRetry/)
  assert.match(styleSource, /\.custom-dashboard-widget-refreshing/)
  assert.match(enMessages, /"widgetDeferred"/)
  assert.match(zhMessages, /"widgetDeferred"/)
})

test('widget preview scheduling helpers gate eligibility and duplicate requests', () => {
  assert.equal(WIDGET_PREVIEW_CONCURRENCY_LIMIT, 2)
  assert.equal(widgetPreviewVisibilityFromRects({
    rootTop: 100,
    rootBottom: 500,
    itemTop: 120,
    itemBottom: 300,
    preloadMargin: 600,
  }), 'visible')
  assert.equal(widgetPreviewVisibilityFromRects({
    rootTop: 100,
    rootBottom: 500,
    itemTop: 700,
    itemBottom: 900,
    preloadMargin: 600,
  }), 'near')
  assert.equal(widgetPreviewVisibilityFromRects({
    rootTop: 100,
    rootBottom: 500,
    itemTop: 1200,
    itemBottom: 1400,
    preloadMargin: 600,
  }), 'offscreen')
  assert.equal(widgetPreviewEligibilityForVisibility({ visibility: 'visible' }), 'visible')
  assert.equal(widgetPreviewEligibilityForVisibility({ visibility: 'near' }), 'near')
  assert.equal(widgetPreviewEligibilityForVisibility({ visibility: 'offscreen' }), '')
  assert.equal(widgetPreviewEligibilityForVisibility({ visibility: 'offscreen', observationAvailable: false }), 'fallback')
  assert.equal(widgetPreviewEligibilityForVisibility({ visibility: 'offscreen', priority: 'fullscreen' }), 'fullscreen')
  assert.equal(shouldRefreshAllWidgetPreviews({ previews: {}, changedVariables: [{ name: 'namespace' }] }), true)
  assert.equal(shouldRefreshAllWidgetPreviews({ previews: { 'draft:w-cpu': { status: 'success' } }, changedVariables: [{ name: 'namespace' }] }), true)
  assert.equal(shouldRefreshAllWidgetPreviews({ previews: { 'w-cpu': { status: 'success' } }, changedVariables: [{ name: 'namespace' }] }), false)
  assert.equal(shouldRefreshAllWidgetPreviews({ previews: { 'w-cpu': { status: 'success' } }, changedVariables: [] }), true)
  assert.equal(shouldRefreshWidgetVisibilityOnScroll({ observationAvailable: true }), false)
  assert.equal(shouldRefreshWidgetVisibilityOnScroll({ observationAvailable: false }), true)

  const currentState = { status: 'success', requestKey: 'same-key', stale: false }
  assert.equal(shouldScheduleWidgetPreview(currentState, 'same-key'), false)
  assert.equal(shouldScheduleWidgetPreview({ ...currentState, stale: true }, 'same-key'), true)
  assert.equal(shouldScheduleWidgetPreview(currentState, 'new-key'), true)
  assert.equal(shouldScheduleWidgetPreview({ status: 'failed', requestKey: 'same-key' }, 'same-key'), false)
  assert.equal(shouldScheduleWidgetPreview({ status: 'failed', requestKey: 'same-key' }, 'same-key', { retry: true }), true)
})

test('widget focus draft previews render charts without visibility observation', async () => {
  const { shouldRenderWidgetChart } = await import('./behavior.mjs')

  assert.equal(typeof shouldRenderWidgetChart, 'function')
  assert.equal(shouldRenderWidgetChart({ draft: true }), true)
  assert.equal(shouldRenderWidgetChart({ draft: false, visibility: '', scrolling: false, alreadyRendered: false }), false)
  assert.equal(shouldRenderWidgetChart({ draft: false, visibility: 'visible', scrolling: false, alreadyRendered: false }), true)
  assert.equal(shouldRenderWidgetChart({ draft: false, visibility: 'visible', scrolling: true, alreadyRendered: false }), false)
  assert.equal(shouldRenderWidgetChart({ draft: false, visibility: 'not-visible', scrolling: false, alreadyRendered: true }), true)

  const editorSource = readFileSync(resolve(currentDir, 'Editor.jsx'), 'utf8')
  assert.match(editorSource, /shouldRenderWidgetChart\(\{[\s\S]*draft,[\s\S]*alreadyRendered:\s*renderedChartWidgets\[widget\.id\],[\s\S]*visibility:\s*widgetVisibilities\[widget\.id\],[\s\S]*scrolling:\s*dashboardScrolling,[\s\S]*\}\)/)
})

test('editor isolates draft widget preview state from saved widget preview state', () => {
  const editorSource = readFileSync(resolve(currentDir, 'Editor.jsx'), 'utf8')
  const previewDashboardWidgetBlock = editorSource.match(/const previewDashboardWidget = async[\s\S]*?\n  }\n\n  const previewWidget/)?.[0] || ''

  assert.match(editorSource, /const draftPreviewKeyForWidget = \(widget\) =>/)
  assert.match(editorSource, /const previewKey = draft \? draftPreviewKeyForWidget\(widget\) : widget\.id/)
  assert.match(editorSource, /preview=\{previews\[previewKey\]\}/)
  assert.match(editorSource, /previewState=\{widgetPreviewStates\[previewKey\]\}/)
  assert.match(editorSource, /const previewDashboardWidget = async \(widget,\s*\{ previewKey = widget\.id \} = \{\}\) =>/)
  assert.match(editorSource, /previewRequestKeysRef\.current\[previewKey\] = requestKey/)
  assert.match(editorSource, /applyPreview\(\(prev\) => \(\{ \.\.\.prev, \[previewKey\]: response \}\)\)/)
  assert.match(editorSource, /clearDraftPreview\(widget\)/)
  assert.doesNotMatch(previewDashboardWidgetBlock, /applyPreview\(\(prev\) => \(\{ \.\.\.prev, \[widget\.id\]: response \}\)\)/)
})

test('editor seeds draft preview from saved preview without inheriting active saved state', () => {
  const editorSource = readFileSync(resolve(currentDir, 'Editor.jsx'), 'utf8')
  const seedDraftPreviewBlock = editorSource.match(/const seedDraftPreview = useCallback[\s\S]*?\n  }, \[[^\]]+\]\)/)?.[0] || ''

  assert.match(editorSource, /const seedDraftPreview = useCallback\(\(widget\) => \{/)
  assert.match(seedDraftPreviewBlock, /const preview = previewsRef\.current\[widget\.id\]/)
  assert.match(seedDraftPreviewBlock, /const draftPreviewKey = draftPreviewKeyForWidget\(widget\)/)
  assert.match(seedDraftPreviewBlock, /applyPreview\(\(prev\) => \(\{ \.\.\.prev, \[draftPreviewKey\]: preview \}\)\)/)
  assert.match(seedDraftPreviewBlock, /updateWidgetPreviewState\(draftPreviewKey,\s*\{[\s\S]*status:\s*previewStateKind\(preview\),[\s\S]*requestKey:\s*'',[\s\S]*stale:\s*false/)
  assert.doesNotMatch(seedDraftPreviewBlock, /widgetPreviewStatesRef\.current\[widget\.id\]/)
  assert.match(editorSource, /const startEditWidget = \(widget\) => \{[\s\S]*seedDraftPreview\(widget\)/)
})

test('editor refreshes the saved widget when an existing draft is closed without saving', () => {
  const editorSource = readFileSync(resolve(currentDir, 'Editor.jsx'), 'utf8')
  const saveWidgetDraftBlock = editorSource.match(/const saveWidgetDraft = \(\) => \{[\s\S]*?\n  }\n\n  const confirmAddWidget/)?.[0] || ''
  const draftCloseVisibilityBlock = editorSource.match(/useEffect\(\(\) => \{[\s\S]*?hadDraftWidgetRef\.current[\s\S]*?\}, \[[^\]]+\]\)/)?.[0] || ''

  assert.match(editorSource, /const refreshSavedWidgetAfterDraftClose = useCallback\(\(widget\) => \{/)
  assert.match(editorSource, /if \(widgetDraftMode\(dashboard\?\.widgets \|\| \[\],\s*widget\) !== 'edit'\) return/)
  assert.match(editorSource, /const savedWidget = dashboard\?\.widgets\?\.find\(\(item\) => item\.id === widget\.id\)/)
  assert.match(editorSource, /if \(savedWidget\) markWidgetsStale\(\[savedWidget\]\)/)
  assert.match(editorSource, /const closeWidgetEditor = \(\) => \{[\s\S]*const widget = draftWidget[\s\S]*clearDraftPreview\(widget\)[\s\S]*refreshSavedWidgetAfterDraftClose\(widget\)/)
  assert.match(editorSource, /const hadDraftWidgetRef = useRef\(false\)/)
  assert.match(draftCloseVisibilityBlock, /const hadDraftWidget = hadDraftWidgetRef\.current/)
  assert.match(draftCloseVisibilityBlock, /hadDraftWidgetRef\.current = Boolean\(draftWidget\)/)
  assert.match(draftCloseVisibilityBlock, /if \(!hadDraftWidget \|\| draftWidget \|\| !dashboard\) return undefined/)
  assert.match(draftCloseVisibilityBlock, /window\.requestAnimationFrame\?\.\(\(\) => refreshWidgetVisibilities\(\)\)/)
  assert.doesNotMatch(saveWidgetDraftBlock, /refreshSavedWidgetAfterDraftClose/)
})

test('editor refreshes the active draft preview when dashboard variables change', () => {
  const editorSource = readFileSync(resolve(currentDir, 'Editor.jsx'), 'utf8')

  assert.match(editorSource, /const draftWidgetRef = useRef\(null\)/)
  assert.match(editorSource, /draftWidgetRef\.current = draftWidget/)
  assert.match(editorSource, /const activeDraftWidget = draftWidgetRef\.current/)
  assert.match(editorSource, /const dashboardWidgets = \(dashboard\.widgets \|\| \[\]\)\.filter\(\(widget\) => \([\s\S]*!activeDraftWidget \|\| widget\.id !== activeDraftWidget\.id[\s\S]*\)\)/)
  assert.match(editorSource, /const draftUsesChangedVariable = activeDraftWidget && \([\s\S]*shouldRefreshAll[\s\S]*changedVariables\.some\(\(variable\) => widgetUsesVariable\(activeDraftWidget,\s*variable,\s*presets\)\)[\s\S]*\)/)
  assert.match(editorSource, /if \(draftUsesChangedVariable\) \{[\s\S]*previewDashboardWidget\(activeDraftWidget,\s*\{ previewKey: draftPreviewKeyForWidget\(activeDraftWidget\) \}\)[\s\S]*\}/)
  assert.match(editorSource, /await previewDashboardWidget\(selectedWidget,\s*\{ previewKey: draftWidget \? draftPreviewKeyForWidget\(selectedWidget\) : selectedWidget\.id \}\)/)
})

test('editor refreshes draft preview key when draft series limit changes', () => {
  const editorSource = readFileSync(resolve(currentDir, 'Editor.jsx'), 'utf8')

  assert.match(editorSource, /const updateWidgetSeriesLimit = useCallback\(\(widget,\s*value,\s*\{ skipRefresh = false \} = \{\}\) =>/)
  assert.match(editorSource, /if \(!skipRefresh\) markWidgetsStale\(\[widget\]\)/)
  assert.match(editorSource, /const changed = updateWidgetSeriesLimit\(widget,\s*value,\s*\{ skipRefresh: draft \}\)[\s\S]*if \(draft && changed\) previewDashboardWidget\(widget,\s*\{ previewKey \}\)/)
})

test('dashboard preview scheduler caps concurrency and replaces queued widget work', async () => {
  const scheduler = createDashboardPreviewScheduler({ limit: 2 })
  const started = []
  const releases = new Map()
  const makeRun = (key) => () => new Promise((resolve) => {
    started.push(key)
    releases.set(key, resolve)
  })

  scheduler.schedule({ widgetId: 'a', requestKey: 'a1', priority: WIDGET_PREVIEW_PRIORITY.visible, run: makeRun('a1') })
  scheduler.schedule({ widgetId: 'b', requestKey: 'b1', priority: WIDGET_PREVIEW_PRIORITY.visible, run: makeRun('b1') })
  scheduler.schedule({ widgetId: 'c', requestKey: 'c1', priority: WIDGET_PREVIEW_PRIORITY.near, run: makeRun('c1') })
  scheduler.schedule({ widgetId: 'c', requestKey: 'c2', priority: WIDGET_PREVIEW_PRIORITY.near, run: makeRun('c2') })
  await flushMicrotasks()

  assert.deepEqual(started, ['a1', 'b1'])
  assert.equal(scheduler.activeCount(), 2)
  assert.equal(scheduler.queuedCount(), 1)

  releases.get('a1')()
  await flushMicrotasks()

  assert.deepEqual(started, ['a1', 'b1', 'c2'])
  assert.equal(started.includes('c1'), false)
})

test('dashboard preview scheduler clears queued work and cancels active work', async () => {
  const scheduler = createDashboardPreviewScheduler({ limit: 1 })
  const signals = []
  scheduler.schedule({
    widgetId: 'a',
    requestKey: 'a1',
    priority: WIDGET_PREVIEW_PRIORITY.visible,
    run: ({ signal }) => new Promise(() => {
      signals.push(signal)
    }),
  })
  scheduler.schedule({
    widgetId: 'b',
    requestKey: 'b1',
    priority: WIDGET_PREVIEW_PRIORITY.visible,
    run: () => Promise.resolve(),
  })
  await flushMicrotasks()

  assert.equal(scheduler.activeCount(), 1)
  assert.equal(scheduler.queuedCount(), 1)

  scheduler.clear()
  assert.equal(scheduler.activeCount(), 0)
  assert.equal(scheduler.queuedCount(), 0)
  assert.equal(signals[0]?.aborted, true)
})

test('preset helpers expose only compatible presets and default parameters', () => {
  const presets = [
    { id: 'line-only', name: 'Line Only', category: 'service', compatibleCharts: ['line'], parameters: [{ name: 'service', default: '$service' }] },
    { id: 'bar-only', name: 'Bar Only', category: 'instance', compatibleCharts: ['bar'], parameters: [{ name: 'pod', default: '$pod' }] },
  ]
  assert.deepEqual(presetOptionsForChart(presets, 'line', (key) => `t:${key}`), [
    { label: 't:presetCategory.service', options: [{ label: 'Line Only', value: 'line-only' }] },
  ])
  assert.deepEqual(mergePresetDefaults(presets[0], { other: 'kept' }), { other: 'kept', service: '$service' })
})

test('catalog templates and preset metadata are translated before display', () => {
  const messages = {
    'templates.service-health.name': '服务健康',
    'templates.service-health.description': '服务指标概览',
    'presetQueries.service-qps.name': '请求量 QPS',
    'presetParameters.service.label': '服务',
    'presetParameters.service.description': '服务变量或具体服务名',
  }
  const t = (key) => messages[key] || key
  const template = { id: 'service-health', name: 'Service Health', description: 'Service health dashboard' }
  const preset = { id: 'service-qps', name: 'Request QPS', category: 'service', compatibleCharts: ['line'] }
  const param = { name: 'service', label: 'Service', description: 'Service variable or concrete service name' }

  assert.equal(templateDisplayName(t, template), '服务健康')
  assert.equal(templateDisplayDescription(t, template), '服务指标概览')
  assert.equal(templateDisplayName(t, { id: 'unknown', name: 'Fallback' }), 'Fallback')
  assert.deepEqual(presetOptionsForChart([preset], 'line', t), [
    { label: 'presetCategory.service', options: [{ label: '请求量 QPS', value: 'service-qps' }] },
  ])
  assert.equal(presetParameterLabel(t, param), '服务')
  assert.equal(presetParameterDescription(t, param), '服务变量或具体服务名')

  const listSource = readFileSync(resolve(currentDir, 'List.jsx'), 'utf8')
  const editorSource = readFileSync(resolve(currentDir, 'Editor.jsx'), 'utf8')
  assert.doesNotMatch(listSource, /templateDisplayName\(t,\s*template\)/)
  assert.doesNotMatch(listSource, /templateDisplayDescription\(t,\s*template\)/)
  assert.match(editorSource, /presetParameterLabel\(t,\s*param\)/)
  assert.match(editorSource, /presetParameterDescription\(t,\s*param\)/)
})

test('template widget titles display localized preset names without overriding custom titles', () => {
  const messages = { 'presetQueries.service-qps.name': '请求量 QPS' }
  const t = (key) => messages[key] || key
  const preset = { id: 'service-qps', name: 'Request QPS' }

  assert.equal(
    widgetDisplayTitle(t, { title: 'Request QPS', query: { presetId: 'service-qps' } }, [preset]),
    '请求量 QPS',
  )
  assert.equal(
    widgetDisplayTitle(t, { title: 'Business QPS', query: { presetId: 'service-qps' } }, [preset]),
    'Business QPS',
  )

  const editorSource = readFileSync(resolve(currentDir, 'Editor.jsx'), 'utf8')
  const viewerSource = readFileSync(resolve(currentDir, 'Viewer.jsx'), 'utf8')
  assert.match(editorSource, /widgetDisplayTitle\(t,\s*widget,\s*presets\)/)
  assert.match(viewerSource, /widgetDisplayTitle\(t,\s*widget,\s*presets\)/)
})

test('viewer option labels and table headers are translated through caller', () => {
  const t = (key) => `t:${key}`
  assert.deepEqual(translatedTimeRangeOptions(t).map((item) => item.label), [
    't:last15Minutes',
    't:last1Hour',
    't:last6Hours',
    't:last24Hours',
  ])
  assert.equal(displayText(t, 'query failed', 3), 'query failed (t:position 3)')
})

test('standard resource variable labels are localized from stored dashboard defaults', () => {
  const t = (key) => ({
    cluster: '集群',
    namespace: '命名空间',
    endpoint: '服务端点',
    instance: '实例',
  }[key] || key)

  assert.equal(variableDisplayLabel(t, { name: 'cluster', type: 'cluster', label: 'Cluster' }), '集群')
  assert.equal(variableDisplayLabel(t, { name: 'namespace', type: 'namespace', label: 'Namespace' }), '命名空间')
  assert.equal(variableDisplayLabel(t, { name: 'endpoint', type: 'endpoint', label: 'Endpoint' }), '服务端点')
  assert.equal(variableDisplayLabel(t, { name: 'instance', type: 'instance', label: 'Instance' }), '实例')
  assert.equal(variableDisplayLabel(t, { name: 'owner', type: 'label', label: 'Owner' }), 'Owner')

  const en = (key) => ({
    cluster: 'Cluster',
    namespace: 'Namespace',
    service: 'Service',
    endpoint: 'Endpoint',
    pod: 'pod',
    instance: 'Instance',
  }[key] || key)
  assert.equal(variableDisplayLabel(en, { name: 'namespace', type: 'namespace', label: '命名空间' }), 'Namespace')
  assert.equal(variableDisplayLabel(en, { name: 'service', type: 'service', label: '服务' }), 'Service')
  assert.equal(variableDisplayLabel(en, { name: 'endpoint', type: 'endpoint', label: '服务端点' }), 'Endpoint')

  const enMessages = JSON.parse(readFileSync(resolve(currentDir, '../../../../public/locales/en/pro/customDashboard.json'), 'utf8'))
  const zhMessages = JSON.parse(readFileSync(resolve(currentDir, '../../../../public/locales/zh/pro/customDashboard.json'), 'utf8'))
  for (const key of ['cluster', 'namespace', 'service', 'pod', 'endpoint', 'instance']) {
    assert.ok(enMessages[key])
    assert.ok(zhMessages[key])
  }
  assert.equal(zhMessages.endpoint, '服务端点')

  const editorSource = readFileSync(resolve(currentDir, 'Editor.jsx'), 'utf8')
  const viewerSource = readFileSync(resolve(currentDir, 'Viewer.jsx'), 'utf8')
  assert.match(editorSource, /variableDisplayLabel\(t,\s*variable\)/)
  assert.match(viewerSource, /variableDisplayLabel\(t,\s*variable\)/)
})

test('known imported dashboard variable labels are localized from stored defaults', () => {
  const en = (key) => ({
    application: 'Application',
    queryMode: 'Query mode',
    queryByK8s: 'Query by K8s',
    relatedNodes: 'Related Nodes',
    resolution: 'Resolution',
    jvmInstance: 'JVM Instance',
    hostMetricsJob: 'Host metrics Job',
    containerMetricsJob: 'Container metrics Job',
  }[key] || key)

  assert.equal(variableDisplayLabel(en, { name: 'service_name', type: 'service', label: '应用名', labels: ['svc_name', 'service_name'] }), 'Application')
  assert.equal(variableDisplayLabel(en, { name: 'service_type', type: 'custom', label: '查询方式' }), 'Query mode')
  assert.equal(variableDisplayLabel(en, { name: 'related_nodes', type: 'instance', label: '相关节点', labels: ['node_name'] }), 'Related Nodes')
  assert.equal(variableDisplayLabel(en, { name: 'resolution', type: 'custom', label: '精度' }), 'Resolution')
  assert.equal(variableDisplayLabel(en, { name: 'jvm_instance', type: 'label', label: 'JVM实例', labels: ['jvm_instance'] }), 'JVM Instance')
  assert.equal(variableDisplayLabel(en, { name: 'job_name', type: 'label', label: '主机指标 Job', labels: ['job_name'] }), 'Host metrics Job')
  assert.equal(variableDisplayLabel(en, { name: 'cadvisor_job_name', type: 'label', label: '容器指标 Job', labels: ['job_name'] }), 'Container metrics Job')
  assert.equal(variableDisplayLabel(en, { name: 'owner', type: 'label', label: '负责人' }), '负责人')
  assert.deepEqual(
    variableSelectOptions(en, { name: 'service_type', type: 'custom' }, [{ label: '按K8s查询', value: '.*', type: 'custom' }]),
    [{ label: 'Query by K8s', value: '.*' }],
  )

  const enMessages = JSON.parse(readFileSync(resolve(currentDir, '../../../../public/locales/en/pro/customDashboard.json'), 'utf8'))
  const zhMessages = JSON.parse(readFileSync(resolve(currentDir, '../../../../public/locales/zh/pro/customDashboard.json'), 'utf8'))
  for (const key of ['application', 'queryByK8s', 'relatedNodes', 'resolution', 'jvmInstance', 'hostMetricsJob', 'containerMetricsJob']) {
    assert.ok(enMessages[key])
    assert.ok(zhMessages[key])
  }
})

test('grid placement uses persisted layout coordinates', () => {
  assert.deepEqual(normalizeGridLayout({ x: 0, y: 0, w: 2, h: 3 }, 2), { x: 0, y: 0, w: 2, h: 3 })
  assert.deepEqual(gridPlacement({ x: 0, y: 0, w: 2, h: 3 }, 2), {
    gridColumn: '1 / span 2',
    gridRow: '1 / span 3',
  })
  assert.deepEqual(gridPlacement({ x: 6, y: 4, w: 3, h: 5 }), {
    gridColumn: '3 / span 1',
    gridRow: '1 / span 1',
  })
  assert.deepEqual(gridPlacement({ x: 11, y: -2, w: 6, h: 0 }), {
    gridColumn: '3 / span 2',
    gridRow: '1 / span 1',
  })
})

test('fixed cell layout helpers move widgets and reject occupied cells', () => {
  const widgets = [
    { id: 'a', layout: { x: 0, y: 0, w: 2, h: 1 } },
    { id: 'b', layout: { x: 2, y: 0, w: 2, h: 1 } },
  ]
  const moved = moveWidgetToGridPosition(widgets, 'b', 0, 0)
  assert.deepEqual(moved, widgets)
  assert.equal(gridLayoutsOverlap(moved[0].layout, moved[1].layout), false)

  const valid = updateWidgetGridLayout(widgets, 'b', { x: 0, y: 1 })
  assert.deepEqual(valid.find((item) => item.id === 'b').layout, { x: 0, y: 1, w: 2, h: 1 })

  const inserted = addWidgetToOpenCell(widgets, { id: 'c', layout: { x: 0, y: 0, w: 2, h: 1 } })
  assert.deepEqual(inserted.find((item) => item.id === 'c').layout, { x: 0, y: 1, w: 2, h: 1 })

  const resized = resizeWidgetGridLayout(widgets, 'a', { w: 2, h: 2 })
  assert.deepEqual(resized.find((item) => item.id === 'a').layout, { x: 0, y: 0, w: 2, h: 2 })
  const blocked = resizeWidgetGridLayout(widgets, 'a', { w: 4, h: 1 })
  assert.deepEqual(blocked, widgets)
})

test('dashboard grid layouts preserve sparse rows until the user moves a panel', () => {
  const layouts = arrangeWidgetLayouts([
    { id: 'first', layout: { x: 0, y: 0, w: 1, h: 1 } },
    { id: 'second', layout: { x: 0, y: 1, w: 1, h: 1 } },
  ])

  assert.deepEqual(layouts.first, { x: 0, y: 0, w: 1, h: 1 })
  assert.deepEqual(layouts.second, { x: 0, y: 1, w: 1, h: 1 })
})

test('dashboard grid sync only persists the panel the user moved', () => {
  const widgets = [
    { id: 'first', layout: { x: 0, y: 0, w: 1, h: 1 } },
    { id: 'second', layout: { x: 0, y: 1, w: 1, h: 1 } },
  ]
  const runtimeLayouts = [
    { i: 'first', x: 1, y: 0, w: 1, h: 1 },
    { i: 'second', x: 2, y: 0, w: 1, h: 1 },
  ]

  assert.deepEqual(
    mergeWidgetGridLayouts(widgets, runtimeLayouts, 4, runtimeLayouts[0]).map((widget) => widget.layout),
    [
      { x: 1, y: 0, w: 1, h: 1 },
      { x: 0, y: 1, w: 1, h: 1 },
    ],
  )
})

test('dashboard grid sync pushes blocking panels down when a panel is dropped into their cell', () => {
  const widgets = [
    { id: 'cpu', layout: { x: 0, y: 0, w: 2, h: 1 } },
    { id: 'throttle', layout: { x: 2, y: 0, w: 2, h: 1 } },
    { id: 'memory', layout: { x: 0, y: 1, w: 4, h: 1 } },
    { id: 'bytes', layout: { x: 0, y: 2, w: 2, h: 1 } },
    { id: 'time', layout: { x: 2, y: 2, w: 2, h: 1 } },
    { id: 'disk', layout: { x: 0, y: 3, w: 2, h: 1 } },
    { id: 'bandwidth', layout: { x: 2, y: 3, w: 2, h: 1 } },
    { id: 'rtt', layout: { x: 0, y: 4, w: 4, h: 1 } },
    { id: 'packets', layout: { x: 0, y: 5, w: 2, h: 1 } },
    { id: 'loss', layout: { x: 2, y: 5, w: 2, h: 1 } },
    { id: 'errors', layout: { x: 0, y: 6, w: 2, h: 1 } },
  ]
  const runtimeLayouts = [
    { i: 'cpu', x: 0, y: 0, w: 2, h: 1 },
    { i: 'throttle', x: 2, y: 0, w: 2, h: 1 },
    { i: 'memory', x: 0, y: 1, w: 4, h: 1 },
    { i: 'bytes', x: 0, y: 2, w: 2, h: 1 },
    { i: 'time', x: 2, y: 2, w: 2, h: 1 },
    { i: 'disk', x: 0, y: 3, w: 2, h: 1 },
    { i: 'bandwidth', x: 0, y: 5, w: 2, h: 1 },
    { i: 'rtt', x: 0, y: 4, w: 4, h: 1 },
    { i: 'packets', x: 0, y: 5, w: 2, h: 1 },
    { i: 'loss', x: 2, y: 5, w: 2, h: 1 },
    { i: 'errors', x: 0, y: 6, w: 2, h: 1 },
  ]

  assert.deepEqual(
    mergeWidgetGridLayouts(widgets, runtimeLayouts, 4, runtimeLayouts[6]).map((widget) => widget.layout),
    [
      { x: 0, y: 0, w: 2, h: 1 },
      { x: 2, y: 0, w: 2, h: 1 },
      { x: 0, y: 1, w: 4, h: 1 },
      { x: 0, y: 2, w: 2, h: 1 },
      { x: 2, y: 2, w: 2, h: 1 },
      { x: 0, y: 3, w: 2, h: 1 },
      { x: 0, y: 5, w: 2, h: 1 },
      { x: 0, y: 4, w: 4, h: 1 },
      { x: 0, y: 6, w: 2, h: 1 },
      { x: 2, y: 5, w: 2, h: 1 },
      { x: 0, y: 7, w: 2, h: 1 },
    ],
  )
})

test('dashboard grid sync removes fully empty rows after a panel is moved away', () => {
  const widgets = [
    { id: 'top', layout: { x: 0, y: 0, w: 4, h: 1 } },
    { id: 'middle', layout: { x: 0, y: 1, w: 4, h: 1 } },
    { id: 'bottom', layout: { x: 0, y: 2, w: 4, h: 1 } },
  ]
  const runtimeLayouts = [
    { i: 'top', x: 0, y: 3, w: 4, h: 1 },
    { i: 'middle', x: 0, y: 1, w: 4, h: 1 },
    { i: 'bottom', x: 0, y: 2, w: 4, h: 1 },
  ]

  assert.deepEqual(
    mergeWidgetGridLayouts(widgets, runtimeLayouts, 4, runtimeLayouts[0]).map((widget) => widget.layout),
    [
      { x: 0, y: 2, w: 4, h: 1 },
      { x: 0, y: 0, w: 4, h: 1 },
      { x: 0, y: 1, w: 4, h: 1 },
    ],
  )
})

test('legacy two-column layouts expand into a four-column grid', () => {
  const layouts = arrangeWidgetLayouts([
    { id: 'a', layout: { x: 0, y: 0 } },
    { id: 'b', layout: { x: 1, y: 3 } },
    { id: 'c', layout: { x: 0, y: 5 } },
  ])
  assert.deepEqual(layouts.a, { x: 0, y: 0, w: 2, h: 1 })
  assert.deepEqual(layouts.b, { x: 2, y: 3, w: 2, h: 1 })
  assert.deepEqual(layouts.c, { x: 0, y: 5, w: 2, h: 1 })
})

test('responsive grid remains a fixed four-column dashboard', () => {
  assert.equal(responsiveGridColumns(900), 4)
  assert.equal(responsiveGridColumns(1280), 4)

  const layouts = arrangeWidgetLayouts([
    { id: 'left', layout: { x: 0, y: 0, w: 6, h: 4 } },
    { id: 'right', layout: { x: 6, y: 0, w: 6, h: 4 } },
  ])
  assert.deepEqual(layouts.left, { x: 0, y: 0, w: 2, h: 1 })
  assert.deepEqual(layouts.right, { x: 2, y: 0, w: 2, h: 1 })
  assert.deepEqual(gridPlacement(layouts.right), {
    gridColumn: '3 / span 2',
    gridRow: '1 / span 1',
  })
})

test('dashboard grid uses explicit row height so cards do not overlap', () => {
  const styleSource = readFileSync(resolve(currentDir, 'index.scss'), 'utf8')
  const gridRule = styleSource.match(/\.custom-dashboard-grid\s*\{[\s\S]*?\n\}/)?.[0] || ''
  const gridLayoutRule = styleSource.match(/\.custom-dashboard-grid-layout\s*\{[\s\S]*?\n\}/)?.[0] || ''
  const widgetRule = styleSource.match(/\.custom-dashboard-widget\s*\{[\s\S]*?\n\}/)?.[0] || ''
  const addWidgetRule = styleSource.match(/\.custom-dashboard-add-widget\s*\{[\s\S]*?\n\}/)?.[0] || ''

  assert.match(gridRule, /--custom-dashboard-widget-height:/)
  assert.match(gridRule, /repeat\(var\(--custom-dashboard-grid-columns,\s*4\)/)
  assert.match(gridRule, /grid-auto-rows:\s*var\(--custom-dashboard-widget-height\)/)
  assert.match(gridLayoutRule, /min-height:\s*0/)
  assert.doesNotMatch(gridLayoutRule, /min-height:\s*var\(--custom-dashboard-widget-height\)/)
  assert.match(widgetRule, /height:\s*100%/)
  assert.doesNotMatch(widgetRule, /aspect-ratio/)
  assert.match(addWidgetRule, /height:\s*100%/)
  assert.doesNotMatch(addWidgetRule, /aspect-ratio/)
})

test('list visibility and save errors are translated for callers', () => {
  const t = (key) => `t:${key}`
  const zhLocale = JSON.parse(readFileSync(resolve(currentDir, '../../../../public/locales/zh/pro/customDashboard.json'), 'utf8'))
  const enLocale = JSON.parse(readFileSync(resolve(currentDir, '../../../../public/locales/en/pro/customDashboard.json'), 'utf8'))

  assert.equal(visibilityLabel(t, 'team'), 't:team')
  assert.equal(zhLocale.personal, '私密')
  assert.equal(zhLocale.unitPresetBytes, '字节 (B)')
  assert.equal(zhLocale.unitPresetMilliseconds, '毫秒 (ms)')
  assert.equal(zhLocale.unitPresetPercent, '百分比 (%)')
  assert.equal(zhLocale.stat, '指标卡')
  assert.equal(zhLocale.barGauge, '条形仪表')
  assert.equal(enLocale.personal, 'Private')
  assert.equal(saveFailureMessage(t, { response: { data: { code: 'A0006' } } }), 't:noEditPermission')
  assert.equal(saveFailureMessage(t, { response: { data: { message: 'server rejected save' } } }), 'server rejected save')
  assert.equal(saveFailureMessage(t, {}), 't:saveFailed')
})

test('editor exposes team selector when dashboard visibility is team scoped', () => {
  const editorSource = readFileSync(resolve(currentDir, 'Editor.jsx'), 'utf8')

  assert.match(editorSource, /getTeamsApi/)
  assert.match(editorSource, /setTeams\(res\?\.teamList\s*\|\|\s*\[\]\)/)
  assert.match(editorSource, /dashboard\.visibility === 'team'/)
  assert.match(editorSource, /value=\{dashboard\.teamIds\s*\|\|\s*\[\]\}/)
  assert.match(editorSource, /updateDashboard\(\{ teamIds \}\)/)
  assert.match(editorSource, /fieldNames=\{\{ label: 'teamName', value: 'teamId' \}\}/)
})

test('label variable helpers discover labels and serialize multi-select values', () => {
  assert.deepEqual(extractMetricLabels('sum(rate(http_requests_total{svc="api",namespace=~"prod|dev"}[5m]))'), ['namespace', 'svc'])
  assert.deepEqual(dashboardLabelOptions([
    { query: { mode: 'custom', expression: 'up{svc_name="checkout"}' } },
  ]), [{ label: 'Service', value: 'service', type: 'service', labels: ['svc_name'] }])
  assert.deepEqual(dashboardLabelOptions([
    { query: { mode: 'custom', expression: 'up{host_name="node-a"}' } },
  ]), [{ label: 'Instance', value: 'instance', type: 'instance', labels: ['host_name'] }])
  assert.deepEqual(dashboardLabelOptions([
    { query: { mode: 'custom', expression: 'up{clusterID="prod"} + up{cluster_name="prod"}' } },
  ]), [{ label: 'Cluster', value: 'cluster', type: 'cluster', labels: ['clusterID', 'cluster_name'] }])
  assert.deepEqual(variableValuesForQuery(
    { service: ['checkout', 'cart', 'foo|bar'], namespace: 'prod' },
    [{ name: 'service' }, { name: 'namespace' }],
  ), { service: ['checkout', 'cart', 'foo|bar'], namespace: ['prod'] })
  assert.equal(
    applyLabelVariableMatchers(
      'sum(rate(http_requests_total{svc_name="old",code!="500"}[5m]))',
      { service: 'checkout|cart' },
      [{ name: 'service', labels: ['svc_name', 'svc'] }],
    ),
    'sum(rate(http_requests_total{svc_name=~"checkout|cart",code!="500"}[5m]))',
  )
  assert.equal(
    applyLabelVariableMatchers(
      'sum(rate(http_requests_total{svc="old"}[5m])) + sum(rate(rpc_total{svc_name="old"}[5m]))',
      { service: ['checkout'] },
      [{ name: 'service', labels: ['svc_name', 'svc'] }],
    ),
    'sum(rate(http_requests_total{svc=~"checkout"}[5m])) + sum(rate(rpc_total{svc_name=~"checkout"}[5m]))',
  )
  assert.equal(
    applyLabelVariableMatchers(
      'up{svc_name="old"}',
      { service: ['api.v1', 'cart+checkout', 'foo|bar'] },
      [{ name: 'service', type: 'service' }],
    ),
    'up{svc_name=~"api\\\\.v1|cart\\\\+checkout|foo\\\\|bar"}',
  )
  assert.equal(
    applyLabelVariableMatchers(
      'up{namespace="old"}',
      { namespace: ['.*'] },
      [{ name: 'namespace', type: 'namespace', includeAll: true, allValue: '.*' }],
    ),
    'up{namespace=~".*"}',
  )
  assert.equal(
    applyLabelVariableMatchers(
      'up{job="apiserver"}',
      { service: ['checkout'] },
      [{ name: 'service', type: 'service' }],
    ),
    'up{job="apiserver"}',
  )
  assert.equal(
    applyLabelVariableMatchers(
      'up{clusterId="old"} + up{cluster_name="old"} + up{cluster="old"}',
      { cluster: ['prod'] },
      [{ name: 'cluster', type: 'cluster', labels: ['clusterId', 'cluster_name'] }],
    ),
    'up{clusterId=~"prod"} + up{cluster_name=~"prod"} + up{cluster="old"}',
  )
  assert.equal(widgetUsesVariable(
    { query: { mode: 'custom', expression: 'up{svc_name="checkout"}' } },
    { name: 'service', type: 'service' },
  ), true)
  assert.equal(widgetUsesVariable(
    { query: { mode: 'custom', expression: 'up{job="apiserver"}' } },
    { name: 'service', type: 'service' },
  ), false)
})

test('supported widget types and historical widget normalization follow compatibility rules', () => {
  assert.deepEqual(chartTypes.map((item) => item.value), ['line', 'area', 'bar', 'table', 'stat', 'barGauge'])
  const dashboard = normalizeDashboard({
    variables: [{ name: 'timeRange', type: 'timeRange' }],
    widgets: [
      { id: 'top', type: 'topn', layout: { x: 0, y: 0 } },
      { id: 'single', type: 'single', layout: { x: 1, y: 0 } },
      { id: 'legacy-query', type: 'line', query: { mode: 'custom', expression: 'up', legend: 'Up' } },
      { id: 'resized', type: 'line', layout: { x: 0, y: 1, w: 2, h: 3 } },
    ],
  })
  assert.deepEqual(dashboard.variables, [])
  assert.equal(dashboard.widgets[0].type, 'table')
  assert.equal(dashboard.widgets[0].legacyType, 'topn')
  assert.equal(dashboard.widgets[1].type, 'single')
  assert.equal(dashboard.widgets[1].unsupportedType, true)
  assert.deepEqual(dashboard.widgets[2].query.queries, [{ id: 'q-1', expression: 'up', legend: 'Up' }])
  assert.deepEqual(dashboard.widgets[3].layout, { x: 0, y: 1, w: 2, h: 3 })
})

test('frontend dashboard normalization defaults loaded builtin resources to multi-select', () => {
  const dashboard = normalizeDashboard({
    variables: [
      { name: 'namespace', label: 'Namespace', type: 'label', labels: ['namespace'] },
      { name: 'service_name', label: 'Service', type: 'label', labels: ['service_name'], value: 'sandbox-default-user-3', defaultValue: 'sandbox-default-user-3' },
      { name: 'pod', label: 'Pod', type: 'label', labels: ['pod'], multi: false },
      { name: 'job_name', label: 'Job', type: 'label', labels: ['job_name'] },
    ],
    widgets: [],
  })
  const namespace = dashboard.variables.find((variable) => variable.name === 'namespace')
  const service = dashboard.variables.find((variable) => variable.name === 'service_name')
  const pod = dashboard.variables.find((variable) => variable.name === 'pod')
  const job = dashboard.variables.find((variable) => variable.name === 'job_name')

  assert.equal(namespace.multi, true)
  assert.equal(namespace.includeAll, true)
  assert.equal(namespace.allValue, '.*')
  assert.equal(variableSelectMode(namespace), 'multiple')
  assert.equal(service.multi, true)
  assert.equal(service.includeAll, true)
  assert.equal(service.value, '.*')
  assert.equal(service.defaultValue, '.*')
  assert.equal(variableSelectMode(service), 'multiple')
  assert.equal(pod.multi, false)
  assert.equal(pod.includeAll, true)
  assert.equal(variableSelectMode(pod), undefined)
  assert.equal(job.multi, true)
  assert.equal(variableSelectMode(job), 'multiple')
  assert.equal(job.includeAll, undefined)
})

test('chart options apply unit and decimal formatting to numeric values', () => {
  const option = buildChartOption(
    { type: 'line', display: { unit: 'ms', decimals: 1, warningThreshold: 80, criticalThreshold: 95 }, query: { legend: '{{svc_name}} latency' } },
    { series: [{ metric: { svc_name: 'checkout' }, values: [{ TimeStamp: 1000, Value: 12.34 }] }] },
  )
  assert.equal(option.series[0].name, 'checkout latency')
  assert.deepEqual(option.series[0].markLine.data.map((item) => item.yAxis), [80, 95])
  assert.equal(option.yAxis.axisLabel.formatter(12.34), '12.3 ms')
  assert.equal(option.yAxis.axisLabel.formatter(1234567.89), '1.23M ms')
  assert.equal(option.yAxis.axisLabel.formatter(-9876.54), '-9.88K ms')
  assert.equal(option.tooltip.valueFormatter(12.34), '12.3 ms')
  assert.equal(option.tooltip.valueFormatter(1234567.89), '1.23M ms')
  assert.equal(option.grid.left, 8)
  assert.equal(option.grid.containLabel, true)
  assert.equal(formatWidgetValue(9.876, { unit: '%', decimals: 2 }), '9.88 %')
  assert.equal(formatWidgetValue('not numeric', { unit: 'ms', decimals: 1 }), 'not numeric')
  assert.equal(seriesDisplayName({ svc_name: 'checkout', pod: 'checkout-0' }, 0, '{{svc_name}} / {{pod}}'), 'checkout / checkout-0')
  assert.equal(seriesDisplayName({ svc_name: 'checkout' }, 0, '$svc_name'), 'checkout')
})

test('preset units format values while custom suffixes stay literal', () => {
  assert.equal(
    formatWidgetValue(1073741824, { unitPreset: 'bytes', unit: 'B', autoScale: true, decimals: 2 }),
    '1.00 GiB',
  )
  assert.equal(
    formatWidgetValue(71862456320, { unitPreset: 'bytes', unit: 'B', autoScale: true, decimals: 2 }),
    '66.93 GiB',
  )
  assert.equal(
    formatWidgetValue(0.8234, { unitPreset: 'percent', unit: '%', autoScale: true, decimals: 1 }),
    '82.3 %',
  )
  assert.equal(
    formatWidgetValue(82.34, { unitPreset: 'percent', unit: '%', autoScale: false, decimals: 1 }),
    '82.3 %',
  )
  assert.equal(
    formatWidgetValue(1234, { unitPreset: 'milliseconds', unit: 'ms', autoScale: true, decimals: 2 }),
    '1.23 s',
  )
  assert.equal(
    formatWidgetValue(0.42, { unitPreset: 'milliseconds', unit: 'ms', autoScale: true, decimals: 2 }),
    '420.00 us',
  )
  assert.equal(
    formatWidgetValue(0.0042, { unitPreset: 'seconds', unit: 's', autoScale: true, decimals: 2 }),
    '4.20 ms',
  )
  assert.equal(
    formatAxisValue(0.0000042, { unitPreset: 'seconds', unit: 's', autoScale: true, decimals: 2 }),
    '4.20 us',
  )
  assert.equal(
    formatWidgetValue(1073741824, { unitPreset: 'custom', unit: 'B', autoScale: true, decimals: 2 }),
    '1073741824.00 B',
  )
  assert.equal(formatWidgetValue('not numeric', { unitPreset: 'bytes', unit: 'B', autoScale: true }), 'not numeric')
  assert.equal(formatAxisValue(1073741824, { unitPreset: 'bytes', unit: 'B', autoScale: true, decimals: 2 }), '1.00 GiB')
})

test('legacy unit strings normalize to presets without changing percent semantics', () => {
  const dashboard = normalizeDashboard({
    widgets: [
      { id: 'legacy-bytes', type: 'stat', display: { unit: 'B', decimals: 2 } },
      { id: 'legacy-percent', type: 'line', display: { unit: '%', decimals: 1 } },
      { id: 'legacy-custom', type: 'table', display: { unit: 'req/s', decimals: 0 } },
    ],
  })

  assert.deepEqual(dashboard.widgets[0].display, {
    unit: 'B',
    unitPreset: 'bytes',
    autoScale: true,
    decimals: 2,
    warningThreshold: undefined,
    criticalThreshold: undefined,
  })
  assert.deepEqual(dashboard.widgets[1].display, {
    unit: '%',
    unitPreset: 'percent',
    autoScale: false,
    decimals: 1,
    legendMode: 'list',
    warningThreshold: undefined,
    criticalThreshold: undefined,
  })
  assert.deepEqual(dashboard.widgets[2].display, {
    unit: 'req/s',
    unitPreset: 'custom',
    autoScale: false,
    decimals: 0,
    warningThreshold: undefined,
    criticalThreshold: undefined,
  })
})

test('percent ratio preset formats chart values and converts displayed thresholds to raw values', () => {
  const option = buildChartOption(
    { type: 'line', display: { unitPreset: 'percent', unit: '%', autoScale: true, decimals: 1, warningThreshold: 80, criticalThreshold: 95 }, query: {} },
    { series: [{ metric: { svc_name: 'checkout' }, values: [{ TimeStamp: 1000, Value: 0.8234 }] }] },
  )

  assert.equal(option.yAxis.axisLabel.formatter(0.8234), '82.3 %')
  assert.equal(option.tooltip.valueFormatter(0.8234), '82.3 %')
  assert.deepEqual(option.series[0].markLine.data.map((item) => item.yAxis), [0.8, 0.95])
})

test('bar gauge progress follows percent ratio display units', () => {
  assert.equal(
    barGaugeProgressPercent(0.2683, { unitPreset: 'percent', unit: '%', autoScale: true, decimals: 2 }),
    26.83,
  )
  assert.equal(
    barGaugeProgressPercent(26.83, { unitPreset: 'percent', unit: '%', autoScale: false, decimals: 2 }),
    26.83,
  )
  assert.equal(barGaugeProgressPercent(125, { unitPreset: 'custom', unit: 'req/s' }), 100)
  assert.equal(barGaugeProgressPercent('not numeric', { unitPreset: 'percent', unit: '%', autoScale: true }), 0)
})

test('editor unit control supports preset selection, free-form custom suffixes, and preset-only behavior switches', () => {
  const editorSource = readFileSync(resolve(currentDir, 'Editor.jsx'), 'utf8')
  const utilsSource = readFileSync(resolve(currentDir, 'utils.mjs'), 'utf8')

  assert.match(utilsSource, /unitPresetOptions/)
  assert.match(editorSource, /AutoComplete/)
  assert.match(editorSource, /value:\s*preset\.unit/)
  assert.match(editorSource, /selectedUnitValue\(selectedWidget\.display\)/)
  assert.match(editorSource, /unitPresetOptions\.find/)
  assert.match(editorSource, /preset\.unit\.toLowerCase\(\) === unitValue\.toLowerCase\(\)/)
  assert.match(editorSource, /unitPreset/)
  assert.match(editorSource, /autoScale/)
  assert.match(editorSource, /unitAutoScaleLabel/)
  assert.match(editorSource, /unitAutoScaleHelp/)
  assert.match(editorSource, /unitRatioToPercent/)
  assert.doesNotMatch(editorSource, /mode="tags"/)
  assert.doesNotMatch(editorSource, /maxCount=\{1\}/)
})

test('series display names resolve query legend templates from response metadata', () => {
  assert.equal(
    seriesDisplayName({ pod: 'apisix-0', container: 'apisix', query_legend: '{{pod}}: {{ container }}' }, 0, ''),
    'apisix-0: apisix',
  )
})

test('series display names resolve imported Grafana auto legends from metric labels', () => {
  assert.equal(
    seriesDisplayName({
      __name__: 'kindling_span_trace_duration_nanoseconds_count',
      content_key: 'GET /**',
      query_id: 'A',
      query_legend: '__auto',
      svc_name: 'sandbox-default-user-3',
    }, 0, ''),
    '{content_key="GET /**", svc_name="sandbox-default-user-3"}',
  )
})

test('editor exposes auto legend as an option instead of internal text input', () => {
  const editorSource = readFileSync(resolve(currentDir, 'Editor.jsx'), 'utf8')
  const utilsSource = readFileSync(resolve(currentDir, 'utils.mjs'), 'utf8')
  const zh = JSON.parse(readFileSync(resolve(currentDir, '../../../../public/locales/zh/pro/customDashboard.json'), 'utf8'))
  const en = JSON.parse(readFileSync(resolve(currentDir, '../../../../public/locales/en/pro/customDashboard.json'), 'utf8'))

  assert.match(utilsSource, /export const autoLegendValue = '__auto'/)
  assert.match(editorSource, /Segmented/)
  assert.match(editorSource, /legendModeForValue\(value\)/)
  assert.match(editorSource, /value:\s*'auto',\s*label:\s*t\('legendAuto'\)/)
  assert.match(editorSource, /mode === 'auto' \? autoLegendValue : ''/)
  assert.match(editorSource, /renderLegendControl\(query\.legend,\s*\(legend\) => updateQueryItem\(query\.id,\s*\{ legend \}\)\)/)
  assert.match(editorSource, /renderLegendControl\(selectedWidget\.query\?\.legend,\s*\(legend\) => updateWidgetNested\('query',\s*\{ legend \}\)\)/)
  assert.equal(zh.legendAuto, '自动')
  assert.equal(zh.legendCustom, '自定义')
  assert.equal(en.legendAuto, 'Auto')
  assert.equal(en.legendCustom, 'Custom')
})

test('editor explains legend and threshold display controls', () => {
  const editorSource = readFileSync(resolve(currentDir, 'Editor.jsx'), 'utf8')
  const styleSource = readFileSync(resolve(currentDir, 'index.scss'), 'utf8')
  const zh = JSON.parse(readFileSync(resolve(currentDir, '../../../../public/locales/zh/pro/customDashboard.json'), 'utf8'))
  const en = JSON.parse(readFileSync(resolve(currentDir, '../../../../public/locales/en/pro/customDashboard.json'), 'utf8'))

  assert.match(editorSource, /tooltip=\{t\('legendHelp'\)\}/)
  assert.match(editorSource, /placeholder=\{t\('legendCustomPlaceholder'\)\}/)
  assert.match(editorSource, /legendMode === 'auto' \? 'legendAutoHelp' : 'legendCustomHelp'/)
  assert.match(editorSource, /tooltip=\{t\('legendDisplayHelp'\)\}/)
  assert.match(editorSource, /tooltip=\{t\('warningThresholdHelp'\)\}/)
  assert.match(editorSource, /tooltip=\{t\('criticalThresholdHelp'\)\}/)
  assert.match(styleSource, /\.custom-dashboard-field-hint/)
  assert.match(zh.legendCustomHelp, /\$label/)
  assert.match(zh.legendAutoHelp, /svc_name/)
  assert.match(zh.warningThresholdHelp, /黄色参考线/)
  assert.match(zh.criticalThresholdHelp, /红色参考线/)
  assert.match(en.legendCustomHelp, /\$label/)
  assert.match(en.legendAutoHelp, /svc_name/)
  assert.match(en.warningThresholdHelp, /yellow reference line/)
  assert.match(en.criticalThresholdHelp, /red reference line/)
})

test('chart tooltips render scan-friendly rows and escape series names', () => {
  const option = buildChartOption(
    { type: 'line', display: { unit: 'ms', decimals: 1 }, query: {} },
    { series: [{ metric: { svc_name: 'checkout' }, values: [{ TimeStamp: 1000, Value: 12 }] }] },
  )
  const tooltip = option.tooltip.formatter({
    name: '10:00:00',
    marker: '<span></span>',
    seriesName: 'checkout <api>',
    value: 1234567.89,
  })

  assert.equal(option.tooltip.trigger, 'item')
  assert.equal(option.tooltip.triggerOn, 'none')
  assert.equal(option.series[0].triggerLineEvent, false)
  assert.equal(option.series[0].showSymbol, true)
  assert.equal(option.series[0].showAllSymbol, false)
  assert.equal(option.series[0].symbolSize, 0)
  assert.equal(option.series[0].emphasis.disabled, true)
  assert.equal(typeof option.tooltip.position, 'function')
  assert.deepEqual(option.tooltip.position([0, 0], null, null, null, { contentSize: [180, 80], viewSize: [400, 240] }), [12, 12])
  assert.deepEqual(option.tooltip.position([390, 230], null, null, null, { contentSize: [180, 80], viewSize: [400, 240] }), [198, 138])
  assert.match(tooltip, /10:00:00/)
  assert.match(tooltip, /grid-template-columns:\s*minmax\(120px,\s*1fr\)\s*auto/)
  assert.match(tooltip, /checkout &lt;api&gt;/)
  assert.match(tooltip, /1\.23M ms/)
  assert.doesNotMatch(tooltip, /cart/)

  const customUnitOption = buildChartOption(
    { type: 'line', display: { unit: 'ns', decimals: 2 }, query: {} },
    { series: [{ metric: { svc_name: 'checkout' }, values: [{ TimeStamp: 1000, Value: 29346000 }] }] },
  )
  const customUnitTooltip = customUnitOption.tooltip.formatter({
    name: '10:00:00',
    marker: '<span></span>',
    seriesName: 'checkout',
    value: 29346000,
  })
  assert.match(customUnitTooltip, /29\.35M ns/)
  assert.doesNotMatch(customUnitTooltip, /29346000\.00 ns/)
})

test('line hover dispatches tooltip for the nearest point on the hovered series', () => {
  const actions = []
  const chart = {
    convertFromPixel: (scope, point) => {
      assert.equal(scope, 'grid')
      assert.deepEqual(point, [12, 34])
      return [1.1, 52]
    },
    convertToPixel: (scope, point) => {
      assert.equal(scope, 'grid')
      assert.deepEqual(point, [1, 50])
      return [12, 35]
    },
    getOption: () => {
      throw new Error('hover should use cached series data instead of reading chart options')
    },
    dispatchAction: (action) => actions.push(action),
  }
  const hoverSeries = [
    { seriesIndex: 0, data: [10, 20, 30], selected: true },
    { seriesIndex: 1, data: [40, 50, 60], selected: true, color: '#22c55e' },
  ]
  const hoverPoints = []

  const handled = showNearestLineTooltip({
    offsetX: 12,
    offsetY: 34,
  }, chart, hoverSeries, (point) => hoverPoints.push(point))

  assert.equal(handled, true)
  assert.deepEqual(actions, [
    { type: 'updateAxisPointer', currTrigger: 'mousemove', x: 12, y: 35 },
    { type: 'showTip', seriesIndex: 1, dataIndex: 1 },
  ])
  assert.deepEqual(hoverPoints, [{ x: 12, y: 35, color: '#22c55e', seriesIndex: 1, dataIndex: 1 }])
  assert.equal(showNearestLineTooltip({
    dataIndex: 1,
    offsetX: 12,
    offsetY: 34,
  }, chart, hoverSeries), false)

  const missActions = []
  const missPoints = []
  const missChart = {
    convertFromPixel: () => [1.1, 52],
    convertToPixel: () => [12, 90],
    dispatchAction: (action) => missActions.push(action),
  }
  assert.equal(showNearestLineTooltip({
    offsetX: 12,
    offsetY: 34,
  }, missChart, hoverSeries, (point) => missPoints.push(point)), false)
  assert.deepEqual(missPoints, [null])
  assert.deepEqual(missActions, [
    { type: 'updateAxisPointer', currTrigger: 'leave' },
    { type: 'hideTip' },
  ])
})

test('custom dashboard line hover work is batched to animation frames', () => {
  const rendererSource = readFileSync(resolve(currentDir, 'WidgetRenderer.jsx'), 'utf8')
  const styleSource = readFileSync(resolve(currentDir, 'index.scss'), 'utf8')

  assert.match(rendererSource, /tooltipFrameRef/)
  assert.match(rendererSource, /hoverPointRef/)
  assert.match(rendererSource, /showLineHoverPoint/)
  assert.match(rendererSource, /hideLineHoverPoint/)
  assert.match(rendererSource, /custom-dashboard-chart-hover-point/)
  assert.match(rendererSource, /requestAnimationFrame/)
  assert.match(rendererSource, /cancelAnimationFrame/)
  assert.match(rendererSource, /getZr\?\.\(\)/)
  assert.match(rendererSource, /\.on\('mousemove'/)
  assert.match(styleSource, /\.custom-dashboard-chart-hover-point/)
  assert.match(styleSource, /pointer-events:\s*none/)
  assert.match(styleSource, /translate\(-50%, -50%\)/)
})

test('line and area charts keep time brushing while bars opt out', () => {
  const option = buildChartOption(
    { type: 'area', display: { unit: '%', decimals: 1 }, query: { legend: '{{svc_name}}' } },
    {
      series: [
        {
          metric: { svc_name: 'checkout' },
          values: [
            { TimeStamp: 1_000_000, Value: 12.34 },
            { TimeStamp: 2_000_000, Value: 18.2 },
          ],
        },
      ],
    },
  )

  assert.equal(chartSupportsTimeBrush('line'), true)
  assert.equal(chartSupportsTimeBrush('area'), true)
  assert.equal(chartSupportsTimeBrush('bar'), false)
  assert.deepEqual(option.xAxis.customDashboardTimeValues, [1_000_000, 2_000_000])
  assert.deepEqual(option.toolbox, { show: false })
  assert.deepEqual(option.brush.toolbox, [])
  assert.equal(option.brush.xAxisIndex, 'all')
  assert.equal(option.brush.brushStyle.borderWidth, 1)

  const barOption = buildChartOption(
    { type: 'bar', display: { unit: '%', decimals: 1 }, query: { legend: '{{svc_name}}' } },
    {
      series: [
        {
          metric: { svc_name: 'checkout' },
          values: [
            { TimeStamp: 1_000_000, Value: 12.34 },
            { TimeStamp: 2_000_000, Value: 18.2 },
          ],
        },
      ],
    },
  )
  assert.equal(barOption.brush, undefined)
})

test('chart brush converts selected category pixels to a global time range', () => {
  const chart = {
    convertFromPixel: (scope, pixel) => {
      assert.deepEqual(scope, { xAxisIndex: 0 })
      return pixel === 12 ? 0.2 : 2.6
    },
  }

  assert.deepEqual(
    chartTimeRangeFromBrush(
      { areas: [{ brushType: 'lineX', range: [12, 42] }] },
      chart,
      [1_000_000, 2_000_000, 3_000_000, 4_000_000],
    ),
    { startTime: 1_000_000, endTime: 4_000_000 },
  )
  assert.equal(chartTimeRangeFromBrush({ areas: [] }, chart, [1_000_000]), null)
})

test('viewer and editor enable chart-level time brushing for supported chart types', () => {
  const viewerSource = readFileSync(resolve(currentDir, 'Viewer.jsx'), 'utf8')
  const editorSource = readFileSync(resolve(currentDir, 'Editor.jsx'), 'utf8')
  const rendererSource = readFileSync(resolve(currentDir, 'WidgetRenderer.jsx'), 'utf8')

  assert.match(viewerSource, /<WidgetRenderer[\s\S]*enableTimeBrush/)
  assert.match(editorSource, /<WidgetRenderer[\s\S]*enableTimeBrush/)
  assert.match(rendererSource, /useDispatch/)
  assert.match(rendererSource, /enableTimeBrush=\{enableTimeBrush && chartSupportsTimeBrush\(renderWidget\.type\)\}/)
  assert.match(rendererSource, /type:\s*'SET_TIMERANGE'/)
  assert.match(rendererSource, /brushEnd/)
  assert.match(rendererSource, /takeGlobalCursor/)
})

test('chart legend rows calculate Grafana-style series statistics', () => {
  const rows = chartLegendRows(
    { type: 'line', display: { unit: '%', decimals: 2 }, query: { legend: '{{svc_name}}' } },
    {
      series: [
        { metric: { svc_name: 'checkout' }, values: [{ Value: 1 }, { Value: 3 }, { Value: 5 }] },
        { metric: { svc_name: 'cart' }, values: [{ value: 2 }, { value: null }, { value: 4 }] },
        { metric: { svc_name: 'payment' }, values: [{ value: 1234567.89 }, { value: 2234567.89 }] },
      ],
    },
  )

  assert.deepEqual(rows.map((row) => row.name), ['checkout', 'cart', 'payment'])
  assert.equal(rows[0].min, '1.00 %')
  assert.equal(rows[0].max, '5.00 %')
  assert.equal(rows[0].mean, '3.00 %')
  assert.equal(rows[1].min, '2.00 %')
  assert.equal(rows[1].max, '4.00 %')
  assert.equal(rows[1].mean, '3.00 %')
  assert.equal(rows[2].min, '1.23M %')
  assert.equal(rows[2].max, '2.23M %')
  assert.equal(rows[2].mean, '1.73M %')
  assert.equal(rows[0].color, '#7eb26d')
  assert.equal(rows[1].color, '#eab839')
})

test('chart rendering uses a front-end TopN series cap for high-cardinality responses', () => {
  const series = Array.from({ length: 60 }, (_, index) => ({
    metric: { svc_name: `svc-${index + 1}` },
    values: [
      { TimeStamp: 1000, Value: index + 1 },
      { TimeStamp: 2000, Value: index + 0.5 },
    ],
  }))
  const widget = {
    type: 'line',
    display: { unit: 'ms', decimals: 1 },
    query: { legend: '{{svc_name}}' },
  }

  const option = buildChartOption(widget, { series })
  const legendRows = chartLegendRows(widget, { series })

  assert.equal(option.series.length, 50)
  assert.deepEqual(option.legend.data.slice(0, 3), ['svc-60', 'svc-59', 'svc-58'])
  assert.equal(option.legend.data.at(-1), 'svc-11')
  assert.equal(legendRows.length, 50)
  assert.deepEqual(legendRows.slice(0, 3).map((row) => row.name), ['svc-60', 'svc-59', 'svc-58'])
})

test('chart renderer exposes TopN controls for truncated high-cardinality panels', () => {
  const rendererSource = readFileSync(resolve(currentDir, 'WidgetRenderer.jsx'), 'utf8')

  assert.match(rendererSource, /Select/)
  assert.match(rendererSource, /seriesLimitOptions/)
  assert.match(rendererSource, /onSeriesLimitChange/)
  assert.match(rendererSource, /custom-dashboard-chart-series-limit-control/)
})

test('chart table legends paginate rows with visible controls outside the scroll body', () => {
  const rendererSource = readFileSync(resolve(currentDir, 'WidgetRenderer.jsx'), 'utf8')
  const styleSource = readFileSync(resolve(currentDir, 'index.scss'), 'utf8')

  assert.match(rendererSource, /Pagination/)
  assert.match(rendererSource, /tableLegendPageSize\s*=\s*10/)
  assert.match(rendererSource, /legendRowsForRender/)
  assert.match(rendererSource, /legendHasPagination/)
  assert.match(rendererSource, /legendBodySeriesLimitInfo/)
  assert.match(rendererSource, /legendFooterSeriesLimitInfo/)
  assert.match(rendererSource, /custom-dashboard-chart-legend-has-pagination/)
  assert.match(rendererSource, /className="custom-dashboard-chart-legend-footer"/)
  assert.match(rendererSource, /className="custom-dashboard-chart-legend-pagination"/)
  assert.match(rendererSource, /pageSize=\{tableLegendPageSize\}/)
  assert.match(rendererSource, /showSizeChanger=\{false\}/)
  assert.match(rendererSource, /simple/)
  assert.match(rendererSource, /legendRowsForRender\.map/)

  assert.match(styleSource, /\.custom-dashboard-chart-legend-table\.custom-dashboard-chart-legend-has-pagination[\s\S]*grid-template-rows:\s*auto minmax\(0,\s*1fr\) auto/)
  assert.match(styleSource, /\.custom-dashboard-chart-legend-footer[\s\S]*justify-content:\s*space-between/)
  assert.match(styleSource, /\.custom-dashboard-chart-legend-footer \.custom-dashboard-chart-series-limit[\s\S]*position:\s*static/)
  assert.match(styleSource, /\.custom-dashboard-chart-legend-pagination/)
  assert.match(styleSource, /justify-content:\s*flex-end/)
})

test('chart options use polished dashboard styling instead of ECharts defaults', () => {
  const option = buildChartOption(
    { type: 'area', display: { unit: '%', decimals: 1 }, query: { legend: '{{svc_name}}' } },
    {
      series: [
        { metric: { svc_name: 'checkout' }, values: [{ TimeStamp: 1000, Value: 12.34 }, { TimeStamp: 2000, Value: 18.2 }] },
        { metric: { svc_name: 'cart' }, values: [{ TimeStamp: 1000, Value: 9.1 }, { TimeStamp: 2000, Value: 13.4 }] },
      ],
    },
  )

  assert.deepEqual(option.color?.slice(0, 5), ['#7eb26d', '#eab839', '#6ed0e0', '#ef843c', '#e24d42'])
  assert.equal(option.tooltip.backgroundColor, 'rgba(255, 255, 255, 0.96)')
  assert.equal(option.tooltip.borderColor, '#d9e2ef')
  assert.deepEqual(option.tooltip.padding, [10, 12])
  assert.equal(option.tooltip.axisPointer.type, 'cross')
  assert.equal(option.tooltip.axisPointer.snap, true)
  assert.equal(option.tooltip.axisPointer.label.show, false)
  assert.equal(option.tooltip.axisPointer.lineStyle.type, 'dashed')
  assert.equal(option.tooltip.axisPointer.crossStyle.type, 'dashed')
  assert.deepEqual(option.toolbox, { show: false })
  assert.match(option.tooltip.extraCssText, /border-radius:\s*12px/)
  assert.match(option.tooltip.extraCssText, /box-shadow:\s*0 12px 32px rgba\(15, 23, 42, 0\.14\)/)
  assert.equal(option.grid.top, 12)
  assert.equal(option.grid.right, 12)
  assert.equal(option.grid.bottom, 24)
  assert.equal(option.legend.show, false)
  assert.deepEqual(option.legend.data, ['checkout', 'cart'])
  assert.deepEqual(option.legend.selected, { checkout: true, cart: true })
  assert.equal(option.legend.icon, 'circle')
  assert.equal(option.legend.itemWidth, 7)
  assert.equal(option.xAxis.axisLine.show, false)
  assert.equal(option.xAxis.axisTick.show, false)
  assert.equal(option.yAxis.splitLine.lineStyle.color, '#edf1f7')
  assert.equal(option.yAxis.splitLine.lineStyle.type, 'dashed')
  assert.equal(option.yAxis.scale, true)
  assert.equal(option.series[0].smooth, 0.35)
  assert.equal(option.series[0].lineStyle.width, 2.4)
  assert.equal(option.series[0].lineStyle.shadowBlur, undefined)
  assert.equal(option.series[0].lineStyle.shadowColor, undefined)
  assert.equal(option.series[0].lineStyle.shadowOffsetY, undefined)
  assert.equal(option.series[0].areaStyle.opacity, 0.24)
  assert.equal(option.series[0].emphasis.disabled, true)

  const barOption = buildChartOption(
    { type: 'bar', display: { unit: '', decimals: 0 }, query: {} },
    { series: [{ metric: { pod: 'checkout-0' }, values: [{ TimeStamp: 1000, Value: 12 }] }] },
  )
  assert.equal(barOption.tooltip.trigger, 'axis')
  assert.equal(barOption.tooltip.triggerOn, 'click')
  assert.equal(barOption.series[0].barMaxWidth, 22)
  assert.equal(barOption.tooltip.axisPointer.type, 'shadow')
  assert.equal(barOption.series[0].barCategoryGap, '48%')
  assert.equal(barOption.series[0].itemStyle.borderRadius[0], 8)
  assert.equal(barOption.series[0].itemStyle.borderRadius[2], 2)
  assert.equal(barOption.series[0].itemStyle.color, '#7eb26d')
  assert.equal(barOption.series[0].itemStyle.shadowBlur, undefined)
  assert.equal(barOption.series[0].emphasis.disabled, true)

  const selectedOption = buildChartOption(
    { type: 'line', display: { unit: '', decimals: 0 }, query: { legend: '{{svc_name}}' } },
    { series: [
      { metric: { svc_name: 'checkout' }, values: [{ TimeStamp: 1000, Value: 12 }] },
      { metric: { svc_name: 'cart' }, values: [{ TimeStamp: 1000, Value: 18 }] },
    ] },
    'light',
    { checkout: false },
  )
  assert.deepEqual(selectedOption.legend.selected, { checkout: false, cart: true })
})

test('line charts render as clean stroked trend lines without area fill', () => {
  const option = buildChartOption(
    { type: 'line', display: { unit: '%', decimals: 1 }, query: {} },
    { series: [{ metric: { svc_name: 'checkout' }, values: [{ TimeStamp: 1000, Value: 12 }, { TimeStamp: 2000, Value: 18 }] }] },
  )

  assert.equal(option.series[0].type, 'line')
  assert.equal(option.legend.show, false)
  assert.equal(option.series[0].smooth, 0.35)
  assert.equal(option.series[0].lineStyle.color, '#7eb26d')
  assert.equal(option.series[0].lineStyle.shadowBlur, undefined)
  assert.equal(option.series[0].lineStyle.shadowColor, undefined)
  assert.equal(option.series[0].areaStyle, undefined)
  assert.equal(option.series[0].itemStyle.borderColor, '#ffffff')
})

test('chart options follow the current light or dark dashboard theme', () => {
  const darkOption = buildChartOption(
    { type: 'area', display: { unit: '%', decimals: 1, warningThreshold: 80, criticalThreshold: 95 }, query: {} },
    { series: [{ metric: { svc_name: 'checkout' }, values: [{ TimeStamp: 1000, Value: 12.34 }] }] },
    'dark',
  )

  assert.equal(darkOption.tooltip.backgroundColor, 'rgba(15, 23, 42, 0.94)')
  assert.equal(darkOption.tooltip.borderColor, '#334155')
  assert.equal(darkOption.legend.textStyle.color, '#a9b4c7')
  assert.equal(darkOption.xAxis.axisLabel.color, '#8b98ad')
  assert.equal(darkOption.yAxis.splitLine.lineStyle.color, 'rgba(148, 163, 184, 0.14)')
  assert.equal(darkOption.yAxis.splitLine.lineStyle.type, 'dashed')
  assert.equal(darkOption.legend.show, false)
  assert.equal(darkOption.series[0].lineStyle.shadowBlur, undefined)
  assert.equal(darkOption.series[0].lineStyle.shadowColor, undefined)
  assert.equal(darkOption.series[0].areaStyle.opacity, 0.2)
  assert.equal(darkOption.series[0].markLine.data[0].lineStyle.color, '#e9b839')
  assert.equal(darkOption.series[0].markLine.data[1].lineStyle.color, '#e24d42')

  const rendererSource = readFileSync(resolve(currentDir, 'WidgetRenderer.jsx'), 'utf8')
  assert.match(rendererSource, /useSelector/)
  assert.match(rendererSource, /state\)\s*=>\s*state\.settingReducer\.theme/)
  assert.match(rendererSource, /buildChartOption\(renderWidget,\s*renderPreview,\s*themeMode,\s*selectedSeries\)/)
})

test('chart renderer replaces old ECharts series when preview series shrink', () => {
  const rendererSource = readFileSync(resolve(currentDir, 'WidgetRenderer.jsx'), 'utf8')

  assert.match(rendererSource, /<ReactECharts[\s\S]*notMerge=\{true\}/)
})

test('table widgets use compact dashboard table styling', () => {
  const rendererSource = readFileSync(resolve(currentDir, 'WidgetRenderer.jsx'), 'utf8')
  const styleSource = readFileSync(resolve(currentDir, 'index.scss'), 'utf8')
  const tableWidgetBlock = rendererSource.slice(
    rendererSource.indexOf("if (widget.type === 'table'"),
    rendererSource.indexOf("if (widget.type === 'stat'"),
  )

  assert.match(tableWidgetBlock, /className="custom-dashboard-table"/)
  assert.match(tableWidgetBlock, /tableColumnsForPreview/)
  assert.match(tableWidgetBlock, /tableRowsForPreview/)
  assert.match(rendererSource, /customDashboardTablePageSize\s*=\s*100/)
  assert.match(rendererSource, /customDashboardTableScrollY\s*=\s*'calc\(100% - 34px\)'/)
  assert.match(rendererSource, /customDashboardTableScroll\s*=\s*\{\s*x:\s*'max-content',\s*y:\s*customDashboardTableScrollY\s*\}/)
  assert.match(rendererSource, /customDashboardThumbnailTableScroll\s*=\s*\{\s*x:\s*'max-content'\s*\}/)
  assert.match(tableWidgetBlock, /rows\.length\s*>\s*customDashboardTablePageSize/)
  assert.match(tableWidgetBlock, /pageSize:\s*customDashboardTablePageSize/)
  assert.match(tableWidgetBlock, /pagination=\{tablePagination\}/)
  assert.match(tableWidgetBlock, /scroll=\{thumbnail \? customDashboardThumbnailTableScroll : customDashboardTableScroll\}/)
  assert.match(styleSource, /\.custom-dashboard-table/)
  assert.match(styleSource, /\.ant-spin-container\s*\{[\s\S]*display:\s*flex/)
  assert.match(styleSource, /\.ant-table\s*\{[\s\S]*flex:\s*1 1 auto/)
  assert.match(styleSource, /\.ant-table-content,\s*\n\s*\.ant-table-body\s*\{[\s\S]*scrollbar-width:\s*thin/)
  assert.match(styleSource, /\.ant-table-body\s*\{[\s\S]*overflow-y:\s*auto !important/)
  assert.match(styleSource, /\.ant-table-content::-webkit-scrollbar,\s*\n\s*\.ant-table-body::-webkit-scrollbar\s*\{[\s\S]*width:\s*6px/)
  assert.match(styleSource, /\.ant-table-content::-webkit-scrollbar-track,\s*\n\s*\.ant-table-body::-webkit-scrollbar-track\s*\{[\s\S]*background:\s*transparent/)
  assert.match(styleSource, /--ant-table-header-bg/)
  assert.match(styleSource, /font-variant-numeric:\s*tabular-nums/)
})

test('table widgets can expand metric labels into configured columns', () => {
  const preview = {
    series: [
      {
        metric: {
          __name__: 'jvm_cpu_count',
          instance: 'f3c47a8b-d372-43ca-bc33-25365a776aac',
          service_name: 'sandbox-default-user',
          service_version: '0.0.1-SNAPSHOT',
          telemetry_sdk_language: 'java',
          container_id: 'ac7ae91b20a3d772b',
          cluster_id: 'dev',
        },
        values: [{ TimeStamp: 1000, Value: 16 }],
      },
    ],
  }
  const widget = {
    type: 'table',
    query: {},
    display: {
      decimals: 2,
      tableColumns: [
        { name: 'service_name', label: '服务名' },
        { name: 'service_version', label: '版本' },
        { name: 'telemetry_sdk_language', label: '语言' },
        { name: 'container_id', label: 'Container Id' },
      ],
      tableHiddenColumns: ['__name__', 'instance'],
      showValueColumn: false,
    },
  }

  assert.deepEqual(tableColumnsForPreview(preview, widget.display), [
    { kind: 'label', name: 'service_name', title: '服务名' },
    { kind: 'label', name: 'service_version', title: '版本' },
    { kind: 'label', name: 'telemetry_sdk_language', title: '语言' },
    { kind: 'label', name: 'container_id', title: 'Container Id' },
    { kind: 'label', name: 'cluster_id', title: 'cluster_id' },
  ])
  assert.deepEqual(tableRowsForPreview(preview, widget), [
    {
      key: 0,
      labels: preview.series[0].metric,
      rawValue: 16,
      name: 'f3c47a8b-d372-43ca-bc33-25365a776aac',
      value: '16.00',
    },
  ])
  assert.deepEqual(tableColumnsForPreview(preview, { decimals: 2 }).map((column) => column.name), [
    'cluster_id',
    'container_id',
    'instance',
    'service_name',
    'service_version',
    'telemetry_sdk_language',
    'value',
  ])
})

test('line, area and bar widgets default legend display to list and preserve table selection', () => {
  assert.deepEqual(chartLegendModes.map((item) => item.value), ['list', 'table'])
  assert.equal(chartSupportsLegendMode('line'), true)
  assert.equal(chartSupportsLegendMode('area'), true)
  assert.equal(chartSupportsLegendMode('bar'), true)
  assert.equal(chartSupportsLegendMode('stat'), false)

  const dashboard = normalizeDashboard({
    widgets: [
      { type: 'line', display: {} },
      { type: 'area', display: { legendMode: 'table' } },
      { type: 'bar', display: { legendMode: 'unknown' } },
      { type: 'stat', display: { legendMode: 'table' } },
    ],
  })

  assert.equal(dashboard.widgets[0].display.legendMode, 'list')
  assert.equal(dashboard.widgets[1].display.legendMode, 'table')
  assert.equal(dashboard.widgets[2].display.legendMode, 'list')
  assert.equal(Object.hasOwn(dashboard.widgets[3].display, 'legendMode'), false)
})

test('editor renders legend display as compact radio buttons', () => {
  const editorSource = readFileSync(resolve(currentDir, 'Editor.jsx'), 'utf8')
  const settingsPanelStart = editorSource.indexOf('const widgetSettingsPanel =')
  const settingsPanelBlock = editorSource.slice(
    settingsPanelStart,
    editorSource.indexOf('\n  return (', settingsPanelStart),
  )
  const legendDisplayBlock = settingsPanelBlock.match(/<Form\.Item label=\{t\('legendDisplay'\)\}[\s\S]*?\n\s*<\/Form\.Item>/)?.[0] || ''

  assert.match(legendDisplayBlock, /<Radio\.Group/)
  assert.match(legendDisplayBlock, /optionType="button"/)
  assert.match(legendDisplayBlock, /options=\{chartLegendModes\.map/)
  assert.match(legendDisplayBlock, /onChange=\{\(event\) => updateWidgetNested\('display', \{ legendMode: event\.target\.value \}\)\}/)
  assert.doesNotMatch(legendDisplayBlock, /block/)
  assert.doesNotMatch(legendDisplayBlock, /<Segmented/)
  assert.doesNotMatch(legendDisplayBlock, /<Select/)
})

test('chart widgets support list and table legend modes outside the plot', () => {
  const rendererSource = readFileSync(resolve(currentDir, 'WidgetRenderer.jsx'), 'utf8')
  const styleSource = readFileSync(resolve(currentDir, 'index.scss'), 'utf8')

  assert.match(rendererSource, /chartLegendRows/)
  assert.match(rendererSource, /custom-dashboard-chart-panel/)
  assert.match(rendererSource, /custom-dashboard-chart-panel-\$\{legendMode\}/)
  assert.match(rendererSource, /custom-dashboard-chart-plot/)
  assert.match(rendererSource, /custom-dashboard-chart-legend/)
  assert.match(rendererSource, /custom-dashboard-chart-legend-\$\{legendMode\}/)
  assert.match(rendererSource, /legendMode\s*===\s*'table'/)
  assert.match(rendererSource, /custom-dashboard-chart-legend-header/)
  assert.match(rendererSource, /custom-dashboard-chart-legend-body/)
  assert.match(rendererSource, /custom-dashboard-chart-legend-row/)
  assert.match(rendererSource, /custom-dashboard-chart-legend-swatch/)
  assert.match(rendererSource, /custom-dashboard-chart-legend-value/)
  assert.match(rendererSource, /hiddenSeriesNames/)
  assert.match(rendererSource, /toggleLegendRow/)
  assert.match(rendererSource, /event\.ctrlKey\s*\|\|\s*event\.metaKey/)
  assert.match(rendererSource, /onClick=\{\(event\)\s*=>\s*toggleLegendRow\(row\.name,\s*event\.ctrlKey\s*\|\|\s*event\.metaKey\)\}/)
  assert.match(rendererSource, /selectedSeries/)
  assert.match(rendererSource, /aria-pressed/)
  assert.match(rendererSource, /buildChartOption\(renderWidget,\s*renderPreview,\s*themeMode,\s*selectedSeries\)/)
  assert.match(rendererSource, /t\('seriesName'\)/)
  assert.match(rendererSource, /t\('legendMin'\)/)
  assert.match(rendererSource, /t\('legendMax'\)/)
  assert.match(rendererSource, /t\('legendMean'\)/)
  assert.match(rendererSource, /row\.min/)
  assert.match(rendererSource, /row\.max/)
  assert.match(rendererSource, /row\.mean/)

  assert.match(styleSource, /\.custom-dashboard-widget-body[\s\S]*container-type:\s*inline-size/)
  assert.match(styleSource, /--custom-dashboard-widget-height:\s*clamp\(320px,\s*32vw,\s*440px\)/)
  assert.match(styleSource, /\.custom-dashboard-widget-body[\s\S]*padding:\s*6px/)
  assert.match(styleSource, /\.custom-dashboard-chart-panel/)
  assert.match(styleSource, /\.custom-dashboard-chart-panel-list/)
  assert.match(styleSource, /\.custom-dashboard-chart-panel-table/)
  assert.match(styleSource, /\.custom-dashboard-chart-legend/)
  assert.match(styleSource, /\.custom-dashboard-chart-legend-list/)
  assert.match(styleSource, /\.custom-dashboard-chart-legend-table/)
  assert.match(styleSource, /\.custom-dashboard-chart-panel[\s\S]*gap:\s*6px/)
  assert.doesNotMatch(styleSource, /\.custom-dashboard-chart-panel\s*\{[^}]*grid-template-rows:\s*minmax\(0,\s*1fr\)\s*auto/)
  assert.match(styleSource, /\.custom-dashboard-chart-legend[\s\S]*overflow:\s*hidden/)
  assert.doesNotMatch(styleSource, /\.custom-dashboard-chart-legend\s*\{[^}]*overflow:\s*auto/)
  assert.match(styleSource, /\.custom-dashboard-chart-legend-body[\s\S]*overflow:\s*auto/)
  assert.match(styleSource, /@container \(min-width:\s*760px\)/)
  assert.match(styleSource, /\.custom-dashboard-chart-panel-list[\s\S]*grid-template-columns:\s*minmax\(0,\s*1fr\)\s*minmax\(160px,\s*22%\)/)
  assert.match(styleSource, /\.custom-dashboard-chart-panel-table[\s\S]*grid-template-columns:\s*minmax\(0,\s*1fr\)\s*minmax\(260px,\s*32%\)/)
  assert.match(styleSource, /grid-template-rows:\s*minmax\(0,\s*1fr\)/)
  assert.match(styleSource, /border-left:\s*1px solid var\(--ant-color-border-secondary\)/)

  const editorSource = readFileSync(resolve(currentDir, 'Editor.jsx'), 'utf8')
  assert.match(editorSource, /const dashboardGridRowHeight = 320/)
  assert.match(editorSource, /rowHeight=\{dashboardGridRowHeight\}/)
  assert.match(editorSource, /chartSupportsLegendMode\(selectedWidget\.type\)/)
  assert.match(editorSource, /chartLegendModes\.map/)
  assert.match(editorSource, /t\('legendDisplay'\)/)
})

test('chart widgets resize ECharts after widget edit focus layout settles', () => {
  const rendererSource = readFileSync(resolve(currentDir, 'WidgetRenderer.jsx'), 'utf8')

  assert.match(rendererSource, /resizeEChartAfterLayout/)
  assert.match(rendererSource, /resizeFrameRef/)
  assert.match(rendererSource, /requestAnimationFrame/)
  assert.match(rendererSource, /cancelAnimationFrame/)
  assert.doesNotMatch(rendererSource, /setTimeout\(resize/)
  assert.match(rendererSource, /ResizeObserver/)
  assert.match(rendererSource, /observer\.observe\(element\)/)
  assert.match(rendererSource, /observer\.disconnect\(\)/)
  assert.match(rendererSource, /onChartReady=\{handleChartReady\}/)
  assert.match(rendererSource, /chartRef\.current\?\.getEchartsInstance\?\.\(\)/)
  assert.match(rendererSource, /width:\s*'100%'/)
  assert.match(rendererSource, /height:\s*'100%'/)
})

test('stat and bar gauge widgets render latest values', () => {
  const rendererSource = readFileSync(resolve(currentDir, 'WidgetRenderer.jsx'), 'utf8')
  const styleSource = readFileSync(resolve(currentDir, 'index.scss'), 'utf8')

  assert.match(rendererSource, /widget\.type === 'stat'/)
  assert.match(rendererSource, /widget\.type === 'barGauge'/)
  assert.match(rendererSource, /custom-dashboard-stat-grid/)
  assert.match(rendererSource, /custom-dashboard-stat-item/)
  assert.match(rendererSource, /custom-dashboard-stat-label/)
  assert.doesNotMatch(rendererSource, /custom-dashboard-stat-card/)
  assert.match(rendererSource, /custom-dashboard-bar-gauge/)
  assert.match(rendererSource, /custom-dashboard-bar-gauge-segments/)
  assert.match(rendererSource, /--bar-gauge-percent/)
  assert.match(rendererSource, /custom-dashboard-bar-gauge-value/)
  assert.match(rendererSource, /barGaugeProgressPercent\(row\.rawValue,\s*renderWidget\.display\)/)
  assert.match(styleSource, /\.custom-dashboard-stat-grid/)
  assert.match(styleSource, /\.custom-dashboard-stat-item/)
  assert.match(styleSource, /\.custom-dashboard-stat-label/)
  assert.doesNotMatch(styleSource, /\.custom-dashboard-stat-card/)
  assert.match(styleSource, /\.custom-dashboard-bar-gauge/)
  assert.match(styleSource, /repeating-linear-gradient/)
  assert.match(styleSource, /custom-dashboard-bar-gauge-value/)
})

test('dashboard normalization preserves widget legend configuration', () => {
  const dashboard = normalizeDashboard({
    widgets: [
      {
        id: 'legend',
        type: 'line',
        query: { mode: 'custom', expression: 'up{svc_name="checkout"}', legend: '{{svc_name}}' },
      },
    ],
  })
  assert.equal(dashboard.widgets[0].query.legend, '{{svc_name}}')
})

test('dashboard removes widget custom state text configuration', () => {
  const dashboard = normalizeDashboard({
    widgets: [
      {
        id: 'state-text',
        display: {
          emptyText: 'Nothing matched this service',
          errorText: 'Metric query failed',
        },
      },
    ],
  })

  assert.equal(dashboard.widgets[0].display.emptyText, undefined)
  assert.equal(dashboard.widgets[0].display.errorText, undefined)

  const editorSource = readFileSync(resolve(currentDir, 'Editor.jsx'), 'utf8')
  const rendererSource = readFileSync(resolve(currentDir, 'WidgetRenderer.jsx'), 'utf8')
  const zhLocaleSource = readFileSync(resolve(currentDir, '../../../../public/locales/zh/pro/customDashboard.json'), 'utf8')
  const enLocaleSource = readFileSync(resolve(currentDir, '../../../../public/locales/en/pro/customDashboard.json'), 'utf8')

  assert.doesNotMatch(editorSource, /emptyText|errorText/)
  assert.doesNotMatch(rendererSource, /widget\.display\?\.emptyText|widget\.display\?\.errorText/)
  assert.doesNotMatch(zhLocaleSource, /"emptyText"|"errorText"/)
  assert.doesNotMatch(enLocaleSource, /"emptyText"|"errorText"/)
})

test('dashboard normalization preserves widget threshold configuration', () => {
  const dashboard = normalizeDashboard({
    widgets: [
      {
        id: 'thresholds',
        display: {
          warningThreshold: 80,
          criticalThreshold: 95,
        },
      },
    ],
  })

  assert.equal(dashboard.widgets[0].display.warningThreshold, 80)
  assert.equal(dashboard.widgets[0].display.criticalThreshold, 95)

  const editorSource = readFileSync(resolve(currentDir, 'Editor.jsx'), 'utf8')
  assert.match(editorSource, /warningThreshold/)
  assert.match(editorSource, /criticalThreshold/)
  assert.match(editorSource, /const thresholdUnitLabel = selectedWidget \? selectedUnitValue\(selectedWidget\.display\) : ''/)
  assert.match(editorSource, /addonAfter=\{thresholdUnitLabel \|\| undefined\}/)
})

test('widget headers expose descriptions through tooltips', () => {
  const editorSource = readFileSync(resolve(currentDir, 'Editor.jsx'), 'utf8')
  const viewerSource = readFileSync(resolve(currentDir, 'Viewer.jsx'), 'utf8')

  assert.match(editorSource, /<Tooltip title=\{widget\.description\}/)
  assert.match(viewerSource, /<Tooltip title=\{widget\.description\}/)
  assert.match(editorSource, /custom-dashboard-widget-description/)
  assert.match(viewerSource, /custom-dashboard-widget-description/)
})

test('viewer keeps configured standard resource variables available as data group filters', () => {
  const variables = [
    { name: 'pod', type: 'pod' },
    { name: 'service', type: 'service' },
    { name: 'cluster', type: 'cluster' },
    { name: 'endpoint', type: 'endpoint' },
    { name: 'job_name', type: 'label', labels: ['job_name'] },
    { name: 'dataGroup', type: 'dataGroup' },
  ]
  const widgets = [
    { query: { mode: 'custom', expression: 'sum by (svc_name) (rate(http_requests_total{svc_name="checkout",cluster_id=~".*"}[1m]))' } },
  ]
  assert.deepEqual(
    enabledVariablesForDashboard(variables, widgets).map((variable) => variable.name),
    ['pod', 'service', 'cluster', 'endpoint'],
  )
})

test('data group aliases are reserved page context variables', () => {
  for (const name of ['dataGroup', 'groupId', 'group_id', 'data_group_id']) {
    assert.equal(isGlobalPageContextVariable({ name }), true)
    assert.deepEqual(enabledVariablesForDashboard([
      { name, type: 'custom' },
      { name: 'namespace', type: 'namespace', labels: ['namespace'] },
    ], [
      { query: { mode: 'custom', expression: 'up{namespace="$namespace"}' } },
    ]).map((item) => item.name), ['namespace'])
  }
})

test('legacy data group variables are normalized out before render and save', () => {
  const dashboard = normalizeDashboard({
    variables: [
      { name: 'dataGroup', type: 'dataGroup' },
      { name: 'timeRange', type: 'timeRange' },
      { name: 'service', type: 'service' },
    ],
    widgets: [],
  })

  assert.deepEqual(dashboard.variables.map((variable) => variable.name), ['service'])
  assert.deepEqual(serializeDashboard(dashboard).variables.map((variable) => variable.name), ['service'])
  assert.deepEqual(dashboardLabelOptions([
    { query: { mode: 'custom', expression: 'up{group_id="7",data_group="prod",time_range="1h"}' } },
  ]), [])
})

test('viewer keeps variables used by other variable queries', () => {
  const variables = [
    { name: 'service_type', type: 'custom' },
    {
      name: 'namespace',
      type: 'label',
      labels: ['namespace'],
      query: 'query_result(group by (namespace) (container_memory_rss{namespace=~"$service_type"}))',
    },
  ]
  const widgets = [
    { query: { mode: 'custom', expression: 'sum(container_memory_rss{namespace="$namespace"})' } },
  ]

  assert.deepEqual(
    enabledVariablesForDashboard(variables, widgets).map((variable) => variable.name),
    ['service_type', 'namespace'],
  )
})

test('editor blocks deleting variables referenced by widget or variable queries', () => {
  const variables = [
    { name: 'namespace', type: 'namespace', labels: ['namespace'] },
    { name: 'service_type', type: 'custom' },
    {
      name: 'pod',
      type: 'pod',
      query: 'label_values(container_cpu_usage_seconds_total{namespace=~"$namespace"}, pod)',
    },
    { name: 'owner', type: 'label', labels: ['owner'] },
  ]
  const widgets = [
    { query: { mode: 'custom', expression: 'sum(container_memory_rss{namespace="$namespace"})' } },
  ]

  assert.deepEqual(variableDeletionState(variables[0], variables, widgets), {
    disabled: true,
    usageCount: 2,
  })
  assert.deepEqual(variableDeletionState(variables[1], variables, widgets), {
    disabled: false,
    usageCount: 0,
  })
  assert.deepEqual(variableDeletionState(variables[3], variables, widgets), {
    disabled: false,
    usageCount: 0,
  })

  const editorSource = readFileSync(resolve(currentDir, 'Editor.jsx'), 'utf8')
  assert.match(editorSource, /variableDeletionState\(variable,\s*dashboard\?\.variables\s*\|\|\s*\[\],\s*dashboard\?\.widgets\s*\|\|\s*\[\],\s*presets\)/)
  assert.match(editorSource, /disabled=\{deleteState\.disabled\}/)
  assert.match(editorSource, /<Tooltip title=\{deleteState\.disabled\s*\?\s*t\('variableDeleteBlocked',\s*\{\s*count:\s*deleteState\.usageCount\s*\}\)\s*:\s*t\('variableDeleteAllowed'\)\}/)
})

test('viewer keeps resolution when dashboard uses grafana interval macros', () => {
  const variables = [
    { name: 'resolution', type: 'custom' },
  ]
  const widgets = [
    { query: { mode: 'custom', expression: 'sum(rate(container_cpu_usage_seconds_total[$__rate_interval]))' } },
  ]

  assert.deepEqual(
    enabledVariablesForDashboard(variables, widgets).map((variable) => variable.name),
    ['resolution'],
  )
})

test('variable candidate errors distinguish missing values from missing permissions', () => {
  const t = (key) => `t:${key}`

  assert.equal(variableCandidateErrorMessage(t, null), 't:noAvailableValues')
  assert.equal(variableCandidateErrorMessage(t, { response: { data: { code: 'A0006' } } }), 't:noVariablePermission')
  assert.equal(variableCandidateErrorMessage(t, { response: { data: { message: 'backend detail' } } }), 'backend detail')
  assert.equal(variableCandidateLabel(t, { type: 'dataGroup', value: '0', label: 'All Data' }), 't:allData')
  assert.equal(variableCandidateLabel(t, { type: 'service', value: 'checkout', label: 'checkout' }), 'checkout')

  const viewerSource = readFileSync(resolve(currentDir, 'Viewer.jsx'), 'utf8')
  assert.match(viewerSource, /variableCandidateErrorMessage\(t,\s*error\)/)
  assert.match(viewerSource, /variableSelectOptions\(t,\s*variable,\s*candidates\[variable\.name\]\s*\|\|\s*\[\]\)/)
})

test('variable discovery includes preset query labels and arbitrary label variables', () => {
  const widgets = [
    { query: { mode: 'preset', presetId: 'service-qps', params: { service: '$service' } } },
    { query: { mode: 'custom', expression: 'sum(rate(api_calls_total{content_key=~".*"}[1m]))' } },
  ]
  const presets = [
    { id: 'service-qps', expression: 'sum(rate(http_requests_total{svc_name="$service"}[1m]))' },
  ]
  assert.deepEqual(dashboardLabelOptions(widgets, presets), [
    { label: 'Service', value: 'service', type: 'service', labels: ['svc_name'] },
    { label: 'Endpoint', value: 'endpoint', type: 'endpoint', labels: ['content_key'] },
  ])
  assert.deepEqual(
    enabledVariablesForDashboard(
      [{ name: 'service', type: 'service' }, { name: 'endpoint', type: 'endpoint' }],
      widgets,
      presets,
    ).map((variable) => variable.name),
    ['service', 'endpoint'],
  )
})

test('promql resource variable hints classify supported, recognized, and ordinary labels', () => {
  const hints = promQLResourceVariableHints(
    'sum(up{clusterId="prod",namespace="default",clusterID2="typo",job="api"})',
  )

  assert.deepEqual(hints.recognized, [
    { name: 'cluster', label: 'Cluster', type: 'cluster', labels: ['clusterId'], snippet: 'cluster_id=~"$cluster"' },
    { name: 'namespace', label: 'Namespace', type: 'namespace', labels: ['namespace'], snippet: 'namespace=~"$namespace"' },
  ])
  assert.deepEqual(hints.unrecognized, [
    { label: 'clusterID2' },
    { label: 'job' },
  ])
  assert.deepEqual(hints.supported.find((item) => item.name === 'cluster')?.aliases, [
    'cluster',
    'cluster_id',
    'clusterId',
    'clusterID',
    'cluster_name',
  ])
  assert.equal(hints.supported.find((item) => item.name === 'service')?.snippet, 'svc_name=~"$service"')
})

test('promql resource variable hints keep parsing labels after grafana regex variables', () => {
  const hints = promQLResourceVariableHints(
    'sum(rate(container_cpu_usage_seconds_total{namespace="$namespace",pod=~"${pod:regex}",cluster_id=~"${cluster:regex}"}[$__rate_interval]))',
  )

  assert.deepEqual(hints.recognized, [
    { name: 'cluster', label: 'Cluster', type: 'cluster', labels: ['cluster_id'], snippet: 'cluster_id=~"$cluster"' },
    { name: 'namespace', label: 'Namespace', type: 'namespace', labels: ['namespace'], snippet: 'namespace=~"$namespace"' },
    { name: 'pod', label: 'Pod', type: 'pod', labels: ['pod'], snippet: 'pod=~"$pod"' },
  ])
  assert.deepEqual(hints.unrecognized, [])
})

test('promql resource variable hints recognize source namespace and pod labels', () => {
  const hints = promQLResourceVariableHints(
    'kindling_network_rtt{src_namespace=~"${namespace:regex}",src_pod=~"${pod:regex}",cluster_id=~"${cluster:regex}"}',
  )

  assert.deepEqual(hints.recognized, [
    { name: 'cluster', label: 'Cluster', type: 'cluster', labels: ['cluster_id'], snippet: 'cluster_id=~"$cluster"' },
    { name: 'namespace', label: 'Namespace', type: 'namespace', labels: ['src_namespace'], snippet: 'namespace=~"$namespace"' },
    { name: 'pod', label: 'Pod', type: 'pod', labels: ['src_pod'], snippet: 'pod=~"$pod"' },
  ])
  assert.deepEqual(hints.unrecognized, [])
})

test('promql resource variable hints explain fixed regex filters', () => {
  const hints = promQLResourceVariableHints(
    'irate(node_disk_reads_completed_total{instance_name=~"${related_nodes:regex}",cluster_id=~"${cluster_id:regex}",job_name="$job_name",device=~"[a-z]+|nvme[0-9]+n[0-9]+"}[$__rate_interval])',
  )

  assert.deepEqual(hints.recognized, [
    { name: 'cluster', label: 'Cluster', type: 'cluster', labels: ['cluster_id'], snippet: 'cluster_id=~"$cluster"' },
    { name: 'instance', label: 'Instance', type: 'instance', labels: ['instance_name'], snippet: 'instance_name=~"$instance"' },
  ])
  assert.deepEqual(hints.fixed, [
    { label: 'device', operator: '=~', value: '[a-z]+|nvme[0-9]+n[0-9]+' },
  ])
  assert.deepEqual(hints.unrecognized, [
    { label: 'job_name' },
  ])
})

test('promql matcher snippet insertion preserves existing selectors and grafana variables', () => {
  assert.equal(insertPromQLMatcherSnippet('', 'cluster_id=~"$cluster"'), 'up{cluster_id=~"$cluster"}')
  assert.equal(insertPromQLMatcherSnippet('up', 'cluster_id=~"$cluster"'), 'up{cluster_id=~"$cluster"}')
  assert.equal(
    insertPromQLMatcherSnippet('up{namespace=~"${namespace:regex}"}', 'cluster_id=~"$cluster"'),
    'up{namespace=~"${namespace:regex}",cluster_id=~"$cluster"}',
  )
})

test('custom promql editor exposes lightweight variable hints without switching to Monaco', () => {
  const editorSource = readFileSync(resolve(currentDir, 'Editor.jsx'), 'utf8')
  const styleSource = readFileSync(resolve(currentDir, 'index.scss'), 'utf8')
  const enMessages = readFileSync(resolve(currentDir, '../../../../public/locales/en/pro/customDashboard.json'), 'utf8')
  const zhMessages = readFileSync(resolve(currentDir, '../../../../public/locales/zh/pro/customDashboard.json'), 'utf8')

  assert.match(editorSource, /Input\.TextArea/)
  assert.match(editorSource, /promQLResourceVariableHints\(query\.expression\)/)
  assert.match(editorSource, /insertQuerySnippet/)
  assert.match(editorSource, /custom-dashboard-promql-hints/)
  assert.match(editorSource, /custom-dashboard-promql-supported/)
  assert.match(editorSource, /custom-dashboard-promql-recognized/)
  assert.match(editorSource, /custom-dashboard-promql-fixed/)
  assert.match(editorSource, /custom-dashboard-promql-unrecognized/)
  assert.doesNotMatch(editorSource, /MonacoEditorWrapper|monaco-editor|@monaco-editor/)

  assert.match(styleSource, /\.custom-dashboard-promql-hints/)
  assert.match(styleSource, /\.custom-dashboard-promql-chip/)
  assert.match(editorSource, /renderPromQLHintTitle/)
  assert.match(editorSource, /supportedParametersHelp/)
  assert.match(editorSource, /fixedFiltersHelp/)
  assert.match(enMessages, /"supportedParameters"/)
  assert.match(enMessages, /"recognizedParameters"/)
  assert.match(enMessages, /"fixedFilters"/)
  assert.match(enMessages, /"unrecognizedParameters"/)
  assert.match(enMessages, /"supportedParameters": "Insertable resource filters"/)
  assert.match(enMessages, /"recognizedParameters": "Bound in this query"/)
  assert.match(enMessages, /"fixedFilters": "Fixed query conditions"/)
  assert.match(enMessages, /"unrecognizedParameters": "Metric labels"/)
  assert.match(enMessages, /"fixedFiltersHelp"/)
  assert.match(zhMessages, /"supportedParameters"/)
  assert.match(zhMessages, /"recognizedParameters"/)
  assert.match(zhMessages, /"fixedFilters"/)
  assert.match(zhMessages, /"unrecognizedParameters"/)
  assert.match(zhMessages, /"supportedParameters": "可插入的资源筛选"/)
  assert.match(zhMessages, /"recognizedParameters": "当前查询已绑定"/)
  assert.match(zhMessages, /"fixedFilters": "查询内固定条件"/)
  assert.match(zhMessages, /"unrecognizedParameters": "普通指标标签"/)
  assert.match(zhMessages, /"fixedFiltersHelp"/)
})

test('custom promql editor uses padded scan-friendly query sections', () => {
  const editorSource = readFileSync(resolve(currentDir, 'Editor.jsx'), 'utf8')
  const styleSource = readFileSync(resolve(currentDir, 'index.scss'), 'utf8')
  const queryPanelBlock = editorSource.slice(
    editorSource.indexOf('const widgetQueryPanel ='),
    editorSource.indexOf('const widgetSettingsPanel ='),
  )

  assert.match(queryPanelBlock, /<Form layout="vertical" className="custom-dashboard-widget-query-form">/)
  assert.match(queryPanelBlock, /<Space className="custom-dashboard-query-list" direction="vertical">/)
  assert.match(queryPanelBlock, /className="custom-dashboard-query-item-header"/)
  assert.match(queryPanelBlock, /className="custom-dashboard-query-title"/)
  assert.match(queryPanelBlock, /className="custom-dashboard-query-expression"/)
  assert.match(queryPanelBlock, /className="custom-dashboard-add-query-button"/)
  assert.match(styleSource, /\.custom-dashboard-widget-query-panel[\s\S]*padding:\s*12px/)
  assert.match(styleSource, /\.custom-dashboard-widget-query-panel-inner[\s\S]*padding:\s*14px 16px 16px/)
  assert.match(styleSource, /\.custom-dashboard-widget-query-form[\s\S]*gap:\s*12px/)
  assert.match(styleSource, /\.custom-dashboard-query-item[\s\S]*padding:\s*14px/)
  assert.match(styleSource, /\.custom-dashboard-promql-hints[\s\S]*padding:\s*10px 2px 2px/)
  assert.match(styleSource, /\.custom-dashboard-promql-chip[\s\S]*padding:\s*3px 8px/)
})
