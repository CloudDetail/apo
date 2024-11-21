import { createContext, ReactNode, useContext, useEffect, useMemo, useState } from 'react'
import { useSelector } from 'react-redux'
import { useParams } from 'react-router-dom'
import { useDebounce } from 'react-use'
import { selectSecondsTimeRange } from 'src/core/store/reducers/timeRangeReducer'

// 定义上下文的数据类型
interface ChartAIType {
  // 分析的入口业务
  service: string | null
  endpoint: string | null
  anormalTypes: string[]
}

// 创建上下文
const ChartAIContext = createContext<ChartAIType | undefined>(undefined)

// 创建自定义 Hook
export const useChartAIContext = () => useContext(ChartAIContext)
// 创建 Provider 组件
export const ChartAIProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const { service = null, endpoint = null } = useParams()
  const { startTime, endTime } = useSelector(selectSecondsTimeRange)

  const memoizedValue = useMemo(
    () => ({
      service,
      endpoint,
      anormalTypes: ['app', 'container', 'infra', 'network', 'error', 'appInstance'],
    }),
    [service, endpoint],
  )
  return <ChartAIContext.Provider value={memoizedValue}>{children}</ChartAIContext.Provider>
}
