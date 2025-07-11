/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useEffect, useState, useRef } from 'react'
import ReactECharts from 'echarts-for-react'
import { convertTime, timeUtils } from 'src/core/utils/time'
import Empty from 'src/core/components/Empty/Empty'
import { ChartColorList } from 'src/constants'
import LoadingSpinner from 'src/core/components/Spinner'
import { useDispatch, useSelector } from 'react-redux'

const MultiLineChart = (props) => {
  const { startTime, endTime, chartData, YFormatter, allowTimeBrush = true, emptyContext } = props
  const chartRef = useRef(null)
  const dispatch = useDispatch()
  const setStoreTimeRange = (value) => {
    dispatch({ type: 'SET_TIMERANGE', payload: value })
  }

  const [option, setOption] = useState({
    title: {},
    color: ChartColorList,
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
        if (!params || !params.length) return ''

        const timeStr = convertTime(params[0]?.data[0] * 1000, 'yyyy-mm-dd hh:mm:ss')

        let result = `<div class="max-h-[200px] flex flex-col"><div class=" flex-0 text-[rgb(102,102,102)] mb-1">${timeStr}</div><div class="w-full flex-1 h-o overflow-auto">`
        params.forEach((param) => {
          const color = param.color
          const name = param.seriesName
          const value = YFormatter(param.data[1])

          result += `
        <div class="flex flex-row items-center justify-between">
          <div class="flex flex-row items-center flex-nowrap flex-shrink flex-1 whitespace-normal break-words">
            <div class="my-2 mr-2 rounded-full w-3 h-3 flex-grow-0 flex-shrink-0" style="background:${color}"></div>
            <div class="flex-1">${name}</div>
          </div>
          <span class="font-bold flex-shrink-0 ml-2">${value}</span>
        </div>`
        })

        result += `</div></div>`
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
      top: '10px',
      containLabel: true,
    },
    series: [],
    toolbox: {
      show: false, // 隐藏 toolbox
    },
    brush: {
      toolbox: ['lineX'],
      brushStyle: {
        borderWidth: 1,
        color: 'rgba(120,140,180,0.3)',
        borderColor: 'rgba(0,0,0,0.5)',
      },
    },
  })
  const [activeSeries, setActiveSeries] = useState([])
  const [loading, setLoading] = useState(false)
  const { theme } = useSelector((state) => state.settingReducer)
  const handleActiveServices = (event, item) => {
    const chartInstance = chartRef.current.getEchartsInstance()

    if (event.metaKey) {
      if (activeSeries?.includes(item.legend)) {
        let newActives = activeSeries.filter((legend) => legend !== item.legend)
        setActiveSeries(newActives)
        chartInstance.dispatchAction({
          type: 'legendUnSelect',
          name: item.legend,
        })
      } else {
        setActiveSeries((prev) => [...prev, item.legend])
        chartInstance.dispatchAction({
          type: 'legendSelect',
          name: item.legend,
        })
      }
    } else {
      if (
        activeSeries.includes(item.legend) &&
        activeSeries?.length === 1 &&
        chartData?.length !== 1
      ) {
        setActiveSeries(chartData?.map((item) => item.legend))
        chartData?.forEach((chartItem) => {
          chartInstance.dispatchAction({
            type: 'legendSelect',
            name: chartItem.legend,
          })
        })
      } else {
        setActiveSeries([item.legend])

        // 先取消所有系列的显示
        chartData?.forEach((item) => {
          chartInstance.dispatchAction({
            type: 'legendUnSelect',
            name: item.legend,
          })
        })

        // 仅显示点击的系列
        chartInstance.dispatchAction({
          type: 'legendSelect',
          name: item.legend,
        })
      }
    }
  }

  useEffect(() => {
    const newOption = {
      ...option,
      yAxis: {
        type: 'value',
        axisLabel: {
          formatter: function (value) {
            return YFormatter(value)
          },
        },
      },
      axisPointer: {
        type: 'line',
        snap: true,
      },
      xAxis: {
        type: 'time',
        boundaryGap: false,
        axisPointer: {
          type: 'line',
          // snap: true
        },
        axisLabel: {
          formatter: function (value) {
            return timeUtils.format(value, 'HH:mm')
          },
          hideOverlap: true,
        },
        // min: startTime / 1000,
        // max: endTime / 1000,
        ...(startTime && endTime
          ? {
              min: startTime / 1000,
              max: endTime / 1000,
            }
          : {}),
      },
      color: ChartColorList,
      series: chartData?.map((item) => ({
        data: item.data,
        type: 'line',
        smooth: true,
        name: item.legend,
        showSymbol: false,
      })),
    }
    if (chartRef.current) {
      const chartInstance = chartRef.current.getEchartsInstance()
      chartInstance.setOption(newOption, true) // 这里通过true来确保完全更新
      onChartReady(chartInstance)
    }
    setActiveSeries(chartData?.map((item) => item.legend))
    setOption(newOption)
    console.log(chartData)
  }, [chartData, theme])
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
    }
  }
  return (
    <>
      <LoadingSpinner loading={loading} />
      {chartData && chartData?.length > 0 && option ? (
        <div className="w-full flex flex-row h-full text-sm">
          <ReactECharts
            ref={chartRef}
            theme={theme}
            option={option}
            style={{ height: '100%', width: '50%' }}
          />
          <div className="w-1/2 h-full overflow-y-auto">
            {chartData?.map((item, index) => (
              <div
                className={'flex break-all p-1 cursor-pointer '}
                onClick={(event) => handleActiveServices(event, item)}
                key={index}
              >
                <div
                  className="w-4 h-2 m-1 rounded flex-shrink-0 "
                  style={{ background: ChartColorList[index] }}
                ></div>
                <span className={activeSeries?.includes(item.legend) ? '' : 'text-stone-400'}>
                  {item.legend}
                </span>
              </div>
            ))}
          </div>
        </div>
      ) : (
        !loading && <Empty context={emptyContext} />
      )}
    </>
  )
}
export default MultiLineChart
