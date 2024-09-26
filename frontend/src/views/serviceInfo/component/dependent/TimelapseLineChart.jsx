import React, { useEffect, useState, useRef } from 'react'
import ReactECharts from 'echarts-for-react'
import { convertTime } from 'src/utils/time'
import { format } from 'date-fns'
import Empty from 'src/components/Empty/Empty'
import { ChartColorList } from 'src/constants'
import { getServiceDsecendantMetricsApi } from 'src/api/serviceInfo'
import { getStep } from 'src/utils/step'
import LoadingSpinner from 'src/components/Spinner'
import { useDispatch } from 'react-redux'
import { useDebounce } from 'react-use'

const TimelapseLineChart = (props) => {
  const { startTime, endTime, serviceName, endpoint } = props
  const chartRef = useRef(null)
  const dispatch = useDispatch()
  const setStoreTimeRange = (value) => {
    dispatch({ type: 'SET_TIMERANGE', payload: value })
  }

  const [option, setOption] = useState({
    title: {},
    color: ChartColorList,
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
              return convertTime(value, 'ms', 2) + 'ms'
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
      className: 'w-[70%]',
      // extraCssText: 'white-space: normal;word-break: break-all;',
      formatter: (params) => {
        let result = `<div class="rgb(102, 102, 102)">${convertTime(params.data[0] * 1000, 'yyyy-mm-dd hh:mm:ss')}<br/></div>
        <div class="overflow-hidden" >`
        result += `<div class="flex flex-row items-center justify-between">
                      <div class="flex flex-row items-center flex-nowrap flex-shrink w-0 flex-1 whitespace-normal break-words">
                        <div class=" my-2 mr-2 rounded-full w-3 h-3 flex-grow-0 flex-shrink-0" style="background:${params.color}"></div>
                        <div class="flex-1 w-0">${params.seriesName}</div>
                      </div>
                      <span class="font-bold flex-shrink-0 ">${convertTime(params.data[1], 'ms', 2)} ms</span>
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
      top: '10px',
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
      },
      axisLabel: {
        formatter: function (value) {
          return format(value, 'HH:mm')
        },
      },
    },
    yAxis: {
      type: 'value',
      axisLabel: {
        formatter: function (value) {
          return convertTime(value, 'ms', 2)
        },
      },
      name: '耗时（ms）', // Y轴说明
      // nameLocation: 'middle', // 说明位置
      // nameTextStyle: {
      //     fontWeight: 'bold',
      //     fontSize: 14,
      //     padding: 10 // 距离轴线的距离
      // }
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
  const [activeSeries, setActiveSeries] = useState(null)
  const [chartData, setChartData] = useState([])
  const [loading, setLoading] = useState(false)
  const handleActiveServices = (item) => {
    const seriesName = item.serviceName + `(${item.endpoint})`
    const chartInstance = chartRef.current.getEchartsInstance()
    if (seriesName === activeSeries) {
      setActiveSeries(null)
      chartData.forEach((item) => {
        chartInstance.dispatchAction({
          type: 'legendSelect',
          name: item.serviceName + `(${item.endpoint})`,
        })
      })
    } else {
      setActiveSeries(seriesName)

      // 先取消所有系列的显示
      chartData.forEach((item) => {
        chartInstance.dispatchAction({
          type: 'legendUnSelect',
          name: item.serviceName + `(${item.endpoint})`,
        })
      })

      // 仅显示点击的系列
      chartInstance.dispatchAction({
        type: 'legendSelect',
        name: seriesName,
      })
    }
  }

  const getChartData = () => {
    getServiceDsecendantMetricsApi({
      startTime: startTime,
      endTime: endTime,
      service: serviceName,
      endpoint: endpoint,
      step: getStep(startTime, endTime),
    })
      .then((res) => {
        // console.log(res)
        setChartData(res ?? [])
        setLoading(false)
      })
      .catch((error) => {
        setChartData([])
        setLoading(false)
      })
  }
  //防抖避免跳转使用旧时间
  useDebounce(
    () => {
      if (serviceName && endpoint && startTime && endTime) {
        setLoading(true)
        getChartData()
      }
    },
    300, // 延迟时间 300ms
    [serviceName, startTime, endTime, endpoint],
  )
  useEffect(() => {
    // console.log(chartData)
    const newOption = {
      ...option,
      xAxis: {
        type: 'time',
        boundaryGap: false,
        axisPointer: {
          type: 'line',
          // snap: true
        },
        axisLabel: {
          formatter: function (value) {
            return format(value, 'HH:mm')
          },
        },
        min: startTime / 1000,
        max: endTime / 1000,
      },
      color: ChartColorList,
      series: chartData.map((item) => {
        return {
          data: item.latencyP90.map((i) => [i.timestamp / 1000, i.value]),
          type: 'line',
          smooth: true,
          name: item.serviceName + `(${item.endpoint})`,
        }
      }),
    }
    if (chartRef.current) {
      const chartInstance = chartRef.current.getEchartsInstance()
      chartInstance.setOption(newOption, true) // 这里通过true来确保完全更新
      onChartReady(chartInstance)
    }
    setOption(newOption)
  }, [chartData])
  const onChartReady = (chart) => {
    setTimeout(() => {
      chart.dispatchAction({
        type: 'takeGlobalCursor',
        key: 'brush',
        brushOption: {
          brushType: 'lineX',
          brushMode: 'single',
        },
      })
    }, 100)
    chart.on('brushEnd', function (params) {
      if (params.areas && params.areas.length > 0) {
        // 获取 brush 选中的区域
        const brushArea = params.areas[0]
        if (brushArea.brushType === 'lineX' && brushArea.range) {
          const range = brushArea.range

          // 获取时间轴的起始和结束时间
          const startTime = chart.convertFromPixel({ xAxisIndex: 0 }, range[0])
          const endTime = chart.convertFromPixel({ xAxisIndex: 0 }, range[1])
          setStoreTimeRange({
            rangeType: null,
            startTime: Math.round(startTime * 1000),
            endTime: Math.round(endTime * 1000),
          })
        }
      }
    })
  }
  return (
    <>
      <LoadingSpinner loading={loading} />
      {chartData && chartData.length > 0 && option ? (
        <div className="w-full flex flex-row h-full text-sm">
          <ReactECharts
            ref={chartRef}
            theme="dark"
            option={option}
            style={{ height: '100%', width: '50%' }}
          />
          <div className="w-1/2 h-full overflow-y-auto">
            {chartData.map((item, index) => (
              <div
                className={'flex break-all p-1 cursor-pointer '}
                onClick={() => handleActiveServices(item)}
                key={index}
              >
                <div
                  className="w-4 h-2 m-1 rounded flex-shrink-0 "
                  style={{ background: ChartColorList[index] }}
                ></div>
                <span
                  className={
                    !activeSeries || item.serviceName + `(${item.endpoint})` === activeSeries
                      ? ''
                      : 'text-stone-400'
                  }
                >
                  {item.serviceName}({item.endpoint})
                </span>
              </div>
            ))}
          </div>
        </div>
      ) : (
        !loading && <Empty context={serviceName + ' 无下游依赖节点'} />
      )}
    </>
  )
}
export default TimelapseLineChart
