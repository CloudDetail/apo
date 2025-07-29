/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useEffect, useRef, useState } from 'react'
import ReactECharts from 'echarts-for-react'
import { getStep } from 'src/core/utils/step'
import { convertTime, timeUtils } from 'src/core/utils/time'
import { DelayLineChartTitleMap, MetricsLineChartColor, YValueMinInterval } from 'src/constants'
import { useDispatch, useSelector } from 'react-redux'
import { useTranslation } from 'react-i18next'

export const adjustAlpha = (color, alpha) => {
  const rgba = color.match(/\d+/g)
  return `rgba(${rgba[0]}, ${rgba[1]}, ${rgba[2]}, ${alpha})`
}

const DelayLineChart = ({ data, timeRange, type, allowTimeBrush = true, needFillZero = true }) => {
  const { t } = useTranslation('common')
  const chartRef = useRef(null)
  const dispatch = useDispatch()
  const setStoreTimeRange = (value) => {
    dispatch({ type: 'SET_TIMERANGE', payload: value })
  }
  const { theme } = useSelector((state) => state.settingReducer)
  const convertYValue = (value) => {
    switch (type) {
      case 'logs':
        return Math.floor(value) + t('units.unit')
      case 'latency':
        if (value > 0 && value < 10) {
          return '< 0.01 ms'
        } else {
          return convertTime(value, 'ms', 2) + 'ms'
        }
      case 'p90':
        if (value > 0 && value < 10) {
          return '< 0.01 ms'
        } else {
          return convertTime(value, 'ms', 2) + 'ms'
        }
      case 'errorRate':
        if (value > 0 && value < 0.01) {
          return '< 0.01%'
        }
        return parseFloat(value.toFixed(2)) + '%'
      case 'tps':
        if (value > 0 && value < 0.01) {
          return '< 0.01' + t('units.timesPerMinute')
        }
        return parseFloat((Math.floor(value * 100) / 100).toString()) + t('units.timesPerMinute')
    }
  }
  const [option, setOption] = useState({
    title: {},
    tooltip: {
      trigger: 'axis',
      confine: true,
      enterable: true,
      // alwaysShowContent: true,
      axisPointer: {
        type: 'line',
        snap: true,
        label: {
          show: false,
        },
      },
      formatter: (params) => {
        let result = `<div class="rgb(102, 102, 102)">${convertTime(params[0]?.data[0] * 1000, 'yyyy-mm-dd hh:mm:ss')}<br/></div>
        <div class="overflow-hidden w-full " >`
        result += `<div class="flex flex-row items-center justify-between">
                      <div class="flex flex-row items-center flex-nowrap flex-shrink flex-1 whitespace-normal break-words">
                        <div class=" my-2 mr-2 rounded-full w-3 h-3 flex-grow-0 flex-shrink-0" style="background:${params[0]?.color}"></div>
                        <div class="flex-1">${params[0]?.seriesName}</div>
                      </div>
                      <span class="font-bold flex-shrink-0 ml-2">${convertYValue(params[0]?.data[1])}</span>
                      </div>`
        // params.forEach((param) => {
        //   result += `<div class="flex flex-row items-center justify-between">
        //               <div class="flex flex-row items-center flex-nowrap flex-shrink w-0 flex-1 whitespace-normal break-words">
        //                 <div class=" my-2 mr-2 rounded-full w-3 h-3 flex-grow-0 flex-shrink-0" style="background:${param.color}"></div>
        //                 <div class="flex-1 w-0">${param.seriesName}</div>
        //               </div>
        //               <span class="font-bold flex-shrink-0 ">${convertTime(param.data[1], 'ms', 2)} ms</span>
        //               </div>`
        // })
        // result+="</div>"
        return result
      },
    },
    backgroundColor: 'rgba(0,0,0,0)',
    legend: {
      type: 'scroll',
      data: [],
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '3%',
      top: '4%',
      containLabel: true,
    },
    // toolbox: {
    //   feature: {
    //     saveAsImage: {},
    //   },
    // },
    xAxis: {
      type: 'time',
      boundaryGap: false,
      axisPointer: {
        type: 'line',
        // snap: true
        interval: 1,
      },
      axisLabel: {
        hideOverlap: true,
        formatter: function (value) {
          return timeUtils.format(value, 'HH:mm')
        },
      },
      // axisLine: {
      //   lineStyle: {
      //     color: '#000000', // 设置 x 轴刻度线颜色
      //   },
      // },
      // axisTick: {
      //   lineStyle: {
      //     color: '#000000', // 设置 x 轴刻度线颜色
      //   },
      // },
    },
    yAxis: {
      type: 'value',
      minInterval: YValueMinInterval[type],
      min: 0,
      axisLabel: {
        formatter: function (value) {
          switch (type) {
            case 'logs':
              return value
            case 'latency':
              return convertTime(value, 'ms', 2)
            case 'p90':
              return convertTime(value, 'ms', 2)
            case 'errorRate':
              return value
            case 'tps':
              return value
          }
        },
      },
      // interval: 0.01, // 设置步长最小为 0.01
      // name: '耗时（ms）', // Y轴说明
      // nameLocation: 'middle', // 说明位置
      // nameTextStyle: {
      //     fontWeight: 'bold',
      //     fontSize: 14,
      //     padding: 10 // 距离轴线的距离
      // }
    },
    series: [],
    toolbox: {
      show: false, // 隐藏 toolbox
    },
    brush: {
      toolbox: ['lineX'],
      xAxisIndex: 'all',
      brushStyle: {
        borderWidth: 1,
        color: 'rgba(120,140,180,0.3)',
        borderColor: 'rgba(0,0,0,0.5)',
      },
    },
  })

  // 处理缺少数据的时间点并补0
  const fillMissingData = () => {
    const filledData = []
    const { startTime, endTime } = timeRange
    const step = getStep(startTime, endTime)

    for (let time = startTime; time <= endTime; time += step) {
      filledData.push({
        timestamp: time, // 用于显示在图例中
        value: data[time] || 0,
      })
    }
    return filledData
  }
  const formatData = () => {
    return Object.keys(data).map((key) => ({
      timestamp: Number(key),
      value: data[key],
    }))
  }
  useEffect(() => {
    if (data) {
      const filledData = needFillZero ? fillMissingData() : formatData()

      setOption({
        ...option,
        xAxis: {
          type: 'time',
          boundaryGap: false,
          axisPointer: {
            type: 'line',
            interval: 0,
          },
          axisLabel: {
            formatter: function (value) {
              return timeUtils.format(value, 'HH:mm')
            },
            hideOverlap: true,
          },
          ...(needFillZero
            ? {
                min: timeRange.startTime / 1000,
                max: timeRange.endTime / 1000,
              }
            : {}),
        },
        series: [
          {
            data: filledData.map((i) => [i.timestamp / 1000, i.value]),
            type: 'line',
            smooth: true,
            name: DelayLineChartTitleMap[type],
            color: MetricsLineChartColor[type],
            areaStyle: {
              color: adjustAlpha(MetricsLineChartColor[type], 0.3), // 设置区域填充颜色
            },
          },
        ],
      })
    }
    if (chartRef.current) {
      const chartInstance = chartRef.current.getEchartsInstance()
      onChartReady(chartInstance)
    }
  }, [data, theme])
  const onChartReady = (chart) => {
    if (allowTimeBrush) {
      setTimeout(() => {
        chart.dispatchAction({
          type: 'takeGlobalCursor',
          key: 'brush',
          brushOption: {
            brushType: 'lineX',
            brushMode: 'single',
          },
        })
        chartRef.current?.getEchartsInstance().resize()
      }, 100)
      chart.on('brushEnd', function (params) {
        if (params.areas && params.areas.length > 0) {
          // 获取 brush 选中的区域
          const brushArea = params.areas[0]
          if (brushArea.brushType === 'lineX' && brushArea.range) {
            const range = brushArea.range

            // 获取时间轴的起始和结束时间
            const startTime = chart.convertFromPixel({ xAxisIndex: 0 }, range[0])
            const endTime = chart.convertFromPixel({ xAxisIndex: 0 }, range[1])
            setStoreTimeRange({
              rangeType: null,
              startTime: Math.round(startTime * 1000),
              endTime: Math.round(endTime * 1000),
            })
          }
        }
      })
    } else {
      chartRef.current?.getEchartsInstance().resize()
    }
  }
  return (
    <ReactECharts
      ref={chartRef}
      theme={theme}
      option={option}
      style={{ height: '100%', width: '100%' }}
    />
  )
}

export default DelayLineChart
