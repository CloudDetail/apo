/**
 * Copyright 2026 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { useDebounce } from 'react-use'
import { Button, Empty, Select, Space, Spin, Tooltip, Typography } from 'antd'
import { LuInfo, LuPencil } from 'react-icons/lu'
import { useNavigate, useParams } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useSelector } from 'react-redux'
import { getDatasourceFilterApiV2 } from 'src/core/api/dataGroup'
import { selectSecondsTimeRange } from 'src/core/store/reducers/timeRangeReducer'
import {
  getCustomDashboard,
  getCustomDashboardPresetQueries,
  getCustomDashboardVariableCandidates,
  previewCustomDashboardQuery,
} from 'src/pro/api/customDashboard'
import WidgetRenderer from './WidgetRenderer'
import CustomDashboardResourceFilter from './CustomDashboardResourceFilter'
import {
  arrangeWidgetLayouts,
  buildPreviewRequest,
  buildResourceVariableCandidatePlan,
  buildVariableCandidateRequest,
  createDashboardPreviewScheduler,
  createVariableCandidateScheduler,
  dashboardLoadFailureMessage,
  dashboardVariableReadinessScopeKey,
  dashboardQueryText,
  enabledVariablesForDashboard,
  gridPlacement,
  isStandardResourceVariable,
  previewRequestSignature,
  previewStateKind,
  resourceVariableCandidatesFromTree,
  responsiveGridColumns,
  shouldRefreshAllWidgetPreviews,
  shouldRefreshWidgetVisibilityOnScroll,
  shouldScheduleWidgetPreview,
  timeRangeRefreshDependency,
  variableCandidatesForDefinition,
  variableCandidatesWithAll,
  variableCandidateErrorMessage,
  variableDisplayLabel,
  normalizeVariableSelection,
  reconcileRenderedChartWidgets,
  syncVariableSelectionWithCandidates,
  shouldRenderWidgetChart,
  variableSelectMode,
  variableSelectOptions,
  variableValuesForQuery,
  widgetDisplayTitle,
  widgetPreviewEligibilityForVisibility,
  widgetPreviewVisibilityFromRects,
  widgetUsesVariable,
  WIDGET_PREVIEW_PRIORITY,
} from './behavior.mjs'
import { chartSeriesLimitForWidget, chartSupportsLegendMode, normalizeChartSeriesLimit, normalizeDashboard } from './utils.mjs'
import './index.scss'

const widgetPreviewPreloadMarginPx = 320
const maxRenderedChartWidgets = 12
const renderedChartMinHoldMs = 2500

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

const CustomDashboardViewer = () => {
  const { id } = useParams()
  const navigate = useNavigate()
  const { t } = useTranslation('pro/customDashboard')
  const [dashboard, setDashboard] = useState(null)
  const [loading, setLoading] = useState(true)
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
  const [presets, setPresets] = useState([])
  const [loadError, setLoadError] = useState('')
  const [gridColumns, setGridColumns] = useState(12)
  const [dashboardVariablesReadyKey, setDashboardVariablesReadyKey] = useState('')
  const [dashboardScrolling, setDashboardScrolling] = useState(false)
  const gridRef = useRef(null)
  const widgetElementsRef = useRef({})
  const widgetElementRefCallbacksRef = useRef({})
  const widgetVisibilityRef = useRef({})
  const dashboardScrollIdleTimerRef = useRef(null)
  const dashboardScrollVisibilityFrameRef = useRef(null)
  const dashboardScrollingRef = useRef(false)
  const observationAvailableRef = useRef(true)
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

  const variablesSignature = JSON.stringify(variables)
  const debouncedVariablesSignature = JSON.stringify(debouncedVariables)
  const widgetVars = useMemo(() => debouncedVariables, [debouncedVariables])
  const queryText = useMemo(() => dashboardQueryText(dashboard?.widgets || [], presets), [dashboard?.widgets, presets])
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

  const arrangedLayouts = useMemo(
    () => arrangeWidgetLayouts(dashboard?.widgets || [], gridColumns),
    [dashboard?.widgets, gridColumns],
  )
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

  const resetPreviewScheduling = useCallback(() => {
    schedulerRef.current?.clear()
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

  const widgetPreviewRequest = useCallback((widget) => {
    const request = buildPreviewRequest({
      widget,
      dashboardId: dashboard?.id,
      variables: variableValuesForQuery(widgetVars, enabledVariables),
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
          const message = error?.response?.data?.message || error?.response?.data?.msg || error?.message || t('statusQueryFailure')
          applyPreview((prev) => ({ ...prev, [widget.id]: { status: 'error', message } }))
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

  const updateWidgetSeriesLimit = useCallback((widget, value) => {
    if (!widget?.id) return
    const nextLimit = normalizeChartSeriesLimit(value, chartSeriesLimitForWidget(widget))
    if (seriesLimitForWidget(widget) === nextLimit) return
    widgetSeriesLimitsRef.current = {
      ...widgetSeriesLimitsRef.current,
      [widget.id]: nextLimit,
    }
    setWidgetSeriesLimits((prev) => ({
      ...prev,
      [widget.id]: nextLimit,
    }))
    markWidgetsStale([widget])
  }, [markWidgetsStale, seriesLimitForWidget])

  const updateWidgetVisibility = useCallback((widgetID, nextVisibility) => {
    if (!widgetID || widgetVisibilityRef.current[widgetID] === nextVisibility) return false
    widgetVisibilityRef.current[widgetID] = nextVisibility
    setWidgetVisibilities((prev) => (
      prev[widgetID] === nextVisibility ? prev : { ...prev, [widgetID]: nextVisibility }
    ))
    return true
  }, [])

  const registerWidgetElement = useCallback((widgetID) => {
    if (!widgetID) return null
    if (!widgetElementRefCallbacksRef.current[widgetID]) {
      widgetElementRefCallbacksRef.current[widgetID] = (node) => {
        if (node) {
          node.dataset.widgetId = widgetID
          widgetElementsRef.current[widgetID] = node
        } else {
          delete widgetElementsRef.current[widgetID]
          updateWidgetVisibility(widgetID, 'offscreen')
        }
      }
    }
    return widgetElementRefCallbacksRef.current[widgetID]
  }, [updateWidgetVisibility])

  const findWidgetElement = useCallback((widgetID) => {
    const cached = widgetElementsRef.current[widgetID]
    if (cached?.isConnected) return cached
    const scope = gridRef.current || document
    const element = [...scope.querySelectorAll('.custom-dashboard-widget[data-widget-id]')]
      .find((node) => node.dataset.widgetId === widgetID)
    if (element) widgetElementsRef.current[widgetID] = element
    return element
  }, [])

  const refreshWidgetVisibilities = useCallback((widgets = dashboard?.widgets || []) => {
    const root = gridRef.current
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

  const widgetDistancesFromViewport = useCallback((widgets = []) => {
    const root = gridRef.current
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

  const loadDashboard = async () => {
    setLoading(true)
    resetPreviewScheduling()
    try {
      const [dataDetail, presetList] = await Promise.all([
        getCustomDashboard(id),
        getCustomDashboardPresetQueries(),
      ])
      const data = normalizeDashboard(dataDetail)
      setDashboard(data)
      setPresets(presetList || [])
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
      setLoadError('')
    } catch (error) {
      setDashboard(null)
      setLoadError(dashboardLoadFailureMessage(t, error))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadDashboard()
  }, [id])

  useEffect(() => () => {
    schedulerRef.current?.clear()
    variableCandidateSchedulerRef.current?.clear()
  }, [id])

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
    const widgets = shouldRefreshAll
      ? (dashboard.widgets || [])
      : (dashboard.widgets || []).filter((widget) => (
        changedVariables.some((variable) => widgetUsesVariable(widget, variable, presets))
      ))
    markWidgetsStale(widgets)
  }, [dashboard?.id, debouncedVariablesSignature, globalRefreshKey, dataGroupId, dashboardVariablesReady])

  useEffect(() => {
    if (!dashboard) return undefined
    const widgets = dashboard.widgets || []
    const root = gridRef.current
    if (typeof window === 'undefined' || !window.IntersectionObserver || !root) {
      observationAvailableRef.current = false
      widgets.forEach((widget) => {
        updateWidgetVisibility(widget.id, 'fallback')
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
        if (widget && nextVisibility !== 'offscreen') scheduleWidgetPreview(widget)
      })
    }, {
      root,
      rootMargin: `${widgetPreviewPreloadMarginPx}px 0px`,
      threshold: 0,
    })

    widgets.forEach((widget) => {
      const element = widgetElementsRef.current[widget.id]
      if (!element) return
      if (!widgetVisibilityRef.current[widget.id]) {
        updateWidgetVisibility(widget.id, 'offscreen')
      }
      observer.observe(element)
    })
    seedWidgetVisibility()
    const frameID = window.requestAnimationFrame?.(seedWidgetVisibility)

    return () => {
      if (frameID !== undefined) window.cancelAnimationFrame?.(frameID)
      observer.disconnect()
    }
  }, [dashboard?.id, dashboard?.widgets, refreshWidgetVisibilities, scheduleWidgetPreview, updateWidgetVisibility, widgetsById])

  const scheduleScrollVisibilityRefresh = useCallback(() => {
    if (typeof window === 'undefined') return
    if (!shouldRefreshWidgetVisibilityOnScroll({ observationAvailable: observationAvailableRef.current })) return
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
      refreshWidgetVisibilities()
      dashboardScrollingRef.current = false
      dashboardScrollIdleTimerRef.current = null
      setDashboardScrolling(false)
    }, 180)
  }, [refreshWidgetVisibilities, scheduleScrollVisibilityRefresh])

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
    setRenderedChartWidgets((prev) => {
      const next = reconcileRenderedChartWidgets({
        widgetIds: chartWidgetIDs,
        rendered: prev,
        visibilities: chartWidgetVisibilities,
        distances,
        now: Date.now(),
        maxActiveCharts: maxRenderedChartWidgets,
        minHoldMs: renderedChartMinHoldMs,
      })
      return renderedChartWidgetsEqual(prev, next) ? prev : next
    })
  }, [dashboard?.widgets, dashboardScrolling, widgetDistancesFromViewport, widgetVisibilities])

  if (loading) return <Spin />
  if (!dashboard) return <Empty description={loadError || t('dashboardNotFound')} />

  return (
    <div className="custom-dashboard-viewer">
      <div className="custom-dashboard-toolbar">
        <div className="custom-dashboard-toolbar-main">
          <Typography.Title level={4} style={{ margin: 0 }}>{dashboard.name}</Typography.Title>
          <Space>
            {dashboard.canEdit && <Button icon={<LuPencil />} onClick={() => navigate(`/custom-dashboard/${dashboard.id}/edit`)} />}
          </Space>
        </div>
        <Space className="custom-dashboard-toolbar-parameters" wrap>
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
        </Space>
      </div>
      <div ref={gridRef} className="custom-dashboard-grid" style={{ '--custom-dashboard-grid-columns': gridColumns }} onScroll={handleDashboardScroll}>
        {dashboard.widgets.map((widget) => {
          const widgetState = widgetPreviewStates[widget.id] || {}
          const widgetRefreshing = Boolean(previews[widget.id]) && (
            widgetState.stale || ['queued', 'loading', 'retrying'].includes(widgetState.status)
          )
          const renderChart = shouldRenderWidgetChart({
            alreadyRendered: renderedChartWidgets[widget.id],
            visibility: widgetVisibilities[widget.id],
            scrolling: dashboardScrolling,
          })
          return (
          <div
            key={widget.id}
            ref={registerWidgetElement(widget.id)}
            data-widget-id={widget.id}
            className="custom-dashboard-widget"
            style={gridPlacement(arrangedLayouts[widget.id] || widget.layout, gridColumns)}
          >
            <div className="custom-dashboard-widget-header">
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
            </div>
            <div className="custom-dashboard-widget-body">
              <WidgetRenderer
                widget={widget}
                preview={previews[widget.id]}
                previewState={widgetPreviewStates[widget.id]}
                renderChart={renderChart}
                enableTimeBrush
                onRetry={() => retryWidgetPreview(widget)}
                seriesLimit={widgetSeriesLimits[widget.id] ?? chartSeriesLimitForWidget(widget)}
                onSeriesLimitChange={(value) => updateWidgetSeriesLimit(widget, value)}
              />
            </div>
          </div>
          )
        })}
      </div>
    </div>
  )
}

export default CustomDashboardViewer
