/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState } from 'react'
import { convertTime } from 'src/core/utils/time'
import { getServiceDsecendantMetricsApi } from 'core/api/serviceInfo'
import { getStep } from 'src/core/utils/step'
import LoadingSpinner from 'src/core/components/Spinner'
import { useDebounce } from 'react-use'
import { useTranslation } from 'react-i18next'
import MultiLineChart from 'src/core/components/Chart/MultiLineChart'
import { useDispatch, useSelector } from 'react-redux'
import { usePropsContext } from 'src/core/contexts/PropsContext'
const convertMetricsData = (data) => {
  return data.map((item) => ({
    data: item.latencyP90.map((i) => [i.timestamp / 1000, i.value]),
    legend: item.serviceName + `(${item.endpoint})`,
  }))
}

const TimelapseLineChart = (props) => {
  const { startTime, endTime, serviceName, endpoint } = props
  const { t } = useTranslation('oss/serviceInfo')
  const dispatch = useDispatch()
  const setStoreTimeRange = (value) => {
    dispatch({ type: 'SET_TIMERANGE', payload: value })
  }
  const { dataGroupId } = useSelector((state) => state.dataGroupReducer)
  const { clusterIds } = usePropsContext()
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
      // extraCssText: 'white-space: normal;word-break: break-all;',
      formatter: (params) => {
        let result = `<div class="rgb(102, 102, 102)">${convertTime(params[0]?.data[0] * 1000, 'yyyy-mm-dd hh:mm:ss')}<br/></div>
        <div class="overflow-hidden" >`
        result += `<div class="flex flex-row items-center justify-between">
                      <div class="flex flex-row items-center flex-nowrap flex-shrink flex-1 break-words">
                        <div class=" my-2 mr-2 rounded-full w-3 h-3 flex-grow-0 flex-shrink-0" style="background:${params[0]?.color}"></div>
                        <div class="flex-1">${params[0]?.seriesName}</div>
                      </div>
                      <span class="font-bold flex-shrink-0 ml-2">${convertTime(params[0]?.data[1], 'ms', 2)} ms</span>
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
        hideOverlap: true,
        formatter: function (value) {
          return timeUtils.format(value, 'HH:mm')
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
      name: t('dependent.dependentTable.timestamp'), // Y轴说明
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

  const getChartData = () => {
    getServiceDsecendantMetricsApi({
      startTime: startTime,
      endTime: endTime,
      service: serviceName,
      endpoint: endpoint,
      step: getStep(startTime, endTime),
      clusterIds: clusterIds,
      groupId: dataGroupId,
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
      if (serviceName && endpoint && startTime && endTime && dataGroupId !== null) {
        setLoading(true)
        getChartData()
      }
    },
    300, // 延迟时间 300ms
    [serviceName, startTime, endTime, endpoint, dataGroupId, clusterIds],
  )
  return (
    <>
      <LoadingSpinner loading={loading} />
      {chartData && (
        <MultiLineChart
          emptyContext={t('dependent.timelapseLineChart.noDownstreamDependencies')}
          chartData={convertMetricsData(chartData)}
          startTime={startTime}
          endTime={endTime}
          YFormatter={(value) => convertTime(value, 'ms', 2) + 'ms'}
        />
      )}
    </>
  )
}
export default TimelapseLineChart
