/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */
import { useEffect, useState } from 'react'
import { useChartsContext } from 'src/core/contexts/ChartsContext'
import TempCell from '../Table/TempCell'
interface ChartData {
  ratio: any
  value: any
  chartData?: any
}
interface ChartTempCellProps {
  type: 'latency' | 'errorRate' | 'tps'
  value: ChartData
  service: string
  endpoint: string
  timeRange: any
}
const ChartTempCell = ({ type, value, service, endpoint, timeRange }: ChartTempCellProps) => {
  const [data, setData] = useState<ChartData>({
    ratio: value.ratio,
    value: value.value,
  })
  const chartsData = useChartsContext((ctx) => ctx.chartsData)
  const chartsLoading = useChartsContext((ctx) => ctx.chartsLoading)
  useEffect(() => {
    if (chartsData?.[service]?.[endpoint]?.[type]) {
      setData({
        ...data,
        chartData: chartsData?.[service]?.[endpoint]?.[type],
      })
    }
  }, [service, endpoint, chartsData])
  return <TempCell type={type} data={data} timeRange={timeRange} chartsLoading={chartsLoading} />
}
export default ChartTempCell
