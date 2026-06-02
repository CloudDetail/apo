/**
 * Copyright 2026 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react'
import ReactECharts from 'echarts-for-react'
import { Button, Pagination, Select, Spin, Table, Tag } from 'antd'
import { useTranslation } from 'react-i18next'
import { useDispatch, useSelector } from 'react-redux'
import { displayText, previewStatusDetails, previewStatusLabel } from './behavior.mjs'
import {
  barGaugeProgressPercent,
  buildChartOption,
  chartLegendRows,
  chartSeriesLimitControlInfo,
  chartSeriesLimitForWidget,
  chartSeriesLimitOptions,
  chartSupportsLegendMode,
  chartSupportsTimeBrush,
  chartTimeRangeFromBrush,
  formatWidgetValue,
  limitChartPreviewSeries,
  normalizeChartSeriesLimit,
  normalizeChartLegendMode,
  seriesDisplayName,
  showNearestLineTooltip,
  tableColumnsForPreview,
  tableRowsForPreview,
} from './utils.mjs'

const latestSeriesRows = (preview, widget) => (preview.series || []).map((item, index) => {
  const values = item.values || []
  const last = values[values.length - 1]
  const rawValue = last ? last.Value ?? last.value : undefined
  return {
    key: index,
    name: seriesDisplayName(item.metric || {}, index, widget.query?.legend),
    rawValue,
    value: last ? formatWidgetValue(rawValue, widget.display) : '-',
  }
})

const barGaugeValueColor = (percent) => {
  if (percent >= 75) return '#e24d42'
  if (percent >= 50) return '#ef843c'
  if (percent >= 25) return '#d6df4e'
  return '#7eb26d'
}

const thumbnailChartStyle = { width: '100%', height: 88, minHeight: 88 }
const fullChartStyle = { width: '100%', height: '100%', minHeight: 0 }
const tableLegendPageSize = 10
const customDashboardTablePageSize = 100
const customDashboardTableScrollY = 'calc(100% - 34px)'
const customDashboardTableScroll = { x: 'max-content', y: customDashboardTableScrollY }
const customDashboardThumbnailTableScroll = { x: 'max-content' }

const resizeEChartAfterLayout = (chart, resizeFrameRef) => {
  if (!chart) return
  const resize = () => {
    try {
      if (!chart.isDisposed?.()) chart.resize()
    } catch {
      // Ignore resize races when the panel is closed before the deferred frame runs.
    }
  }
  if (typeof window === 'undefined') {
    resize()
    return
  }
  if (resizeFrameRef?.current) {
    window.cancelAnimationFrame(resizeFrameRef.current)
  }
  resizeFrameRef.current = window.requestAnimationFrame(() => {
    resizeFrameRef.current = null
    resize()
  })
}

const ResponsiveEChart = ({ option, style, enableTimeBrush = false }) => {
  const chartRef = React.useRef(null)
  const resizeFrameRef = React.useRef(null)
  const brushTimerRef = React.useRef(null)
  const tooltipFrameRef = React.useRef(null)
  const pendingTooltipRef = React.useRef(null)
  const lineHoverCleanupRef = React.useRef(null)
  const hoverPointRef = React.useRef(null)
  const dispatch = useDispatch()
  const timeValues = React.useMemo(() => {
    const xAxis = Array.isArray(option?.xAxis) ? option.xAxis[0] : option?.xAxis
    return xAxis?.customDashboardTimeValues || []
  }, [option])
  const lineHoverSeries = React.useMemo(() => {
    if (!Array.isArray(option?.series)) return []
    const legend = Array.isArray(option?.legend) ? option.legend[0] : option?.legend
    const selected = legend?.selected || {}
    return option.series
      .map((series, seriesIndex) => ({
        seriesIndex,
        data: Array.isArray(series?.data) ? series.data : [],
        selected: selected[series?.name] !== false,
        type: series?.type,
        color: series?.itemStyle?.color || series?.lineStyle?.color || series?.color,
      }))
      .filter((series) => series.type === 'line' && series.data.length > 0)
  }, [option])
  const hideLineHoverPoint = React.useCallback(() => {
    const element = hoverPointRef.current
    if (!element) return
    element.style.opacity = '0'
  }, [])
  const showLineHoverPoint = React.useCallback((point) => {
    const element = hoverPointRef.current
    const x = Number(point?.x)
    const y = Number(point?.y)
    if (!element || !Number.isFinite(x) || !Number.isFinite(y)) {
      hideLineHoverPoint()
      return
    }
    element.style.setProperty('--custom-dashboard-chart-hover-point-color', point?.color || 'var(--ant-color-primary)')
    element.style.transform = `translate3d(${x}px, ${y}px, 0) translate(-50%, -50%)`
    element.style.opacity = '1'
  }, [hideLineHoverPoint])
  const activateTimeBrush = React.useCallback((chart) => {
    if (!enableTimeBrush || !timeValues.length || !chart?.dispatchAction) return
    const run = () => {
      brushTimerRef.current = null
      try {
        if (chart.isDisposed?.()) return
        chart.dispatchAction({
          type: 'takeGlobalCursor',
          key: 'brush',
          brushOption: {
            brushType: 'lineX',
            brushMode: 'single',
          },
        })
      } catch {
        // Ignore chart lifecycle races while widgets are virtualized or resized.
      }
    }
    if (typeof window === 'undefined') {
      run()
      return
    }
    if (brushTimerRef.current) window.clearTimeout(brushTimerRef.current)
    brushTimerRef.current = window.setTimeout(run, 100)
  }, [enableTimeBrush, timeValues])
  const handleBrushEnd = React.useCallback((params, chart) => {
    if (!enableTimeBrush) return
    const nextRange = chartTimeRangeFromBrush(params, chart, timeValues)
    if (!nextRange) return
    dispatch({
      type: 'SET_TIMERANGE',
      payload: {
        rangeType: null,
        ...nextRange,
      },
    })
  }, [dispatch, enableTimeBrush, timeValues])
  const flushPendingLineTooltip = React.useCallback(() => {
    tooltipFrameRef.current = null
    const pending = pendingTooltipRef.current
    pendingTooltipRef.current = null
    if (!pending) return
    showNearestLineTooltip(pending.params, pending.chart, lineHoverSeries, showLineHoverPoint)
  }, [lineHoverSeries, showLineHoverPoint])
  const handleLineMouseMove = React.useCallback((params, chart) => {
    if (typeof window === 'undefined' || typeof window.requestAnimationFrame !== 'function') {
      return showNearestLineTooltip(params, chart, lineHoverSeries, showLineHoverPoint)
    }
    pendingTooltipRef.current = { params, chart }
    if (tooltipFrameRef.current) return true
    tooltipFrameRef.current = window.requestAnimationFrame(flushPendingLineTooltip)
    return true
  }, [flushPendingLineTooltip, lineHoverSeries, showLineHoverPoint])
  const handleLineGlobalOut = React.useCallback((_params, chart) => {
    pendingTooltipRef.current = null
    if (tooltipFrameRef.current && typeof window !== 'undefined') {
      window.cancelAnimationFrame(tooltipFrameRef.current)
      tooltipFrameRef.current = null
    }
    hideLineHoverPoint()
    chart?.dispatchAction?.({ type: 'updateAxisPointer', currTrigger: 'leave' })
    chart?.dispatchAction?.({ type: 'hideTip' })
  }, [hideLineHoverPoint])
  const detachLineHover = React.useCallback(() => {
    if (!lineHoverCleanupRef.current) return
    lineHoverCleanupRef.current()
    lineHoverCleanupRef.current = null
  }, [])
  const attachLineHover = React.useCallback((chart) => {
    detachLineHover()
    if (!lineHoverSeries.length) return
    const zr = chart?.getZr?.()
    if (!zr?.on || !zr?.off) return
    const handleMove = (params) => handleLineMouseMove(params, chart)
    const handleOut = (params) => handleLineGlobalOut(params, chart)
    zr.on('mousemove', handleMove)
    zr.on('globalout', handleOut)
    lineHoverCleanupRef.current = () => {
      zr.off('mousemove', handleMove)
      zr.off('globalout', handleOut)
    }
  }, [detachLineHover, handleLineGlobalOut, handleLineMouseMove, lineHoverSeries.length])
  const chartEvents = React.useMemo(() => ({
    ...(enableTimeBrush ? { brushEnd: handleBrushEnd } : {}),
  }), [enableTimeBrush, handleBrushEnd])
  const resizeCurrentChart = React.useCallback(() => {
    const chart = chartRef.current?.getEchartsInstance?.()
    resizeEChartAfterLayout(chart, resizeFrameRef)
    activateTimeBrush(chart)
  }, [activateTimeBrush])
  const handleChartReady = React.useCallback((chart) => {
    resizeEChartAfterLayout(chart, resizeFrameRef)
    activateTimeBrush(chart)
    attachLineHover(chart)
  }, [activateTimeBrush, attachLineHover])

  React.useEffect(() => {
    hideLineHoverPoint()
    resizeCurrentChart()
  }, [hideLineHoverPoint, option, resizeCurrentChart])

  React.useEffect(() => {
    attachLineHover(chartRef.current?.getEchartsInstance?.())
    return detachLineHover
  }, [attachLineHover, detachLineHover])

  React.useEffect(() => () => {
    detachLineHover()
    hideLineHoverPoint()
    if (resizeFrameRef.current && typeof window !== 'undefined') {
      window.cancelAnimationFrame(resizeFrameRef.current)
    }
    if (brushTimerRef.current && typeof window !== 'undefined') {
      window.clearTimeout(brushTimerRef.current)
    }
    if (tooltipFrameRef.current && typeof window !== 'undefined') {
      window.cancelAnimationFrame(tooltipFrameRef.current)
    }
  }, [detachLineHover, hideLineHoverPoint])

  React.useEffect(() => {
    if (typeof ResizeObserver === 'undefined') return undefined
    const chart = chartRef.current?.getEchartsInstance?.()
    const element = chart?.getDom?.()
    if (!element) return undefined
    const observer = new ResizeObserver(() => resizeCurrentChart())
    observer.observe(element)
    return () => observer.disconnect()
  }, [resizeCurrentChart])

  return (
    <div className="custom-dashboard-chart-hover-surface" style={style}>
      <ReactECharts
        ref={chartRef}
        option={option}
        notMerge={true}
        style={{ width: '100%', height: '100%' }}
        onEvents={chartEvents}
        onChartReady={handleChartReady}
      />
      <span aria-hidden="true" className="custom-dashboard-chart-hover-point" ref={hoverPointRef} />
    </div>
  )
}

const renderWidgetState = ({ kind = 'empty', tagColor, tagText, title, description, action }) => {
  const className = kind === 'error'
    ? 'custom-dashboard-state custom-dashboard-state-error'
    : 'custom-dashboard-state custom-dashboard-state-empty'
  return (
    <div className={className}>
      {tagText && <Tag color={tagColor}>{tagText}</Tag>}
      <span className="custom-dashboard-state-content">
        <span className="custom-dashboard-state-title">{title}</span>
        {description && <span className="custom-dashboard-state-suggestion">{description}</span>}
      </span>
      {action && <span className="custom-dashboard-state-action">{action}</span>}
    </div>
  )
}

const renderEmptyState = (t) => renderWidgetState({
  kind: 'empty',
  title: t('noData'),
})

const WidgetRenderer = ({
  widget,
  preview,
  previewState,
  compact = false,
  thumbnail = false,
  loadingOnMissingPreview = true,
  renderChart = true,
  onRetry,
  seriesLimit,
  onSeriesLimitChange,
  enableTimeBrush = false,
}) => {
  const { t } = useTranslation('pro/customDashboard')
  const themeMode = useSelector((state) => state.settingReducer.theme)
  const [hiddenSeriesNames, setHiddenSeriesNames] = React.useState(() => new Set())
  const [legendPage, setLegendPage] = React.useState(1)
  const renderWidget = React.useMemo(() => {
    if (!chartSupportsLegendMode(widget.type)) return widget
    const activeLimit = normalizeChartSeriesLimit(
      seriesLimit ?? widget.display?.seriesLimit ?? widget.query?.seriesLimit,
      chartSeriesLimitForWidget(widget),
    )
    return {
      ...widget,
      display: {
        ...(widget.display || {}),
        seriesLimit: activeLimit,
      },
    }
  }, [seriesLimit, widget])
  const activeSeriesLimit = React.useMemo(() => chartSeriesLimitForWidget(renderWidget), [renderWidget])
  const seriesLimitOptions = React.useMemo(() => (
    Array.from(new Set([...chartSeriesLimitOptions, activeSeriesLimit]))
      .sort((a, b) => a - b)
      .map((value) => ({ value, label: `Top ${value}` }))
  ), [activeSeriesLimit])
  const renderPreview = React.useMemo(() => limitChartPreviewSeries(renderWidget, preview), [renderWidget, preview])
  const legendSeriesNames = React.useMemo(() => (
    (renderPreview?.series || []).map((item, index) => seriesDisplayName(item.metric || {}, index, renderWidget.query?.legend))
  ), [renderPreview, renderWidget.query?.legend])
  const selectedSeries = React.useMemo(() => (
    Object.fromEntries(legendSeriesNames.map((name) => [name, !hiddenSeriesNames.has(name)]))
  ), [legendSeriesNames, hiddenSeriesNames])
  const legendRows = React.useMemo(() => (
    chartLegendRows(renderWidget, renderPreview, themeMode)
  ), [renderWidget, renderPreview, themeMode])
  const legendMode = React.useMemo(() => (
    normalizeChartLegendMode(renderWidget.display?.legendMode)
  ), [renderWidget.display?.legendMode])
  const tableLegend = legendMode === 'table'
  const legendHasPagination = tableLegend && legendRows.length > tableLegendPageSize
  const legendRowsForRender = React.useMemo(() => {
    if (!legendHasPagination) return legendRows
    const start = (legendPage - 1) * tableLegendPageSize
    return legendRows.slice(start, start + tableLegendPageSize)
  }, [legendHasPagination, legendPage, legendRows])
  const toggleLegendRow = React.useCallback((name, multiSelect = false) => {
    if (!legendSeriesNames.includes(name)) return
    setHiddenSeriesNames((prev) => {
      if (multiSelect) {
        const next = new Set(prev)
        if (next.has(name)) {
          next.delete(name)
          return next
        }
        const visibleCount = legendSeriesNames.reduce((count, seriesName) => (
          count + (next.has(seriesName) ? 0 : 1)
        ), 0)
        if (visibleCount <= 1) {
          return prev
        }
        next.add(name)
        return next
      }
      const onlyCurrentVisible = legendSeriesNames.length > 0
        && legendSeriesNames.every((seriesName) => (
          seriesName === name ? !prev.has(seriesName) : prev.has(seriesName)
        ))
      if (onlyCurrentVisible) {
        return prev.size === 0 ? prev : new Set()
      }
      return new Set(legendSeriesNames.filter((seriesName) => seriesName !== name))
    })
  }, [legendSeriesNames])
  const handleLegendKeyDown = React.useCallback((event, name) => {
    if (event.key !== 'Enter' && event.key !== ' ') return
    event.preventDefault()
    toggleLegendRow(name, event.ctrlKey || event.metaKey)
  }, [toggleLegendRow])

  React.useEffect(() => {
    const currentNames = new Set(legendSeriesNames)
    setHiddenSeriesNames((prev) => {
      const next = new Set([...prev].filter((name) => currentNames.has(name)))
      return next.size === prev.size ? prev : next
    })
  }, [legendSeriesNames])

  React.useEffect(() => {
    if (!legendHasPagination) {
      setLegendPage((prev) => (prev === 1 ? prev : 1))
      return
    }
    const maxPage = Math.max(1, Math.ceil(legendRows.length / tableLegendPageSize))
    setLegendPage((prev) => Math.min(prev, maxPage))
  }, [legendHasPagination, legendRows.length])

  if (!preview) {
    if (previewState?.status === 'deferred') {
      return renderWidgetState({
        kind: 'empty',
        title: t('widgetDeferred'),
        description: t('widgetDeferredSuggestion'),
      })
    }
    if (loadingOnMissingPreview) {
      return <div className="custom-dashboard-state"><Spin size="small" /></div>
    }
    return renderEmptyState(t)
  }
  const status = renderPreview.status
  if (status && status !== 'success') {
    const details = previewStatusDetails(status)
    if (status === 'no_data') {
      return renderEmptyState(t)
    }
    return renderWidgetState({
      kind: 'error',
      tagColor: details.color,
      tagText: previewStatusLabel(t, status),
      title: displayText(t, renderPreview.message || t('statusQueryFailure'), renderPreview.position),
      description: details.suggestionKey ? t(details.suggestionKey) : '',
      action: onRetry ? <Button size="small" onClick={onRetry}>{t('retry')}</Button> : null,
    })
  }
  if (widget.unsupportedType || widget.type === 'single') {
    return renderWidgetState({
      kind: 'error',
      tagColor: 'warning',
      tagText: t('unsupportedWidgetType'),
      title: t('unsupportedWidgetHelp'),
    })
  }
  if (widget.type === 'table' || widget.type === 'topn') {
    const rows = tableRowsForPreview(renderPreview, renderWidget)
    const columns = tableColumnsForPreview(renderPreview, renderWidget.display).map((column) => {
      if (column.kind === 'value') {
        return { title: t(column.title), dataIndex: 'value', width: 120, align: 'right', className: 'custom-dashboard-table-value' }
      }
      if (column.kind === 'name') {
        return { title: t(column.title), dataIndex: 'name', ellipsis: true }
      }
      return {
        title: column.title,
        dataIndex: ['labels', column.name],
        ellipsis: true,
        render: (value) => value ?? '-',
      }
    })
    if (rows.length === 0) return renderEmptyState(t)
    const tablePagination = !thumbnail && rows.length > customDashboardTablePageSize
      ? { pageSize: customDashboardTablePageSize, showSizeChanger: false, size: 'small' }
      : false
    return (
      <Table
        className="custom-dashboard-table"
        size="small"
        pagination={tablePagination}
        scroll={thumbnail ? customDashboardThumbnailTableScroll : customDashboardTableScroll}
        dataSource={thumbnail ? rows.slice(0, 2) : rows}
        columns={columns}
      />
    )
  }
  if (widget.type === 'stat') {
    const rows = latestSeriesRows(renderPreview, renderWidget)
    if (rows.length === 0) return renderEmptyState(t)
    return (
      <div className="custom-dashboard-stat-grid">
        {rows.map((row) => (
          <div className="custom-dashboard-stat-item" key={row.key}>
            <span className="custom-dashboard-stat-label">{row.name}</span>
            <span className="custom-dashboard-stat-value">{row.value}</span>
          </div>
        ))}
      </div>
    )
  }
  if (widget.type === 'barGauge') {
    const rows = latestSeriesRows(renderPreview, renderWidget)
    if (rows.length === 0) return renderEmptyState(t)
    return (
      <div className="custom-dashboard-bar-gauge">
        {rows.map((row) => {
          const percent = barGaugeProgressPercent(row.rawValue, renderWidget.display)
          return (
            <div className={`custom-dashboard-bar-gauge-row ${rows.length === 1 ? 'single' : ''}`} key={row.key}>
              {rows.length > 1 && <span className="custom-dashboard-bar-gauge-name">{row.name}</span>}
              <div className="custom-dashboard-bar-gauge-body">
                <div className="custom-dashboard-bar-gauge-segments" style={{ '--bar-gauge-percent': `${percent}%` }} />
                <strong className="custom-dashboard-bar-gauge-value" style={{ color: barGaugeValueColor(percent) }}>{row.value}</strong>
              </div>
            </div>
          )
        })}
      </div>
    )
  }
  const hasChartSeries = (renderPreview.series || []).length > 0
  if (!hasChartSeries) {
    return renderEmptyState(t)
  }
  if (!renderChart && !thumbnail) {
    return renderWidgetState({
      kind: 'empty',
      title: t('widgetDeferred'),
      description: t('widgetDeferredSuggestion'),
    })
  }
  const chartOption = buildChartOption(renderWidget, renderPreview, themeMode, selectedSeries)
  if (thumbnail) {
    return <ResponsiveEChart option={chartOption} style={thumbnailChartStyle} />
  }
  const seriesLimitInfo = chartSeriesLimitControlInfo(renderWidget, renderPreview)
  const legendBodySeriesLimitInfo = legendHasPagination ? null : seriesLimitInfo
  const legendFooterSeriesLimitInfo = legendHasPagination
    ? (seriesLimitInfo || { total: legendRows.length, shown: legendRows.length, limit: activeSeriesLimit })
    : null
  const renderSeriesLimitInfo = (info) => (
    <div className="custom-dashboard-chart-series-limit">
      <span
        className="custom-dashboard-chart-series-limit-summary"
        title={t('seriesLimitNotice', { total: info.total, shown: info.shown })}
      >
        {t('seriesLimitSummary', { total: info.total })}
      </span>
      {seriesLimitInfo && onSeriesLimitChange && (
        <Select
          className="custom-dashboard-chart-series-limit-control"
          size="small"
          title={t('seriesLimitNotice', { total: seriesLimitInfo.total, shown: activeSeriesLimit })}
          value={activeSeriesLimit}
          options={seriesLimitOptions}
          popupMatchSelectWidth={false}
          onChange={onSeriesLimitChange}
          onClick={(event) => event.stopPropagation()}
        />
      )}
    </div>
  )
  return (
    <div className={`custom-dashboard-chart-panel custom-dashboard-chart-panel-${legendMode}`} style={{ '--custom-dashboard-chart-min-height': compact ? '140px' : '160px' }}>
      <div className="custom-dashboard-chart-plot">
        <ResponsiveEChart
          option={chartOption}
          style={fullChartStyle}
          enableTimeBrush={enableTimeBrush && chartSupportsTimeBrush(renderWidget.type)}
        />
      </div>
      <div className={`custom-dashboard-chart-legend custom-dashboard-chart-legend-${legendMode} ${legendHasPagination ? 'custom-dashboard-chart-legend-has-pagination' : ''}`}>
        {tableLegend && (
          <div className="custom-dashboard-chart-legend-header">
            <span>{t('seriesName')}</span>
            <span>{t('legendMin')}</span>
            <span>{t('legendMax')}</span>
            <span>{t('legendMean')}</span>
          </div>
        )}
        <div className="custom-dashboard-chart-legend-body">
          {legendBodySeriesLimitInfo && renderSeriesLimitInfo(legendBodySeriesLimitInfo)}
          {legendRowsForRender.map((row) => {
            const visible = selectedSeries[row.name] !== false
            return (
              <button
                aria-pressed={visible}
                className={`custom-dashboard-chart-legend-row ${visible ? '' : 'is-hidden'}`}
                key={row.key}
                onClick={(event) => toggleLegendRow(row.name, event.ctrlKey || event.metaKey)}
                onKeyDown={(event) => handleLegendKeyDown(event, row.name)}
                type="button"
              >
                <span className="custom-dashboard-chart-legend-name" title={row.name}>
                  <span className="custom-dashboard-chart-legend-swatch" style={{ backgroundColor: row.color }} />
                  <span>{row.name}</span>
                </span>
                {tableLegend && (
                  <>
                    <span className="custom-dashboard-chart-legend-value">{row.min}</span>
                    <span className="custom-dashboard-chart-legend-value">{row.max}</span>
                    <span className="custom-dashboard-chart-legend-value">{row.mean}</span>
                  </>
                )}
              </button>
            )
          })}
        </div>
        {legendHasPagination && (
          <div className="custom-dashboard-chart-legend-footer">
            {legendFooterSeriesLimitInfo && renderSeriesLimitInfo(legendFooterSeriesLimitInfo)}
            <div className="custom-dashboard-chart-legend-pagination" onClick={(event) => event.stopPropagation()}>
              <Pagination
                simple
                size="small"
                current={legendPage}
                pageSize={tableLegendPageSize}
                total={legendRows.length}
                showSizeChanger={false}
                onChange={setLegendPage}
              />
            </div>
          </div>
        )}
      </div>
    </div>
  )
}

export default WidgetRenderer
