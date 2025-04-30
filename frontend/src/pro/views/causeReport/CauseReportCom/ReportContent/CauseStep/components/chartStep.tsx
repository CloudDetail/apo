import React, { useEffect, useState } from 'react'
import { cloneDeep } from 'lodash'

import { basiclineOption, buildLineSeriesData, stackDataHandle } from './service'
import { timeUtils } from 'src/core/utils/time'
import { ThemeColor, ThemeStyle } from 'src/constants'
import StackChart from './StackChart'
import CopyButton from 'src/core/components/CopyButton'
import GrafanaChart from './GrafanaChart'
import { Card } from 'antd'
import FlamegraphCanvas from './Flamegraph'
import { useTranslation } from 'react-i18next'

// const testPQL = 'sum by (device) (rate(node_disk_io_time_seconds_total{instance=\"10.10.103.219:9100\"}[5m]))*100';
interface Props {
  data: any
  occurrenceAt?: number
}
interface ChartOption {
  name: string
  action: any
  option: any
}
const i18nKey = 'causeStep'
function ChartStep({ data, occurrenceAt }: Props) {
  const { t } = useTranslation('pro/causeReport')
  const theme = ThemeStyle['dark']
  const [chartData, setChartData] = useState<any>()
  const [chartOptionList, setChartOptionList] = useState<ChartOption[]>()
  const [stackList, setStackList] = useState<any[]>([])
  const [timeRange, setTimeRange] = useState(null)

  const prepareMetricLegend = (metric: any) => {
    if (metric) {
      return JSON.stringify(metric)
    } else {
      return 'No labels'
    }
  }

  const chartOptionHandle = (chartData: any) => {
    if (data.chart === 'line') {
      let chartOptionList: ChartOption[] = []
      chartData.forEach((chart: any) => {
        let option: any
        option = cloneDeep(basiclineOption)
        option.xAxis[0].data = chart.metricList[0].values.map((item: any) =>
          timeUtils.format(parseInt((item[0] * 1000).toString(), 10), 'MM-dd HH:mm'),
        )
        // option.xAxis[0].data = chart.metricList[0].values.map((item: any) =>item[0] * 1000  );
        chart.metricList.forEach((metric: any) => {
          const name = prepareMetricLegend(metric.metric)
          option.legend.data.push(name)
          option.series.push(
            buildLineSeriesData(
              name,
              metric.values.map((item: any) => (isNaN(item[1]) ? null : item[1])),
              { connectNulls: false },
            ),
          )
        })
        if (occurrenceAt) {
          option.series.push({
            markLine: {
              data: [
                {
                  xAxis: timeUtils.format(occurrenceAt / 1000000, 'MM-dd HH:mm'),
                  lineStyle: {
                    color: ThemeColor.red,
                    cap: 'round',
                  },
                  label: {
                    show: true,
                    position: 'end',
                    fontsize: 12,
                    overflow: 'breakAll',
                    formatter: t(`${i18nKey}.requestTime`),
                    color: theme.colors.text.primary,
                    distance: 18,
                  },
                },
              ],
              symbol: ['pin', 'pin'],
              emphasis: {
                label: {
                  show: true,
                  formatter: timeUtils.format(occurrenceAt / 1000000, 'MM-dd HH:mm'),
                },
              },
            },
          })
        }

        chartOptionList.push({
          name: chart.name,
          action: chart.action,
          option: option,
        })
      })
      setChartOptionList(chartOptionList)
    }
  }
  const dataHandle = () => {
    if (data.source === 'text' && data.text?.flame_graph) {
      // const result = flameDataHandle(data.text.flame_graph)
      setChartData(JSON.parse(data.text.flame_graph || ''))
    } else if (data.source === 'object') {
      setStackList(stackDataHandle(data.object?.stack))
    } else {
      // 数据已经直接返回的处理逻辑
      let chartData: any[] = []
      Object.keys(data.metrics).forEach((key) => {
        if (data.metrics[key].result.length > 0) {
          let cdItem = {
            name: key,
            action: data.metrics[key].action,
            metricList: data.metrics[key].result,
          }
          chartData.push(cdItem)
        }
      })
      chartData.length > 0 && chartOptionHandle(chartData)
    }
    if (data.source === 'text' && data.text?.flameInfo) {
      try {
        const flameInfo = JSON.parse(data.text.flameInfo)
        if (flameInfo && typeof flameInfo === 'object') {
          const startTime = flameInfo.startTime || null
          const endTime = flameInfo.endTime || null

          if (startTime && endTime) {
            setTimeRange({ startTime, endTime })
          } else {
            console.warn('flameInfo miss timerange ')
          }
        } else {
          console.warn('flameInfo wrong')
        }
      } catch (error) {
        console.error('flameInfo JSON error:', error)
      }
    }
  }

  useEffect(() => {
    dataHandle()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [data])

  const ChartComponent = () => {
    if (data.chart === 'flame_graph') {
      return (
        chartData && (
          <div className="h-[500px] overflow-hidden">
            <FlamegraphCanvas {...chartData} timeRange={timeRange} />
          </div>
        )
      )
    } else if (data.chart === 'stack') {
      return <StackChart stackList={stackList} />
    } else {
      return (
        <div>
          {chartOptionList &&
            chartOptionList.map((chartOption) => {
              return (
                <div key={chartOption.name}>
                  <div className="chart-title">
                    {chartOption.action.type + '：' + chartOption.action.text}
                    <CopyButton value={chartOption.action.text} />
                  </div>
                  <GrafanaChart option={chartOption.option} height={260} />
                </div>
              )
            })}
        </div>
      )
    }
  }

  return (
    <Card title={data.report_name}>
      {/* <TextTag data={data.text}/> */}
      {ChartComponent()}
      {data?.conclusion?.length > 0 && (
        <div>
          <span>{t(`${i18nKey}.analysisConclusion`)}：</span>
          {/* <a className='conclusion_tag' href={`/d/c902c778-2a67-463c-8f08-ae495df7fd28/2?var-pql=${testPQL}&viewPanel=1`}>{data.conclusion}</a> */}
          <span className="conclusion_tag">{data.conclusion}</span>
        </div>
      )}
    </Card>
  )
}
// const getStyles = (theme: GrafanaTheme2) => ({
//   chartWarp: css`
//     .chart-title {
//       color: ${theme.colors.text.primary};
//       font-size: 14px;
//       font-weight: bolder;
//       white-space: normal;
//       word-break: break-all;
//       padding: 0 10px;
//       button {
//         display: inline-block;
//         vertical-align: middle;
//       }
//     }
//   `,
// })

export default ChartStep
