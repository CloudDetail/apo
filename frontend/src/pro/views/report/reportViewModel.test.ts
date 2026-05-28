import test from 'node:test'
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'

import {
  createReportViewModelFromViewResponse,
  type AlertEventReportResponse,
  type AlertEventReportViewResponse,
} from './reportViewModel.ts'

const cpuReport = (): AlertEventReportResponse => ({
  reportType: 'infra',
  reportData: {
    overview: {
      timestamp: 1779258240000000,
      detail: '错误日志数升高，初步定位 CPU 方向异常。',
      reason: 'CPU 使用率与错误日志数同步升高。',
      tags: {
        service: 'checkout',
      },
    },
    rootCauseAnalysis: {
      cause: '应用错误日志异常增长，更可能为日志打印或重试逻辑导致 CPU 方向异常',
      other: [
        {
          level: 'major',
          type: 'cpu',
          reason: 'CPU 使用率和错误日志数在告警窗口同步升高。',
        },
      ],
    },
    suggest: [
      {
        action: '降低 checkout 服务异常日志输出，并临时关闭高频重试。',
        level: 'high',
        verification: 'CPU 使用率和错误日志数回落。',
      },
    ],
    analyzeProcess: '排查发现 CPU 与错误日志同步升高。',
    evidence: [
      {
        name: '服务 RED 指标',
        type: 'red',
        description: '平均延迟和错误率升高。',
        unit: 'percent',
        data: [
          {
            legend: 'error rate',
            chartData: {
              '1779256980000000': 0.6,
              '1779258240000000': 4.8,
            },
          },
        ],
      },
      {
        name: 'CPU 使用率',
        type: 'metrics',
        description: 'CPU 使用率在告警窗口持续升高。',
        unit: 'percent',
        data: [
          {
            legend: 'cpu usage',
            chart: {
              chartData: {
                '1779256980000000': 45,
                '1779258240000000': 96,
              },
            },
          },
        ],
      },
      {
        name: '错误日志数',
        type: 'metrics',
        description: '错误日志数在告警窗口同步升高。',
        unit: 'count',
        data: [
          {
            legend: 'error logs',
            chart: {
              chartData: {
                '1779256980000000': 30,
                '1779258240000000': 55377,
              },
            },
          },
        ],
      },
    ],
    reportView: {
      headline: '应用错误日志异常增长，更可能为日志打印或重试逻辑导致 CPU 异常',
      conclusionSummary: 'CPU 使用率与错误日志数在告警窗口同步升高，RED 只作为影响现象。',
      faultLocation: {
        label: '异常位置',
        target: 'checkout',
        summary: 'CPU 异常集中在 checkout 服务实例。',
      },
      summaryCards: [
        { type: 'symptom', title: '告警症状', summary: '错误日志数异常增长。' },
        { type: 'rootCause', title: '快速判断原因', summary: 'CPU 与错误日志同步升高。' },
        { type: 'excluded', title: '已排除方向', summary: '没有下游 5xx 同步升高。' },
      ],
      rootCauseMetricRefs: [
        { evidenceIndex: 1, title: 'CPU 使用率', summary: '45% -> 96%' },
        { evidenceIndex: 2, title: '错误日志数', summary: '30 -> 55377' },
        { evidenceIndex: 0, title: 'RED 指标', summary: '不应展示' },
      ],
      businessImpact: null,
      actionTitles: [{ suggestIndex: 0, title: '先降低异常日志输出' }],
      evidenceHighlights: [
        { evidenceIndex: 1, title: 'CPU 使用率持续升高', summary: '峰值接近 96%。' },
      ],
    },
  },
})

const downstreamReportViewResponse = (): AlertEventReportViewResponse => ({
  snapshot: {
    reportType: 'slow',
    reportData: {
      overview: {
        timestamp: 1779797547335000,
        detail: 'ts-gateway-service 请求延时升高，初步指向下游依赖慢调用。',
        reason: '当前实例自身 RED 与资源证据正常，下游接口耗时明显升高。',
        tags: {
          service: 'ts-gateway-service',
        },
      },
      rootCauseAnalysis: {
        cause: 'ts-travel-plan-service 下游依赖问题',
        other: [
          {
            level: 'major',
            reason: 'ts-travel-plan-service 下游响应变慢，导致请求延时升高。',
            type: 'downstream',
          },
          {
            level: 'normal',
            reason: '当前实例 RED 错误率正常，不作为根因支撑指标。',
            type: 'red',
          },
        ],
      },
      suggest: [
        {
          action: '对比全量实例并定位下游慢接口。',
          level: 'high',
        },
      ],
      analyzeProcess: '当前实例自身指标正常，下游慢接口耗时升高。',
      evidence: [
        {
          data: [],
          description: '应用日志未见异常。',
          name: '应用日志',
          type: 'log',
        },
        {
          data: [
            {
              chartData: {
                '1779796620000000': 0,
                '1779797547000000': 0,
              },
              name: 'error rate',
              unit: 'percent',
            },
          ],
          description: '错误率始终为 0。',
          name: '服务 RED 指标',
          type: 'red',
          unit: 'percent',
        },
        {
          data: [],
          description: '未检测到 RTT 异常数据。',
          name: '下游 RTT',
          type: 'metrics',
          unit: 'ms',
        },
        {
          data: [
            {
              chartData: {
                '1779796620000000': 2.19,
                '1779797547000000': 4.39,
              },
              name: 'POST /api/v1/travelplanservice/travelPlan/cheapest',
              unit: 's',
            },
          ],
          description: '下游慢接口耗时从 2.19s 升至 4.39s。',
          name: '下游慢接口耗时',
          type: 'metrics',
          unit: 's',
        },
      ],
      reportView: {
        headline: '请求延时升高，更可能由下游慢接口传导。',
        conclusionSummary: '当前实例自身证据正常，下游调用耗时明显升高。',
        faultLocation: {
          label: '故障位置',
          target: '服务',
          summary: '更可能位于下游依赖服务。',
        },
        rootCauseEvidenceRefs: [
          {
            evidenceIndex: 3,
            title: '下游慢接口耗时',
            summary: '2.19s -> 4.39s',
          },
        ],
        rootCauseMetricRefs: [
          {
            evidenceIndex: 3,
            title: '下游慢接口耗时',
            summary: '2.19s -> 4.39s',
          },
        ],
        businessImpact: null,
        timeline: [],
        analysisSteps: [],
        actionTitles: [{ suggestIndex: 0, title: '对比全量实例' }],
        evidenceHighlights: [
          {
            evidenceIndex: 3,
            title: '下游慢接口耗时',
            summary: '慢点集中在 ts-travel-plan-service 的 cheapest 接口。',
          },
        ],
      },
    },
  },
})

test('prefers additive reportView headline, summary, fault location, and action title', () => {
  const model = createReportViewModelFromViewResponse({ snapshot: cpuReport() })

  assert.equal(model.headline, '应用错误日志异常增长，更可能为日志打印或重试逻辑导致 CPU 异常')
  assert.equal(
    model.analysisSummary,
    'CPU 使用率与错误日志数在告警窗口同步升高，RED 只作为影响现象。',
  )
  assert.equal(model.faultLocation.label, '异常位置')
  assert.equal(model.faultLocation.service, 'checkout')
  assert.equal(model.actions[0].displayTitle, '先降低异常日志输出')
})

test('sorts actions by priority while preserving original title mapping', () => {
  const report = cpuReport()
  report.reportData.suggest = [
    { action: 'medium action', level: 'medium' },
    { action: 'high action 1', level: 'high' },
    { action: 'low action', level: 'low' },
    { action: 'unknown action', level: 'normal' },
    { action: 'high action 2', level: 'high' },
  ]
  report.reportData.reportView = {
    ...report.reportData.reportView,
    actionTitles: [
      { suggestIndex: 0, title: '中优先级动作' },
      { suggestIndex: 1, title: '高优先级动作 1' },
      { suggestIndex: 2, title: '低优先级动作' },
      { suggestIndex: 3, title: '未知优先级动作' },
      { suggestIndex: 4, title: '高优先级动作 2' },
    ],
  }

  const model = createReportViewModelFromViewResponse({ snapshot: report })

  assert.deepEqual(
    model.actions.map((action) => action.displayTitle),
    ['高优先级动作 1', '高优先级动作 2', '中优先级动作', '低优先级动作', '未知优先级动作'],
  )
})

test('maps reportView timeline timestamp and summary into display fields', () => {
  const report = cpuReport()
  const timestamp = 1779811334000000
  report.reportData.reportView = {
    ...report.reportData.reportView,
    timeline: [
      {
        timestamp,
        title: '首次集中超时',
        summary:
          '日志出现 dispatcherServlet 请求处理失败，根因指向调用 http://localhost:58081/api/jpa-demo/get 读超时。',
      },
    ],
  }

  const model = createReportViewModelFromViewResponse({ snapshot: report })

  assert.deepEqual(model.timeline, [
    {
      time: new Date(timestamp / 1000).toLocaleTimeString('zh-CN', {
        hour12: false,
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
      }),
      title: '首次集中超时',
      detail:
        '日志出现 dispatcherServlet 请求处理失败，根因指向调用 http://localhost:58081/api/jpa-demo/get 读超时。',
    },
  ])
})

test('formats reportView timeline timestamp with resolved local timezone', () => {
  const originalDateTimeFormat = Intl.DateTimeFormat
  const originalToLocaleTimeString = Date.prototype.toLocaleTimeString

  try {
    ;(Intl as any).DateTimeFormat = function DateTimeFormat(
      locales?: Intl.LocalesArgument,
      options?: Intl.DateTimeFormatOptions,
    ) {
      const formatter = new originalDateTimeFormat(locales, options)
      return {
        format: formatter.format.bind(formatter),
        resolvedOptions: () => ({ ...formatter.resolvedOptions(), timeZone: 'Asia/Shanghai' }),
      }
    }
    Date.prototype.toLocaleTimeString = function toLocaleTimeString(
      _locales?: Intl.LocalesArgument,
      options?: Intl.DateTimeFormatOptions,
    ) {
      return options?.timeZone || 'missing-time-zone'
    }

    const report = cpuReport()
    report.reportData.reportView = {
      ...report.reportData.reportView,
      timeline: [
        {
          timestamp: 1779811334000000,
          title: '首次集中超时',
          summary: '本地时间应由前端按浏览器时区转换。',
        },
      ],
    }

    const model = createReportViewModelFromViewResponse({ snapshot: report })

    assert.equal(model.timeline[0].time, 'Asia/Shanghai')
  } finally {
    ;(Intl as any).DateTimeFormat = originalDateTimeFormat
    Date.prototype.toLocaleTimeString = originalToLocaleTimeString
  }
})

test('does not use reportView analysisSteps when timeline is empty', () => {
  const report = cpuReport()
  report.reportData.reportView = {
    ...report.reportData.reportView,
    timeline: [],
    analysisSteps: [
      {
        title: '先排除本地瓶颈',
        summary: '无真实时间点的排查步骤不应展示在事件时间线里。',
      },
    ],
  }

  const model = createReportViewModelFromViewResponse({ snapshot: report })

  assert.deepEqual(model.timeline, [])
})

test('only shows root-cause evidence metrics and rejects RED for cpu root cause', () => {
  const model = createReportViewModelFromViewResponse({ snapshot: cpuReport() })

  assert.deepEqual(
    model.coreMetrics.map((metric) => metric.name),
    ['CPU 使用率', '错误日志数'],
  )
  assert.deepEqual(
    model.coreMetrics.map((metric) => metric.valueSummary),
    ['45% -> 96%', '30 -> 55377'],
  )
})

test('allows RED as support metric for downstream root cause', () => {
  const report = cpuReport()
  report.reportType = 'error'
  report.reportData.rootCauseAnalysis.other[0].type = 'downstream'
  report.reportData.reportView = {
    ...report.reportData.reportView,
    rootCauseMetricRefs: [{ evidenceIndex: 0, title: '下游错误率', summary: '0.6% -> 4.8%' }],
  }

  const model = createReportViewModelFromViewResponse({ snapshot: report })

  assert.deepEqual(
    model.coreMetrics.map((metric) => metric.name),
    ['下游错误率'],
  )
})

test('keeps support metrics empty when no evidence matches the root cause direction', () => {
  const report = cpuReport()
  report.reportData.reportView = {
    ...report.reportData.reportView,
    rootCauseMetricRefs: [{ evidenceIndex: 0, title: 'RED 指标', summary: '不应展示' }],
  }
  report.reportData.evidence = [report.reportData.evidence[0]]

  const model = createReportViewModelFromViewResponse({ snapshot: report })

  assert.equal(model.coreMetrics.length, 0)
})

test('respects explicit empty root-cause metric refs from reportView', () => {
  const report = cpuReport()
  report.reportType = 'error'
  report.reportData.rootCauseAnalysis.other[0].type = 'downstream'
  report.reportData.reportView = {
    ...report.reportData.reportView,
    rootCauseMetricRefs: [],
  }

  const model = createReportViewModelFromViewResponse({ snapshot: report })

  assert.equal(model.coreMetrics.length, 0)
})

test('does not use highlighted normal RED evidence as support metric when root-cause refs are empty', () => {
  const report = cpuReport()
  report.reportType = 'slow'
  report.reportData.rootCauseAnalysis.other[0].type = 'downstream'
  report.reportData.reportView = {
    ...report.reportData.reportView,
    rootCauseMetricRefs: [],
    evidenceHighlights: [
      { evidenceIndex: 0, title: '网关 RED 指标表现', summary: '延时升高但错误率为 0。' },
    ],
  }

  const model = createReportViewModelFromViewResponse({ snapshot: report })

  assert.equal(model.coreMetrics.length, 0)
})

test('uses referenced log evidence as root-cause support when no metric chart exists', () => {
  const report = cpuReport()
  report.reportType = 'slow'
  report.reportData.rootCauseAnalysis.other[0] = {
    level: 'major',
    type: 'downstream',
    reason: 'requesttemplate-demo 调用本地依赖 localhost:58082 时持续读超时。',
  }
  report.reportData.evidence = [
    {
      name: '服务 RED 指标',
      type: 'red',
      description: '错误率为 0，延时已回落，整体运行稳定。',
      unit: 'percent',
      data: [
        {
          legend: 'error rate',
          chartData: {
            '1779256980000000': 0,
            '1779258240000000': 0,
          },
        },
      ],
    },
    {
      name: '应用日志',
      type: 'log',
      description:
        'compile-worker-25 节点连续出现 requesttemplate-demo 调用 http://localhost:58082/api/jpa-demo/sleep 的 ResourceAccessException 和 SocketTimeoutException: Read timed out，线程为 http-nio-58081-exec-6/7。',
      data: [],
    },
  ]
  report.reportData.reportView = {
    ...report.reportData.reportView,
    rootCauseMetricRefs: [],
    rootCauseEvidenceRefs: [
      {
        evidenceIndex: 1,
        title: 'localhost:58082 读超时',
        summary:
          'requesttemplate-demo 调用 /api/jpa-demo/sleep 持续出现 ResourceAccessException 和 SocketTimeoutException: Read timed out，集中在 compile-worker-25 的 http-nio-58081-exec-6/7 线程。',
      },
    ],
    evidenceHighlights: [
      { evidenceIndex: 0, title: 'RED 表现正常', summary: '错误率为 0，整体运行稳定。' },
    ],
  }

  const model = createReportViewModelFromViewResponse({ snapshot: report })

  assert.deepEqual(
    model.rootCauseEvidence.map((item) => item.name),
    ['localhost:58082 读超时'],
  )
  assert.equal(model.rootCauseEvidence[0].typeLabel, '日志')
  assert.equal(model.rootCauseEvidence[0].chartData, undefined)
  assert.equal(
    model.rootCauseEvidence[0].description,
    'requesttemplate-demo 调用 /api/jpa-demo/sleep 持续出现 ResourceAccessException 和 SocketTimeoutException: Read timed out，集中在 compile-worker-25 的 http-nio-58081-exec-6/7 线程。',
  )
  assert.equal(model.coreMetrics.length, 0)
})

test('uses referenced mysql database evidence as downstream root-cause support', () => {
  const report = cpuReport()
  report.reportType = 'slow'
  report.reportData.rootCauseAnalysis.other[0] = {
    level: 'major',
    type: 'downstream',
    reason: 'Mysql 下游依赖性能异常，导致 ts-config-service 数据库调用显著变慢。',
  }
  report.reportData.evidence = [
    {
      name: 'Mysql 数据库等待时间',
      type: 'metrics',
      description:
        'train-ticket-mysql-master-0 SQL 查询变慢，锁等待和连接池等待升高，ts-config-service 受数据库资源瓶颈影响。',
      unit: 'ms',
      data: [
        {
          legend: 'mysql wait',
          chart: {
            chartData: {
              '1779256980000000': 120,
              '1779258240000000': 1840,
            },
          },
        },
      ],
    },
  ]
  report.reportData.reportView = {
    ...report.reportData.reportView,
    rootCauseEvidenceRefs: [
      {
        evidenceIndex: 0,
        title: 'Mysql 数据库等待时间',
        summary: '120 ms -> 1840 ms',
      },
    ],
  }

  const model = createReportViewModelFromViewResponse({ snapshot: report })

  assert.deepEqual(
    model.rootCauseEvidence.map((item) => item.name),
    ['Mysql 数据库等待时间'],
  )
  assert.deepEqual(
    model.coreMetrics.map((metric) => metric.name),
    ['Mysql 数据库等待时间'],
  )
})

test('keeps prose root-cause evidence summaries out of the metric value slot', () => {
  const report = cpuReport()
  const proseSummary =
    '同一节点上的 ts-gateway-service、ts-basic-service、ts-route-plan-service 与 ts-seat-service 均有 RTT 指标，异常集中在 large-server-6 上的 ts-seat-service 两个实例。'
  report.reportType = 'slow'
  report.reportData.rootCauseAnalysis.other[0] = {
    level: 'major',
    type: 'downstream',
    reason: '下游 ts-seat-service 实例 RTT 持续偏高。',
  }
  report.reportData.evidence = [
    {
      name: '下游 RTT',
      type: 'metrics',
      description: '下游 RTT 出现持续偏高。',
      unit: 's',
      data: [
        {
          legend: 'ts-seat-service RTT',
          chart: {
            chartData: {
              '1779256980000000': 0.12,
              '1779258240000000': 0.2,
            },
          },
        },
      ],
    },
  ]
  report.reportData.reportView = {
    ...report.reportData.reportView,
    rootCauseEvidenceRefs: [
      {
        evidenceIndex: 0,
        title: '异常集中在 large-server-6 节点下游',
        summary: proseSummary,
      },
    ],
  }

  const model = createReportViewModelFromViewResponse({ snapshot: report })

  assert.equal(model.rootCauseEvidence[0].name, '异常集中在 large-server-6 节点下游')
  assert.equal(model.rootCauseEvidence[0].valueSummary, '')
  assert.equal(model.rootCauseEvidence[0].description, proseSummary)
})

test('focuses referenced root-cause evidence series instead of the first chart series', () => {
  const report = cpuReport()
  report.reportType = 'slow'
  report.reportData.rootCauseAnalysis.other[0] = {
    level: 'major',
    type: 'downstream',
    reason: '下游 ts-seat-service 实例 RTT 持续偏高。',
  }
  report.reportData.evidence = [
    {
      name: '应用网络质量指标',
      type: 'metrics',
      description: '下游 RTT 出现持续偏高。',
      unit: 's',
      data: [
        {
          legend:
            'ts-travel-plan-service-7cf74b9649-vbqns -> ts-gateway-service-687c769b7b-cwbgq (10.244.0.11)',
          chart: {
            chartData: {
              '1779256980000000': 0.0011,
              '1779258240000000': 0.0012,
            },
          },
        },
        {
          legend:
            'ts-travel-plan-service-7cf74b9649-vbqns -> ts-seat-service-5b88897f7b-6bg5d (10.244.236.77)',
          chart: {
            chartData: {
              '1779256980000000': 0.202,
              '1779258240000000': 0.209,
            },
          },
        },
      ],
    },
  ]
  report.reportData.reportView = {
    ...report.reportData.reportView,
    rootCauseEvidenceRefs: [
      {
        evidenceIndex: 0,
        title: '实例 10.244.236.77 RTT 持续偏高',
        summary:
          'ts-travel-plan-service -> ts-seat-service-5b88897f7b-6bg5d(10.244.236.77) RTT 稳定在约 0.20s。',
      },
    ],
  }

  const model = createReportViewModelFromViewResponse({ snapshot: report })

  assert.deepEqual(
    model.rootCauseEvidence[0].chartData?.map((series) => [series.legend, series.focus]),
    [
      [
        'ts-travel-plan-service-7cf74b9649-vbqns -> ts-gateway-service-687c769b7b-cwbgq (10.244.0.11)',
        false,
      ],
      [
        'ts-travel-plan-service-7cf74b9649-vbqns -> ts-seat-service-5b88897f7b-6bg5d (10.244.236.77)',
        true,
      ],
    ],
  )
})

test('rejects referenced normal evidence from root-cause support', () => {
  const report = cpuReport()
  report.reportType = 'slow'
  report.reportData.rootCauseAnalysis.other[0] = {
    level: 'major',
    type: 'downstream',
    reason: '下游依赖响应慢。',
  }
  report.reportData.evidence = [
    {
      name: '下游 RED 指标',
      type: 'red',
      description: '下游接口错误率始终为 0，延时已回落，整体运行正常。',
      unit: 'percent',
      data: [
        {
          legend: 'latency',
          chartData: {
            '1779256980000000': 120,
            '1779258240000000': 118,
          },
        },
      ],
    },
    {
      name: '应用日志',
      type: 'log',
      description: '调用 localhost:58082 正常完成，未见异常，未见报错。',
      data: [],
    },
  ]
  report.reportData.reportView = {
    ...report.reportData.reportView,
    rootCauseEvidenceRefs: [
      { evidenceIndex: 0, title: '下游 RED 表现正常', summary: '错误率始终为 0，整体运行正常。' },
      { evidenceIndex: 1, title: '调用日志正常', summary: '未见异常，未见报错。' },
    ],
    evidenceHighlights: [
      { evidenceIndex: 0, title: '下游 RED 表现正常', summary: '错误率始终为 0，整体运行正常。' },
    ],
  }

  const model = createReportViewModelFromViewResponse({ snapshot: report })

  assert.equal(model.rootCauseEvidence.length, 0)
  assert.equal(model.coreMetrics.length, 0)
})

test('hides business impact when reportView marks it meaningless', () => {
  const model = createReportViewModelFromViewResponse({ snapshot: cpuReport() })

  assert.equal(model.businessImpact.summary, '')
  assert.deepEqual(model.impactGroups, [])
})

test('uses evidence description as raw evidence conclusion', () => {
  const model = createReportViewModelFromViewResponse({ snapshot: cpuReport() })

  assert.equal(model.evidence[1].description, 'CPU 使用率在告警窗口持续升高。')
  assert.equal(model.evidenceSummary[0].name, 'CPU 使用率持续升高')
})

test('preserves raw evidence index for highlighted evidence summaries', () => {
  const report = cpuReport()
  report.reportData.reportView = {
    ...report.reportData.reportView,
    evidenceHighlights: [
      { evidenceIndex: 2, title: '错误日志数同步升高', summary: '错误日志数峰值达到 55377。' },
    ],
  }

  const model = createReportViewModelFromViewResponse({ snapshot: report })

  assert.equal(model.evidenceSummary[0].name, '错误日志数同步升高')
  assert.equal(model.evidenceSummary[0].evidenceIndex, 2)
})

test('report UI jumps evidence summaries by raw evidence index', () => {
  const source = readFileSync('src/pro/views/report/index.tsx', 'utf8')
  const styles = readFileSync('src/pro/views/report/index.module.scss', 'utf8')

  assert.equal(source.includes('handleEvidenceSummaryClick(item.evidenceIndex)'), true)
  assert.match(
    source,
    /const jumpTargetId\s*=\s*canJump\s*\?\s*getEvidenceAnchorId\(item\.evidenceIndex\)\s*:\s*undefined/,
  )
  assert.equal(source.includes('aria-controls={jumpTargetId}'), true)
  assert.equal(source.includes('detailEvidence[index]'), false)
  assert.equal(source.includes('styles.evidenceSummaryItemInteractive'), true)
  assert.equal(source.includes("role={canJump ? 'button' : undefined}"), true)
  assert.equal(source.includes('tabIndex={canJump ? 0 : undefined}'), true)
  assert.equal(source.includes("event.key !== 'Enter' && event.key !== ' '"), true)
  assert.equal(
    source.includes('<span className={styles.evidenceJumpButton}>查看原始证据</span>'),
    true,
  )
  assert.equal(styles.includes('.evidenceSummaryItemInteractive'), true)
})

test('raw evidence section is collapsed by default and expands from header or summary jump', () => {
  const reportSource = readFileSync('src/pro/views/report/index.tsx', 'utf8')
  const evidenceSource = readFileSync(
    'src/pro/views/report/components/SupportingEvidence/index.tsx',
    'utf8',
  )

  assert.equal(
    reportSource.includes('const [isRawEvidenceOpen, setIsRawEvidenceOpen] = useState(false)'),
    true,
  )
  assert.equal(reportSource.includes('setIsRawEvidenceOpen(true)'), true)
  assert.equal(reportSource.includes('aria-expanded={isRawEvidenceOpen}'), true)
  assert.equal(reportSource.includes('setIsRawEvidenceOpen((current) => !current)'), true)
  assert.equal(reportSource.includes('{isRawEvidenceOpen && ('), true)
  assert.equal(reportSource.includes('expandedEvidenceIndexes'), false)
  assert.equal(reportSource.includes('onToggleEvidence'), false)

  assert.equal(evidenceSource.includes('expandedEvidenceIndexes'), false)
  assert.equal(evidenceSource.includes('onToggleEvidence'), false)
  assert.equal(evidenceSource.includes('aria-expanded'), false)
})

test('sparkline tooltip escapes dynamic HTML content', () => {
  const source = readFileSync('src/pro/views/report/index.tsx', 'utf8')

  assert.equal(source.includes('function escapeTooltipHtml'), true)
  assert.equal(source.includes('escapeTooltipHtml(item.seriesName)'), true)
  assert.equal(
    source.includes('escapeTooltipHtml(formatSparklineValue(item.data[1], matchedSeries?.unit))'),
    true,
  )
})

test('raw evidence renders trace samples explicitly instead of falling back to topology', () => {
  const source = readFileSync('src/pro/views/report/components/SupportingEvidence/index.tsx', 'utf8')

  assert.equal(source.includes("type: 'log' | 'red' | 'metrics' | 'endpoints' | 'trace' | 'traces'"), true)
  assert.equal(source.includes('TraceEvidence'), true)
  assert.equal(source.includes("type === 'trace' || type === 'traces'"), true)
  assert.equal(source.includes("PropTypes.oneOf(['log', 'red', 'metrics', 'endpoints', 'trace', 'traces'])"), true)
  assert.equal(source.includes(': <AffectedEndpoints topology={data} />'), false)
})

test('raw evidence metric renderer accepts both nested and direct chartData', () => {
  const source = readFileSync('src/pro/views/report/components/SupportingEvidence/index.tsx', 'utf8')

  assert.equal(source.includes('Array.isArray(data) ? data : []'), true)
  assert.equal(source.includes('item?.chart?.chartData || item?.chartData || {}'), true)
  assert.equal(source.includes("item?.legend || item?.name || item?.type || 'series'"), true)
})

test('pro routes do not expose report mock page', () => {
  const source = readFileSync('src/pro/routes.js', 'utf8')

  assert.equal(source.includes('/report/mock'), false)
  assert.equal(source.includes('ReportMockPage'), false)
  assert.equal(source.includes('src/pro/views/report/mock'), false)
})

test('report UI hides the support evidence panel when no root-cause evidence exists', () => {
  const source = readFileSync('src/pro/views/report/index.tsx', 'utf8')

  assert.equal(source.includes('报告 ID'), false)
  assert.equal(source.includes('生成时间'), false)
  assert.equal(source.includes('分析窗口'), false)
  assert.equal(source.includes('支撑证据'), true)
  assert.equal(source.includes('const hasSnapshotAside = hasRootCauseEvidence'), true)
  assert.equal(source.includes('const hasSnapshotAside = true'), false)
  assert.equal(source.includes('未返回根因支撑证据'), false)
})

test('report UI renders root-cause cause as a compact tag beside quick judgement', () => {
  const source = readFileSync('src/pro/views/report/index.tsx', 'utf8')

  assert.equal(source.includes('styles.rootCauseCauseTag'), true)
  assert.equal(source.includes("card.tone === 'rootCause' && view.rootCause"), true)
  assert.equal(source.includes('styles.summaryCardHeader'), false)
})

test('report UI keeps event timeline panel disabled', () => {
  const source = readFileSync('src/pro/views/report/index.tsx', 'utf8')

  assert.equal(source.includes('const hasTimeline = false'), true)
  assert.equal(source.includes('const hasTimeline = view.timeline.length > 0'), false)
})

test('report UI renders analyzeProcess as a default-collapsed markdown panel', () => {
  const source = readFileSync('src/pro/views/report/index.tsx', 'utf8')
  const styles = readFileSync('src/pro/views/report/index.module.scss', 'utf8')

  assert.equal(source.includes('分析过程'), true)
  assert.equal(source.includes('hasAnalyzeProcess'), true)
  assert.equal(
    source.includes('const [isAnalyzeProcessOpen, setIsAnalyzeProcessOpen] = useState(false)'),
    true,
  )
  assert.equal(source.includes('aria-expanded={isAnalyzeProcessOpen}'), true)
  assert.equal(source.includes('setIsAnalyzeProcessOpen((current) => !current)'), true)
  assert.equal(source.includes('{isAnalyzeProcessOpen && ('), true)
  assert.equal(
    source.includes('<MarkdownRender content={view.analyzeProcess} theme={currentTheme} />'),
    true,
  )
  assert.equal(styles.includes('.analysisProcessPanel'), true)
  assert.equal(styles.includes('padding: 0;'), true)
  assert.equal(styles.includes('padding: 14px 16px;'), true)
  assert.equal(styles.includes('min-height: 52px;'), true)
  assert.equal(styles.includes('font-size: 14px;'), true)
  assert.equal(styles.includes('font-weight: 840;'), true)
  assert.equal(styles.includes('box-shadow: 0 10px 22px'), true)
  assert.equal(styles.includes('&:hover,'), true)
  assert.equal(styles.includes("&[aria-expanded='true']"), true)
})

test('report UI follows app dark theme for page tokens, markdown, and sparklines', () => {
  const source = readFileSync('src/pro/views/report/index.tsx', 'utf8')
  const styles = readFileSync('src/pro/views/report/index.module.scss', 'utf8')

  assert.equal(source.includes('const currentTheme'), true)
  assert.equal(
    source.includes('<MarkdownRender content={view.analyzeProcess} theme={currentTheme} />'),
    true,
  )
  assert.equal(source.includes('theme="light"'), false)
  assert.equal(source.includes('const sparklineTheme'), true)
  assert.equal(source.includes("backgroundColor: 'transparent'"), true)
  assert.equal(source.includes('theme={sparklineTheme}'), false)
  assert.match(styles, /:global\(\[data-coreui-theme='dark'\]\) \.reportPage/)
  assert.match(styles, /--page:\s*oklch\(16/)
  assert.match(styles, /--panel:\s*oklch\(20/)
})

test('report dark mode keeps summary cards and support sparkline visually quiet', () => {
  const styles = readFileSync('src/pro/views/report/index.module.scss', 'utf8')

  assert.match(styles, /:global\(\[data-coreui-theme='dark'\]\) \.summarySymptom,/)
  assert.match(styles, /:global\(\[data-coreui-theme='dark'\]\) \.summaryRootCause,/)
  assert.match(styles, /:global\(\[data-coreui-theme='dark'\]\) \.summaryExcluded,/)
  assert.match(styles, /:global\(\[data-coreui-theme='dark'\]\) \.summaryLocation/)
  assert.match(styles, /border-left-width:\s*1px;/)
  assert.match(styles, /border-left-style:\s*solid;/)
  assert.match(styles, /border-left-color:\s*color-mix\(in oklch, var\(--warning\) 76%, var\(--line\)\);/)
  assert.match(styles, /border-left-color:\s*color-mix\(in oklch, var\(--danger\) 74%, var\(--line\)\);/)
  assert.equal(styles.includes('border-left-color: color-mix(in oklch, var(--success)'), false)
  assert.match(
    styles,
    /background:\s*color-mix\(in oklch, var\(--panel-soft\) 22%, var\(--panel\)\);/,
  )
  assert.match(styles, /\.metricSparkline\s*\{[^}]*background:\s*transparent;/s)
})

test('report UI places actions in the full-width flow instead of a right rail', () => {
  const source = readFileSync('src/pro/views/report/index.tsx', 'utf8')
  const styles = readFileSync('src/pro/views/report/index.module.scss', 'utf8')

  assert.equal(source.includes('styles.actionPanel'), true)
  assert.equal(source.includes('styles.reportSideColumn'), false)
  assert.equal(styles.includes('.reportSideColumn'), false)
  assert.equal(
    styles.includes('grid-template-columns: minmax(0, 1fr) minmax(340px, 420px);'),
    false,
  )
})

test('report UI collapses long action lists to three visible rows by default', () => {
  const source = readFileSync('src/pro/views/report/index.tsx', 'utf8')
  const styles = readFileSync('src/pro/views/report/index.module.scss', 'utf8')

  assert.equal(source.includes('DEFAULT_VISIBLE_ACTION_COUNT = 3'), true)
  assert.equal(source.includes('visibleActions'), true)
  assert.equal(source.includes('hiddenActionCount'), true)
  assert.equal(source.includes('展开剩余'), true)
  assert.equal(styles.includes('.actionListCollapsed'), true)
  assert.match(
    styles,
    /\.actionList \{\s+display: grid;\s+grid-template-columns: minmax\(0, 1fr\);/,
  )
})

test('report UI toggles action details from the whole action row', () => {
  const source = readFileSync('src/pro/views/report/index.tsx', 'utf8')
  const styles = readFileSync('src/pro/views/report/index.module.scss', 'utf8')

  assert.equal(source.includes('className={styles.actionRowToggle}'), true)
  assert.equal(source.includes('className={styles.actionDetailToggle}'), false)
  assert.equal(source.includes('aria-expanded={isActionOpen}'), true)
  assert.equal(source.includes('setOpenActionIndex(isActionOpen ? null : index)'), true)
  assert.equal(styles.includes('.actionRowToggle'), true)
  assert.equal(styles.includes('.actionDetailToggle'), false)
  assert.match(styles, /\.actionRowToggle,\s*\.actionRowStatic \{\s+display: grid;/)
  assert.match(styles, /width:\s*100%;/)
  assert.match(styles, /text-align:\s*left;/)
})

test('evidence summary cards use content-sized type labels', () => {
  const styles = readFileSync('src/pro/views/report/index.module.scss', 'utf8')
  const articleBlock =
    styles.match(/\.evidenceSummaryList article\s*\{(?<body>[^}]*)\}/s)?.groups?.body || ''

  assert.match(articleBlock, /grid-template-columns:\s*auto minmax\(0,\s*1fr\);/)
  assert.equal(articleBlock.includes('grid-template-columns: 72px minmax(0, 1fr);'), false)
})

test('root-cause evidence without chart does not reserve chart column', () => {
  const source = readFileSync('src/pro/views/report/index.tsx', 'utf8')
  const styles = readFileSync('src/pro/views/report/index.module.scss', 'utf8')

  assert.equal(source.includes('styles.metricSummaryCardNoChart'), true)
  assert.equal(styles.includes('.metricSummaryCardNoChart'), true)
  assert.equal(styles.includes("grid-template-areas: 'header' 'desc';"), true)
})

test('root-cause evidence header avoids one-character wrapping in narrow support panel', () => {
  const source = readFileSync('src/pro/views/report/index.tsx', 'utf8')
  const styles = readFileSync('src/pro/views/report/index.module.scss', 'utf8')
  const headerBlock =
    styles.match(/\.metricSummaryHeader\s*\{(?<body>[^}]*)\}/s)?.groups?.body || ''
  const titleBlock = styles.match(/\.metricSummaryTitle\s*\{(?<body>[^}]*)\}/s)?.groups?.body || ''
  const valueBlock = styles.match(/>\s*strong\s*\{(?<body>[^}]*)\}/s)?.groups?.body || ''

  assert.equal(source.includes('className={styles.metricSummaryTitle}'), true)
  assert.match(headerBlock, /flex-wrap:\s*wrap;/)
  assert.match(titleBlock, /flex:\s*1 1 auto;/)
  assert.match(titleBlock, /min-width:\s*0;/)
  assert.match(titleBlock, /overflow-wrap:\s*break-word;/)
  assert.match(titleBlock, /word-break:\s*normal;/)
  assert.equal(titleBlock.includes('overflow-wrap: anywhere'), false)
  assert.match(valueBlock, /flex:\s*0 1 auto;/)
  assert.match(valueBlock, /max-width:\s*100%;/)
  assert.match(valueBlock, /overflow-wrap:\s*break-word;/)
  assert.match(valueBlock, /white-space:\s*normal;/)
})

test('root-cause evidence sparkline omits the inline legend while preserving tooltip colors', () => {
  const source = readFileSync('src/pro/views/report/index.tsx', 'utf8')
  const styles = readFileSync('src/pro/views/report/index.module.scss', 'utf8')

  assert.equal(source.includes('appendToBody: true'), true)
  assert.equal(source.includes('confine: false'), true)
  assert.match(source, /position:\s*\([^)]*size/)
  assert.match(source, /size\.contentSize/)
  assert.match(source, /size\.viewSize/)
  assert.match(source, /const color = getMetricSparklineColor\(index\)/)
  assert.match(
    source,
    /color:\s*visibleSeries\.map\(\(_, index\) => getMetricSparklineColor\(index\)\)/,
  )
  assert.match(source, /lineStyle:\s*\{\s*color,/s)
  assert.equal(source.includes('styles.metricSparklineLegend'), false)
  assert.equal(source.includes('styles.metricSparklineLabel'), false)
  assert.equal(source.includes('displayedLegendSeries'), false)
  assert.equal(styles.includes('.metricSparklineLegend'), false)
  assert.equal(styles.includes('.metricSparklineLabel'), false)
})

test('downstream report uses root-cause downstream metrics instead of normal RED evidence', () => {
  const response = downstreamReportViewResponse()
  const model = createReportViewModelFromViewResponse(response)
  const reportView = response.snapshot.reportData.reportView

  assert.equal(reportView?.rootCauseEvidenceRefs?.[0].evidenceIndex, 3)
  assert.equal(model.reportType, 'slow')
  assert.equal(model.status.state, 'snapshot')
  assert.equal(model.faultLocation.service, '服务')
  assert.equal(model.actions[0].displayTitle, '对比全量实例')
  assert.equal(model.evidenceSummary[0].name, '下游慢接口耗时')
  assert.deepEqual(
    model.rootCauseEvidence.map((item) => item.name),
    ['下游慢接口耗时'],
  )
  assert.deepEqual(
    model.coreMetrics.map((metric) => metric.name),
    ['下游慢接口耗时'],
  )
  assert.equal(model.impactGroups.length, 0)
  assert.equal(
    model.evidence.some((item) => item.type === 'red'),
    true,
  )
  assert.equal(
    model.coreMetrics.some((metric) => metric.name.includes('RED')),
    false,
  )
})
