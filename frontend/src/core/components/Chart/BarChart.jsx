import React from 'react'
import { Bar } from 'react-chartjs-2'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
} from 'chart.js'
ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Title, Tooltip, Legend)

export const adjustAlpha = (color, alpha) => {
  const rgba = color.match(/\d+/g)
  return `rgba(${rgba[0]}, ${rgba[1]}, ${rgba[2]}, ${alpha})`
}

const BarChart = () => {
  const color = 'rgba(229, 56, 53, 1)'
  const data = {
    labels: Array.from({ length: 30 }, (_, index) => {
      // 这里的公式是元素值等于其索引的平方
      return index * index
    }),
    datasets: [
      {
        label: 'Dataset 1',
        data: Array.from({ length: 30 }, (_, index) => {
          // 这里的公式是元素值等于其索引的平方
          return 100 - index
        }),
        // fill: true, // 填充区域
        backgroundColor: color,
        borderColor: color,
        // backgroundColor: adjustAlpha(color, 0.2),
        // pointRadius: 0, // 不显示数据点
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
        display: false, // 不显示横坐标轴
      },
      y: {
        display: false, // 不显示纵坐标轴
      },
    },
  }

  return <Bar data={data} options={options} />
}
// BarChart.propTypes = {
//   color: PropTypes.string.isRequired,
// };
export default BarChart
