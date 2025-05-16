export default {
  color: [
    '#7eb26d',
    '#eab839',
    '#6ed0e0',
    '#ef843c',
    '#e24d42',
    '#1f78c1',
    '#ba43a9',
    '#705da0',
    '#4e8340',
  ],
  backgroundColor: '#ffffff',
  textStyle: {},
  title: {
    textStyle: {
      color: '#2c2c2c',
    },
    subtextStyle: {
      color: '#444444',
    },
  },
  line: {
    itemStyle: { borderWidth: 1 },
    lineStyle: { width: 2 },
    symbolSize: 4,
    symbol: 'circle',
    smooth: false,
  },
  radar: {
    itemStyle: { borderWidth: 1 },
    lineStyle: { width: 2 },
    symbolSize: 4,
    symbol: 'circle',
    smooth: false,
  },
  bar: {
    itemStyle: {
      barBorderWidth: 0,
      barBorderColor: '#eeeeee',
    },
  },
  pie: {
    itemStyle: {
      borderWidth: 0,
      borderColor: '#eeeeee',
    },
  },
  scatter: {
    itemStyle: {
      borderWidth: 0,
      borderColor: '#eeeeee',
    },
  },
  boxplot: {
    itemStyle: {
      borderWidth: 0,
      borderColor: '#eeeeee',
    },
  },
  parallel: {
    itemStyle: {
      borderWidth: 0,
      borderColor: '#eeeeee',
    },
  },
  candlestick: {
    itemStyle: {
      color: '#e24d42',
      color0: '#7eb26d',
      borderColor: '#e24d42',
      borderColor0: '#7eb26d',
      borderWidth: 1,
    },
  },
  categoryAxis: {
    axisLine: {
      show: true,
      lineStyle: { color: '#cccccc' },
    },
    axisTick: {
      show: true,
      lineStyle: { color: '#cccccc' },
    },
    axisLabel: {
      show: true,
      color: '#555555',
    },
    splitLine: {
      show: true,
      lineStyle: { color: ['#f0f0f0'] },
    },
    splitArea: { show: false },
  },
  valueAxis: {
    axisLine: {
      show: true,
      lineStyle: { color: '#cccccc' },
    },
    axisTick: {
      show: true,
      lineStyle: { color: '#cccccc' },
    },
    axisLabel: {
      show: true,
      color: '#555555',
    },
    splitLine: {
      show: true,
      lineStyle: { color: ['#f0f0f0'] },
    },
    splitArea: { show: false },
  },
  tooltip: {
    axisPointer: {
      lineStyle: {
        color: '#aaaaaa',
        width: 1,
      },
      crossStyle: {
        color: '#aaaaaa',
        width: 1,
      },
    },
  },
  legend: {
    textStyle: {
      color: '#333333',
    },
  },
  toolbox: {
    iconStyle: {
      borderColor: '#666666',
    },
    emphasis: {
      iconStyle: {
        borderColor: '#333333',
      },
    },
  },
  visualMap: {
    color: ['#e24d42', '#f0c29b', '#f6f6f6'],
  },
  dataZoom: {
    backgroundColor: 'rgba(255,255,255,0)',
    dataBackgroundColor: 'rgba(200,200,200,0.3)',
    fillerColor: 'rgba(167,183,204,0.4)',
    handleColor: '#999999',
    textStyle: {
      color: '#333333',
    },
  },
  markPoint: {
    label: {
      color: '#000000',
    },
    emphasis: {
      label: {
        color: '#000000',
      },
    },
  },
}
