import React, { useEffect, useRef, useState } from 'react'
import ReactECharts from 'echarts-for-react'
import dayjs from 'dayjs' // 用来格式化时间
import { useLogsContext } from 'src/contexts/LogsContext'
import { convertTime } from 'src/utils/time'
import { Empty } from 'antd'
import { useDispatch } from 'react-redux'

const BarChart = () => {
  const chartRef = useRef(null)
  const { logsChartData } = useLogsContext()
  const dispatch = useDispatch()
  const setStoreTimeRange = (value) => {
    dispatch({ type: 'SET_TIMERANGE', payload: value })
  }

  const [option, setOption] = useState({
    tooltip: {
      show: true,
      confine: 'true',
      trigger: 'axis',
      // axisPointer: {
      //   type: 'none', // 禁用坐标辅助线
      // },
      axisPointer: {
        type: 'shadow', // makes entire x-axis area clickable
      },
      textStyle: {
        overflow: 'breakAll',
        width: 40,
      },
      formatter: (params) => {
        const { data } = params[0]
        return `<div>
            <div>开始时间：${convertTime(data.from, 'yyyy-mm-dd hh:mm:ss')}</div>
            <div>开始时间：${convertTime(data.to, 'yyyy-mm-dd hh:mm:ss')}</div>
            <div>次数:${data.value}</div>
          </div>`
      },
    },

    yAxis: {
      show: false, // 隐藏纵坐标
    },
    grid: {
      show: false, // 隐藏网格背景线
    },
  })
  useEffect(() => {
    // console.log(chartData)
    const newOption = {
      ...option,
      xAxis: {
        type: 'category',
        data: logsChartData.map((item) => convertTime(item.from, 'yyyy-mm-dd hh:mm:ss')), // 格式化时间显示
        axisPointer: {
          type: 'shadow', // 让整个x轴区域响应点击
        },
        axisLabel: {
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
          // rotate: 45, // 旋转角度，防止x轴标签重叠
        },
      },

      series: [
        {
          name: 'Value',
          type: 'bar',
          data: logsChartData.map((item) => ({
            from: item.from,
            to: item.to,
            count: item.count,
            value: item.count,
          })), // 将整个对象放入 series data
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

    // 确保在组件卸载时移除事件监听
    return () => {
      zr.off('click', handleZrClick)
    }
  }

  return (
    <div className="h-[100px]">
      {logsChartData?.length > 0 ? (
        <ReactECharts
          ref={chartRef}
          option={option}
          style={{ height: 100, width: '100%' }}
          onChartReady={onChartReady}
        />
      ) : (
        <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无直方图数据" />
      )}
    </div>
  )
}

export default BarChart