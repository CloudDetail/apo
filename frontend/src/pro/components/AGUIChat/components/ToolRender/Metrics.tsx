import React, { useEffect, useState, useRef } from 'react'
import ReactECharts from 'echarts-for-react'
import { convertTime, timeUtils } from 'src/core/utils/time'
import Empty from 'src/core/components/Empty/Empty'
import { ChartColorList } from 'src/constants'
import { useSelector } from 'react-redux'

const ChartComponent = (props) => {
  const { chartData, YFormatter, emptyContext } = props
  const chartRef = useRef(null)

  const [option, setOption] = useState({
    title: {},
    color: ChartColorList,
    tooltip: {
      trigger: 'axis',
      confine: true,
      enterable: true,
      // alwaysShowContent: true,
      axisPointer: {
        type: 'line',
        snap: true,
        label: {
          show: false,
        },
      },
      formatter: (params) => {
        if (!params || !params.length) return ''

        const timeStr = convertTime(params[0]?.data[0] * 1000, 'yyyy-mm-dd hh:mm:ss')

        let result = `<div class="max-h-[200px] flex flex-col"><div class=" flex-0 text-[rgb(102,102,102)] mb-1">${timeStr}</div><div class="w-full flex-1 h-o overflow-auto">`
        params.forEach((param) => {
          const color = param.color
          const name = param.seriesName
          const value = YFormatter(param.data[1])

          result += `
        <div class="flex flex-row items-center justify-between">
          <div class="flex flex-row items-center flex-nowrap flex-shrink flex-1 whitespace-normal break-words">
            <div class="my-2 mr-2 rounded-full w-3 h-3 flex-grow-0 flex-shrink-0" style="background:${color}"></div>
            <div class="flex-1">${name}</div>
          </div>
          <span class="font-bold flex-shrink-0 ml-2">${value}</span>
        </div>`
        })

        result += `</div></div>`
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
      bottom: '15%',
      top: '5%',
      containLabel: true,
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
  const { theme } = useSelector((state) => state.settingReducer)

  useEffect(() => {
    const newOption = {
      ...option,
      yAxis: {
        type: 'value',
        axisLabel: {
          formatter: function (value) {
            return YFormatter(value)
          },
        },
      },
      axisPointer: {
        type: 'line',
        snap: true,
      },
      xAxis: {
        type: 'time',
        boundaryGap: false,
        axisPointer: {
          type: 'line',
          // snap: true
        },
        axisLabel: {
          formatter: function (value) {
            return timeUtils.format(value, 'HH:mm')
          },
          hideOverlap: true,
        },
      },
      color: ChartColorList,
      series: chartData?.map((item) => ({
        data: item.data,
        type: 'line',
        smooth: true,
        name: item.legend,
        showSymbol: false,
      })),
      legend: {
        show: true,
        icon: 'roundRect',
        type: 'scroll',
        left: 10,
        bottom: 0,
        itemWidth: 10,
        itemHeight: 10,
        data: chartData?.map((item) => item.legend),
        itemGap: 30,
        textStyle: {
          width: 800,
          overflow: 'breakAll',
        },
        tooltip: {
          show: true,
        },
      },
    }
    if (chartRef.current) {
      const chartInstance = chartRef.current.getEchartsInstance()
      chartInstance.setOption(newOption, true)
    }
    setOption(newOption)
  }, [chartData, theme])
  return (
    <>
      {chartData && chartData?.length > 0 && option ? (
        <div className="w-full flex flex-row h-full text-sm">
          <ReactECharts
            ref={chartRef}
            theme={theme}
            option={option}
            style={{ height: '100%', width: '100%' }}
          />
        </div>
      ) : (
        <Empty context={emptyContext} />
      )}
    </>
  )
}

const convertMetricsData = (data) => {
  return data.map((item) => ({
    data: Object.entries(item.chart.chartData)
      ?.map(([key, value]) => [Number(key) / 1000, value])
      .sort((a, b) => a[0] - b[0]),
    legend: item.legend,
  }))
}
const Metrics = ({ args }: { args: any }) => {
  return (
    <div className="grid grid-cols-3 gap-4 max-h-[500px] overflow-y-auto">
      {args?.map((item) => (
        <div
          className="h-[260px] flex flex-col bg-[var(--ant-color-fill-quaternary)] p-2 rounded-xl "
          key={item.title}
        >
          <div className="pb-2 px-2 font-bold">{item.title}</div>
          {/* <LineChart data={data} unit={unit} /> */}
          <ChartComponent
            chartData={convertMetricsData(item.timeseries)}
            unit={item.unit}
            YFormatter={(value) => value.toLocaleString() + ' ' + item.unit}
          />
        </div>
      ))}
    </div>
  )
}

export default Metrics
