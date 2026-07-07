import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import {
  Alert,
  Button,
  message,
  Popover,
  Select,
  Spin,
  Tag,
} from 'antd'
import {
  CircleAlert,
  ListTodo,
  Settings,
  TriangleAlert,
} from 'lucide-react'
import { useSelector } from 'react-redux'
import { useLocation } from 'react-router-dom'
import { useDebounce } from 'react-use'
import {
  analyzeSystemHealthPendingIssueApi,
  getSystemHealthOverviewApi,
  getSystemHealthPendingIssuesApi,
  getSystemHealthServicesApi,
  getSystemHealthSystemDetailApi,
  getSystemHealthSystemsApi,
  getSystemHealthSystemsTrendApi,
  getSystemHealthTrendApi,
  setSystemHealthConfigApi,
} from 'src/pro/api/serviceHealthy'
import {
  buildOverviewSummaryCards,
  buildStatusDistribution,
  DEFAULT_HEALTH_SCORE_THRESHOLDS,
  normalizeHealthScoreThresholds,
} from './metricVisuals.mjs'
import { buildPendingIssueQueueRows } from './pendingIssueQueue.mjs'
import {
  DEFAULT_PENDING_ISSUE_PAGE,
  DEFAULT_PENDING_ISSUE_PAGE_SIZE,
} from './pendingIssuePagination.mjs'
import {
  buildAbnormalServiceRatio,
  buildDetailHeaderRequestKey,
  DEFAULT_HEALTH_FILTER_VALUE,
  DEFAULT_HEALTH_FILTER_VALUES,
  shouldSyncDetailHeaderFromRows,
  shouldRenderDetailBody,
} from './detailViewState.mjs'
import {
  buildOverviewRequestDecision,
  isLatestOverviewRequest,
  overviewServiceEmptyMessage,
  overviewSystemEmptyMessage,
  shouldRenderInitialOverviewLoading,
} from './overviewViewState.mjs'
import {
  buildOverviewRequestKey,
  buildOverviewRequestParams,
  makeTimeRangeSignature,
} from './timeParams.mjs'
import {
  buildDetailTrendRows,
  buildTrendChartRows,
  readStoredTrendRange,
  trendMetricGroups,
  trendRangeOptions,
  writeStoredTrendRange,
} from './trendViewState.mjs'
import { DetailLoadingPanel } from './components/DetailLoading.jsx'
import DetailHeader from './components/DetailHeader.jsx'
import DetailTrendSection from './components/DetailTrendSection.jsx'
import {
  Legend,
  statusMeta,
} from './components/HealthStatus.jsx'
import {
  HealthScoreCalculationTooltip,
  HealthScoreColumnTitle,
  HealthScoreRangeHelpIcon,
} from './components/HealthScoreHelp.jsx'
import MetricVisualTile from './components/MetricVisualTile.jsx'
import OverviewInitialLoading from './components/OverviewInitialLoading.jsx'
import OverviewSummary from './components/OverviewSummary.jsx'
import OverviewSystemTable from './components/OverviewSystemTable.jsx'
import PendingIssueQueue from './components/PendingIssueQueue.jsx'
import ServiceHealthTable from './components/ServiceHealthTable.jsx'
import ThresholdSettingsContent from './components/ThresholdSettingsContent.jsx'
import { getTrendRangeTitleLabel } from './components/TrendCommon.jsx'
import './index.css'

const metricLabels = [
  ['resource', '资源水位'],
  ['resourceShift', '资源波动'],
  ['errorRate', '错误率'],
  ['latency', '响应延时'],
  ['errorLog', '错误日志'],
  ['badRequests', '活跃告警'],
  ['lockCost', '锁等待'],
]

const OVERVIEW_REQUEST_DEBOUNCE_MS = 150
const SYSTEM_SEARCH_DEBOUNCE_MS = 300
const SERVICE_SEARCH_DEBOUNCE_MS = 300
const DEFAULT_SYSTEM_PAGE = 1
const DEFAULT_SYSTEM_PAGE_SIZE = 50
const DEFAULT_SERVICE_PAGE = 1
const DEFAULT_SERVICE_PAGE_SIZE = 50
const systemHealthFilterOptions = [
  { label: '异常系统', value: DEFAULT_HEALTH_FILTER_VALUES.abnormal },
  { label: '关注系统', value: DEFAULT_HEALTH_FILTER_VALUES.attention },
  { label: '健康系统', value: DEFAULT_HEALTH_FILTER_VALUES.healthy },
  { label: '无数据系统', value: DEFAULT_HEALTH_FILTER_VALUES.noData },
  { label: '全部', value: DEFAULT_HEALTH_FILTER_VALUES.all },
]
const serviceHealthFilterOptions = [
  { label: '异常服务', value: DEFAULT_HEALTH_FILTER_VALUES.abnormal },
  { label: '关注服务', value: DEFAULT_HEALTH_FILTER_VALUES.attention },
  { label: '健康服务', value: DEFAULT_HEALTH_FILTER_VALUES.healthy },
  { label: '无数据服务', value: DEFAULT_HEALTH_FILTER_VALUES.noData },
  { label: '全部', value: DEFAULT_HEALTH_FILTER_VALUES.all },
]

const summaryCardIcons = {
  abnormalSystems: (
    <TriangleAlert className="sho-summary-title-icon sho-summary-title-icon-critical" size={17} />
  ),
  attentionSystems: (
    <CircleAlert className="sho-summary-title-icon sho-summary-title-icon-warning" size={17} />
  ),
  pendingIssues: (
    <ListTodo className="sho-summary-title-icon sho-summary-title-icon-critical" size={17} />
  ),
}

const summaryTrendGroupByKey = trendMetricGroups.reduce((result, group) => {
  result[group.key] = group
  return result
}, {})

const positiveNumber = (value) => {
  const number = Number(value)
  return Number.isFinite(number) && number > 0 ? number : 0
}

const scoreThresholdsEqual = (left, right) => {
  const normalizedLeft = normalizeHealthScoreThresholds(left)
  const normalizedRight = normalizeHealthScoreThresholds(right)
  return (
    normalizedLeft.severeAbnormalScoreThreshold === normalizedRight.severeAbnormalScoreThreshold &&
    normalizedLeft.generalAbnormalScoreThreshold === normalizedRight.generalAbnormalScoreThreshold
  )
}

const normalizeSystem = (system = {}) => ({
  id: system.id ?? system.groupId,
  groupId: system.groupId ?? system.id,
  name: system.name || '未命名系统',
  score: Number.isFinite(Number(system.score)) ? Number(system.score) : 0,
  status: system.status || 'gray',
  metrics: normalizeMetrics(system.metrics),
  metricDetails: system.metricDetails || {},
  services: Array.isArray(system.services) ? system.services.map(normalizeService) : [],
  serviceTotal: Number.isFinite(Number(system.serviceTotal))
    ? Number(system.serviceTotal)
    : Array.isArray(system.services)
      ? system.services.length
      : 0,
  abnormalServiceCount: Number.isFinite(Number(system.abnormalServiceCount))
    ? Number(system.abnormalServiceCount)
    : 0,
  attentionServiceCount: Number.isFinite(Number(system.attentionServiceCount))
    ? Number(system.attentionServiceCount)
    : 0,
  source: system.source || 'data_group',
})

const normalizePendingIssue = (issue = {}) => ({
  id: issue.id || `${issue.groupId || ''}:${issue.serviceName || ''}:${issue.metricKey || ''}`,
  groupId: issue.groupId,
  systemName: issue.systemName || '未命名系统',
  serviceName: issue.serviceName || '未命名服务',
  metricKey: issue.metricKey || '',
  metricLabel: issue.metricLabel || issue.metricKey || '未知指标',
  status: issue.status || 'gray',
  value: issue.value || '--',
  note: issue.note || '暂无指标说明',
})

const buildSystemServiceRiskCounts = (
  system,
  scoreThresholds = DEFAULT_HEALTH_SCORE_THRESHOLDS,
) => {
  if (!system) return { abnormal: 0, attention: 0, total: 0 }
  const total = positiveNumber(system.serviceTotal)
  const abnormal = positiveNumber(system.abnormalServiceCount)
  const attention = positiveNumber(system.attentionServiceCount)
  if (total || abnormal || attention) {
    return { abnormal, attention, total }
  }
  return buildAbnormalServiceRatio(system.services, scoreThresholds)
}

const normalizeService = (service = {}) => ({
  name: service.name || '未命名服务',
  desc: service.desc || '',
  score: Number.isFinite(Number(service.score)) ? Number(service.score) : 0,
  status: service.status || 'gray',
  metrics: normalizeMetrics(service.metrics),
  metricDetails: service.metricDetails || {},
})

const normalizeMetrics = (metrics = {}) =>
  metricLabels.reduce((result, [key]) => {
    result[key] = metrics[key] || 'gray'
    return result
  }, {})

const statusTags = (metrics, status, emptyText) => {
  const tags = metricLabels
    .map(([key, label]) => ({ label, status: metrics[key] }))
    .filter((item) => item.status === status)

  if (!tags.length) return <Tag>{emptyText}</Tag>
  return tags.map((item) => (
    <Tag className={`sho-tag sho-tag-${status}`} key={item.label}>
      {item.label}
    </Tag>
  ))
}

const metricDetail = (target, key) => {
  const detail = target?.metricDetails?.[key]
  if (detail) return detail
  const status = target?.metrics?.[key] || 'gray'
  return {
    status,
    value: '--',
    note: status === 'gray' ? '暂无数据' : `${statusMeta[status]?.label || '未知'}状态`,
  }
}

export default function SystemHealthOverviewPage() {
  const location = useLocation()
  const timeRange = useSelector((state) => state.timeRange)
  const pageLoadedAtRef = useRef(Date.now() * 1000)
  const initialTimeRangeSignatureRef = useRef(null)
  const overviewRequestSeqRef = useRef(0)
  const lastOverviewRequestKeyRef = useRef('')
  const overviewRequestTimerRef = useRef(null)
  const [systems, setSystems] = useState([])
  const [systemRowsRequestKey, setSystemRowsRequestKey] = useState('')
  const [systemPage, setSystemPage] = useState(DEFAULT_SYSTEM_PAGE)
  const [systemPageSize, setSystemPageSize] = useState(DEFAULT_SYSTEM_PAGE_SIZE)
  const [systemPageTotal, setSystemPageTotal] = useState(0)
  const [systemRowsLoading, setSystemRowsLoading] = useState(false)
  const [systemRowsError, setSystemRowsError] = useState('')
  const [systemRowsRefreshToken, setSystemRowsRefreshToken] = useState(0)
  const [summary, setSummary] = useState({})
  const [scoreThresholds, setScoreThresholds] = useState(DEFAULT_HEALTH_SCORE_THRESHOLDS)
  const [thresholdDraft, setThresholdDraft] = useState(DEFAULT_HEALTH_SCORE_THRESHOLDS)
  const [thresholdSaving, setThresholdSaving] = useState(false)
  const [thresholdSettingsOpen, setThresholdSettingsOpen] = useState(false)
  const [pendingIssueItems, setPendingIssueItems] = useState([])
  const [pendingIssuePage, setPendingIssuePage] = useState(DEFAULT_PENDING_ISSUE_PAGE)
  const [pendingIssuePageSize, setPendingIssuePageSize] = useState(DEFAULT_PENDING_ISSUE_PAGE_SIZE)
  const [pendingIssueTotal, setPendingIssueTotal] = useState(0)
  const [pendingIssueLoading, setPendingIssueLoading] = useState(false)
  const [pendingIssueError, setPendingIssueError] = useState('')
  const [issueAnalyses, setIssueAnalyses] = useState({})
  const [expandedIssueAnalyses, setExpandedIssueAnalyses] = useState({})
  const [overviewLoading, setOverviewLoading] = useState(false)
  const [overviewError, setOverviewError] = useState('')
  const [trendRange, setTrendRange] = useState(() => readStoredTrendRange())
  const [trendData, setTrendData] = useState(null)
  const [trendLoading, setTrendLoading] = useState(false)
  const [trendError, setTrendError] = useState('')
  const [overviewTrendRefreshToken, setOverviewTrendRefreshToken] = useState(0)
  const [detailTrendRefreshToken, setDetailTrendRefreshToken] = useState(0)
  const [systemTrendRows, setSystemTrendRows] = useState({})
  const [systemTrendGranularity, setSystemTrendGranularity] = useState('')
  const [systemTrendLoading, setSystemTrendLoading] = useState(false)
  const [systemTrendError, setSystemTrendError] = useState('')
  const [detailTrendRows, setDetailTrendRows] = useState([])
  const [detailTrendGranularity, setDetailTrendGranularity] = useState('')
  const [detailTrendLoading, setDetailTrendLoading] = useState(false)
  const [detailTrendError, setDetailTrendError] = useState('')
  const [selectedSystemId, setSelectedSystemId] = useState('')
  const [selectedSystem, setSelectedSystem] = useState(null)
  const [detailHeaderLoading, setDetailHeaderLoading] = useState(false)
  const [serviceRows, setServiceRows] = useState([])
  const [servicePage, setServicePage] = useState(DEFAULT_SERVICE_PAGE)
  const [servicePageSize, setServicePageSize] = useState(DEFAULT_SERVICE_PAGE_SIZE)
  const [servicePageTotal, setServicePageTotal] = useState(0)
  const [serviceRowsLoading, setServiceRowsLoading] = useState(false)
  const [serviceRowsError, setServiceRowsError] = useState('')
  const [systemHealthFilter, setSystemHealthFilter] = useState(DEFAULT_HEALTH_FILTER_VALUE)
  const [systemKeyword, setSystemKeyword] = useState('')
  const [debouncedSystemKeyword, setDebouncedSystemKeyword] = useState('')
  const [serviceKeyword, setServiceKeyword] = useState('')
  const [debouncedServiceKeyword, setDebouncedServiceKeyword] = useState('')
  const [serviceHealthFilter, setServiceHealthFilter] = useState(DEFAULT_HEALTH_FILTER_VALUE)
  const lastDetailHeaderRequestKeyRef = useRef('')

  const currentTimeRangeSignature = makeTimeRangeSignature(timeRange)
  if (initialTimeRangeSignatureRef.current === null) {
    initialTimeRangeSignatureRef.current = currentTimeRangeSignature
  }
  const timeRangeChangedAfterMount =
    currentTimeRangeSignature !== initialTimeRangeSignatureRef.current
  const timeRangeRefreshedAfterMount =
    positiveNumber(timeRange?.refreshTimestamp) >= pageLoadedAtRef.current
  const queryParams = useMemo(
    () =>
      buildOverviewRequestParams(location.search, timeRange, {
        preferStore: timeRangeChangedAfterMount || timeRangeRefreshedAfterMount,
        requireResolvedTimeRange: true,
      }),
    [
      location.search,
      timeRange?.rangeTypeKey,
      timeRange?.startTime,
      timeRange?.endTime,
      timeRange?.refreshTimestamp,
      timeRangeChangedAfterMount,
      timeRangeRefreshedAfterMount,
    ],
  )
  const overviewRequestKey = useMemo(() => buildOverviewRequestKey(queryParams), [queryParams])

  useDebounce(
    () => {
      setDebouncedSystemKeyword(systemKeyword.trim())
    },
    SYSTEM_SEARCH_DEBOUNCE_MS,
    [systemKeyword],
  )

  useDebounce(
    () => {
      setDebouncedServiceKeyword(serviceKeyword.trim())
    },
    SERVICE_SEARCH_DEBOUNCE_MS,
    [serviceKeyword],
  )

  const loadOverview = useCallback(async () => {
    if (!queryParams) return
    const requestId = overviewRequestSeqRef.current + 1
    overviewRequestSeqRef.current = requestId
    setOverviewLoading(true)
    setOverviewError('')
    try {
      const resp = await getSystemHealthOverviewApi({
        ...queryParams,
        skipNewAbnormalMetrics: true,
        summaryOnly: true,
      })
      if (!isLatestOverviewRequest(requestId, overviewRequestSeqRef.current)) return
      const nextThresholds = normalizeHealthScoreThresholds(resp?.config)
      setScoreThresholds(nextThresholds)
      setThresholdDraft(nextThresholds)
      setSummary(resp?.summary || {})
    } catch (error) {
      if (!isLatestOverviewRequest(requestId, overviewRequestSeqRef.current)) return
      setOverviewError(error?.message || '健康度加载失败')
      setSummary({})
    } finally {
      if (isLatestOverviewRequest(requestId, overviewRequestSeqRef.current)) {
        setOverviewLoading(false)
      }
    }
  }, [queryParams])

  useEffect(() => {
    const decision = buildOverviewRequestDecision({
      requestKey: overviewRequestKey,
      lastRequestKey: lastOverviewRequestKeyRef.current,
      overviewLoading,
    })
    if (!decision.shouldLoad) {
      lastOverviewRequestKeyRef.current = decision.nextLastRequestKey
      return
    }
    if (overviewRequestTimerRef.current) {
      window.clearTimeout(overviewRequestTimerRef.current)
    }
    overviewRequestTimerRef.current = window.setTimeout(() => {
      overviewRequestTimerRef.current = null
      lastOverviewRequestKeyRef.current = overviewRequestKey
      loadOverview()
    }, OVERVIEW_REQUEST_DEBOUNCE_MS)
    return () => {
      if (overviewRequestTimerRef.current) {
        window.clearTimeout(overviewRequestTimerRef.current)
        overviewRequestTimerRef.current = null
      }
    }
  }, [loadOverview, overviewLoading, overviewRequestKey])

  useEffect(() => {
    setSystemPage(DEFAULT_SYSTEM_PAGE)
  }, [queryParams, systemHealthFilter, debouncedSystemKeyword])

  useEffect(() => {
    if (!queryParams) return undefined

    let cancelled = false
    const requestKey = overviewRequestKey
    const loadSystemRows = async () => {
      setSystemRowsLoading(true)
      setSystemRowsError('')
      try {
        const resp = await getSystemHealthSystemsApi({
          ...queryParams,
          page: systemPage,
          pageSize: systemPageSize,
          keyword: debouncedSystemKeyword,
          status: systemHealthFilter || DEFAULT_HEALTH_FILTER_VALUES.all,
          sort: 'score_asc',
        })
        if (cancelled) return
        setSystems(Array.isArray(resp?.systems) ? resp.systems.map(normalizeSystem) : [])
        setSystemRowsRequestKey(requestKey)
        setSystemPageTotal(positiveNumber(resp?.total))
      } catch (error) {
        if (!cancelled) {
          setSystemRowsError(error?.message || '系统列表加载失败')
          setSystems([])
          setSystemRowsRequestKey(requestKey)
          setSystemPageTotal(0)
        }
      } finally {
        if (!cancelled) {
          setSystemRowsLoading(false)
        }
      }
    }

    loadSystemRows()
    return () => {
      cancelled = true
    }
  }, [
    queryParams,
    systemHealthFilter,
    debouncedSystemKeyword,
    overviewRequestKey,
    systemPage,
    systemPageSize,
    systemRowsRefreshToken,
  ])

  useEffect(() => {
    if (selectedSystemId) {
      setTrendLoading(false)
      return undefined
    }

    let cancelled = false
    const loadTrend = async () => {
      setTrendLoading(true)
      setTrendError('')
      try {
        const resp = await getSystemHealthTrendApi({
          range: trendRange,
          judgement: 'snapshot',
          view: 'summary',
        })
        if (!cancelled) {
          setTrendData(resp || null)
        }
      } catch (error) {
        if (!cancelled) {
          setTrendError(error?.message || '历史趋势加载失败')
          setTrendData(null)
        }
      } finally {
        if (!cancelled) {
          setTrendLoading(false)
        }
      }
    }

    loadTrend()
    return () => {
      cancelled = true
    }
  }, [trendRange, overviewTrendRefreshToken, selectedSystemId])

  useEffect(() => {
    setPendingIssuePage(DEFAULT_PENDING_ISSUE_PAGE)
  }, [queryParams])

  useEffect(() => {
    if (!queryParams) return undefined

    let cancelled = false
    const loadPendingIssues = async () => {
      setPendingIssueLoading(true)
      setPendingIssueError('')
      try {
        const resp = await getSystemHealthPendingIssuesApi({
          ...queryParams,
          page: pendingIssuePage,
          pageSize: pendingIssuePageSize,
          status: 'abnormal',
        })
        if (cancelled) return
        setPendingIssueItems(Array.isArray(resp?.issues) ? resp.issues.map(normalizePendingIssue) : [])
        setPendingIssueTotal(positiveNumber(resp?.total))
      } catch (error) {
        if (!cancelled) {
          setPendingIssueError(error?.message || '待处理问题加载失败')
          setPendingIssueItems([])
          setPendingIssueTotal(0)
        }
      } finally {
        if (!cancelled) {
          setPendingIssueLoading(false)
        }
      }
    }

    loadPendingIssues()
    return () => {
      cancelled = true
    }
  }, [pendingIssuePage, pendingIssuePageSize, queryParams])

  useEffect(() => {
    setServicePage(DEFAULT_SERVICE_PAGE)
  }, [queryParams, selectedSystemId, serviceHealthFilter, debouncedServiceKeyword])

  useEffect(() => {
    if (!selectedSystemId || !queryParams) {
      setServiceRows([])
      setServicePageTotal(0)
      setServiceRowsError('')
      setServiceRowsLoading(false)
      return undefined
    }

    let cancelled = false
    const loadServiceRows = async () => {
      setServiceRowsLoading(true)
      setServiceRowsError('')
      try {
        const resp = await getSystemHealthServicesApi(selectedSystemId, {
          ...queryParams,
          page: servicePage,
          pageSize: servicePageSize,
          keyword: debouncedServiceKeyword,
          status: serviceHealthFilter || DEFAULT_HEALTH_FILTER_VALUES.all,
          sort: 'score_asc',
        })
        if (cancelled) return
        setServiceRows(Array.isArray(resp?.services) ? resp.services.map(normalizeService) : [])
        setServicePageTotal(positiveNumber(resp?.total))
      } catch (error) {
        if (!cancelled) {
          setServiceRowsError(error?.message || '服务列表加载失败')
          setServiceRows([])
          setServicePageTotal(0)
        }
      } finally {
        if (!cancelled) {
          setServiceRowsLoading(false)
        }
      }
    }

    loadServiceRows()
    return () => {
      cancelled = true
    }
  }, [
    queryParams,
    selectedSystemId,
    serviceHealthFilter,
    debouncedServiceKeyword,
    servicePage,
    servicePageSize,
  ])

  const overviewSystems = useMemo(
    () => systems,
    [systems],
  )
  const selectedSystemFromRows = useMemo(
    () =>
      selectedSystemId
        ? systems.find((system) => String(system.groupId) === selectedSystemId) || null
        : null,
    [selectedSystemId, systems],
  )
  const overviewSystemTrendGroupIds = useMemo(
    () =>
      overviewSystems
        .map((system) => system.groupId)
        .filter((groupId) => groupId !== undefined && groupId !== null && `${groupId}` !== '')
        .map((groupId) => Number(groupId))
        .filter(Number.isFinite),
    [overviewSystems],
  )
  const overviewSystemTrendGroupKey = useMemo(
    () => overviewSystemTrendGroupIds.join(','),
    [overviewSystemTrendGroupIds],
  )
  const systemTrendRangeLabel = getTrendRangeTitleLabel(trendRange)
  const pendingIssuePageData = useMemo(
    () => ({
      issues: pendingIssueItems,
      total: pendingIssueTotal,
      page: pendingIssuePage,
      pageSize: pendingIssuePageSize,
    }),
    [pendingIssueItems, pendingIssuePage, pendingIssuePageSize, pendingIssueTotal],
  )
  const pendingIssues = pendingIssuePageData.issues
  const pendingIssueRows = useMemo(() => buildPendingIssueQueueRows(pendingIssues), [pendingIssues])

  useEffect(() => {
    if (
      !shouldSyncDetailHeaderFromRows(
        selectedSystemFromRows,
        systemRowsRequestKey,
        overviewRequestKey,
      )
    ) {
      return
    }
    setSelectedSystem(selectedSystemFromRows)
    lastDetailHeaderRequestKeyRef.current = buildDetailHeaderRequestKey(
      selectedSystemId,
      queryParams,
    )
  }, [
    overviewRequestKey,
    queryParams,
    selectedSystemFromRows,
    selectedSystemId,
    systemRowsRequestKey,
  ])

  useEffect(() => {
    if (!selectedSystemId || !queryParams) {
      setDetailHeaderLoading(false)
      return undefined
    }

    const detailHeaderRequestKey = buildDetailHeaderRequestKey(selectedSystemId, queryParams)
    if (!detailHeaderRequestKey || lastDetailHeaderRequestKeyRef.current === detailHeaderRequestKey) {
      setDetailHeaderLoading(false)
      return undefined
    }

    let cancelled = false
    const loadDetailHeader = async () => {
      setDetailHeaderLoading(true)
      try {
        const resp = await getSystemHealthSystemDetailApi(selectedSystemId, {
          ...queryParams,
          includeServices: false,
        })
        if (cancelled || !resp?.system) return
        setSelectedSystem({ ...normalizeSystem(resp.system), services: [] })
        lastDetailHeaderRequestKeyRef.current = detailHeaderRequestKey
      } catch {
        if (!cancelled) {
          lastDetailHeaderRequestKeyRef.current = ''
        }
      } finally {
        if (!cancelled) {
          setDetailHeaderLoading(false)
        }
      }
    }

    loadDetailHeader()
    return () => {
      cancelled = true
    }
  }, [queryParams, selectedSystemId])

  useEffect(() => {
    if (selectedSystemId || !overviewSystemTrendGroupIds.length) {
      setSystemTrendRows({})
      setSystemTrendGranularity('')
      setSystemTrendError('')
      setSystemTrendLoading(false)
      return undefined
    }

    let cancelled = false
    const loadSystemTrendRows = async () => {
      setSystemTrendLoading(true)
      setSystemTrendError('')
      try {
        const resp = await getSystemHealthSystemsTrendApi({
          range: trendRange,
          judgement: 'snapshot',
          groupIds: overviewSystemTrendGroupIds,
        })
        if (cancelled) return

        const nextRows = {}
        const nextGranularity = resp?.granularity || ''
        const groups = Array.isArray(resp?.groups) ? resp.groups : []
        groups.forEach((group) => {
          nextRows[String(group.groupId)] = buildTrendChartRows(group.points, 'systemHealthScore')
        })
        setSystemTrendRows(nextRows)
        setSystemTrendGranularity(nextGranularity)
      } catch (error) {
        if (!cancelled) {
          setSystemTrendError(error?.message || '健康度历史趋势加载失败')
          setSystemTrendRows({})
          setSystemTrendGranularity('')
        }
      } finally {
        if (!cancelled) {
          setSystemTrendLoading(false)
        }
      }
    }

    loadSystemTrendRows()
    return () => {
      cancelled = true
    }
  }, [
    overviewSystemTrendGroupIds,
    overviewSystemTrendGroupKey,
    trendRange,
    overviewTrendRefreshToken,
    selectedSystemId,
  ])

  useEffect(() => {
    if (!selectedSystemId) {
      setDetailTrendRows([])
      setDetailTrendGranularity('')
      setDetailTrendError('')
      setDetailTrendLoading(false)
      return undefined
    }

    let cancelled = false
    const loadDetailTrendRows = async () => {
      setDetailTrendLoading(true)
      setDetailTrendError('')
      try {
        const resp = await getSystemHealthTrendApi({
          range: trendRange,
          groupId: selectedSystemId,
          judgement: 'snapshot',
          view: 'detail',
        })
        if (!cancelled) {
          setDetailTrendRows(buildDetailTrendRows(resp?.points, resp?.granularity || ''))
          setDetailTrendGranularity(resp?.granularity || '')
        }
      } catch (error) {
        if (!cancelled) {
          setDetailTrendError(error?.message || '详情趋势加载失败')
          setDetailTrendRows([])
          setDetailTrendGranularity('')
        }
      } finally {
        if (!cancelled) {
          setDetailTrendLoading(false)
        }
      }
    }

    loadDetailTrendRows()
    return () => {
      cancelled = true
    }
  }, [selectedSystemId, trendRange, detailTrendRefreshToken])

  const pendingServiceCount = useMemo(
    () => new Set(pendingIssueRows.map((issue) => issue.serviceName)).size,
    [pendingIssueRows],
  )
  const abnormalServiceRatio = useMemo(
    () => buildSystemServiceRiskCounts(selectedSystem, scoreThresholds),
    [selectedSystem, scoreThresholds],
  )

  const filteredServices = serviceRows
  const serviceEmptyMessage = overviewServiceEmptyMessage({
    services: filteredServices,
    servicePageTotal,
    serviceHealthFilter,
    serviceKeyword,
    debouncedServiceKeyword,
    allFilterValue: DEFAULT_HEALTH_FILTER_VALUES.all,
  })

  const handleTrendRangeChange = useCallback((nextRange) => {
    setTrendRange(writeStoredTrendRange(nextRange))
  }, [])

  const handleOverviewTrendRetry = () => {
    setOverviewTrendRefreshToken((current) => current + 1)
  }

  const handleDetailTrendRetry = () => {
    setDetailTrendRefreshToken((current) => current + 1)
  }

  const openDetail = (system) => {
    const systemId = String(system.groupId)
    lastDetailHeaderRequestKeyRef.current = buildDetailHeaderRequestKey(systemId, queryParams)
    setSelectedSystemId(systemId)
    setSelectedSystem(system)
    setServiceKeyword('')
    setServiceHealthFilter(DEFAULT_HEALTH_FILTER_VALUE)
    setServicePage(DEFAULT_SERVICE_PAGE)
    setServiceRows([])
    setServicePageTotal(0)
    setServiceRowsError('')
  }

  const closeDetail = () => {
    lastDetailHeaderRequestKeyRef.current = ''
    setSelectedSystemId('')
    setSelectedSystem(null)
    setServiceRows([])
    setServicePageTotal(0)
    setServiceRowsError('')
  }

  const toggleIssueAnalysis = (analysisKey) => {
    setExpandedIssueAnalyses((current) => ({
      ...current,
      [analysisKey]: !(current[analysisKey] !== false),
    }))
  }

  const handlePendingIssuePageChange = (page, pageSize) => {
    setPendingIssuePage(page)
    setPendingIssuePageSize(pageSize)
  }

  const handleSystemPageChange = (page, pageSize) => {
    setSystemPage(page)
    setSystemPageSize(pageSize)
  }

  const handleServicePageChange = (page, pageSize) => {
    setServicePage(page)
    setServicePageSize(pageSize)
  }

  const handleThresholdSettingsOpenChange = (open) => {
    setThresholdSettingsOpen(open)
    if (!open) {
      setThresholdDraft(scoreThresholds)
    }
  }

  const saveAbnormalScoreThreshold = async () => {
    const severeThreshold = Number(thresholdDraft.severeAbnormalScoreThreshold)
    const generalThreshold = Number(thresholdDraft.generalAbnormalScoreThreshold)
    if (
      !Number.isFinite(severeThreshold) ||
      !Number.isFinite(generalThreshold) ||
      severeThreshold < 1 ||
      severeThreshold > 99 ||
      generalThreshold < 2 ||
      generalThreshold > 100
    ) {
      message.error('请输入 1-100 内的健康度分值')
      return
    }
    if (severeThreshold >= generalThreshold) {
      message.error('异常分值必须低于关注系统阈值')
      return
    }
    const nextThresholds = normalizeHealthScoreThresholds({
      severeAbnormalScoreThreshold: severeThreshold,
      generalAbnormalScoreThreshold: generalThreshold,
    })
    setThresholdDraft(nextThresholds)
    if (scoreThresholdsEqual(nextThresholds, scoreThresholds)) return

    setThresholdSaving(true)
    try {
      const resp = await setSystemHealthConfigApi({
        abnormalScoreThreshold: nextThresholds.generalAbnormalScoreThreshold,
        severeAbnormalScoreThreshold: nextThresholds.severeAbnormalScoreThreshold,
        generalAbnormalScoreThreshold: nextThresholds.generalAbnormalScoreThreshold,
      })
      const savedThresholds = normalizeHealthScoreThresholds(resp)
      setScoreThresholds(savedThresholds)
      setThresholdDraft(savedThresholds)
      message.success('异常判定已更新')
      await loadOverview()
      setSystemRowsRefreshToken((current) => current + 1)
      setOverviewTrendRefreshToken((current) => current + 1)
    } catch (error) {
      message.error(error?.message || '异常判定保存失败')
      setThresholdDraft(scoreThresholds)
    } finally {
      setThresholdSaving(false)
    }
  }

  const analyzeIssue = async (issue, analysisKey) => {
    setIssueAnalyses((current) => ({
      ...current,
      [analysisKey]: {
        ...(current[analysisKey] || {}),
        loading: true,
        error: '',
      },
    }))
    try {
      const data = await analyzeSystemHealthPendingIssueApi({
        groupId: issue.groupId,
        serviceName: issue.serviceName,
        metricKey: issue.metricKey,
        ...queryParams,
      })
      setIssueAnalyses((current) => ({
        ...current,
        [analysisKey]: {
          data,
          error: '',
          loading: false,
        },
      }))
      if (data?.aiStatus === 'succeeded') {
        setExpandedIssueAnalyses((current) => ({
          ...current,
          [analysisKey]: true,
        }))
      }
    } catch (error) {
      setIssueAnalyses((current) => ({
        ...current,
        [analysisKey]: {
          ...(current[analysisKey] || {}),
          error: error?.message || 'AI 分析失败',
          loading: false,
        },
      }))
    }
  }

  const renderOverviewState = () => {
    if (!queryParams || overviewLoading || systemRowsLoading) {
      return (
        <div className="sho-state-panel">
          <Spin />
          <span>正在加载健康度</span>
        </div>
      )
    }
    const currentError = overviewError || systemRowsError
    if (currentError) {
      return (
        <Alert
          action={
            <Button
              size="small"
              onClick={() => {
                loadOverview()
                setSystemRowsRefreshToken((current) => current + 1)
              }}
            >
              重试
            </Button>
          }
          message="健康度加载失败"
          description={currentError}
          type="error"
          showIcon
        />
      )
    }
    const emptyMessage = overviewSystemEmptyMessage({
      systems,
      systemPageTotal,
      systemHealthFilter,
      systemKeyword,
      debouncedSystemKeyword,
      allFilterValue: DEFAULT_HEALTH_FILTER_VALUES.all,
    })
    if (emptyMessage) {
      return <div className="sho-empty-block">{emptyMessage}</div>
    }
    return null
  }

  const renderOverview = () => {
    const initialOverviewLoading = shouldRenderInitialOverviewLoading({
      overviewLoading: !queryParams || overviewLoading,
      systems,
      pendingIssues,
      summary,
    })
    const overviewState = renderOverviewState()
    const statusDistribution = buildStatusDistribution([], scoreThresholds)
    const summaryCards = buildOverviewSummaryCards(summary, [], scoreThresholds, {
      pendingIssueTotal,
    })
    const trendPoints = Array.isArray(trendData?.points) ? trendData.points : []
    const trendGranularity = trendData?.granularity || ''
    const normalizedThresholdDraft = normalizeHealthScoreThresholds(thresholdDraft)
    const hasThresholdChange = !scoreThresholdsEqual(normalizedThresholdDraft, scoreThresholds)

    return (
      <>
        <div className="sho-toolbar">
          <div>
            <div className="sho-title">健康度概览</div>
          </div>
          <div className="sho-toolbar-actions">
            <div className="sho-trend-range-control">
              <span className="sho-trend-range-label">趋势图历史范围</span>
              <Select
                className="sho-trend-range-select"
                options={trendRangeOptions.map((option) => ({
                  label: option.label,
                  value: option.key,
                }))}
                size="small"
                value={trendRange}
                onChange={handleTrendRangeChange}
              />
            </div>
            <Popover
              content={
                <ThresholdSettingsContent
                  hasThresholdChange={hasThresholdChange}
                  normalizedDraft={normalizedThresholdDraft}
                  thresholdDraft={thresholdDraft}
                  thresholdSaving={thresholdSaving}
                  onDraftChange={setThresholdDraft}
                  onSave={saveAbnormalScoreThreshold}
                />
              }
              onOpenChange={handleThresholdSettingsOpenChange}
              open={thresholdSettingsOpen}
              placement="bottomRight"
              trigger="click"
            >
              <Button
                aria-label="健康异常判定设置"
                className="sho-threshold-toolbar-trigger"
                icon={<Settings size={15} />}
              >
                健康异常判定
              </Button>
            </Popover>
          </div>
        </div>

        {initialOverviewLoading ? (
          <OverviewInitialLoading metricLabels={metricLabels} />
        ) : (
          <>
            <OverviewSummary
              statusDistribution={statusDistribution}
              summaryCardIcons={summaryCardIcons}
              summaryCards={summaryCards}
              summaryTrendGroupByKey={summaryTrendGroupByKey}
              trendError={trendError}
              trendGranularity={trendGranularity}
              trendLoading={trendLoading}
              trendPoints={trendPoints}
              trendRangeLabel={systemTrendRangeLabel}
              onTrendRetry={handleOverviewTrendRetry}
            />

            <OverviewSystemTable
              HealthScoreCalculationTooltip={HealthScoreCalculationTooltip}
              HealthScoreColumnTitle={HealthScoreColumnTitle}
              metricLabels={metricLabels}
              onOpenDetail={openDetail}
              onPageChange={handleSystemPageChange}
              onSystemHealthFilterChange={setSystemHealthFilter}
              onSystemKeywordChange={setSystemKeyword}
              overviewState={overviewState}
              overviewSystems={overviewSystems}
              scoreThresholds={scoreThresholds}
              systemHealthFilter={systemHealthFilter}
              systemHealthFilterOptions={systemHealthFilterOptions}
              systemKeyword={systemKeyword}
              systemPage={systemPage}
              systemPageSize={systemPageSize}
              systemPageTotal={systemPageTotal}
              systemTrendError={systemTrendError}
              systemTrendGranularity={systemTrendGranularity}
              systemTrendLoading={systemTrendLoading}
              systemTrendRangeLabel={systemTrendRangeLabel}
              systemTrendRows={systemTrendRows}
            />
            <PendingIssueQueue
              expandedIssueAnalyses={expandedIssueAnalyses}
              issueAnalyses={issueAnalyses}
              pendingIssueError={pendingIssueError}
              pendingIssueLoading={pendingIssueLoading}
              pendingIssuePage={pendingIssuePage}
              pendingIssuePageSize={pendingIssuePageSize}
              pendingIssueRows={pendingIssueRows}
              pendingIssueTotal={pendingIssueTotal}
              pendingServiceCount={pendingServiceCount}
              queryParams={queryParams}
              onAnalyzeIssue={analyzeIssue}
              onPageChange={handlePendingIssuePageChange}
              onToggleIssueAnalysis={toggleIssueAnalysis}
            />
          </>
        )}
      </>
    )
  }

  const renderDetail = () => {
    if (!selectedSystem) return null

    const shouldShowDetailBody = shouldRenderDetailBody({ selectedSystem, detailHeaderLoading })

    return (
      <>
        <DetailHeader
          HealthScoreRangeHelpIcon={HealthScoreRangeHelpIcon}
          abnormalServiceRatio={abnormalServiceRatio}
          detailHeaderLoading={detailHeaderLoading}
          onCloseDetail={closeDetail}
          redMetricTags={statusTags(selectedSystem.metrics, 'red', '暂无红灯指标')}
          scoreThresholds={scoreThresholds}
          selectedSystem={selectedSystem}
        />

        {detailHeaderLoading ? <DetailLoadingPanel /> : shouldShowDetailBody && (
          <>
            <DetailTrendSection
              detailTrendError={detailTrendError}
              detailTrendGranularity={detailTrendGranularity}
              detailTrendLoading={detailTrendLoading}
              detailTrendRows={detailTrendRows}
              scoreThresholds={scoreThresholds}
              trendRange={trendRange}
              onRangeChange={handleTrendRangeChange}
              onRetry={handleDetailTrendRetry}
            />

            <section className="sho-panel sho-system-metrics-panel">
              <div className="sho-panel-head">
                <div className="sho-panel-title">健康指标</div>
                <Legend />
              </div>
              <div className="sho-metric-grid">
                {metricLabels.map(([key, label]) => {
                  const status = selectedSystem.metrics[key] || 'gray'
                  const detail = metricDetail(selectedSystem, key)
                  return (
                    <MetricVisualTile
                      detail={detail}
                      key={key}
                      label={label}
                      metricKey={key}
                      status={status}
                    />
                  )
                })}
              </div>
            </section>

            <ServiceHealthTable
              HealthScoreColumnTitle={HealthScoreColumnTitle}
              filteredServices={filteredServices}
              metricLabels={metricLabels}
              onPageChange={handleServicePageChange}
              onServiceHealthFilterChange={setServiceHealthFilter}
              onServiceKeywordChange={setServiceKeyword}
              scoreThresholds={scoreThresholds}
              serviceEmptyMessage={serviceEmptyMessage}
              serviceHealthFilter={serviceHealthFilter}
              serviceHealthFilterOptions={serviceHealthFilterOptions}
              serviceKeyword={serviceKeyword}
              servicePage={servicePage}
              servicePageSize={servicePageSize}
              servicePageTotal={servicePageTotal}
              serviceRowsError={serviceRowsError}
              serviceRowsLoading={serviceRowsLoading}
            />
          </>
        )}
      </>
    )
  }

  return (
    <div
      className={`system-health-overview-page${selectedSystemId ? ' system-health-overview-page-detail' : ''}`}
    >
      {selectedSystemId ? renderDetail() : renderOverview()}
    </div>
  )
}
