import { useEffect, useMemo, useRef, useState } from 'react'
import ReactECharts from 'echarts-for-react'
import { Activity, AlertCircle, CircleGauge, Clock3, FileSearch, ListChecks } from 'lucide-react'
import { useSearchParams } from 'react-router-dom'
import { useSelector } from 'react-redux'
import MarkdownRender from 'src/core/components/Markdown/MarkdownRender'
import { getReportViewApi } from 'src/pro/api/report'

import AnalysisProcess from './components/AnalysisProcess'
import SupportingEvidence from './components/SupportingEvidence'
import styles from './index.module.scss'
import {
  createReportViewModelFromViewResponse,
  type AlertEventReportViewResponse,
  type CandidateState,
  type CoreMetricChartSeries,
  type ReportViewModel,
  type RootCauseEvidenceItem,
  type ReportStatus,
  type ReportSummaryCardTone,
  type SuggestLevel,
} from './reportViewModel'
import { buildReportViewRequest, type ReportViewRequest } from './reportRequest'

const candidateStateClass: Record<CandidateState, string> = {
  primary: styles.candidatePrimary,
  secondary: styles.candidateSecondary,
  excluded: styles.candidateExcluded,
  unknown: styles.candidateUnknown,
}

const statusClass: Record<ReportStatus, string> = {
  active: styles.statusActive,
  recovering: styles.statusWarning,
  resolved: styles.statusResolved,
  snapshot: styles.statusSnapshot,
  unknown: styles.statusSnapshot,
}

const actionLevelLabel: Record<string, string> = {
  high: 'P1',
  medium: 'P2',
  low: 'P3',
}

const actionLevelClass: Record<string, string> = {
  high: styles.actionHigh,
  medium: styles.actionMedium,
  low: styles.actionLow,
}

const DEFAULT_VISIBLE_ACTION_COUNT = 3

type AppTheme = 'light' | 'dark'

const summaryCardClass: Record<ReportSummaryCardTone, string> = {
  symptom: styles.summarySymptom,
  rootCause: styles.summaryRootCause,
  excluded: styles.summaryExcluded,
}

function getAppTheme(theme?: string): AppTheme {
  return theme === 'dark' ? 'dark' : 'light'
}

function getActionLevelLabel(level?: SuggestLevel) {
  if (!level) return '建议'
  return actionLevelLabel[level] || level.toUpperCase()
}

function getActionLevelClass(level?: SuggestLevel) {
  return actionLevelClass[level || 'low'] || styles.actionLow
}

function getRequestReportId(request: ReportViewRequest) {
  return request.tags.alertEventId || request.tags.agentChatId || ''
}

function getLoadErrorMessage(error: unknown) {
  if (error && typeof error === 'object') {
    const responseMessage = (error as { response?: { data?: { message?: string } } }).response?.data
      ?.message
    if (responseMessage) return responseMessage
  }

  if (error instanceof Error && error.message) {
    return error.message
  }

  return '报告数据加载失败'
}

function escapeTooltipHtml(value: unknown) {
  if (value === null || value === undefined) return ''

  return String(value).replace(/[&<>"']/g, (char) => {
    switch (char) {
      case '&':
        return '&amp;'
      case '<':
        return '&lt;'
      case '>':
        return '&gt;'
      case '"':
        return '&quot;'
      case "'":
        return '&#39;'
      default:
        return char
    }
  })
}

function getEvidenceAnchorId(index: number) {
  return `report-raw-evidence-${index}`
}

function formatSparklineValue(value: number | string, unit?: string) {
  if (typeof value === 'string') return value
  if (unit === 's' || unit === 'second' || unit === 'seconds') {
    if (Math.abs(value) < 1) return `${Number((value * 1000).toFixed(2))} ms`
    return `${Number(value.toFixed(2))} s`
  }
  if (unit === 'percent') return `${Number(value.toFixed(2))}%`
  if (unit === 'bytes') return `${Number((value / 1024 / 1024).toFixed(2))} MB`
  if ((unit === 'core' || unit === 'cores') && value !== 0 && Math.abs(value) < 0.01) {
    return `<0.01 ${unit}`
  }
  if (unit) return `${Number(value.toFixed(2))} ${unit}`
  return `${Number(value.toFixed(2))}`
}

const metricSparklineColors = ['#d14343', '#256f8f', '#6b7a24', '#7b5aa6', '#9b6b24']

function getMetricSparklineColor(index: number) {
  return metricSparklineColors[index % metricSparklineColors.length]
}

function getVisibleMetricSeries(seriesList: CoreMetricChartSeries[]) {
  const dataSeries = seriesList.filter((series) => series.data.length)
  const focusedSeries = dataSeries.filter((series) => series.focus)
  const backgroundSeries = dataSeries.filter((series) => !series.focus)

  return [...focusedSeries, ...backgroundSeries].slice(0, 5)
}

function MetricSparkline({ seriesList }: { seriesList: CoreMetricChartSeries[] }) {
  const sparklineTheme = useSelector((state: { settingReducer?: { theme?: string } }) =>
    getAppTheme(state.settingReducer?.theme),
  )
  const visibleSeries = useMemo(() => getVisibleMetricSeries(seriesList), [seriesList])
  const option = useMemo(
    () => ({
      animation: false,
      backgroundColor: 'transparent',
      color: visibleSeries.map((_, index) => getMetricSparklineColor(index)),
      tooltip: {
        trigger: 'axis',
        appendToBody: true,
        confine: false,
        backgroundColor:
          sparklineTheme === 'dark' ? 'oklch(20% 0.018 250)' : 'oklch(99.2% 0.002 250)',
        borderColor: sparklineTheme === 'dark' ? 'oklch(36% 0.018 250)' : 'oklch(86.5% 0.008 250)',
        textStyle: {
          color: sparklineTheme === 'dark' ? 'oklch(92% 0.008 250)' : 'oklch(22% 0.018 250)',
        },
        extraCssText: 'max-width:360px;white-space:normal;z-index:10000;',
        position: (
          _point: [number, number],
          _params: unknown,
          _element: unknown,
          _rect: unknown,
          size: { contentSize: [number, number]; viewSize: [number, number] },
        ) => {
          const contentWidth = size.contentSize[0] || 0
          const viewWidth = size.viewSize[0] || 0
          const viewHeight = size.viewSize[1] || 0

          return [Math.min(0, viewWidth - contentWidth), viewHeight + 8]
        },
        formatter: (
          params: Array<{ seriesName: string; data: [number, number | string]; color: string }>,
        ) => {
          if (!params?.length) return ''

          return params
            .map((item) => {
              const matchedSeries = visibleSeries.find(
                (series) => series.legend === item.seriesName,
              )
              const seriesName = escapeTooltipHtml(item.seriesName)
              const value = escapeTooltipHtml(formatSparklineValue(item.data[1], matchedSeries?.unit))
              return `<div style="display:flex;gap:8px;align-items:center;justify-content:space-between;">
                  <span style="display:inline-flex;gap:6px;align-items:center;">
                    <i style="width:7px;height:7px;border-radius:999px;background:${item.color};display:inline-block;"></i>
                    ${seriesName}
                  </span>
                  <strong>${value}</strong>
                </div>`
            })
            .join('')
        },
      },
      grid: {
        left: 0,
        right: 0,
        top: 5,
        bottom: 4,
        containLabel: false,
      },
      xAxis: {
        type: 'time',
        show: false,
        boundaryGap: false,
      },
      yAxis: {
        type: 'value',
        show: false,
        scale: true,
      },
      series: [
        ...visibleSeries.map((series, index) => {
          const color = getMetricSparklineColor(index)

          return {
            name: series.legend,
            type: 'line',
            data: series.data,
            smooth: true,
            showSymbol: false,
            itemStyle: {
              color,
            },
            lineStyle: {
              color,
              width: series.focus || index === 0 ? 2 : 1.2,
              opacity: series.focus || index === 0 ? 1 : 0.36,
            },
            emphasis: {
              focus: 'series',
            },
          }
        }),
      ],
    }),
    [sparklineTheme, visibleSeries],
  )

  if (!visibleSeries.length) return null

  return (
    <ReactECharts
      option={option}
      style={{ height: visibleSeries.length > 1 ? 52 : 32, width: '100%' }}
    />
  )
}

function MetricSparklineGroup({ metric }: { metric: RootCauseEvidenceItem }) {
  if (!metric.chartData?.length) return null
  const visibleSeries = getVisibleMetricSeries(metric.chartData)
  if (!visibleSeries.length) return null

  return (
    <div className={styles.metricSparklineGrid}>
      <div className={styles.metricSparkline}>
        <MetricSparkline seriesList={visibleSeries} />
      </div>
    </div>
  )
}

function RootCauseEvidenceCard({ item }: { item: RootCauseEvidenceItem }) {
  const hasChart = Boolean(item.chartData?.length)

  return (
    <article
      className={hasChart ? styles.metricSummaryCardWithChart : styles.metricSummaryCardNoChart}
    >
      <div className={styles.metricSummaryHeader}>
        <span className={styles.metricSummaryTitle}>{item.name}</span>
        {item.valueSummary ? <strong>{item.valueSummary}</strong> : <em>{item.typeLabel}</em>}
      </div>
      <p>{item.description}</p>
      {hasChart ? <MetricSparklineGroup metric={item} /> : null}
    </article>
  )
}

export function ReportDocument({ view }: { view: ReportViewModel }) {
  const currentTheme = useSelector((state: { settingReducer?: { theme?: string } }) =>
    getAppTheme(state.settingReducer?.theme),
  )
  const [activeEvidenceIndex, setActiveEvidenceIndex] = useState<number | null>(null)
  const [openActionIndex, setOpenActionIndex] = useState<number | null>(null)
  const [showAllActions, setShowAllActions] = useState(false)
  const [isAnalyzeProcessOpen, setIsAnalyzeProcessOpen] = useState(false)
  const [isRawEvidenceOpen, setIsRawEvidenceOpen] = useState(false)
  const activeEvidenceTimeoutRef = useRef<number | null>(null)
  const evidenceScrollFrameRef = useRef<number | null>(null)

  useEffect(() => {
    return () => {
      if (activeEvidenceTimeoutRef.current !== null) {
        window.clearTimeout(activeEvidenceTimeoutRef.current)
      }
      if (evidenceScrollFrameRef.current !== null) {
        window.cancelAnimationFrame(evidenceScrollFrameRef.current)
      }
    }
  }, [])

  const detailEvidence = view.evidence.filter((item) => item.type !== 'endpoints')
  const canJumpToEvidence = (index: number) =>
    Boolean(
      Number.isInteger(index) && view.evidence[index] && view.evidence[index].type !== 'endpoints',
    )

  const scrollToEvidenceItem = (evidenceIndex: number) => {
    const target = document.getElementById(getEvidenceAnchorId(evidenceIndex))
    if (!target) return

    target.scrollIntoView({ behavior: 'smooth', block: 'start' })
    setActiveEvidenceIndex(evidenceIndex)

    if (activeEvidenceTimeoutRef.current !== null) {
      window.clearTimeout(activeEvidenceTimeoutRef.current)
    }

    activeEvidenceTimeoutRef.current = window.setTimeout(() => {
      setActiveEvidenceIndex((current) => (current === evidenceIndex ? null : current))
      activeEvidenceTimeoutRef.current = null
    }, 1800)
  }

  const handleEvidenceSummaryClick = (evidenceIndex: number) => {
    setIsRawEvidenceOpen(true)

    if (evidenceScrollFrameRef.current !== null) {
      window.cancelAnimationFrame(evidenceScrollFrameRef.current)
    }

    evidenceScrollFrameRef.current = window.requestAnimationFrame(() => {
      scrollToEvidenceItem(evidenceIndex)
      evidenceScrollFrameRef.current = null
    })
  }

  const hasRootCauseEvidence = view.rootCauseEvidence.length > 0
  const hasSnapshotAside = hasRootCauseEvidence
  const hasImpact = view.impactGroups.length > 0
  const hasTimeline = false
  const hasAnalyzeProcess = Boolean(view.analyzeProcess?.trim())
  const hasActions = view.actions.length > 0
  const visibleActions = showAllActions
    ? view.actions
    : view.actions.slice(0, DEFAULT_VISIBLE_ACTION_COUNT)
  const hiddenActionCount = Math.max(view.actions.length - DEFAULT_VISIBLE_ACTION_COUNT, 0)
  const hasHiddenActions = hiddenActionCount > 0
  const hasEvidence = view.evidenceSummary.length > 0
  const hasEvidencePanel = hasEvidence || hasImpact
  const hasRawEvidence = detailEvidence.length > 0
  const hasCandidates = view.candidates.length > 0
  const impactSummary =
    view.businessImpact.summary && view.businessImpact.summary !== view.overview
      ? view.businessImpact.summary
      : ''

  return (
    <main className={styles.reportPage}>
      <section className={styles.conclusionSection}>
        <div
          className={`${styles.snapshotGrid} ${!hasSnapshotAside ? styles.snapshotGridSingle : ''}`}
        >
          <section className={styles.snapshotMain}>
            <div className={styles.snapshotHeader}>
              <div>
                <div className={styles.metaLine}>
                  <span className={styles.typeBadge}>{view.typeLabel}</span>
                  {view.status.label && (
                    <span className={`${styles.statusPill} ${statusClass[view.status.state]}`}>
                      {view.status.label}
                    </span>
                  )}
                </div>
                <h2>{view.headline}</h2>
                <p>{view.analysisSummary}</p>
                {impactSummary && (
                  <p className={styles.impactInsight}>
                    <span>影响判断</span>
                    {impactSummary}
                  </p>
                )}
              </div>
            </div>
            <div className={styles.summaryCards}>
              {view.summaryCards.map((card) => (
                <article
                  className={`${styles.summaryCard} ${summaryCardClass[card.tone]}`}
                  key={card.label}
                >
                  <span className={styles.summaryCardLabel}>{card.label}</span>
                  {card.tone === 'rootCause' && view.rootCause && (
                    <em className={styles.rootCauseCauseTag} title={view.rootCause}>
                      {view.rootCause}
                    </em>
                  )}
                  <strong>{card.title}</strong>
                  {card.description && <p>{card.description}</p>}
                </article>
              ))}
              {view.faultLocation.service && (
                <article className={`${styles.summaryCard} ${styles.summaryLocation}`}>
                  <span>{view.faultLocation.label || '故障位置'}</span>
                  <strong>{view.faultLocation.service}</strong>
                  {view.faultLocation.description && <p>{view.faultLocation.description}</p>}
                </article>
              )}
            </div>
          </section>

          {hasSnapshotAside && (
            <aside className={styles.snapshotAside}>
              <section className={styles.snapshotPanel}>
                <div className={styles.panelTitle}>
                  <CircleGauge size={16} />
                  <span>支撑证据</span>
                </div>
                <div className={styles.metricSummaryList}>
                  {view.rootCauseEvidence.slice(0, 3).map((item) => (
                    <RootCauseEvidenceCard item={item} key={`${item.type}-${item.name}`} />
                  ))}
                </div>
              </section>
            </aside>
          )}
        </div>
      </section>

      {hasActions && (
        <section className={`${styles.detailPanel} ${styles.actionPanel}`}>
          <div className={styles.sectionHeader}>
            <span>
              <Activity size={16} />
              处置建议
            </span>
            <h2>按优先级执行</h2>
          </div>
          <div
            className={`${styles.actionList} ${!showAllActions ? styles.actionListCollapsed : ''}`}
          >
            {visibleActions.map((action, index) => {
              const isActionOpen = openActionIndex === index

              return (
                <article key={`${action.action}-${index}`}>
                  {action.details.length > 0 ? (
                    <button
                      type="button"
                      className={styles.actionRowToggle}
                      aria-expanded={isActionOpen}
                      onClick={() => setOpenActionIndex(isActionOpen ? null : index)}
                    >
                      <span className={getActionLevelClass(action.level)}>
                        {getActionLevelLabel(action.level)}
                      </span>
                      <strong>{action.displayTitle}</strong>
                      <em>{isActionOpen ? '收起' : '详情'}</em>
                    </button>
                  ) : (
                    <div className={styles.actionRowStatic}>
                      <span className={getActionLevelClass(action.level)}>
                        {getActionLevelLabel(action.level)}
                      </span>
                      <strong>{action.displayTitle}</strong>
                    </div>
                  )}
                  {isActionOpen && (
                    <div className={styles.actionDetailList}>
                      {action.details.map((detail) => (
                        <p key={`${detail.label}-${detail.text}`}>
                          <span>{detail.label}</span>
                          {detail.text}
                        </p>
                      ))}
                    </div>
                  )}
                </article>
              )
            })}
            {hasHiddenActions && (
              <button
                type="button"
                className={styles.actionListToggle}
                onClick={() => {
                  setShowAllActions((current) => !current)
                  setOpenActionIndex(null)
                }}
              >
                {showAllActions ? '收起剩余建议' : `展开剩余 ${hiddenActionCount} 条建议`}
              </button>
            )}
          </div>
        </section>
      )}

      <section className={styles.reportBody}>
        <div className={styles.reportMainColumn}>
          {hasTimeline && (
            <section className={styles.detailPanel}>
              <div className={styles.sectionHeader}>
                <span>
                  <Clock3 size={16} />
                  事件时间线
                </span>
              </div>
              <AnalysisProcess timeline={view.timeline} />
            </section>
          )}

          {hasAnalyzeProcess && (
            <section className={`${styles.detailPanel} ${styles.analysisProcessPanel}`}>
              <button
                type="button"
                className={styles.analysisProcessToggle}
                aria-expanded={isAnalyzeProcessOpen}
                onClick={() => setIsAnalyzeProcessOpen((current) => !current)}
              >
                <span>
                  <FileSearch size={16} />
                  分析过程
                </span>
                <em>{isAnalyzeProcessOpen ? '收起' : '展开'}</em>
              </button>
              {isAnalyzeProcessOpen && (
                <div className={styles.analysisProcessMarkdown}>
                  <MarkdownRender content={view.analyzeProcess} theme={currentTheme} />
                </div>
              )}
            </section>
          )}

          {hasEvidencePanel && (
            <section className={styles.detailPanel}>
              <div className={styles.sectionHeader}>
                <span>
                  <FileSearch size={16} />
                  关键证据
                </span>
                <h2>{hasEvidence ? '证据摘要' : '定位依据'}</h2>
              </div>
              {hasEvidence && (
                <div className={styles.evidenceSummaryList}>
                  {view.evidenceSummary.map((item) => {
                    const canJump = canJumpToEvidence(item.evidenceIndex)
                    const jumpTargetId = canJump
                      ? getEvidenceAnchorId(item.evidenceIndex)
                      : undefined
                    const handleSummaryJump = () => handleEvidenceSummaryClick(item.evidenceIndex)

                    return (
                      <article
                        className={canJump ? styles.evidenceSummaryItemInteractive : ''}
                        key={`${item.evidenceIndex}-${item.type}-${item.name}`}
                        role={canJump ? 'button' : undefined}
                        tabIndex={canJump ? 0 : undefined}
                        aria-controls={jumpTargetId}
                        onClick={canJump ? handleSummaryJump : undefined}
                        onKeyDown={
                          canJump
                            ? (event) => {
                                if (event.key !== 'Enter' && event.key !== ' ') return
                                event.preventDefault()
                                handleSummaryJump()
                              }
                            : undefined
                        }
                      >
                        <span>{item.typeLabel}</span>
                        <div>
                          <strong>{item.name}</strong>
                          {item.description && <p>{item.description}</p>}
                          {item.valueSummary && <p>摘要：{item.valueSummary}</p>}
                          {canJump && (
                            <span className={styles.evidenceJumpButton}>查看原始证据</span>
                          )}
                        </div>
                      </article>
                    )
                  })}
                </div>
              )}
              {hasImpact && (
                <div className={styles.impactBasisInline} aria-label="定位依据">
                  <span className={styles.impactBasisTitle}>定位依据</span>
                  <div className={styles.impactBasisGroups}>
                    {view.impactGroups.map((group) => {
                      const visibleValues = group.values.slice(0, 3)
                      const hiddenCount = group.values.length - visibleValues.length

                      return (
                        <span className={styles.impactBasisGroup} key={group.label}>
                          <strong>{group.label}</strong>
                          <span>{visibleValues.join('、')}</span>
                          {hiddenCount > 0 && <em>+{hiddenCount}</em>}
                        </span>
                      )
                    })}
                  </div>
                </div>
              )}
              {hasRawEvidence && (
                <section className={styles.rawEvidenceInline} aria-label="原始证据">
                  <button
                    type="button"
                    className={styles.rawEvidenceInlineHeader}
                    aria-expanded={isRawEvidenceOpen}
                    aria-controls="report-raw-evidence-list"
                    onClick={() => setIsRawEvidenceOpen((current) => !current)}
                  >
                    <span>原始证据</span>
                    <em>{detailEvidence.length} 项</em>
                    <strong>{isRawEvidenceOpen ? '收起' : '展开'}</strong>
                  </button>
                  {isRawEvidenceOpen && (
                    <div id="report-raw-evidence-list">
                      <SupportingEvidence
                        evidence={view.evidence as any}
                        startTime={view.startTime}
                        endTime={view.endTime}
                        getEvidenceItemId={getEvidenceAnchorId}
                        activeEvidenceIndex={activeEvidenceIndex}
                        activeEvidenceClassName={styles.rawEvidenceActive}
                      />
                    </div>
                  )}
                </section>
              )}
            </section>
          )}

          {hasCandidates && (
            <section className={`${styles.candidatePanel} ${styles.mutedPanel}`}>
              <div className={styles.sectionHeader}>
                <span>
                  <ListChecks size={16} />
                  排查方向排序
                </span>
                <h2>候选方向与排除项</h2>
              </div>

              <div className={styles.candidateTable}>
                <div className={styles.candidateHead}>
                  <span>排名</span>
                  <span>方向</span>
                  <span>判断</span>
                  <span>依据</span>
                  <span>结论</span>
                </div>
                {view.candidates.map((candidate) => (
                  <div
                    className={styles.candidateRow}
                    key={`${candidate.rank}-${candidate.direction}`}
                  >
                    <span>{candidate.rank}</span>
                    <strong>{candidate.direction}</strong>
                    <em className={candidateStateClass[candidate.state]}>{candidate.judgement}</em>
                    <p>{candidate.evidence}</p>
                    <span>{candidate.conclusion}</span>
                  </div>
                ))}
              </div>
            </section>
          )}
        </div>
      </section>
    </main>
  )
}

function Report() {
  const [searchParams] = useSearchParams()
  const timeRange = useSelector(
    (state: {
      timeRange?: { startTime?: number | string | null; endTime?: number | string | null }
    }) => state.timeRange,
  )
  const request = useMemo(
    () => buildReportViewRequest(searchParams, timeRange),
    [searchParams, timeRange?.startTime, timeRange?.endTime],
  )
  const [view, setView] = useState<ReportViewModel>()
  const [loading, setLoading] = useState(true)
  const [errorMessage, setErrorMessage] = useState('')

  useEffect(() => {
    let ignored = false

    if (!request) {
      setView(undefined)
      setErrorMessage('缺少报告参数，需要 alertEventId、startTime 和 endTime。')
      setLoading(false)
      return () => {
        ignored = true
      }
    }

    setLoading(true)
    setErrorMessage('')

    getReportViewApi(request)
      .then((response: AlertEventReportViewResponse) => {
        if (ignored) return

        setView(
          createReportViewModelFromViewResponse(response, {
            id: getRequestReportId(request),
            startTime: request.startTime,
            endTime: request.endTime,
            dataSourceLabel: '报告快照',
          }),
        )
      })
      .catch((error: unknown) => {
        if (ignored) return

        setView(undefined)
        setErrorMessage(getLoadErrorMessage(error))
      })
      .finally(() => {
        if (!ignored) {
          setLoading(false)
        }
      })

    return () => {
      ignored = true
    }
  }, [request])

  if (loading) {
    return (
      <main className={styles.reportPage}>
        <section className={styles.statePanel}>
          <FileSearch size={22} />
          <strong>正在加载诊断报告</strong>
          <p>正在读取报告快照和页面视图数据。</p>
        </section>
      </main>
    )
  }

  if (!view) {
    return (
      <main className={styles.reportPage}>
        <section className={styles.statePanel}>
          <AlertCircle size={22} />
          <strong>无法展示诊断报告</strong>
          <p>{errorMessage || '未获取到报告数据。'}</p>
        </section>
      </main>
    )
  }

  return <ReportDocument view={view} />
}

export default Report
