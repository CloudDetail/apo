import React, { useEffect, useState } from 'react'
import ReactECharts from 'echarts-for-react'
import { useLocation } from 'react-router-dom'
import { getStep } from 'src/utils/step'
import { convertTime } from 'src/utils/time'
import { format } from 'date-fns'
import { DelayLineChartTitleMap, MetricsLineChartColor, YValueMinInterval } from 'src/constants'

export const adjustAlpha = (color, alpha) => {
  const rgba = color.match(/\d+/g)
  return `rgba(${rgba[0]}, ${rgba[1]}, ${rgba[2]}, ${alpha})`
}

const DelayLineChart = ({ color, multiple = false, data, timeRange, type }) => {
  const location = useLocation()
  const searchParams = new URLSearchParams(location.search)

  const serviceName = searchParams.get('service-name')
  const convertYValue = (value) => {
    switch (type) {
      case 'logs':
        return Math.floor(value) + '个'
      case 'latency':
        let result = convertTime(value, 'ms', 2)
        if (result > 0 && result < 0.01) {
          return '< 0.01ms'
        }
        return result + 'ms'
      case 'errorRate':
        if (value > 0 && value < 0.01) {
          return '< 0.01%'
        }
        return parseFloat(value.toFixed(2)) + '%'
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
          return format(value, 'hh:mm')
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
            case 'errorRate':
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
  }, [data])

  return <ReactECharts theme="dark" option={option} style={{ height: '200px', width: '250px' }} />
}

export default DelayLineChart
