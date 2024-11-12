import React, { useEffect, useRef, useState } from 'react'
import ReactECharts from 'echarts-for-react'
import { getStep } from 'src/core/utils/step'
import { convertTime } from 'src/core/utils/time'
import { format } from 'date-fns'
import { DelayLineChartTitleMap, MetricsLineChartColor, YValueMinInterval } from 'src/constants'
import { useDispatch } from 'react-redux'

export const adjustAlpha = (color, alpha) => {
  const rgba = color.match(/\d+/g)
  return `rgba(${rgba[0]}, ${rgba[1]}, ${rgba[2]}, ${alpha})`
}

const DelayLineChart = ({ data, timeRange, type }) => {
  const chartRef = useRef(null)
  const dispatch = useDispatch()
  const setStoreTimeRange = (value) => {
    dispatch({ type: 'SET_TIMERANGE', payload: value })
  }
  const convertYValue = (value) => {
    switch (type) {
      case 'logs':
        return Math.floor(value) + '个'
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
          return '< 0.01次/分'
        }
        return parseFloat((Math.floor(value * 100) / 100).toString()) + '次/分'
    }
  }
  const [option, setOption] = useState({
    title: {},
    tooltip: {
      trigger: 'item',
      confine: true,
      enterable: true,
      // alwaysShowContent: true,
      axisPointer: {
        type: 'cross',
        label: {
          formatter: function (params) {
            // 自定义格式化函数，params.value 是轴上指示的值
            const { axisDimension, value } = params
            if (axisDimension === 'y') {
              return convertYValue(value)
            } else {
              return convertTime(value * 1000, 'yyyy-mm-dd hh:mm:ss')
            }
            // return `自定义格式: ${params.value}`;
          },
        },
      },
      //   position: function (point, params, dom, rect, size) {
      //     // point 是鼠标位置 [x, y]
      //     // size 包含 dom 的宽高 {contentSize: [width, height], viewSize: [width, height]}
      //     var x = point[0];
      //     var y = point[1];
      //     var viewWidth = size.viewSize[0];
      //     var viewHeight = size.viewSize[1];
      //     var boxWidth = size.contentSize[0];
      //     var boxHeight = size.contentSize[1];

      //     var posX = x + 20; // 偏移量
      //     var posY = y + 20; // 偏移量

      //     // 防止 tooltip 超出右边界
      //     if (posX + boxWidth > viewWidth) {
      //         posX = x - boxWidth - 20;
      //     }

      //     // 防止 tooltip 超出下边界
      //     if (posY + boxHeight > viewHeight) {
      //         posY = y - boxHeight - 20;
      //     }

      //     return [posX, posY];
      // },
      // appendToBody: true,
      // className: 'w-[70%] overflow-x-hidden overflow-y-auto fixed ',
      // extraCssText: 'white-space: normal;word-break: break-all;',
      formatter: (params) => {
        let result = `<div class="rgb(102, 102, 102)">${convertTime(params.data[0] * 1000, 'yyyy-mm-dd hh:mm:ss')}<br/></div>
        <div class="overflow-hidden w-full " >`
        result += `<div class="flex flex-row items-center justify-between">
                      <div class="flex flex-row items-center flex-nowrap flex-shrink w-0 flex-1 whitespace-normal break-words">
                        <div class=" my-2 mr-2 rounded-full w-3 h-3 flex-grow-0 flex-shrink-0" style="background:${params.color}"></div>
                        <div class="flex-1 w-0">${params.seriesName}</div>
                      </div>
                      <span class="font-bold flex-shrink-0 ">${convertYValue(params.data[1])}</span>
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
          return format(value, 'HH:mm')
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

  useEffect(() => {
    if (data) {
      const filledData = fillMissingData()

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
              return format(value, 'HH:mm')
            },
            hideOverlap: true,
          },
          min: timeRange.startTime / 1000,
          max: timeRange.endTime / 1000,
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
  }, [data])
  const onChartReady = (chart) => {
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
  return (
    <ReactECharts
      ref={chartRef}
      theme="dark"
      option={option}
      style={{ height: '200px', width: '250px' }}
    />
  )
}

export default DelayLineChart
