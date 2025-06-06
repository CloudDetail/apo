/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useEffect, useRef, useState } from 'react'
import ReactECharts from 'echarts-for-react'
import { useLogsContext } from 'src/core/contexts/LogsContext'
import { convertTime, timeUtils } from 'src/core/utils/time'
import { Empty } from 'antd'
import { useDispatch } from 'react-redux'
import { useTranslation } from 'react-i18next' // 引入i18n

const BarChart = () => {
  const { t } = useTranslation('oss/fullLogs')
  const chartRef = useRef(null)
  const { logsChartData } = useLogsContext()
  const logsChartDataRef = useRef(logsChartData)
  const dispatch = useDispatch()
  const setStoreTimeRange = (value) => {
    dispatch({ type: 'SET_TIMERANGE', payload: value })
  }

  const [option, setOption] = useState({
    tooltip: {
      show: true,
      confine: true,
      trigger: 'axis',
      axisPointer: {
        type: 'shadow',
      },
      textStyle: {
        overflow: 'breakAll',
        width: 40,
      },
      formatter: (params) => {
        const { data } = params[0]
        return `<div>
            <div>${t('histogram.startTimeText')}：${convertTime(data.from, 'yyyy-mm-dd hh:mm:ss')}</div>
            <div>${t('histogram.endTimeText')}：${convertTime(data.to, 'yyyy-mm-dd hh:mm:ss')}</div>
            <div>${t('histogram.times')}:${data.value}</div>
          </div>`
      },
    },
    brush: {
      xAxisIndex: 'all',
      brushLink: 'all',
      outOfBrush: {
        colorAlpha: 0.3,
      },
    },
    xAxis: {
      type: 'category',
      data: [],
    },
    yAxis: {
      show: false,
    },
    grid: {
      show: false,
    },
  })

  useEffect(() => {
    logsChartDataRef.current = logsChartData
    if (!logsChartData || logsChartData.length === 0) return
    const newOption = {
      ...option,
      xAxis: {
        type: 'category',
        data: logsChartData.map((item) => convertTime(item.from, 'yyyy-mm-dd hh:mm:ss')),
        axisPointer: {
          type: 'shadow', // 让整个x轴区域响应点击
        },
        axisLabel: {
          formatter: (value) => {
            const time = new Date(value)
            const now = new Date()

            // 根据时间差判断是否需要显示年份，月份等
            if (timeUtils.getDiff(time, now, 'years') !== 0) {
              return timeUtils.format(time, 'yyyy/MM/dd HH:mm')
            } else if (timeUtils.getDiff(time, now, 'days') !== 0) {
              return timeUtils.format(time, 'MM/dd HH:mm')
            } else if (timeUtils.getDiff(time, now, 'hours') !== 0) {
              return timeUtils.format(time, 'HH:mm')
            } else if (timeUtils.getDiff(time, now, 'minutes') !== 0) {
              return timeUtils.format(time, 'HH:mm:ss') // 如果分钟不同，展示到秒
            }
            return timeUtils.format(time, 'HH:mm:ss') // 如果秒不同，展示到秒
          },
          // rotate: 45, // 旋转角度，防止x轴标签重叠
        },
      },
      series: [
        {
          type: 'bar',
          data: logsChartData.map((item) => ({
            from: item.from,
            to: item.to,
            count: item.count,
            value: item.count,
          })),
        },
      ],
    }

    if (chartRef.current) {
      const chartInstance = chartRef.current.getEchartsInstance()
      chartInstance.setOption(newOption, true) // 这里通过true来确保完全更新
      //   onChartReady(chartInstance)
    }

    setOption(newOption)
  }, [logsChartData])

  const onChartReady = (chartInstance) => {
    const zr = chartInstance.getZr()

    //拖动
    const handleMouseDown = () => {
      chartInstance.dispatchAction({
        type: 'takeGlobalCursor',
        key: 'brush',
        brushOption: {
          brushType: 'lineX',
        },
      })
    }
    const handleMouseUp = () => {
      const brushComponent = chartInstance.getModel().getComponent('brush')
      const areas = brushComponent?.areas || []

      if (areas.length > 0) {
        const currentLogsChartData = logsChartDataRef.current
        const range = areas[0]?.coordRange
        if (range && range[0] >= 0 && range[1] < currentLogsChartData.length) {
          const startTime = currentLogsChartData[range[0]]?.from
          const endTime = currentLogsChartData[range[1]]?.to
          if (startTime && endTime) {
            if (range[1] - range[0] + 1 === currentLogsChartData.length) {
              chartInstance.setOption(chartInstance.getOption(), true)
            } else {
              setStoreTimeRange({
                rangeType: null,
                startTime: Math.round(startTime),
                endTime: Math.round(endTime),
              })
            }
          }
        }
      }
    }

    //点击
    const handleZrClick = (params) => {
      const pointInPixel = [params.offsetX, params.offsetY]
      const pointInGrid = chartInstance.convertFromPixel('grid', pointInPixel)

      if (pointInGrid) {
        const categoryIndex = Math.round(pointInGrid[0]) // 获取x轴索引
        const chartOption = chartInstance.getOption() // 获取当前图表的所有数据
        const seriesData = chartOption.series[0].data[categoryIndex] // 获取点击的对象数据

        setStoreTimeRange({
          rangeType: null,
          startTime: Math.round(seriesData.from),
          endTime: Math.round(seriesData.to),
        })
      }
    }

    zr.on('click', handleZrClick)
    zr.on('mousedown', handleMouseDown)
    zr.on('mouseup', handleMouseUp)

    return () => {
      zr.off('click', handleZrClick)
      zr.off('mousedown', handleMouseDown)
      zr.off('mouseup', handleMouseUp)
    }
  }

  return (
    <div className="h-[80px]">
      {logsChartData?.length > 0 ? (
        <ReactECharts
          ref={chartRef}
          option={option}
          style={{ height: 80, width: '100%' }}
          onChartReady={onChartReady}
        />
      ) : (
        <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={t('histogram.noDataText')} />
      )}
    </div>
  )
}

export default BarChart
