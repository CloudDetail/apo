/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { createContext, useContextSelector } from '@fluentui/react-context-selector'
import { ReactNode, useEffect, useMemo, useState } from 'react'
import { getServiceChartsApi } from '../api/service'
import { useSelector } from 'react-redux'
import { selectSecondsTimeRange } from '../store/reducers/timeRangeReducer'
import { getStep } from '../utils/step'

interface ChartsContextType {}
const ChartsContext = createContext<ChartsContextType>({} as ChartsContextType)

export const useChartsContext = <T,>(selector: (context: any) => T): T =>
  useContextSelector(ChartsContext, selector)
export const ChartsProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const [chartsData, setChartsData] = useState(null)
  const [chartsLoading, setChartsLoading] = useState(true)
  const { startTime, endTime } = useSelector(selectSecondsTimeRange)
  const getChartsData = async (serviceList, endpointList) => {
    try {
      setChartsLoading(true)
      if (serviceList.length === 0 || endpointList.length === 0) {
        return
      }
      const res = await getServiceChartsApi({
        startTime: startTime,
        endTime: endTime,
        step: getStep(startTime, endTime),
        serviceList,
        endpointList,
      })
      setChartsData(res)
    } finally {
      setChartsLoading(false)
    }
  }
  const finalValue = {
    chartsData,
    setChartsData,
    chartsLoading,
    setChartsLoading,
    getChartsData,
  }
  return <ChartsContext.Provider value={finalValue}>{children}</ChartsContext.Provider>
}
