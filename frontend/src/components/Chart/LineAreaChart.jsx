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
import { getStep } from 'src/utils/step'
import { AiOutlineLineChart } from "react-icons/ai";

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

const AreaChart = ({ color, data, timeRange }) => {
  const [chartData, setChartData] = useState({
    labels: [],
    datasets: [
      {
        label: 'My Dataset',
        data: [],
        fill: true, // 填充区域
        borderColor: color,
        backgroundColor: adjustAlpha(color, 0.2),
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
            borderColor: color,
            backgroundColor: adjustAlpha(color, 0.2),
            pointRadius: 0, // 不显示数据点
            borderWidth: 1,
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

  return chartData && data ? <Line data={chartData} options={options} /> : <div className="w-full h-full flex items-center"><AiOutlineLineChart size={30} color='#98a2b3'/></div>
}
export default AreaChart
