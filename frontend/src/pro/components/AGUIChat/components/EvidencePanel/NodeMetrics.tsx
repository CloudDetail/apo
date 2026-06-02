import React, { useEffect, useState, useRef } from 'react'
import { theme } from 'antd'
import { ExclamationCircleFilled, CheckCircleFilled, InfoCircleOutlined } from '@ant-design/icons'
import ReactECharts from 'echarts-for-react'
import { timeUtils } from 'src/core/utils/time'
import Empty from 'src/core/components/Empty/Empty'
import { ChartColorList } from 'src/constants'
import { useSelector } from 'react-redux'

const { useToken } = theme

interface MetricItem {
  title: string
  unit: string
  timeseries: Array<{
    labels: Record<string, any>
    chart: {
      chartData: Record<string, number>
    }
  }>
  detection?: string
}

interface NodeMetricsProps {
  metrics: MetricItem[]
  type: 'abnormal' | 'normal'
}

interface InlineChartProps {
  chartData: any
  YFormatter: (value: number) => string
  showLegend?: boolean
}

const SAFE_GAP = 8 // 视口边缘安全间距

const InlineChart: React.FC<InlineChartProps> = ({ chartData, YFormatter, showLegend = true }) => {
  const chartRef = useRef<any>(null)
  const { theme: currentTheme } = useSelector((state: any) => state.settingReducer)
  const tooltipPosition = (
    point: number[],
    _params: any,
    dom: HTMLElement,
    _rect: any,
    size: { contentSize?: number[] } = {},
  ) => {
    // 容器 -> 视口坐标
    const host = chartRef.current?.getEchartsInstance?.()?.getDom?.()
    const hostRect = host ? host.getBoundingClientRect() : { left: 0, top: 0 }

    const cw = (size.contentSize && size.contentSize[0]) || dom?.offsetWidth || 0
    const ch = (size.contentSize && size.contentSize[1]) || dom?.offsetHeight || 0

    const px = hostRect.left + point[0] // 指示点(视口X)
    const py = hostRect.top + point[1] // 指示点(视口Y)

    // —— 水平：优先向左，放不下再向右 ——
    const spaceLeft = px - SAFE_GAP
    const spaceRight = window.innerWidth - px - SAFE_GAP
    let x = spaceLeft >= cw ? px - cw - POINTER_GAP : px + POINTER_GAP

    // —— 垂直：优先向下，放不下再向上 ——
    const spaceTop = py - SAFE_GAP
    const spaceBottom = window.innerHeight - py - SAFE_GAP
    let y = spaceBottom >= ch ? py + POINTER_GAP : py - ch - POINTER_GAP

    // 最终钳位（视口坐标，因为 position: fixed）
    x = Math.max(SAFE_GAP, Math.min(x, window.innerWidth - SAFE_GAP - cw))
    y = Math.max(SAFE_GAP, Math.min(y, window.innerHeight - SAFE_GAP - ch))
    return [x, y]
  }
  const [option, setOption] = useState<any>({
    title: {},
    color: ChartColorList,
    tooltip: {
      trigger: 'axis',
      confine: false,
      appendToBody: true,
      enterable: true,
      extraCssText: [
        'position: fixed', // 结合我们返回的视口坐标
        'max-width: 380px',
        'max-height: 400px',
        'white-space: normal',
        'word-break: break-word',
        'overflow: auto',
        'box-sizing: border-box',
        'padding: 8px 10px',
        'z-index: 2147483646', // 足够大，防止被更高层遮盖
        'pointer-events: auto',
      ].join(';'),
      position: 'left',
    },
    backgroundColor: 'rgba(0,0,0,0)',
    legend: { type: 'scroll', data: [] },
    grid: {
      left: '3%',
      right: '4%',
      bottom: showLegend ? '15%' : '5%',
      top: '5%',
      containLabel: true,
    },
    series: [],
    toolbox: { show: false },
    brush: {
      toolbox: ['lineX'],
      brushStyle: {
        borderWidth: 1,
        color: 'rgba(120,140,180,0.3)',
        borderColor: 'rgba(0,0,0,0.5)',
      },
    },
  })

  useEffect(() => {
    const newOption = {
      ...option,
      yAxis: {
        type: 'value',
        axisLabel: { formatter: (value: number) => YFormatter(value) },
      },
      axisPointer: { type: 'line', snap: true },
      xAxis: {
        type: 'time',
        boundaryGap: false,
        axisPointer: { type: 'line' },
        axisLabel: {
          formatter: (value: number) => timeUtils.format(value, 'HH:mm'),
          hideOverlap: true,
        },
      },
      color: ChartColorList,
      series: chartData?.map((item: any) => ({
        data: item.data,
        type: 'line',
        smooth: true,
        name: item.labels,
        showSymbol: false,
      })),
      legend: {
        show: showLegend,
        icon: 'roundRect',
        type: 'scroll',
        left: 10,
        bottom: 0,
        itemWidth: 10,
        itemHeight: 10,
        data: chartData?.map((item: any) => item.labels),
        itemGap: 30,
        textStyle: { width: 200, overflow: 'hidden' },
        tooltip: { show: true },
      },
    }
    if (chartRef.current) {
      const chartInstance = chartRef.current.getEchartsInstance()
      chartInstance.setOption(newOption, true)
    }
    setOption(newOption)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [chartData, currentTheme])

  return (
    <>
      {chartData && chartData?.length > 0 && option ? (
        <div className="w-full flex-1 min-h-0 text-sm">
          <ReactECharts
            ref={chartRef}
            theme={currentTheme}
            option={option}
            style={{ height: '100%', minHeight: 180, width: '100%' }}
          />
        </div>
      ) : (
        <Empty context="暂无数据" />
      )}
    </>
  )
}

// 转换指标数据格式
const convertMetricsData = (data: any[]) => {
  return data.map((item) => ({
    data: Object.entries(item.chart.chartData)
      ?.map(([key, value]) => [Number(key) / 1000, value as number])
      .sort((a, b) => (a[0] as number) - (b[0] as number)),
    labels: JSON.stringify(item.labels),
  }))
}

const NodeMetrics: React.FC<NodeMetricsProps> = ({ metrics, type }) => {
  const { token } = useToken()
  const isAbnormal = type === 'abnormal'

  return (
    <div className="flex flex-col gap-3">
      {metrics?.map((item) => (
        <div
          className="min-h-[280px] flex flex-col bg-[var(--ant-color-fill-quaternary)] p-2 rounded-xl relative"
          key={item.title}
        >
          {/* 右上角图标 */}
          <div className="absolute top-2 right-2 flex items-center justify-center w-6 h-6 rounded-full">
            {isAbnormal ? (
              <ExclamationCircleFilled style={{ color: token.colorError, fontSize: 14 }} />
            ) : (
              <CheckCircleFilled style={{ color: token.colorSuccess, fontSize: 14 }} />
            )}
          </div>

          <div className="pb-2 px-2 pr-8 font-bold shrink-0">{item.title}</div>
          <div className="flex-1 min-h-[180px]">
            <InlineChart
              chartData={convertMetricsData(item.timeseries)}
              YFormatter={(value) => value.toLocaleString() + ' ' + item.unit}
            />
          </div>
          {item.detection && (
            <div
              className="flex items-start gap-2 mx-2 mt-2 px-3 py-2 rounded-lg text-xs leading-relaxed shrink-0"
              style={{
                backgroundColor: isAbnormal ? 'rgba(255, 77, 79, 0.08)' : 'rgba(82, 196, 26, 0.08)',
                border: `1px solid ${isAbnormal ? token.colorErrorBorder : token.colorSuccessBorder}`,
                color: token.colorTextSecondary,
                borderLeft: `5px solid ${isAbnormal ? token.colorError : token.colorSuccess}`,
              }}
            >
              <InfoCircleOutlined
                style={{
                  color: isAbnormal ? token.colorError : token.colorSuccess,
                  fontSize: 14,
                  marginTop: 2,
                  flexShrink: 0,
                }}
              />
              <span className="flex-1">{item.detection}</span>
            </div>
          )}
        </div>
      ))}
    </div>
  )
}

export default NodeMetrics
