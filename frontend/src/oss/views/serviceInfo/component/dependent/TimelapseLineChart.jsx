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
const convertMetricsData = (data) => {
  return data.map((item) => ({
    data: item.latencyP90.map((i) => [i.timestamp / 1000, i.value]),
    legend: item.serviceName + `(${item.endpoint})`,
  }))
}
const TimelapseLineChart = (props) => {
  const { startTime, endTime, serviceName, endpoint } = props
  const { t } = useTranslation('oss/serviceInfo')

  const [chartData, setChartData] = useState([])
  const [loading, setLoading] = useState(false)

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
