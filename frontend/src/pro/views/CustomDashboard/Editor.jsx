/**
 * Copyright 2026 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { useDebounce } from 'react-use'
import { Alert, AutoComplete, Button, Dropdown, Empty, Form, Input, InputNumber, Modal, Radio, Segmented, Select, Space, Spin, Splitter, Switch, Tooltip, Typography, message } from 'antd'
import GridLayout, { WidthProvider } from 'react-grid-layout/legacy'
import { LuArrowLeft, LuCopy, LuEllipsisVertical, LuEye, LuEyeOff, LuGripVertical, LuInfo, LuPlus, LuSave, LuSettings, LuTrash2 } from 'react-icons/lu'
import { useNavigate, useParams, useSearchParams } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useSelector } from 'react-redux'
import { getDatasourceFilterApiV2 } from 'src/core/api/dataGroup'
import { selectSecondsTimeRange } from 'src/core/store/reducers/timeRangeReducer'
import { getTeamsApi } from 'src/core/api/team'
import { useUserContext } from 'src/core/contexts/UserContext'
import {
  getCustomDashboard,
  getCustomDashboardPresetQueries,
  getCustomDashboardVariableCandidates,
  previewCustomDashboardQuery,
  updateCustomDashboard,
} from 'src/pro/api/customDashboard'
import WidgetRenderer from './WidgetRenderer'
import CustomDashboardResourceFilter from './CustomDashboardResourceFilter'
import {
  buildPreviewRequest,
  buildResourceVariableCandidatePlan,
  buildVariableCandidateRequest,
  createDashboardPreviewScheduler,
  createVariableCandidateScheduler,
  addWidgetToOpenCell,
  addWidgetPlaceholderCopy,
  addWidgetPlaceholderMode,
  applyWidgetDraftToDashboard,
  arrangeWidgetLayouts,
  cloneWidgetDraft,
  copyWidgetTitle,
  dashboardLoadFailureMessage,
  dashboardVariableReadinessScopeKey,
  dashboardVariableOptionGroups,
  dashboardVariableDefinition,
  dashboardQueryText,
  editorPanelMode,
  enabledVariablesForDashboard,
  findOpenGridLayout,
  isStandardResourceVariable,
  insertPromQLMatcherSnippet,
  labelsForVariable,
  mergeWidgetGridLayouts,
  mergePresetDefaults,
  normalizeGridLayout,
  presetOptionsForChart,
  presetParameterDescription,
  presetParameterLabel,
  previewRequestSignature,
  previewStateKind,
  promQLResourceVariableHints,
  reconcileRenderedChartWidgets,
  resourceVariableCandidatesFromTree,
  responsiveGridColumns,
  runDashboardMutationWithPermissionRefresh,
  saveFailureMessage,
  shouldPromptWidgetDraftDiscard,
  shouldRenderWidgetChart,
  shouldRefreshAllWidgetPreviews,
  shouldScheduleWidgetPreview,
  timeRangeRefreshDependency,
  variableCandidatesForDefinition,
  variableCandidatesWithAll,
  variableCandidateErrorMessage,
  variableDeletionState,
  variableDisplayLabel,
  normalizeVariableSelection,
  syncVariableSelectionWithCandidates,
  variableSelectMode,
  variableSelectOptions,
  variableValuesForQuery,
  widgetDisplayTitle,
  widgetDraftMode,
  widgetPreviewEligibilityForVisibility,
  widgetPreviewVisibilityFromRects,
  widgetUsesVariable,
  WIDGET_PREVIEW_PRIORITY,
} from './behavior.mjs'
import {
  autoLegendValue,
  chartLegendModes,
  chartSeriesLimitForWidget,
  chartSupportsLegendMode,
  chartTypes,
  defaultWidget,
  displayForUnitPreset,
  normalizeDashboard,
  normalizeChartSeriesLimit,
  normalizeWidgetQuery,
  queryItemsForQuery,
  serializeDashboard,
  unitPresetOptions,
  unitPresetSupportsAutoScale,
} from './utils.mjs'
import 'react-grid-layout/css/styles.css'
import 'react-resizable/css/styles.css'
import './index.scss'

const gridColumnCount = 4
const DashboardGridLayout = WidthProvider(GridLayout)
const dashboardGridRowHeight = 320
const dashboardGridMargin = [12, 12]
const addWidgetGridItemID = '__custom-dashboard-add-widget__'
const emptyWidgetPreview = { status: 'success', series: [] }
const widgetFocusLayoutStorageKey = 'customDashboard.widgetFocusLayout'
const defaultWidgetFocusLayout = { settingsWidth: 360, previewHeight: '48%' }
const widgetPreviewPreloadMarginPx = 320
const maxRenderedChartWidgets = 12
const renderedChartMinHoldMs = 2500
const draftPreviewKeyForWidget = (widget) => (widget?.id ? `draft:${widget.id}` : '')
const warnPermissionRefreshFailure = (error) => {
  console.warn('[custom-dashboard] Failed to refresh permission menu after dashboard mutation.', error)
}

const toDashboardGridItem = (id, layout = {}, columns = gridColumnCount, options = {}) => {
  const normalized = normalizeGridLayout(layout, columns)
  const safeColumns = Math.max(1, Number(columns) || gridColumnCount)
  return {
    i: id,
    x: normalized.x,
    y: normalized.y,
    w: normalized.w,
    h: normalized.h,
    minW: 1,
    minH: 1,
    maxW: Math.max(1, safeColumns - normalized.x),
    ...options,
  }
}

const clampNumber = (value, min, max) => Math.min(max, Math.max(min, value))

const renderedChartWidgetsEqual = (current = {}, next = {}) => {
  const currentKeys = Object.keys(current)
  const nextKeys = Object.keys(next)
  if (currentKeys.length !== nextKeys.length) return false
  return currentKeys.every((key) => (
    next[key]
    && current[key]?.renderedAt === next[key].renderedAt
    && current[key]?.lastVisibleAt === next[key].lastVisibleAt
  ))
}

const normalizeWidgetFocusLayout = (layout = {}) => {
  const settingsWidth = Number(layout.settingsWidth)
  const previewHeight = Number(layout.previewHeight)
  return {
    settingsWidth: Number.isFinite(settingsWidth)
      ? clampNumber(settingsWidth, 280, 560)
      : defaultWidgetFocusLayout.settingsWidth,
    previewHeight: Number.isFinite(previewHeight)
      ? Math.max(280, previewHeight)
      : defaultWidgetFocusLayout.previewHeight,
  }
}

const readWidgetFocusLayout = () => {
  if (typeof window === 'undefined') return defaultWidgetFocusLayout
  try {
    return normalizeWidgetFocusLayout(JSON.parse(window.localStorage.getItem(widgetFocusLayoutStorageKey) || '{}'))
  } catch {
    return defaultWidgetFocusLayout
  }
}

const writeWidgetFocusLayout = (layout) => {
  if (typeof window === 'undefined') return
  try {
    window.localStorage.setItem(widgetFocusLayoutStorageKey, JSON.stringify(layout))
  } catch {
    // Ignore unavailable local storage; resizing should still work for this session.
  }
}

const variableSelectionForMode = (value, multi) => {
  if (multi) {
    if (Array.isArray(value)) return value
    return value === undefined || value === null || value === '' ? [] : [value]
  }
  if (Array.isArray(value)) return value[0] ?? undefined
  return value ?? undefined
}

const CustomDashboardEditor = () => {
  const { id } = useParams()
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const { t } = useTranslation('pro/customDashboard')
  const { getUserPermission } = useUserContext()
  const widgetIDFromSearch = searchParams.get('widgetId') || ''
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [dashboard, setDashboard] = useState(null)
  const [selectedWidgetID, setSelectedWidgetID] = useState('')
  const [dashboardPanelOpen, setDashboardPanelOpen] = useState(false)
  const [draftWidget, setDraftWidget] = useState(null)
  const [widgetDraftDirty, setWidgetDraftDirty] = useState(false)
  const [widgetLeavePromptOpen, setWidgetLeavePromptOpen] = useState(false)
  const [presets, setPresets] = useState([])
  const [teams, setTeams] = useState([])
  const [variables, setVariables] = useState({})
  const [debouncedVariables, setDebouncedVariables] = useState({})
  const [openMultiVariableSelectors, setOpenMultiVariableSelectors] = useState({})
  const [candidates, setCandidates] = useState({})
  const [candidateErrors, setCandidateErrors] = useState({})
  const [previews, setPreviews] = useState({})
  const [widgetPreviewStates, setWidgetPreviewStates] = useState({})
  const [widgetVisibilities, setWidgetVisibilities] = useState({})
  const [renderedChartWidgets, setRenderedChartWidgets] = useState({})
  const [widgetSeriesLimits, setWidgetSeriesLimits] = useState({})
  const [dashboardVariablesReadyKey, setDashboardVariablesReadyKey] = useState('')
  const [dashboardScrolling, setDashboardScrolling] = useState(false)
  const [dirty, setDirty] = useState(false)
  const widgetElementsRef = useRef({})
  const widgetElementRefCallbacksRef = useRef({})
  const widgetVisibilityRef = useRef({})
  const observationAvailableRef = useRef(true)
  const editorBodyRef = useRef(null)
  const dashboardScrollIdleTimerRef = useRef(null)
  const dashboardScrollVisibilityFrameRef = useRef(null)
  const dashboardScrollingRef = useRef(false)
  const refreshWidgetVisibilitiesRef = useRef(null)
  const draftWidgetRef = useRef(null)
  const hadDraftWidgetRef = useRef(false)
  const previousVariablesRef = useRef({})
  const previewsRef = useRef({})
  const widgetPreviewStatesRef = useRef({})
  const widgetSeriesLimitsRef = useRef({})
  const previewRequestKeysRef = useRef({})
  const candidateRequestKeysRef = useRef({})
  const variableCandidateCacheRef = useRef({})
  const resourceTreeCacheRef = useRef({})
  const resourceTreeScopeKeyRef = useRef('')
  const dashboardVariableScopeKeyRef = resourceTreeScopeKeyRef
  const schedulerRef = useRef(null)
  const variableCandidateSchedulerRef = useRef(null)
  const previewing = false
  const [saveError, setSaveError] = useState('')
  const [loadError, setLoadError] = useState('')
  const [gridColumns, setGridColumns] = useState(gridColumnCount)
  const [widgetFocusLayout, setWidgetFocusLayout] = useState(() => readWidgetFocusLayout())
  const globalTimeRange = useSelector(selectSecondsTimeRange)
  const { dataGroupId } = useSelector((state) => state.dataGroupReducer)

  if (!schedulerRef.current) {
    schedulerRef.current = createDashboardPreviewScheduler()
  }
  if (!variableCandidateSchedulerRef.current) {
    variableCandidateSchedulerRef.current = createVariableCandidateScheduler()
  }

  const updateVariableSelectorOpen = useCallback((variable, open) => {
    if (variableSelectMode(variable) !== 'multiple') return
    setOpenMultiVariableSelectors((prev) => {
      if (open) return { ...prev, [variable.name]: true }
      const next = { ...prev }
      delete next[variable.name]
      return next
    })
  }, [])

  useEffect(() => {
    draftWidgetRef.current = draftWidget
  }, [draftWidget])

  const variablesSignature = JSON.stringify(variables)
  const debouncedVariablesSignature = JSON.stringify(debouncedVariables)
  const widgetVars = useMemo(() => debouncedVariables, [debouncedVariables])
  const queryText = useMemo(() => dashboardQueryText(dashboard?.widgets || [], presets), [dashboard?.widgets, presets])
  const selectedWidget = useMemo(
    () => draftWidget || dashboard?.widgets?.find((widget) => widget.id === selectedWidgetID),
    [dashboard, selectedWidgetID, draftWidget],
  )
  const arrangedLayouts = useMemo(
    () => arrangeWidgetLayouts(dashboard?.widgets || [], gridColumns),
    [dashboard?.widgets, gridColumns],
  )
  const variableOptionGroups = useMemo(
    () => dashboardVariableOptionGroups(dashboard?.widgets || [], presets, dashboard?.variables || [])
      .map((group) => ({
        label: t(group.labelKey),
        options: group.options.map((option) => ({
          ...option,
          label: option.type === 'label' ? option.label : t(option.value),
        })),
      })),
    [dashboard?.widgets, dashboard?.variables, presets, t],
  )
  const flatVariableOptions = useMemo(
    () => variableOptionGroups.flatMap((group) => group.options),
    [variableOptionGroups],
  )
  const enabledVariables = useMemo(
    () => enabledVariablesForDashboard(dashboard?.variables || [], dashboard?.widgets || [], presets),
    [dashboard?.variables, dashboard?.widgets, presets],
  )
  const customLabelVariables = useMemo(
    () => enabledVariables.filter((variable) => !isStandardResourceVariable(variable)),
    [enabledVariables],
  )
  const activeMultiVariableSelectorNames = useMemo(() => {
    const multiVariableNames = new Set(
      enabledVariables.filter((variable) => variableSelectMode(variable) === 'multiple').map((variable) => variable.name),
    )
    return Object.keys(openMultiVariableSelectors).filter((name) => multiVariableNames.has(name))
  }, [enabledVariables, openMultiVariableSelectors])
  useEffect(() => {
    setOpenMultiVariableSelectors((prev) => {
      const prevNames = Object.keys(prev)
      if (
        activeMultiVariableSelectorNames.length === prevNames.length
        && activeMultiVariableSelectorNames.every((name) => prev[name])
      ) return prev
      return Object.fromEntries(activeMultiVariableSelectorNames.map((name) => [name, true]))
    })
  }, [activeMultiVariableSelectorNames])
  const variableQueryPaused = activeMultiVariableSelectorNames.length > 0

  useDebounce(() => {
    if (!variableQueryPaused) {
      setDebouncedVariables(variables)
    }
  }, 300, [variables, variableQueryPaused])

  const globalRefreshKey = useMemo(
    () => timeRangeRefreshDependency(globalTimeRange),
    [globalTimeRange],
  )
  const dashboardVariableScopeKey = useMemo(
    () => dashboardVariableReadinessScopeKey({
      dashboardId: dashboard?.id,
      dataGroupId,
      globalRefreshKey,
      queryText,
      variables: enabledVariables,
      selectedValues: variableValuesForQuery(debouncedVariables, enabledVariables),
    }),
    [dashboard?.id, dataGroupId, globalRefreshKey, queryText, enabledVariables, debouncedVariablesSignature],
  )
  const hasDashboardVariables = useMemo(
    () => enabledVariables.length > 0,
    [enabledVariables],
  )
  const dashboardVariablesReady = (
    dataGroupId !== null
    && dataGroupId !== undefined
    && variablesSignature === debouncedVariablesSignature
    && (!hasDashboardVariables || dashboardVariablesReadyKey === dashboardVariableScopeKey)
  )
  const widgetsById = useMemo(
    () => Object.fromEntries((dashboard?.widgets || []).map((widget) => [widget.id, widget])),
    [dashboard?.widgets],
  )
  const widgetDraftKind = useMemo(
    () => widgetDraftMode(dashboard?.widgets || [], draftWidget),
    [dashboard?.widgets, draftWidget],
  )
  const widgetDraftIsNew = widgetDraftKind === 'new'

  const updateVariableSelection = (variable, value) => {
    setVariables((prev) => ({
      ...prev,
      [variable.name]: normalizeVariableSelection(variable, value, prev[variable.name], candidates[variable.name] || []),
    }))
  }

  const applyVariableCandidates = (variable, nextCandidates, requestKey = '') => {
    if (requestKey && candidateRequestKeysRef.current[variable.name] !== requestKey) return
    setCandidates((prev) => ({ ...prev, [variable.name]: nextCandidates }))
    setCandidateErrors((prev) => ({ ...prev, [variable.name]: '' }))
    setVariables((prev) => {
      const nextValue = syncVariableSelectionWithCandidates(variable, prev[variable.name], nextCandidates)
      if (JSON.stringify(prev[variable.name]) === JSON.stringify(nextValue)) return prev
      return { ...prev, [variable.name]: nextValue }
    })
  }

  const applyVariableCandidateError = (variable, error, requestKey = '') => {
    if (requestKey && candidateRequestKeysRef.current[variable.name] !== requestKey) return
    setCandidates((prev) => ({ ...prev, [variable.name]: [] }))
    setCandidateErrors((prev) => ({ ...prev, [variable.name]: variableCandidateErrorMessage(t, error) }))
  }

  const applyPreview = useCallback((updater) => {
    setPreviews((prev) => {
      const next = typeof updater === 'function' ? updater(prev) : updater
      previewsRef.current = next
      return next
    })
  }, [])

  const updateWidgetPreviewState = useCallback((widgetID, updater) => {
    const previous = widgetPreviewStatesRef.current[widgetID] || {}
    const nextState = typeof updater === 'function'
      ? updater(previous)
      : { ...previous, ...updater }
    const nextStates = {
      ...widgetPreviewStatesRef.current,
      [widgetID]: nextState,
    }
    widgetPreviewStatesRef.current = nextStates
    setWidgetPreviewStates(nextStates)
  }, [])

  const clearPreviewKey = useCallback((previewKey) => {
    if (!previewKey) return
    delete previewRequestKeysRef.current[previewKey]
    applyPreview((prev) => {
      if (!Object.prototype.hasOwnProperty.call(prev, previewKey)) return prev
      const next = { ...prev }
      delete next[previewKey]
      return next
    })
    if (!Object.prototype.hasOwnProperty.call(widgetPreviewStatesRef.current, previewKey)) return
    const nextStates = { ...widgetPreviewStatesRef.current }
    delete nextStates[previewKey]
    widgetPreviewStatesRef.current = nextStates
    setWidgetPreviewStates(nextStates)
  }, [applyPreview])

  const clearDraftPreview = useCallback((widget) => {
    clearPreviewKey(draftPreviewKeyForWidget(widget))
  }, [clearPreviewKey])

  const seedDraftPreview = useCallback((widget) => {
    if (!widget?.id) return
    const preview = previewsRef.current[widget.id]
    if (!preview) return
    const draftPreviewKey = draftPreviewKeyForWidget(widget)
    applyPreview((prev) => ({ ...prev, [draftPreviewKey]: preview }))
    updateWidgetPreviewState(draftPreviewKey, {
      status: previewStateKind(preview),
      requestKey: '',
      stale: false,
    })
  }, [applyPreview, updateWidgetPreviewState])

  const resetPreviewScheduling = useCallback(() => {
    schedulerRef.current?.clear()
    if (dashboardScrollVisibilityFrameRef.current !== null && typeof window !== 'undefined') {
      window.cancelAnimationFrame?.(dashboardScrollVisibilityFrameRef.current)
      dashboardScrollVisibilityFrameRef.current = null
    }
    previewsRef.current = {}
    widgetPreviewStatesRef.current = {}
    widgetVisibilityRef.current = {}
    widgetElementsRef.current = {}
    previewRequestKeysRef.current = {}
    previousVariablesRef.current = {}
    widgetSeriesLimitsRef.current = {}
    setPreviews({})
    setWidgetPreviewStates({})
    setWidgetVisibilities({})
    setRenderedChartWidgets({})
    setWidgetSeriesLimits({})
  }, [])

  const seriesLimitForWidget = useCallback((widget) => (
    normalizeChartSeriesLimit(widgetSeriesLimitsRef.current[widget?.id], chartSeriesLimitForWidget(widget))
  ), [])

  const widgetPreviewRequest = useCallback((widget, variableState = widgetVars) => {
    const request = buildPreviewRequest({
      widget,
      dashboardId: dashboard?.id,
      variables: variableValuesForQuery(variableState, enabledVariables),
      variableDefinitions: enabledVariables,
      timeRange: globalTimeRange,
      groupId: dataGroupId,
      seriesLimit: seriesLimitForWidget(widget),
    })
    return {
      request,
      requestKey: previewRequestSignature(request),
    }
  }, [dashboard?.id, dataGroupId, enabledVariables, globalTimeRange, seriesLimitForWidget, widgetVars])

  const scheduleWidgetPreview = useCallback((widget, options = {}) => {
    if (!widget?.id || !dashboard) return
    const { priority = '', retry = false } = options
    const eligibility = widgetPreviewEligibilityForVisibility({
      visibility: widgetVisibilityRef.current[widget.id] || 'offscreen',
      observationAvailable: observationAvailableRef.current,
      priority,
    })
    const { request, requestKey } = widgetPreviewRequest(widget)
    const currentState = widgetPreviewStatesRef.current[widget.id] || {}

    if (dataGroupId === null || dataGroupId === undefined || !dashboardVariablesReady) {
      updateWidgetPreviewState(widget.id, { status: 'deferred', requestKey, stale: false })
      return
    }

    if (!eligibility) {
      if (!currentState.status) {
        updateWidgetPreviewState(widget.id, { status: 'deferred', requestKey, stale: false })
      }
      return
    }
    if (!shouldScheduleWidgetPreview(currentState, requestKey, { retry })) return

    previewRequestKeysRef.current[widget.id] = requestKey
    updateWidgetPreviewState(widget.id, {
      status: retry ? 'retrying' : 'queued',
      requestKey,
      stale: Boolean(currentState.stale),
    })
    schedulerRef.current.schedule({
      widgetId: widget.id,
      requestKey,
      priority: WIDGET_PREVIEW_PRIORITY[eligibility] ?? WIDGET_PREVIEW_PRIORITY.near,
      run: async ({ isCancelled }) => {
        if (isCancelled()) return
        updateWidgetPreviewState(widget.id, (previous) => ({
          ...previous,
          status: retry ? 'retrying' : 'loading',
          requestKey,
        }))
        try {
          const response = await previewCustomDashboardQuery(request)
          if (isCancelled() || previewRequestKeysRef.current[widget.id] !== requestKey) return
          applyPreview((prev) => ({ ...prev, [widget.id]: response }))
          updateWidgetPreviewState(widget.id, {
            status: previewStateKind(response),
            requestKey,
            stale: false,
          })
        } catch (error) {
          if (isCancelled() || previewRequestKeysRef.current[widget.id] !== requestKey) return
          const errorMessage = error?.response?.data?.message || error?.response?.data?.msg || error?.message || t('statusQueryFailure')
          applyPreview((prev) => ({ ...prev, [widget.id]: { status: 'error', message: errorMessage } }))
          updateWidgetPreviewState(widget.id, {
            status: 'failed',
            requestKey,
            stale: false,
          })
        }
      },
    })
  }, [applyPreview, dashboard, dataGroupId, dashboardVariablesReady, t, updateWidgetPreviewState, widgetPreviewRequest])

  const scheduleEligibleWidgets = useCallback((widgets = []) => {
    widgets.forEach((widget) => scheduleWidgetPreview(widget))
  }, [scheduleWidgetPreview])

  const markWidgetsStale = useCallback((widgets = []) => {
    widgets.forEach((widget) => {
      const { requestKey } = widgetPreviewRequest(widget)
      const hasPreview = Boolean(previewsRef.current[widget.id])
      previewRequestKeysRef.current[widget.id] = requestKey
      updateWidgetPreviewState(widget.id, (previous) => ({
        ...previous,
        status: hasPreview ? (previous.status || 'success') : 'deferred',
        requestKey,
        stale: hasPreview,
      }))
    })
    scheduleEligibleWidgets(widgets)
  }, [scheduleEligibleWidgets, updateWidgetPreviewState, widgetPreviewRequest])

  const retryWidgetPreview = useCallback((widget) => {
    scheduleWidgetPreview(widget, { priority: 'retry', retry: true })
  }, [scheduleWidgetPreview])

  const refreshSavedWidgetAfterDraftClose = useCallback((widget) => {
    if (widgetDraftMode(dashboard?.widgets || [], widget) !== 'edit') return
    const savedWidget = dashboard?.widgets?.find((item) => item.id === widget.id)
    if (savedWidget) markWidgetsStale([savedWidget])
  }, [dashboard?.widgets, markWidgetsStale])

  const updateWidgetSeriesLimit = useCallback((widget, value, { skipRefresh = false } = {}) => {
    if (!widget?.id) return
    const nextLimit = normalizeChartSeriesLimit(value, chartSeriesLimitForWidget(widget))
    if (seriesLimitForWidget(widget) === nextLimit) return false
    widgetSeriesLimitsRef.current = {
      ...widgetSeriesLimitsRef.current,
      [widget.id]: nextLimit,
    }
    setWidgetSeriesLimits((prev) => ({
      ...prev,
      [widget.id]: nextLimit,
    }))
    if (!skipRefresh) markWidgetsStale([widget])
    return true
  }, [markWidgetsStale, seriesLimitForWidget])

  const updateWidgetVisibility = useCallback((widgetID, nextVisibility) => {
    if (!widgetID || widgetVisibilityRef.current[widgetID] === nextVisibility) return false
    widgetVisibilityRef.current[widgetID] = nextVisibility
    const renderVisibility = nextVisibility === 'visible' ? 'visible' : 'not-visible'
    setWidgetVisibilities((prev) => (
      prev[widgetID] === renderVisibility ? prev : { ...prev, [widgetID]: renderVisibility }
    ))
    return true
  }, [])

  const schedulePostLayoutVisibilityRefresh = useCallback(() => {
    const refresh = () => refreshWidgetVisibilitiesRef.current?.()
    if (typeof window === 'undefined' || !window.requestAnimationFrame) {
      refresh()
      return
    }
    if (dashboardScrollVisibilityFrameRef.current !== null) return
    dashboardScrollVisibilityFrameRef.current = window.requestAnimationFrame(() => {
      dashboardScrollVisibilityFrameRef.current = window.requestAnimationFrame(() => {
        dashboardScrollVisibilityFrameRef.current = null
        refresh()
      })
    })
  }, [])

  const registerWidgetElement = useCallback((widgetID) => {
    if (!widgetID) return null
    if (!widgetElementRefCallbacksRef.current[widgetID]) {
      widgetElementRefCallbacksRef.current[widgetID] = (node) => {
        if (node) {
          node.dataset.widgetId = widgetID
          widgetElementsRef.current[widgetID] = node
          schedulePostLayoutVisibilityRefresh()
        } else {
          delete widgetElementsRef.current[widgetID]
          updateWidgetVisibility(widgetID, 'offscreen')
        }
      }
    }
    return widgetElementRefCallbacksRef.current[widgetID]
  }, [schedulePostLayoutVisibilityRefresh, updateWidgetVisibility])

  const findWidgetElement = useCallback((widgetID) => {
    const cached = widgetElementsRef.current[widgetID]
    if (cached?.isConnected) return cached
    const scope = editorBodyRef.current || document
    const element = [...scope.querySelectorAll('.custom-dashboard-widget[data-widget-id]')]
      .find((node) => node.dataset.widgetId === widgetID)
    if (element) widgetElementsRef.current[widgetID] = element
    return element
  }, [])

  const refreshWidgetVisibilities = useCallback((widgets = dashboard?.widgets || []) => {
    const root = editorBodyRef.current
    if (!root) return
    const rootRect = root.getBoundingClientRect()
    widgets.forEach((widget) => {
      const element = findWidgetElement(widget.id)
      if (!element) return
      const rect = element.getBoundingClientRect()
      const nextVisibility = widgetPreviewVisibilityFromRects({
        rootTop: rootRect.top,
        rootBottom: rootRect.bottom,
        itemTop: rect.top,
        itemBottom: rect.bottom,
        preloadMargin: widgetPreviewPreloadMarginPx,
      })
      if (updateWidgetVisibility(widget.id, nextVisibility) && nextVisibility !== 'offscreen') {
        scheduleWidgetPreview(widget)
      }
    })
  }, [dashboard?.widgets, findWidgetElement, scheduleWidgetPreview, updateWidgetVisibility])
  refreshWidgetVisibilitiesRef.current = refreshWidgetVisibilities

  const widgetDistancesFromViewport = useCallback((widgets = []) => {
    const root = editorBodyRef.current
    if (!root) return {}
    const rootRect = root.getBoundingClientRect()
    return widgets.reduce((distances, widget) => {
      const element = findWidgetElement(widget.id)
      if (!element) return distances
      const rect = element.getBoundingClientRect()
      if (rect.bottom > rootRect.top && rect.top < rootRect.bottom) {
        distances[widget.id] = 0
      } else if (rect.bottom <= rootRect.top) {
        distances[widget.id] = rootRect.top - rect.bottom
      } else {
        distances[widget.id] = rect.top - rootRect.bottom
      }
      return distances
    }, {})
  }, [findWidgetElement])

  useEffect(() => {
    const hadDraftWidget = hadDraftWidgetRef.current
    hadDraftWidgetRef.current = Boolean(draftWidget)
    if (!hadDraftWidget || draftWidget || !dashboard) return undefined
    if (typeof window === 'undefined') {
      refreshWidgetVisibilities()
      return undefined
    }
    const frameID = window.requestAnimationFrame?.(() => refreshWidgetVisibilities())
    if (frameID === undefined) refreshWidgetVisibilities()
    return () => {
      if (frameID !== undefined) window.cancelAnimationFrame?.(frameID)
    }
  }, [dashboard, draftWidget, refreshWidgetVisibilities])

  const updateDashboard = (patch) => {
    setDashboard((prev) => ({ ...prev, ...patch }))
    setDirty(true)
  }

  const updateVariableMulti = (name, multi) => {
    updateDashboard({
      variables: (dashboard.variables || []).map((variable) => (
        variable.name === name ? { ...variable, multi } : variable
      )),
    })
    setVariables((prev) => {
      const nextValue = variableSelectionForMode(prev[name], multi)
      if (JSON.stringify(prev[name]) === JSON.stringify(nextValue)) return prev
      return { ...prev, [name]: nextValue }
    })
  }

  const updateWidget = (patch) => {
    if (draftWidget) {
      setDraftWidget((prev) => ({ ...prev, ...patch }))
      setWidgetDraftDirty(true)
      return
    }
  }

  const updateWidgetNested = (key, patch) => {
    updateWidget({ [key]: { ...(selectedWidget?.[key] || {}), ...patch } })
  }

  const selectedUnitValue = (display = {}) => {
    if (display.unit) return display.unit
    if (display.unitPreset && display.unitPreset !== 'custom') {
      const preset = unitPresetOptions.find((item) => item.value === display.unitPreset)
      return preset?.unit || display.unitPreset
    }
    return ''
  }

  const displayForUnitInput = (unitValue) => {
    const preset = unitPresetOptions.find((preset) => (
      preset.value === unitValue || preset.unit.toLowerCase() === unitValue.toLowerCase()
    ))
    return displayForUnitPreset(preset?.value || unitValue, unitValue)
  }

  const updateWidgetUnit = (values = []) => {
    const value = (Array.isArray(values) ? values[values.length - 1] : values) || ''
    const unitValue = String(value).trim()
    if (!unitValue) {
      updateWidgetNested('display', { unit: '', unitPreset: '', autoScale: false })
      return
    }
    updateWidgetNested('display', displayForUnitInput(unitValue))
  }

  const updateWidgetQuery = (query) => {
    updateWidget({ query: normalizeWidgetQuery(query) })
  }

  const updateQueryItem = (queryID, patch) => {
    const queries = queryItemsForQuery(selectedWidget?.query).map((query) => (
      query.id === queryID ? { ...query, ...patch } : query
    ))
    updateWidgetQuery({ ...(selectedWidget?.query || {}), queries })
  }

  const addQuery = () => {
    const queries = queryItemsForQuery(selectedWidget?.query)
    updateWidgetQuery({
      ...(selectedWidget?.query || {}),
      queries: [...queries, { id: `q-${Date.now()}`, expression: 'up', legend: '' }],
    })
  }

  const deleteQuery = (queryID) => {
    const queries = queryItemsForQuery(selectedWidget?.query)
    if (queries.length <= 1) return
    updateWidgetQuery({
      ...(selectedWidget?.query || {}),
      queries: queries.filter((query) => query.id !== queryID),
    })
  }

  const toggleQueryHidden = (queryID) => {
    const queries = queryItemsForQuery(selectedWidget?.query)
    const target = queries.find((query) => query.id === queryID)
    if (!target) return
    updateQueryItem(queryID, { hidden: target.hidden !== true })
  }

  const insertQuerySnippet = (query, snippet) => {
    updateQueryItem(query.id, {
      expression: insertPromQLMatcherSnippet(query.expression, snippet),
    })
  }

  const closeWidgetEditor = () => {
    const widget = draftWidget
    clearDraftPreview(widget)
    refreshSavedWidgetAfterDraftClose(widget)
    setDraftWidget(null)
    setWidgetDraftDirty(false)
    setWidgetLeavePromptOpen(false)
    setSelectedWidgetID('')
    setDashboardPanelOpen(false)
    if (widgetIDFromSearch && dashboard?.id) navigate(`/custom-dashboard/${dashboard.id}/edit`, { replace: true })
  }

  const requestReturnToDashboard = () => {
    if (shouldPromptWidgetDraftDiscard(draftWidget, widgetDraftDirty)) {
      setWidgetLeavePromptOpen(true)
      return
    }
    closeWidgetEditor()
  }

  const startEditWidget = (widget) => {
    clearDraftPreview(draftWidgetRef.current)
    setSelectedWidgetID(widget.id)
    setDraftWidget(cloneWidgetDraft(widget))
    seedDraftPreview(widget)
    setWidgetDraftDirty(false)
    setWidgetLeavePromptOpen(false)
    setDashboardPanelOpen(false)
  }

  const openDashboardPanel = () => {
    closeWidgetEditor()
    setDashboardPanelOpen(true)
  }

  const updateWidgetFocusLayout = (patch) => {
    setWidgetFocusLayout((prev) => {
      const nextLayout = normalizeWidgetFocusLayout({ ...prev, ...patch })
      writeWidgetFocusLayout(nextLayout)
      return nextLayout
    })
  }

  useEffect(() => {
    const load = async () => {
      setLoading(true)
      resetPreviewScheduling()
      try {
        const [detail, presetList, res] = await Promise.all([
          getCustomDashboard(id),
          getCustomDashboardPresetQueries(),
          getTeamsApi({ currentPage: 1, pageSize: 1000 }),
        ])
        const data = normalizeDashboard(detail)
        const widgetFromSearch = data.widgets?.find((widget) => widget.id === widgetIDFromSearch)
        setDashboard(data)
        if (widgetFromSearch) {
          setSelectedWidgetID(widgetIDFromSearch)
          setDraftWidget(cloneWidgetDraft(widgetFromSearch))
        } else {
          setSelectedWidgetID('')
          setDraftWidget(null)
        }
        setDashboardPanelOpen(false)
        setWidgetDraftDirty(false)
        setWidgetLeavePromptOpen(false)
        setPresets(presetList || [])
        setTeams(res?.teamList || [])
        const initialVars = {}
        const activeVariables = enabledVariablesForDashboard(data.variables || [], data.widgets || [], presetList || [])
        activeVariables.forEach((variable) => {
          const value = variable.value || variable.defaultValue || ''
          initialVars[variable.name] = variable.multi !== false
            ? (Array.isArray(value) ? value : (value ? [value] : []))
            : value
        })
        setVariables(initialVars)
        setDebouncedVariables(initialVars)
        setCandidates({})
        setCandidateErrors({})
        setLoadError('')
      } catch (error) {
        setDashboard(null)
        setLoadError(dashboardLoadFailureMessage(t, error))
      } finally {
        setLoading(false)
      }
    }
    load()
  }, [id, t, widgetIDFromSearch, resetPreviewScheduling])

  useEffect(() => () => {
    schedulerRef.current?.clear()
    variableCandidateSchedulerRef.current?.clear()
  }, [id])

  useEffect(() => {
    const beforeUnload = (event) => {
      if (!dirty && !widgetDraftDirty) return
      event.preventDefault()
      event.returnValue = ''
    }
    window.addEventListener('beforeunload', beforeUnload)
    return () => window.removeEventListener('beforeunload', beforeUnload)
  }, [dirty, widgetDraftDirty])

  useEffect(() => {
    const syncColumns = () => setGridColumns(responsiveGridColumns(window.innerWidth))
    syncColumns()
    window.addEventListener('resize', syncColumns)
    return () => window.removeEventListener('resize', syncColumns)
  }, [])

  useEffect(() => {
    if (!dashboard) return
    const variableCandidateScheduler = variableCandidateSchedulerRef.current
    variableCandidateScheduler?.clear()
    let cancelled = false
    if (dashboardVariableScopeKeyRef.current !== dashboardVariableScopeKey) {
      resourceTreeCacheRef.current = {}
      variableCandidateCacheRef.current = {}
      dashboardVariableScopeKeyRef.current = dashboardVariableScopeKey
      setDashboardVariablesReadyKey('')
    }
    const selectedValues = variableValuesForQuery(debouncedVariables, enabledVariables)
    const pendingCandidateRequestKeys = new Set()
    const markVariableCandidateReady = (requestKey) => {
      pendingCandidateRequestKeys.delete(requestKey)
      if (!cancelled && pendingCandidateRequestKeys.size === 0 && dashboardVariableScopeKeyRef.current === dashboardVariableScopeKey) {
        setDashboardVariablesReadyKey(dashboardVariableScopeKey)
      }
    }
    const resourceVariables = []
    enabledVariables.forEach((variable) => {
      const staticCandidates = variableCandidatesForDefinition(variable)
      if (staticCandidates.length) {
        candidateRequestKeysRef.current[variable.name] = 'static'
        applyVariableCandidates(variable, staticCandidates)
        return
      }
      if (isStandardResourceVariable(variable)) {
        if (dataGroupId === null || dataGroupId === undefined) {
          candidateRequestKeysRef.current[variable.name] = 'data-group-pending'
          applyVariableCandidates(variable, [])
          return
        }
        resourceVariables.push(variable)
        return
      }
      const request = buildVariableCandidateRequest({
        variable,
        queryText,
        variables: debouncedVariables,
        enabledVariables,
        groupId: dataGroupId,
        timeRange: globalTimeRange,
      })
      const requestKey = JSON.stringify(request)
      candidateRequestKeysRef.current[variable.name] = requestKey
      pendingCandidateRequestKeys.add(requestKey)
      variableCandidateScheduler?.schedule({
        variableName: variable.name,
        requestKey,
        run: async ({ signal, isCancelled }) => {
          let cacheEntry = null
          const requestConfig = signal && typeof signal.addEventListener === 'function' ? { signal } : {}
          try {
            const cached = variableCandidateCacheRef.current[requestKey]
            let items = cached?.items
            if (items === undefined) {
              cacheEntry = {
                promise: getCustomDashboardVariableCandidates(request, requestConfig),
              }
              variableCandidateCacheRef.current[requestKey] = cacheEntry
              items = await cacheEntry.promise
              if (isCancelled()) {
                if (variableCandidateCacheRef.current[requestKey] === cacheEntry) {
                  delete variableCandidateCacheRef.current[requestKey]
                }
                return
              }
              if (variableCandidateCacheRef.current[requestKey] === cacheEntry) {
                variableCandidateCacheRef.current[requestKey] = { items }
              }
            }
            if (isCancelled()) return
            applyVariableCandidates(variable, variableCandidatesWithAll(variable, items || []), requestKey)
            markVariableCandidateReady(requestKey)
          } catch (error) {
            if (cacheEntry && variableCandidateCacheRef.current[requestKey] === cacheEntry) {
              delete variableCandidateCacheRef.current[requestKey]
            }
            if (isCancelled() || signal?.aborted) return
            applyVariableCandidateError(variable, error, requestKey)
            markVariableCandidateReady(requestKey)
          }
        },
      })
    })
    const { treeRequests, variablePlans } = buildResourceVariableCandidatePlan({
      variables: resourceVariables,
      selectedValues,
      groupId: dataGroupId,
      timeRange: globalTimeRange,
    })
    variablePlans.forEach(({ variable, candidateRequestKey }) => {
      candidateRequestKeysRef.current[variable.name] = candidateRequestKey
      pendingCandidateRequestKeys.add(candidateRequestKey)
    })
    Promise.all(treeRequests.map(async ({ requestKey, request }) => {
      const plans = variablePlans.filter((plan) => plan.treeRequestKey === requestKey)
      try {
        const cached = resourceTreeCacheRef.current[requestKey]
        let view = cached?.view
        if (view === undefined) {
          const promise = cached?.promise || getDatasourceFilterApiV2(request).then((response) => response?.view)
          resourceTreeCacheRef.current[requestKey] = { promise }
          view = await promise
          resourceTreeCacheRef.current[requestKey] = { view }
        }
        if (cancelled || dashboardVariableScopeKeyRef.current !== dashboardVariableScopeKey) return
        plans.forEach(({ variable, candidateRequestKey }) => {
          const items = resourceVariableCandidatesFromTree(view, variable, selectedValues)
          applyVariableCandidates(variable, variableCandidatesWithAll(variable, items), candidateRequestKey)
          markVariableCandidateReady(candidateRequestKey)
        })
      } catch (error) {
        if (cancelled || dashboardVariableScopeKeyRef.current !== dashboardVariableScopeKey) return
        delete resourceTreeCacheRef.current[requestKey]
        plans.forEach(({ variable, candidateRequestKey }) => {
          applyVariableCandidateError(variable, error, candidateRequestKey)
          markVariableCandidateReady(candidateRequestKey)
        })
      }
    })).then(() => {
      markVariableCandidateReady('')
    }).catch(() => {})
    markVariableCandidateReady('')
    return () => {
      cancelled = true
      variableCandidateScheduler?.clear()
    }
  }, [dashboard?.id, queryText, debouncedVariablesSignature, enabledVariables, dataGroupId, globalTimeRange, dashboardVariableScopeKey, t])

  const startAddWidget = (type = 'line') => {
    clearDraftPreview(draftWidgetRef.current)
    const widget = { ...defaultWidget(dashboard.widgets.length), type, title: t(type) }
    applyPreview((prev) => ({ ...prev, [widget.id]: emptyWidgetPreview }))
    setSelectedWidgetID('')
    setDraftWidget(widget)
    setWidgetDraftDirty(false)
    setWidgetLeavePromptOpen(false)
  }

  const saveWidgetDraft = () => {
    if (!draftWidget) return
    const widget = draftWidget
    const draftKind = widgetDraftMode(dashboard.widgets || [], draftWidget)
    setDashboard((prev) => ({
      ...prev,
      widgets: applyWidgetDraftToDashboard(prev.widgets, draftWidget, gridColumns),
    }))
    setDirty(true)
    setSelectedWidgetID('')
    setDraftWidget(null)
    setWidgetDraftDirty(false)
    setWidgetLeavePromptOpen(false)
    clearDraftPreview(widget)
    if (draftKind === 'new') {
      applyPreview((prev) => ({ ...prev, [widget.id]: emptyWidgetPreview }))
      updateWidgetPreviewState(widget.id, { status: 'success', requestKey: '', stale: false })
    } else {
      markWidgetsStale([widget])
    }
    message.success(t('widgetSavedToDraft'))
  }

  const confirmAddWidget = saveWidgetDraft

  const copyWidget = () => {
    if (!selectedWidget || draftWidget) return
    const widget = {
      ...selectedWidget,
      id: `w-${Date.now()}`,
      title: copyWidgetTitle(selectedWidget.title, t),
      layout: { x: 0, y: dashboard.widgets.length },
    }
    updateDashboard({ widgets: addWidgetToOpenCell(dashboard.widgets, widget, gridColumns) })
    setSelectedWidgetID(widget.id)
  }

  const deleteWidget = (widgetID = selectedWidget?.id) => {
    if (draftWidget && (!widgetID || widgetID === draftWidget.id)) {
      requestReturnToDashboard()
      return
    }
    if (!widgetID) return
    const next = dashboard.widgets.filter((widget) => widget.id !== widgetID)
    updateDashboard({ widgets: next })
    if (selectedWidgetID === widgetID) closeWidgetEditor()
  }

  const addVariable = (name) => {
    const exists = (dashboard.variables || []).some((variable) => variable.name === name)
    if (exists) return
    const option = flatVariableOptions.find((item) => item.value === name)
    updateDashboard({
      variables: [
        ...(dashboard.variables || []),
        dashboardVariableDefinition(name, option, t),
      ],
    })
  }

  const deleteVariable = (name) => {
    const variable = (dashboard.variables || []).find((item) => item.name === name)
    if (variableDeletionState(variable, dashboard?.variables || [], dashboard?.widgets || [], presets).disabled) return
    updateDashboard({
      variables: (dashboard.variables || []).filter((variable) => variable.name !== name),
    })
  }

  const syncWidgetGridLayouts = (layouts = [], _oldItem, changedItem) => {
    setDashboard((prev) => {
      if (!prev) return prev
      return {
        ...prev,
        widgets: mergeWidgetGridLayouts(prev.widgets, layouts, gridColumns, changedItem),
      }
    })
    setDirty(true)
  }

  const previewDashboardWidget = async (widget, { previewKey = widget.id } = {}) => {
    const { request, requestKey } = widgetPreviewRequest(widget, variables)
    if (dataGroupId === null || dataGroupId === undefined || !dashboardVariablesReady) {
      updateWidgetPreviewState(previewKey, { status: 'deferred', requestKey, stale: Boolean(previewsRef.current[previewKey]) })
      return null
    }
    previewRequestKeysRef.current[previewKey] = requestKey
    updateWidgetPreviewState(previewKey, {
      status: 'loading',
      requestKey,
      stale: Boolean(previewsRef.current[previewKey]),
    })
    try {
      const response = await previewCustomDashboardQuery(request)
      if (previewRequestKeysRef.current[previewKey] !== requestKey) return null
      applyPreview((prev) => ({ ...prev, [previewKey]: response }))
      updateWidgetPreviewState(previewKey, {
        status: previewStateKind(response),
        requestKey,
        stale: false,
      })
      return response
    } catch (error) {
      if (previewRequestKeysRef.current[previewKey] !== requestKey) return null
      const errorMessage = error?.response?.data?.message || error?.response?.data?.msg || error?.message || t('statusQueryFailure')
      applyPreview((prev) => ({ ...prev, [previewKey]: { status: 'error', message: errorMessage } }))
      updateWidgetPreviewState(previewKey, {
        status: 'failed',
        requestKey,
        stale: false,
      })
      return null
    }
  }

  const previewWidget = async () => {
    if (!selectedWidget) return
    await previewDashboardWidget(selectedWidget, { previewKey: draftWidget ? draftPreviewKeyForWidget(selectedWidget) : selectedWidget.id })
  }

  useEffect(() => {
    if (!dashboard || dataGroupId === null || dataGroupId === undefined || !dashboardVariablesReady) return
    const previous = previousVariablesRef.current
    const changedVariables = enabledVariables.filter((variable) => (
      JSON.stringify(previous[variable.name]) !== JSON.stringify(widgetVars[variable.name])
    ))
    previousVariablesRef.current = widgetVars
    const shouldRefreshAll = shouldRefreshAllWidgetPreviews({
      previews: previewsRef.current,
      changedVariables,
    })
    const activeDraftWidget = draftWidgetRef.current
    const dashboardWidgets = (dashboard.widgets || []).filter((widget) => (
      !activeDraftWidget || widget.id !== activeDraftWidget.id
    ))
    const widgets = shouldRefreshAll
      ? dashboardWidgets
      : dashboardWidgets.filter((widget) => (
        changedVariables.some((variable) => widgetUsesVariable(widget, variable, presets))
      ))
    markWidgetsStale(widgets)
    const draftUsesChangedVariable = activeDraftWidget && (
      shouldRefreshAll
      || changedVariables.some((variable) => widgetUsesVariable(activeDraftWidget, variable, presets))
    )
    if (draftUsesChangedVariable) {
      previewDashboardWidget(activeDraftWidget, { previewKey: draftPreviewKeyForWidget(activeDraftWidget) })
    }
  }, [dashboard?.id, debouncedVariablesSignature, globalRefreshKey, dataGroupId, dashboardVariablesReady])

  useEffect(() => {
    if (!dashboard) return undefined
    const widgets = dashboard.widgets || []
    const root = editorBodyRef.current
    if (typeof window === 'undefined' || !window.IntersectionObserver || !root) {
      observationAvailableRef.current = false
      widgets.forEach((widget) => {
        widgetVisibilityRef.current[widget.id] = 'fallback'
        scheduleWidgetPreview(widget)
      })
      return undefined
    }

    observationAvailableRef.current = true
    const seedWidgetVisibility = () => refreshWidgetVisibilities(widgets)

    const observer = new window.IntersectionObserver((entries) => {
      entries.forEach((entry) => {
        const widgetID = entry.target?.dataset?.widgetId
        if (!widgetID) return
        const rootRect = root.getBoundingClientRect()
        const rect = entry.boundingClientRect
        const nextVisibility = entry.isIntersecting
          ? widgetPreviewVisibilityFromRects({
            rootTop: rootRect.top,
            rootBottom: rootRect.bottom,
            itemTop: rect.top,
            itemBottom: rect.bottom,
            preloadMargin: widgetPreviewPreloadMarginPx,
          })
          : 'offscreen'
        if (!updateWidgetVisibility(widgetID, nextVisibility)) return
        const widget = widgetsById[widgetID]
        if (widget) scheduleWidgetPreview(widget)
      })
    }, {
      root,
      rootMargin: `${widgetPreviewPreloadMarginPx}px 0px`,
      threshold: 0,
    })

    widgets.forEach((widget) => {
      const element = findWidgetElement(widget.id)
      if (!element) return
      if (!widgetVisibilityRef.current[widget.id]) {
        widgetVisibilityRef.current[widget.id] = 'offscreen'
      }
      observer.observe(element)
    })
    seedWidgetVisibility()
    const frameID = window.requestAnimationFrame?.(seedWidgetVisibility)

    return () => {
      if (frameID !== undefined) window.cancelAnimationFrame?.(frameID)
      observer.disconnect()
    }
  }, [dashboard?.id, dashboard?.widgets, findWidgetElement, refreshWidgetVisibilities, scheduleWidgetPreview, updateWidgetVisibility, widgetsById])

  const scheduleScrollVisibilityRefresh = useCallback(() => {
    if (typeof window === 'undefined') return
    if (!window.requestAnimationFrame) {
      refreshWidgetVisibilities()
      return
    }
    if (dashboardScrollVisibilityFrameRef.current !== null) return
    dashboardScrollVisibilityFrameRef.current = window.requestAnimationFrame(() => {
      dashboardScrollVisibilityFrameRef.current = null
      refreshWidgetVisibilities()
    })
  }, [refreshWidgetVisibilities])

  const handleDashboardScroll = useCallback(() => {
    if (typeof window === 'undefined') return
    scheduleScrollVisibilityRefresh()
    if (!dashboardScrollingRef.current) {
      dashboardScrollingRef.current = true
      setDashboardScrolling(true)
    }
    if (dashboardScrollIdleTimerRef.current) {
      window.clearTimeout(dashboardScrollIdleTimerRef.current)
    }
    dashboardScrollIdleTimerRef.current = window.setTimeout(() => {
      dashboardScrollingRef.current = false
      dashboardScrollIdleTimerRef.current = null
      setDashboardScrolling(false)
    }, 180)
  }, [scheduleScrollVisibilityRefresh])

  useEffect(() => () => {
    if (dashboardScrollVisibilityFrameRef.current !== null && typeof window !== 'undefined') {
      window.cancelAnimationFrame?.(dashboardScrollVisibilityFrameRef.current)
      dashboardScrollVisibilityFrameRef.current = null
    }
    if (dashboardScrollIdleTimerRef.current && typeof window !== 'undefined') {
      window.clearTimeout(dashboardScrollIdleTimerRef.current)
    }
  }, [])

  useEffect(() => {
    if (dashboardScrolling) return
    const widgets = dashboard?.widgets || []
    const chartWidgetIDs = widgets
      .filter((widget) => chartSupportsLegendMode(widget.type))
      .map((widget) => widget.id)
    if (!chartWidgetIDs.length) return
    const distances = widgetDistancesFromViewport(widgets)
    const chartWidgetVisibilities = chartWidgetIDs.reduce((next, widgetID) => {
      next[widgetID] = widgetVisibilityRef.current[widgetID] || widgetVisibilities[widgetID] || 'offscreen'
      return next
    }, {})
    const pinnedWidgetIDs = selectedWidgetID ? [selectedWidgetID] : []
    setRenderedChartWidgets((prev) => {
      const next = reconcileRenderedChartWidgets({
        widgetIds: chartWidgetIDs,
        rendered: prev,
        visibilities: chartWidgetVisibilities,
        distances,
        pinnedWidgetIds: pinnedWidgetIDs,
        now: Date.now(),
        maxActiveCharts: maxRenderedChartWidgets,
        minHoldMs: renderedChartMinHoldMs,
      })
      return renderedChartWidgetsEqual(prev, next) ? prev : next
    })
  }, [dashboard?.widgets, dashboardScrolling, selectedWidgetID, widgetDistancesFromViewport, widgetVisibilities])

  const save = async () => {
    setSaving(true)
    setSaveError('')
    try {
      const saved = normalizeDashboard(await runDashboardMutationWithPermissionRefresh(
        () => updateCustomDashboard(serializeDashboard(dashboard)),
        getUserPermission,
        warnPermissionRefreshFailure,
      ))
      setDashboard(saved)
      setDirty(false)
      message.success(t('saved'))
      return true
    } catch (error) {
      const reason = saveFailureMessage(t, error)
      setSaveError(reason)
      message.error(reason)
      return false
    } finally {
      setSaving(false)
    }
  }

  const saveDashboardPanel = async () => {
    if (await save()) setDashboardPanelOpen(false)
  }

  if (loading) return <Spin />
  if (!dashboard) return <Empty description={loadError || t('dashboardNotFound')} />
  if (!dashboard.canEdit) {
    return (
      <div className="custom-dashboard-editor">
        <Alert
          type="warning"
          showIcon
          message={t('noEditPermission')}
          action={<Button onClick={() => navigate(`/custom-dashboard/${dashboard.id}`)}>{t('backToViewer')}</Button>}
        />
      </div>
    )
  }
  const selectedPreset = presets.find((preset) => preset.id === selectedWidget?.query?.presetId)
  const panelMode = editorPanelMode(selectedWidget)
  const widgetFocusMode = !previewing && panelMode === 'widget'
  const visibleWidgets = widgetFocusMode
    ? (draftWidget ? [] : dashboard.widgets.filter((widget) => widget.id === selectedWidgetID))
    : dashboard.widgets
  const placeholderMode = addWidgetPlaceholderMode(dashboard.widgets || [], Boolean(previewing || draftWidget))
  const placeholderCopy = placeholderMode !== 'hidden' ? addWidgetPlaceholderCopy(placeholderMode, t) : null
  const unitAutoScaleVisible = Boolean(selectedWidget && unitPresetSupportsAutoScale(selectedWidget.display))
  const unitAutoScaleLabel = selectedWidget?.display?.unitPreset === 'percent' ? t('unitRatioToPercent') : t('unitAutoScaleLabel')
  const unitAutoScaleHelpVisible = unitAutoScaleVisible && selectedWidget?.display?.unitPreset !== 'percent'
  const thresholdUnitLabel = selectedWidget ? selectedUnitValue(selectedWidget.display) : ''

  const renderWidgetCard = (widget, { draft = false } = {}) => {
    const previewKey = draft ? draftPreviewKeyForWidget(widget) : widget.id
    const widgetState = widgetPreviewStates[previewKey] || {}
    const widgetRefreshing = Boolean(previews[previewKey]) && (
      widgetState.stale || ['queued', 'loading', 'retrying'].includes(widgetState.status)
    )
    const renderChart = shouldRenderWidgetChart({
      draft,
      alreadyRendered: renderedChartWidgets[widget.id],
      visibility: widgetVisibilities[widget.id],
      scrolling: dashboardScrolling,
    })
    const widgetDraggable = !previewing && !widgetFocusMode && !draft
    return (
      <div
        key={widget.id}
        ref={draft ? null : registerWidgetElement(widget.id)}
        data-widget-id={draft ? undefined : widget.id}
        className={`custom-dashboard-widget ${selectedWidgetID === widget.id || draft ? 'selected' : ''} ${draft ? 'draft' : ''}`}
        onClick={draft ? (event) => {
          event.stopPropagation()
          setSelectedWidgetID('')
        } : undefined}
      >
      <div className={`custom-dashboard-widget-header ${widgetDraggable ? 'custom-dashboard-widget-drag-region' : ''}`}>
        {widgetDraggable && (
          <button
            type="button"
            className="custom-dashboard-widget-drag-handle"
            aria-label={t('position')}
            onClick={(event) => event.stopPropagation()}
          >
            <LuGripVertical />
          </button>
        )}
        <span className="custom-dashboard-widget-heading">
          <span className="custom-dashboard-widget-title">{widgetDisplayTitle(t, widget, presets)}</span>
          {widget.description && (
            <Tooltip title={widget.description}>
              <LuInfo className="custom-dashboard-widget-description" />
            </Tooltip>
          )}
          {widgetRefreshing && (
            <span className="custom-dashboard-widget-refreshing">
              {t(widgetState.status === 'retrying' ? 'retrying' : 'refreshing')}
            </span>
          )}
        </span>
        {widgetDraggable && (
          <Dropdown
            trigger={['click']}
            menu={{
              items: [
                { key: 'edit', label: t('edit') },
                { key: 'delete', label: t('delete'), danger: true },
              ],
              onClick: ({ key, domEvent }) => {
                domEvent?.stopPropagation()
                if (key === 'edit') startEditWidget(widget)
                if (key === 'delete') deleteWidget(widget.id)
              },
            }}
          >
            <Button
              type="text"
              size="small"
              className="custom-dashboard-widget-action-trigger"
              icon={<LuEllipsisVertical />}
              onPointerDown={(event) => event.stopPropagation()}
              onClick={(event) => event.stopPropagation()}
            />
          </Dropdown>
        )}
      </div>
      <div className="custom-dashboard-widget-body">
        <WidgetRenderer
          widget={widget}
          preview={previews[previewKey]}
          previewState={widgetPreviewStates[previewKey]}
          renderChart={renderChart}
          enableTimeBrush
          loadingOnMissingPreview={!draft}
          onRetry={() => (draft ? previewDashboardWidget(widget, { previewKey }) : retryWidgetPreview(widget))}
          seriesLimit={widgetSeriesLimits[widget.id] ?? chartSeriesLimitForWidget(widget)}
          onSeriesLimitChange={(value) => {
            const changed = updateWidgetSeriesLimit(widget, value, { skipRefresh: draft })
            if (draft && changed) previewDashboardWidget(widget, { previewKey })
          }}
        />
      </div>
    </div>
    )
  }

  const occupiedGridLayouts = (dashboard.widgets || []).map((widget) => ({
    id: widget.id,
    layout: arrangedLayouts[widget.id] || normalizeGridLayout(widget.layout, gridColumns),
  }))
  const placeholderInGrid = placeholderMode === 'empty'
  const placeholderLayout = placeholderInGrid ? findOpenGridLayout(
    { x: 0, y: 0, w: gridColumns, h: 1 },
    occupiedGridLayouts,
    gridColumns,
  ) : null
  const dashboardGridLayout = [
    ...visibleWidgets.map((widget) => toDashboardGridItem(widget.id, arrangedLayouts[widget.id] || widget.layout, gridColumns)),
    ...(!widgetFocusMode && placeholderInGrid && placeholderLayout ? [
      toDashboardGridItem(addWidgetGridItemID, placeholderLayout, gridColumns, {
        static: true,
        isDraggable: false,
        isResizable: false,
      }),
    ] : []),
  ]
  const placeholderAction = placeholderMode === 'hidden' ? null : (
    <button
      key={addWidgetGridItemID}
      type="button"
      className={`custom-dashboard-add-widget ${placeholderMode}`}
      onClick={(event) => {
        event.stopPropagation()
        startAddWidget('line')
      }}
    >
      <span className="custom-dashboard-add-widget-title">
        {placeholderCopy.title}
      </span>
      <span className="custom-dashboard-add-widget-description">
        {placeholderCopy.description}
      </span>
      <span className="custom-dashboard-add-widget-action">
        <LuPlus />
        {placeholderCopy.action}
      </span>
    </button>
  )

  const widgetGrid = widgetFocusMode ? (
    <div
      className="custom-dashboard-grid custom-dashboard-focus-grid"
      style={{ '--custom-dashboard-grid-columns': gridColumns }}
      onClick={(event) => {
        if (event.target === event.currentTarget) requestReturnToDashboard()
      }}
    >
      {visibleWidgets.map((widget) => renderWidgetCard(widget))}
      {draftWidget && renderWidgetCard(draftWidget, { draft: true })}
    </div>
  ) : (
    <>
    <DashboardGridLayout
      className="custom-dashboard-grid custom-dashboard-grid-layout"
      style={{ '--custom-dashboard-grid-columns': gridColumns }}
      cols={gridColumns}
      rowHeight={dashboardGridRowHeight}
      margin={dashboardGridMargin}
      containerPadding={[0, 0]}
      compactType={null}
      preventCollision={false}
      isBounded
      isDraggable={!previewing && !widgetFocusMode}
      isResizable={!previewing && !widgetFocusMode}
      draggableHandle=".custom-dashboard-widget-drag-region"
      draggableCancel=".custom-dashboard-widget-action-trigger, .custom-dashboard-add-widget"
      resizeHandles={['se']}
      layout={dashboardGridLayout}
      onDragStop={syncWidgetGridLayouts}
      onResizeStop={syncWidgetGridLayouts}
      onClick={(event) => {
        if (event.target === event.currentTarget) requestReturnToDashboard()
      }}
    >
      {visibleWidgets.map((widget) => renderWidgetCard(widget))}
      {placeholderInGrid && placeholderAction}
    </DashboardGridLayout>
    {placeholderMode === 'append' && placeholderAction}
    </>
  )

  const dashboardSettingsModal = (
    <Modal
      open={dashboardPanelOpen}
      className="custom-dashboard-settings-modal"
      title={t('dashboard')}
      onCancel={() => setDashboardPanelOpen(false)}
      footer={
        <Button type="primary" loading={saving} onClick={saveDashboardPanel}>
          {t('saveDashboard')}
        </Button>
      }
      width={520}
      styles={{
        body: {
          maxHeight: 'calc(100vh - 180px)',
          overflowY: 'auto',
          overflowX: 'hidden',
        },
      }}
    >
      <Form layout="vertical" className="custom-dashboard-settings-form">
        <div className="custom-dashboard-settings-section">
          <Form.Item label={t('name')}>
            <Input value={dashboard.name} onChange={(event) => updateDashboard({ name: event.target.value })} />
          </Form.Item>
          <Form.Item label={t('description')}>
            <Input.TextArea rows={3} value={dashboard.description} onChange={(event) => updateDashboard({ description: event.target.value })} />
          </Form.Item>
          <Form.Item label={t('visibility')}>
            <Select
              value={dashboard.visibility}
              onChange={(visibility) => updateDashboard({
                visibility,
                teamIds: visibility === 'team' ? (dashboard.teamIds || []) : [],
              })}
              options={[
                { label: t('everyone'), value: 'everyone' },
                { label: t('team'), value: 'team' },
                { label: t('personal'), value: 'personal' },
              ]}
            />
          </Form.Item>
          {dashboard.visibility === 'team' && (
            <Form.Item label={t('team')}>
              <Select
                mode="multiple"
                showSearch
                allowClear
                value={dashboard.teamIds || []}
                options={teams}
                optionFilterProp="teamName"
                fieldNames={{ label: 'teamName', value: 'teamId' }}
                onChange={(teamIds) => updateDashboard({ teamIds })}
              />
            </Form.Item>
          )}
        </div>
        <div className="custom-dashboard-settings-section">
          <Typography.Text className="custom-dashboard-settings-section-title">{t('variables')}</Typography.Text>
          {enabledVariables.map((variable) => {
            const deleteState = variableDeletionState(variable, dashboard?.variables || [], dashboard?.widgets || [], presets)
            return (
              <Space key={variable.name} className="custom-dashboard-variable-row" align="center">
                <Input
                  className="custom-dashboard-variable-name"
                  value={variableDisplayLabel(t, variable)}
                  readOnly
                  title={labelsForVariable(variable).join(', ')}
                />
                <Tooltip title={t('variableSelectionMode')}>
                  <span className="custom-dashboard-variable-mode">
                    <span className="custom-dashboard-variable-mode-label">
                      {variable.multi !== false ? t('multiSelect') : t('singleSelect')}
                    </span>
                    <Switch
                      size="small"
                      checked={variable.multi !== false}
                      onChange={(multi) => updateVariableMulti(variable.name, multi)}
                    />
                  </span>
                </Tooltip>
                <Tooltip title={deleteState.disabled ? t('variableDeleteBlocked', { count: deleteState.usageCount }) : t('variableDeleteAllowed')}>
                  <span>
                    <Button
                      danger
                      icon={<LuTrash2 />}
                      disabled={deleteState.disabled}
                      onClick={() => deleteVariable(variable.name)}
                    />
                  </span>
                </Tooltip>
              </Space>
            )
          })}
          <Select
            value={null}
            placeholder={t('addVariable')}
            options={variableOptionGroups}
            onChange={addVariable}
            style={{ width: '100%' }}
          />
          {!flatVariableOptions.length && <Typography.Text type="secondary">{t('labelsAvailableAfterQuery')}</Typography.Text>}
        </div>
      </Form>
    </Modal>
  )

  const widgetLeavePrompt = (
    <Modal
      open={widgetLeavePromptOpen}
      title={t('unsavedWidgetChanges')}
      onCancel={() => setWidgetLeavePromptOpen(false)}
      footer={(
        <div className="custom-dashboard-widget-leave-actions">
          <Button onClick={() => setWidgetLeavePromptOpen(false)}>
            {t('continueEditing')}
          </Button>
          <Button danger onClick={closeWidgetEditor}>
            {t('discardWidgetChanges')}
          </Button>
          <Button type="primary" onClick={saveWidgetDraft}>
            {t('saveWidgetAndReturn')}
          </Button>
        </div>
      )}
    >
      <Typography.Text>{t('leaveWidgetWithoutSaving')}</Typography.Text>
    </Modal>
  )

  const renderPromQLHints = (query) => {
    const hints = promQLResourceVariableHints(query.expression)
    const renderPromQLHintTitle = (labelKey, helpKey) => (
      <div className="custom-dashboard-promql-section-heading">
        <Typography.Text className="custom-dashboard-promql-section-title">
          {t(labelKey)}
        </Typography.Text>
        <Tooltip title={t(helpKey)}>
          <span className="custom-dashboard-promql-section-help" tabIndex={0} aria-label={t(helpKey)}>
            <LuInfo />
          </span>
        </Tooltip>
      </div>
    )
    return (
      <div className="custom-dashboard-promql-hints">
        <div className="custom-dashboard-promql-section custom-dashboard-promql-supported">
          {renderPromQLHintTitle('supportedParameters', 'supportedParametersHelp')}
          <div className="custom-dashboard-promql-chip-list">
            {hints.supported.map((item) => (
              <Tooltip key={item.name} title={item.aliases.join(' / ')}>
                <Button
                  size="small"
                  className="custom-dashboard-promql-chip"
                  onClick={() => insertQuerySnippet(query, item.snippet)}
                >
                  <span>{t(item.name)}</span>
                  <span className="custom-dashboard-promql-chip-meta">{item.snippet}</span>
                </Button>
              </Tooltip>
            ))}
          </div>
        </div>
        {hints.recognized.length > 0 && (
          <div className="custom-dashboard-promql-section custom-dashboard-promql-recognized">
            {renderPromQLHintTitle('recognizedParameters', 'recognizedParametersHelp')}
            <div className="custom-dashboard-promql-chip-list">
              {hints.recognized.flatMap((item) => item.labels.map((label) => (
                <span className="custom-dashboard-promql-chip readonly" key={`${item.name}:${label}`}>
                  <span>{label}</span>
                  <span className="custom-dashboard-promql-chip-meta">-&gt; {t(item.name)}</span>
                </span>
              )))}
            </div>
          </div>
        )}
        {hints.fixed.length > 0 && (
          <div className="custom-dashboard-promql-section custom-dashboard-promql-fixed">
            {renderPromQLHintTitle('fixedFilters', 'fixedFiltersHelp')}
            <div className="custom-dashboard-promql-chip-list">
              {hints.fixed.map((item) => (
                <span className="custom-dashboard-promql-chip readonly" key={`${item.label}:${item.operator}:${item.value}`}>
                  <span>{item.label}</span>
                  <span className="custom-dashboard-promql-chip-meta">{item.operator} &quot;{item.value}&quot;</span>
                </span>
              ))}
            </div>
          </div>
        )}
        {hints.unrecognized.length > 0 && (
          <div className="custom-dashboard-promql-section custom-dashboard-promql-unrecognized">
            {renderPromQLHintTitle('unrecognizedParameters', 'unrecognizedParametersHelp')}
            <div className="custom-dashboard-promql-chip-list">
              {hints.unrecognized.map((item) => (
                <span className="custom-dashboard-promql-chip readonly" key={item.label}>
                  <span>{item.label}</span>
                  <span className="custom-dashboard-promql-chip-meta">{t('labelVariableFallback')}</span>
                </span>
              ))}
            </div>
          </div>
        )}
      </div>
    )
  }

  const legendModeForValue = (value) => (String(value || '').trim() === autoLegendValue ? 'auto' : 'custom')

  const renderLegendControl = (value, onChange) => {
    const legendMode = legendModeForValue(value)
    return (
      <Space direction="vertical" size={8} style={{ width: '100%' }}>
        <Segmented
          size="small"
          value={legendMode}
          options={[
            { value: 'custom', label: t('legendCustom') },
            { value: 'auto', label: t('legendAuto') },
          ]}
          onChange={(mode) => onChange(mode === 'auto' ? autoLegendValue : '')}
        />
        {legendMode === 'custom' ? (
          <Input
            value={value || ''}
            placeholder={t('legendCustomPlaceholder')}
            onChange={(event) => onChange(event.target.value)}
          />
        ) : null}
        <Typography.Text type="secondary" className="custom-dashboard-field-hint">
          {t(legendMode === 'auto' ? 'legendAutoHelp' : 'legendCustomHelp')}
        </Typography.Text>
      </Space>
    )
  }

  const widgetQueryPanel = widgetFocusMode && selectedWidget ? (
    <Form layout="vertical" className="custom-dashboard-widget-query-form">
      <Form.Item label={t('queryMode')}>
        <Select
          value={selectedWidget.query?.mode}
          onChange={(mode) => updateWidgetNested('query', { mode })}
          options={[
            { label: t('preset'), value: 'preset' },
            { label: t('custom'), value: 'custom' },
          ]}
        />
      </Form.Item>
      {selectedWidget.query?.mode === 'preset' ? (
        <Form.Item label={t('presetQuery')}>
          <Select
            value={selectedWidget.query?.presetId}
            showSearch
            options={presetOptionsForChart(presets, selectedWidget.type, t)}
            onChange={(presetId) => {
              const preset = presets.find((item) => item.id === presetId)
              updateWidgetNested('query', {
                presetId,
                params: mergePresetDefaults(preset, selectedWidget.query?.params || {}),
              })
            }}
          />
        </Form.Item>
      ) : null}
      {selectedWidget.query?.mode === 'preset' && selectedPreset ? (
        (selectedPreset.parameters || []).map((param) => (
          <Form.Item
            key={param.name}
            label={presetParameterLabel(t, param)}
            extra={presetParameterDescription(t, param)}
            required={param.required}
          >
            <Input
              value={selectedWidget.query?.params?.[param.name] ?? param.default ?? ''}
              onChange={(event) => updateWidgetNested('query', {
                params: {
                  ...(selectedWidget.query?.params || {}),
                  [param.name]: event.target.value,
                },
              })}
            />
          </Form.Item>
        ))
      ) : selectedWidget.query?.mode === 'custom' ? (
        <Space className="custom-dashboard-query-list" direction="vertical">
          {queryItemsForQuery(selectedWidget.query).map((query, index, queries) => (
            <div className={`custom-dashboard-query-item ${query.hidden ? 'is-hidden' : ''}`} key={query.id}>
              <div className="custom-dashboard-query-item-header">
                <Typography.Text className="custom-dashboard-query-title" type="secondary">{t('query')} {index + 1}</Typography.Text>
                <Space className="custom-dashboard-query-actions" size={4}>
                  <Tooltip title={t(query.hidden ? 'showQuery' : 'hideQuery')}>
                    <Button
                      size="small"
                      type="text"
                      icon={query.hidden ? <LuEyeOff /> : <LuEye />}
                      onClick={() => toggleQueryHidden(query.id)}
                    />
                  </Tooltip>
                  <Button danger size="small" type="text" icon={<LuTrash2 />} disabled={queries.length <= 1} onClick={() => deleteQuery(query.id)} />
                </Space>
              </div>
              <Form.Item className="custom-dashboard-query-expression-item" label={t('expression')}>
                <Input.TextArea
                  className="custom-dashboard-query-expression"
                  rows={4}
                  value={query.expression}
                  onChange={(event) => updateQueryItem(query.id, { expression: event.target.value })}
                />
                {renderPromQLHints(query)}
              </Form.Item>
              <Form.Item label={t('legend')} tooltip={t('legendHelp')}>
                {renderLegendControl(query.legend, (legend) => updateQueryItem(query.id, { legend }))}
              </Form.Item>
            </div>
          ))}
          <Button className="custom-dashboard-add-query-button" icon={<LuPlus />} onClick={addQuery}>{t('addQuery')}</Button>
        </Space>
      ) : null}
      {selectedWidget.query?.mode === 'preset' && (
        <Form.Item label={t('legend')} tooltip={t('legendHelp')}>
          {renderLegendControl(selectedWidget.query?.legend, (legend) => updateWidgetNested('query', { legend }))}
        </Form.Item>
      )}
    </Form>
  ) : null

  const widgetSettingsPanel = widgetFocusMode && selectedWidget ? (
    <>
      <div className="custom-dashboard-panel-header">
        <Typography.Title level={5} style={{ margin: 0 }}>
          {widgetDraftIsNew ? t('newWidget') : t('widget')}
        </Typography.Title>
        <Button type="link" icon={<LuArrowLeft />} onClick={requestReturnToDashboard}>
          {t('backToDashboard')}
        </Button>
      </div>
      <Form layout="vertical">
        <Form.Item label={t('name')}>
          <Input value={selectedWidget.title} onChange={(event) => updateWidget({ title: event.target.value })} />
        </Form.Item>
        <Form.Item label={t('description')}>
          <Input value={selectedWidget.description} onChange={(event) => updateWidget({ description: event.target.value })} />
        </Form.Item>
        <Form.Item label={t('chartType')}>
          <Select
            value={selectedWidget.type}
            options={chartTypes.map((type) => ({ label: t(type.labelKey), value: type.value }))}
            onChange={(type) => {
              const currentPreset = presets.find((preset) => preset.id === selectedWidget.query?.presetId)
              updateWidget({
                type,
                query: currentPreset?.compatibleCharts?.includes(type)
                  ? selectedWidget.query
                  : { ...selectedWidget.query, presetId: '', params: {} },
              })
            }}
          />
        </Form.Item>
        {chartSupportsLegendMode(selectedWidget.type) && (
          <Form.Item label={t('legendDisplay')} tooltip={t('legendDisplayHelp')}>
            <Radio.Group
              optionType="button"
              value={selectedWidget.display?.legendMode || 'list'}
              options={chartLegendModes.map((mode) => ({ label: t(mode.labelKey), value: mode.value }))}
              onChange={(event) => updateWidgetNested('display', { legendMode: event.target.value })}
            />
          </Form.Item>
        )}
        <Space>
          <Form.Item label={t('unit')}>
            <AutoComplete
              allowClear
              placeholder={t('unitCustomPlaceholder')}
              optionFilterProp="label"
              style={{ minWidth: 156 }}
              value={selectedUnitValue(selectedWidget.display)}
              options={unitPresetOptions.map((preset) => ({ label: t(preset.labelKey), value: preset.unit }))}
              onChange={updateWidgetUnit}
            />
          </Form.Item>
          <Form.Item label={t('decimals')}>
            <InputNumber min={0} max={8} value={selectedWidget.display?.decimals} onChange={(decimals) => updateWidgetNested('display', { decimals })} />
          </Form.Item>
        </Space>
        {unitAutoScaleVisible && (
          <Form.Item label={t('unitDisplay')}>
            <Space wrap>
              <Switch checked={Boolean(selectedWidget.display?.autoScale)} onChange={(autoScale) => updateWidgetNested('display', { autoScale })} />
              <span>{unitAutoScaleLabel}</span>
              {unitAutoScaleHelpVisible && (
                <Typography.Text type="secondary" className="custom-dashboard-field-hint">
                  {t('unitAutoScaleHelp')}
                </Typography.Text>
              )}
            </Space>
          </Form.Item>
        )}
        <Space className="custom-dashboard-widget-threshold-row" wrap>
          <Form.Item label={t('warningThreshold')} tooltip={t('warningThresholdHelp')}>
            <InputNumber addonAfter={thresholdUnitLabel || undefined} value={selectedWidget.display?.warningThreshold} onChange={(warningThreshold) => updateWidgetNested('display', { warningThreshold })} />
          </Form.Item>
          <Form.Item label={t('criticalThreshold')} tooltip={t('criticalThresholdHelp')}>
            <InputNumber addonAfter={thresholdUnitLabel || undefined} value={selectedWidget.display?.criticalThreshold} onChange={(criticalThreshold) => updateWidgetNested('display', { criticalThreshold })} />
          </Form.Item>
        </Space>
        <Space className="custom-dashboard-widget-action-row" wrap>
          <Button onClick={previewWidget}>{t('preview')}</Button>
          {draftWidget ? (
            <>
              <Button danger onClick={requestReturnToDashboard}>{t('cancel')}</Button>
              <Button type="primary" onClick={confirmAddWidget}>{t('saveWidget')}</Button>
            </>
          ) : (
            <>
              <Button icon={<LuCopy />} onClick={copyWidget} />
              <Button danger icon={<LuTrash2 />} onClick={deleteWidget} />
            </>
          )}
        </Space>
      </Form>
    </>
  ) : null

  return (
    <div className="custom-dashboard-editor">
      <div className="custom-dashboard-toolbar">
        <div className="custom-dashboard-toolbar-main">
          <Space>
            <Typography.Title level={4} style={{ margin: 0 }}>{dashboard.name}</Typography.Title>
            {dirty && <Typography.Text type="secondary">{t('unsaved')}</Typography.Text>}
            {widgetDraftDirty && <Typography.Text type="secondary">{t('unsavedWidget')}</Typography.Text>}
          </Space>
          {widgetFocusMode ? null : (
            <Space>
              <Button icon={<LuSettings />} onClick={openDashboardPanel}>{t('settings')}</Button>
              <Button icon={<LuSave />} type="primary" loading={saving} onClick={save}>{t('saveDashboard')}</Button>
            </Space>
          )}
        </div>
        <div className="custom-dashboard-toolbar-parameters">
          <CustomDashboardResourceFilter
            t={t}
            variables={enabledVariables}
            values={variables}
            candidates={candidates}
            candidateErrors={candidateErrors}
            onChange={updateVariableSelection}
            onDropdownVisibleChange={updateVariableSelectorOpen}
          />
          {customLabelVariables.map((variable) => (
            <span className="custom-dashboard-toolbar-parameter" key={variable.name}>
              <span className="custom-dashboard-toolbar-parameter-label">
                {variableDisplayLabel(t, variable)}
              </span>
              <Select
                placeholder={variable.name}
                value={variables[variable.name] || undefined}
                onChange={(value) => updateVariableSelection(variable, value)}
                style={{ width: 180 }}
                allowClear
                mode={variableSelectMode(variable)}
                onDropdownVisibleChange={(open) => updateVariableSelectorOpen(variable, open)}
                notFoundContent={candidateErrors[variable.name] || t('noAvailableValues')}
                options={variableSelectOptions(t, variable, candidates[variable.name] || [])}
              />
            </span>
          ))}
        </div>
      </div>
      {dashboardSettingsModal}
      {widgetLeavePrompt}
      {saveError && (
        <Alert
          type="error"
          showIcon
          message={saveError}
          action={<Button size="small" onClick={save} loading={saving}>{t('retry')}</Button>}
          closable
          onClose={() => setSaveError('')}
        />
      )}
      <div ref={editorBodyRef} className={`custom-dashboard-editor-body ${widgetFocusMode ? 'widget-focus' : ''}`} onScroll={handleDashboardScroll}>
        {widgetFocusMode ? (
          <Splitter
            className="custom-dashboard-widget-focus-splitter"
            onResize={(sizes) => updateWidgetFocusLayout({ settingsWidth: Math.round(sizes[1]) })}
          >
            <Splitter.Panel className="custom-dashboard-widget-focus-main-panel" min={420}>
              <Splitter
                layout="vertical"
                className="custom-dashboard-widget-focus-main"
                onResize={(sizes) => updateWidgetFocusLayout({ previewHeight: Math.round(sizes[0]) })}
              >
                <Splitter.Panel className="custom-dashboard-widget-focus-preview" min={280} size={widgetFocusLayout.previewHeight}>
                  {widgetGrid}
                </Splitter.Panel>
                <Splitter.Panel className="custom-dashboard-widget-query-panel" min={200}>
                  <div className="custom-dashboard-widget-query-panel-inner">
                    {widgetQueryPanel}
                  </div>
                </Splitter.Panel>
              </Splitter>
            </Splitter.Panel>
            <Splitter.Panel className="custom-dashboard-panel custom-dashboard-widget-settings-panel" min={280} max={560} size={widgetFocusLayout.settingsWidth}>
              {widgetSettingsPanel}
            </Splitter.Panel>
          </Splitter>
        ) : (
          widgetGrid
        )}
      </div>
    </div>
  )
}

export default CustomDashboardEditor
