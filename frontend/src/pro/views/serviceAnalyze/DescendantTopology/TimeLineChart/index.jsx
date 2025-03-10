import React, { useEffect, useRef, useState } from 'react'
import { useSelector } from 'react-redux'
import DelayLineChart from 'src/core/components/Chart/DelayLineChart'
import { useAlertAnalyzeContext } from 'src/pro/contexts/AlertAnalyzeContext'
import { selectSecondsTimeRange, toNearestSecond } from 'src/core/store/reducers/timeRangeReducer'
import { getStep } from 'src/core/utils/step'
import { convertTime } from 'src/core/utils/time'
import ReactECharts from 'echarts-for-react'
import dayjs from 'dayjs' // 用来格式化时间
import _ from 'lodash'
import { useUpdate, useUpdateEffect } from 'react-use'
import { t } from 'i18next'
const TimeLineChart = ({ rightWidth }) => {
  const {
    timeLineChartData,
    marklineTimestamp,
    deltaStartTime,
    deltaEndTime,
    changeAlertDescendantDelta,
  } = useAlertAnalyzeContext()
  const { startTime, endTime } = useSelector(selectSecondsTimeRange)
  const [currentMarkLineX, setCurrentMarkLineX] = useState(0)
  const chartRef = useRef(null)
  // useEffect(() => {
  //   console.log(deltaStartTime, deltaEndTime, startTime, endTime, '时间变了，即将update')
  //   requestAnimationFrame(updateGraphic)
  //   requestAnimationFrame(updateMarkLine)
  // }, [deltaStartTime, deltaEndTime])
  const [option, setOption] = useState({
    title: {},
    tooltip: {
      trigger: 'axis',
      confine: true,
      enterable: true,
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
    },
    yAxis: {
      type: 'value',
      minInterval: 1,
      min: 0,
    },
    series: [],
    toolbox: {
      show: false, // 隐藏 toolbox
    },
    animationDurationUpdate: 1000, // 每次更新的动画时长，单位ms
    animationEasingUpdate: 'linear', // 使用线性动画效果
  })
  const [brushStartTime, setBrushStartTime] = useState(null)
  const [brushEndTime, setBrushEndTime] = useState(null)
  useEffect(() => {
    if (timeLineChartData) {
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
            hideOverlap: true,
            formatter: (value) => {
              const time = dayjs(value)
              // 根据时间差判断是否需要显示年份，月份等
              if (time.diff(dayjs(), 'year') !== 0) {
                return time.format('YYYY/MM/DD HH:mm')
              } else if (time.diff(dayjs(), 'day') !== 0) {
                return time.format('MM/DD HH:mm')
              } else if (time.diff(dayjs(), 'hour') !== 0) {
                return time.format('HH:mm')
              } else if (time.diff(dayjs(), 'minute') !== 0) {
                return time.format('HH:mm:ss') // 如果分钟不同，展示到秒
              }
              return time.format('HH:mm:ss') // 如果秒不同，展示到秒
            },
          },
          min: startTime / 1000,
          max: endTime / 1000,
        },
        series: [
          {
            data: Object.entries(timeLineChartData).map(([key, value]) => [key / 1000, value]),
            // name: DelayLineChartTitleMap[type],
            // color: MetricsLineChartColor[type],
            // areaStyle: {
            //   color: adjustAlpha(MetricsLineChartColor[type], 0.3), // 设置区域填充颜色
            // },
            name: t('pro/serviceAnalyze:abnormalCount'),
            type: 'line',
            smooth: true,
            // markLine: {
            //   symbol: 'none', // 不显示箭头或其他符号
            //   animation: true, // 开启动画
            //   label: {
            //     show: false, // 不显示label标签
            //   },
            //   lineStyle: {
            //     color: 'red', // 竖线的颜色
            //     width: 2, // 竖线的宽度
            //     type: 'solid',
            //   },
            //   data: [
            //     {
            //       xAxis: marklineTimestamp / 1000, // 初始位置的时间戳，按需求设置为 x 轴的开始时间
            //     },
            //   ],
            // },
          },
        ],
        brush: {
          toolbox: ['lineX', 'rect', 'clear'], // 工具箱选项：矩形选择和清除选择
          throttleType: 'debounce', // 防抖处理，减少频繁更新
          throttleDelay: 300, // 延迟
          brushType: 'lineX', // 限制为矩形选择
          xAxisIndex: 0,
          brushLink: 'all',
          outOfBrush: {
            color: 'rgba(0,0,0,0.1)', // 非选择区域的颜色
          },
        },
      })
    }
    if (chartRef.current) {
      const chartInstance = chartRef.current.getEchartsInstance()
      onChartReady(chartInstance)
      chartInstance.resize()
    }
  }, [timeLineChartData, deltaStartTime, deltaEndTime])
  const handleResize = () => {
    if (chartRef.current) {
      const chartInstance = chartRef.current.getEchartsInstance()
      onChartReady()
      chartInstance.resize({
        width: chartRef.current.offsetWidth,
        height: chartRef.current.offsetHeight,
      })
    }
  }
  useEffect(() => {
    window.addEventListener('resize', handleResize)
    window.addEventListener('resize', updateMarkLine) // 监听窗口大小变化
    // window.addEventListener('resize', updateBrush) // 监听窗口大小变化
    return () => {
      window.removeEventListener('resize', handleResize)
      window.removeEventListener('resize', updateMarkLine) // 监听窗口大小变化
      // window.removeEventListener('resize', updateBrush) // 监听窗口大小变化
    }
  }, [])
  useEffect(() => {
    handleResize()
    // updateMarkLine()
  }, [rightWidth])
  const onChartReady = (chart) => {
    // 确保下一帧再更新
    // 确保下一帧再更新
    if (chartRef.current) {
      const chart = chartRef.current.getEchartsInstance()
      setTimeout(() => {
        chart.dispatchAction({
          type: 'takeGlobalCursor',
          key: 'brush',
          brushOption: {
            brushType: 'lineX',
            brushMode: 'single',
          },
        })
        requestAnimationFrame(updateBrush)
        requestAnimationFrame(updateMarkLine)
      }, 100)
      chart.on(
        'brushEnd',
        _.throttle((params) => {
          if (params.areas && params.areas.length > 0) {
            const brushArea = params.areas[0]
            if (brushArea.brushType === 'lineX' && brushArea.range) {
              const range = brushArea.range
              const brushStartTime = chart.convertFromPixel({ xAxisIndex: 0 }, range[0])
              const brushEndTime = chart.convertFromPixel({ xAxisIndex: 0 }, range[1])

              setBrushStartTime(Math.floor(Math.max(brushStartTime * 1000, startTime)))
              setBrushEndTime(Math.floor(Math.min(brushEndTime * 1000, endTime)))
            }
          }
        }, 500), // 500ms 节流间隔，可以根据需求调整
      )
      // const start = chart.convertToPixel({ xAxisIndex: 0 }, deltaStartTime / 1000)
      // const end = chart.convertToPixel({ xAxisIndex: 0 }, deltaEndTime / 1000)
      // chart.dispatchAction({
      //   type: 'brush',
      //   areas: [
      //     {
      //       brushType: 'lineX',
      //       xAxisIndex: 0,
      //       coordRange: [start, end], // 设置默认的 x 范围
      //     },
      //   ],
      // })
    }
  }
  const updateMarkLine = () => {
    if (chartRef.current) {
      const chartInstance = chartRef.current.getEchartsInstance()
      console.log(chartInstance)
      if (!chartInstance || chartInstance._disposed) {
        console.warn('ECharts 实例未初始化或已被销毁')
        return
      }

      console.log('ECharts 实例', chartInstance)

      // if (typeof chartInstance.convertToPixel === 'function') {
      //   const x = chartInstance.convertToPixel({ xAxisIndex: 0 }, marklineTimestamp / 1000)
      //   setCurrentMarkLineX(x)
      // } else {
      //   console.warn('convertToPixel 方法不存在')
      // }
      // const yAxis = chartInstance.getModel().getComponent('yAxis').axis.scale.getExtent()
      // const y = chartInstance.convertToPixel({ yAxisIndex: 0 }, yAxis[0]) // 获取Y轴最小值的像素位置（即Y轴的0值）
      // const yTop = chartInstance.convertToPixel({ yAxisIndex: 0 }, yAxis[1]) // 获取Y轴最大值的像素位置（图表顶部）

      // //细线
      // chartInstance.setOption({
      //   // graphic: [
      //   //   {
      //   //     type: 'line',
      //   //     id: 'markLine', // 给标记线设定唯一ID，避免重复添加
      //   //     shape: {
      //   //       x1: x, // 动态更新X轴位置
      //   //       x2: x,
      //   //       y1: y,
      //   //       y2: yTop,
      //   //     },
      //   //     style: {
      //   //       stroke: 'blue',
      //   //       lineWidth: 2,
      //   //     },
      //   //     transition: currentMarkLineX <= x ? ['position', 'shape'] : [''], // 定义过渡的属性
      //   //   },
      //   // ],
      // })
    }
  }
  const updateBrush = () => {
    if (chartRef.current) {
      const chart = chartRef.current.getEchartsInstance()
      chart.dispatchAction({
        type: 'brush',
        areas: [
          {
            brushType: 'lineX',

            xAxisIndex: 0,
            coordRange: [deltaStartTime / 1000, deltaEndTime / 1000], // 设置默认的 x 范围
          },
        ],
      })
    }
  }
  useEffect(() => {
    updateMarkLine()
  }, [marklineTimestamp])
  useUpdateEffect(() => {
    console.log(brushStartTime, brushEndTime)
    changeAlertDescendantDelta({
      deltaStartTime: brushStartTime,
      deltaEndTime: brushEndTime,
    })
  }, [brushStartTime, brushEndTime])
  return (
    <ReactECharts
      ref={chartRef}
      theme="dark"
      option={option}
      style={{ height: '200px', width: '100%' }}
      onChartReady={onChartReady}
    />
  )
}
export default TimeLineChart
