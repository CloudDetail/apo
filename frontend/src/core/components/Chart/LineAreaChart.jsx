/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState, useEffect } from 'react'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler,
  TimeScale,
} from 'chart.js'
import { Line } from 'react-chartjs-2'
import { getStep } from 'src/core/utils/step'
import { AiOutlineLineChart } from 'react-icons/ai'
import { Popover } from 'antd'
import DelayLineChart from './DelayLineChart'
import { MetricsLineChartColor } from 'src/constants'

ChartJS.register(
  Filler,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  TimeScale,
)

export const adjustAlpha = (color, alpha) => {
  const rgba = color.match(/\d+/g)
  return `rgba(${rgba[0]}, ${rgba[1]}, ${rgba[2]}, ${alpha})`
}

const AreaChart = ({ type, data, timeRange }) => {
  const [chartData, setChartData] = useState({
    labels: [],
    datasets: [
      {
        label: 'My Dataset',
        data: [],
        fill: true, // 填充区域
        borderColor: MetricsLineChartColor[type],
        backgroundColor: adjustAlpha(MetricsLineChartColor[type], 0.2),
        pointRadius: 0, // 不显示数据点
        borderWidth: 1,
      },
    ],
  })
  // 处理缺少数据的时间点并补0
  const fillMissingData = () => {
    const filledData = []
    const { startTime, endTime } = timeRange
    const step = getStep(startTime, endTime)

    for (let time = startTime; time <= endTime; time += step) {
      const point = data[time]
      if (point) {
        filledData.push({ time: time, value: data[time] })
      } else {
        filledData.push({ time: time, value: 0 })
      }
    }
    return filledData
  }
  useEffect(() => {
    if (data && timeRange) {
      const filledData = fillMissingData()
      setChartData({
        labels: filledData.map((d) => d.time),
        datasets: [
          {
            label: 'My Dataset',
            data: filledData.map((d) => d.value),
            fill: true, // 填充区域
            borderColor: MetricsLineChartColor[type],
            backgroundColor: adjustAlpha(MetricsLineChartColor[type], 0.2),
            pointRadius: 0, // 不显示数据点
            borderWidth: 1,
            pointHoverRadius: 0, // 隐藏悬停时的数据点
          },
        ],
      })
    }
  }, [data, timeRange])

  const options = {
    plugins: {
      legend: {
        display: false, // 不显示图例
      },
      tooltip: {
        enabled: false, // 禁用 tooltip
      },
    },
    scales: {
      x: {
        display: false, // 不显示横坐标轴
      },
      y: {
        display: false, // 不显示纵坐标轴
        beginAtZero: true,
      },
    },
    elements: {
      line: {
        // tension: 0.4, // 曲线张力，使曲线平滑
      },
    },
  }

  return chartData && data ? (
    <Popover
      content={
        <div onClickCapture={(e) => e.stopPropagation()}>
          <DelayLineChart data={data} timeRange={timeRange} type={type} />
        </div>
      }
      title=""
      zIndex={1060}
    >
      <Line data={chartData} options={options} />
    </Popover>
  ) : (
    <div className="w-full h-full flex items-center justify-center">
      <AiOutlineLineChart size={30} color="#98a2b3" />
    </div>
  )
}
export default AreaChart
