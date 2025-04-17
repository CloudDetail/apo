/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react'
import ReactECharts from 'echarts-for-react'

interface PieChartProps {
  data: { name: string; value: number }[]
  title?: string
}

const PieChart: React.FC<PieChartProps> = ({ data, title = '' }) => {
  const option = {
    title: {
      text: title,
      left: 'center',
      top: 10,
      textStyle: {
        fontSize: 18,
      },
    },
    tooltip: {
      trigger: 'item',
      formatter: '{b}: {c} ({d}%)',
    },
    legend: {
      show: false,
    },
    label: {
      show: true,
      color: '#ffffff',
      fontSize: 12,
      margin: 20,
      formatter: '{b}: {d}%',
      overflow: 'break', // or 'truncate' or 'none'
    },
    series: [
      {
        name: title,
        type: 'pie',
        radius: ['40%', '60%'],
        center: ['50%', '50%'],
        data: data.map(item => ({
          ...item,
          itemStyle: {
            color: item.type === 'error' ? '#ee6666' : '#91cc75'
          }
        })),
        emphasis: {
          itemStyle: {
            shadowBlur: 10,
            shadowOffsetX: 0,
            shadowColor: 'rgba(0, 0, 0, 0.3)',
          },
        },
      },
    ],
  }

  return <ReactECharts option={option} style={{ height: '150px' }} />
}

export default PieChart
