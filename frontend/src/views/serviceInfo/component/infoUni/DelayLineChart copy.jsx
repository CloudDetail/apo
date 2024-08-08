import React from 'react'
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
} from 'chart.js'
import { Line } from 'react-chartjs-2'

ChartJS.register(
  Filler,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
)

export const adjustAlpha = (color, alpha) => {
  const rgba = color.match(/\d+/g)
  return `rgba(${rgba[0]}, ${rgba[1]}, ${rgba[2]}, ${alpha})`
}

const DelayLineChart = ({ color }) => {
  const data = {
    labels: ['12:00', '12:05', '12:10', '12:15', '12:20', '12:25', '12:30'],
    datasets: [
      {
        label: 'Dataset 1',
        data: [65, 59, 80, 81, 56, 55, 40],
        fill: true, // 填充区域
        borderColor: color,
        backgroundColor: adjustAlpha(color, 0.2),
        pointRadius: 0, // 不显示数据点
        borderWidth: 1,
      },
    ],
  }

  const options = {
    plugins: {
      legend: {
        display: false, // 不显示图例
      },
    },
    scales: {
      x: {
        // display: false, // 不显示横坐标轴

        ticks: {
          minRotation: 0, // 最小旋转角度
          maxRotation: 0, // 最大旋转角度
        },
      },
      y: {
        // display: false, // 不显示纵坐标轴
      },
    },
    elements: {
      line: {
        // tension: 0.4, // 曲线张力，使曲线平滑
      },
    },
  }

  return <Line data={data} options={options} />
}
export default DelayLineChart
