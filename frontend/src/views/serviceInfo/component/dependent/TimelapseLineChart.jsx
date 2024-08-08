import { color } from 'd3'
import React, { useEffect, useState, useRef } from 'react'
import { usePropsContext } from 'src/contexts/PropsContext'
import ReactECharts from 'echarts-for-react'
import { convertTime } from 'src/utils/time'
import { format } from 'date-fns'
import Empty from 'src/components/Empty/Empty'

function TimelapseLineChart(props) {
  const { data, startTime, endTime } = props
  const { serviceName } = usePropsContext()
  const [option, setOption] = useState({
    title: {},
    tooltip: {
      trigger: 'item',
      confine: true,
      enterable: true,
      // alwaysShowContent: true,
      axisPointer: {
        type: 'cross',
        label: {
          formatter: function(params) {
            // 自定义格式化函数，params.value 是轴上指示的值
            const{axisDimension, value} = params
            if(axisDimension === 'y'){
              return convertTime(value,'ms',2) + 'ms'
            } else{
              return convertTime(value * 1000,'yyyy-mm-dd hh:mm:ss')
            }
            // return `自定义格式: ${params.value}`;
          }
        }
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
          return format(value, 'hh:mm')
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
  })

  useEffect(() => {
    console.log(data)
    setOption({
      ...option,
      legend: {
        type: 'scroll',
        data: data.map((item) => {
          return item.serviceName + `(${item.endpoint})`
        }),
      },
      series: data.map((item) => {
        return {
          data: item.latencyP90.map((i) => [i.timestamp / 1000, i.value]),
          type: 'line',
          smooth: true,
          name: item.serviceName + `(${item.endpoint})`,
        }
      }),
    })
  }, [data])

  return data && data.length > 0 && option ? (
    <ReactECharts theme="dark" option={option} style={{ height: '100%', width: '100%' }} />
  ) : (
    <Empty />
  )
}
export default TimelapseLineChart
