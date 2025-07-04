/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { createContext, useContext, ReactNode, useState } from 'react'

interface LogsTraceFilterContextType {
  clusterIds: string[]
  services: string[]
  instance: string | null
  traceId: string | null
  namespaces: string[]
  startTime: number | null
  endTime: number | null
  minDuration: number | null
  maxDuration: number | null
  inputEndpoint: string | null
  endpoint: string | null
  faultTypeList: string[]
  setClusterIds: (ids: string[]) => void
  setServices: (services: string[]) => void
  setInstance: (instance: object[]) => void
  setTraceId: (traceId: string | null) => void
  setNamespaces: (namespaces: string[]) => void
  setStartTime: (startTime: number | null) => void
  setEndTime: (endTime: number | null) => void
  setMinDuration: (minDuration: number | null) => void
  setMaxDuration: (maxDuration: number | null) => void
  setInputEndpoint: (inputEndpoint: string | null) => void
  setEndpoint: (endpoint: string | null) => void
  setFaultTypeList: (faultTypeList: string[]) => void
  instanceOption: any[]
  setInstanceOption: (instanceOption: any[]) => void
  isFilterDone: boolean
  setIsFilterDone: (isFilterDone: boolean) => void
}

const LogsTraceFilterContext = createContext<LogsTraceFilterContextType>(
  {} as LogsTraceFilterContextType,
)

export const useLogsTraceFilterContext = <T,>(
  selector: (context: LogsTraceFilterContextType) => T,
): T => {
  const context = useContext(LogsTraceFilterContext)
  return selector(context)
}

export const LogsTraceFilterProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const [clusterIds, setClusterIds] = useState([])
  const [services, setServices] = useState([])
  const [instance, setInstance] = useState(null)
  const [traceId, setTraceId] = useState(null)
  const [namespaces, setNamespaces] = useState([])
  const [startTime, setStartTime] = useState(null)
  const [endTime, setEndTime] = useState(null)
  const [instanceOption, setInstanceOption] = useState([])
  const [isFilterDone, setIsFilterDone] = useState(false)
  const [minDuration, setMinDuration] = useState(null)
  const [maxDuration, setMaxDuration] = useState(null)
  const [inputEndpoint, setInputEndpoint] = useState(null)
  const [endpoint, setEndpoint] = useState(null)
  const [faultTypeList, setFaultTypeList] = useState([])
  const finalValue = {
    clusterIds,
    services,
    instance,
    traceId,
    namespaces,
    startTime,
    endTime,
    minDuration,
    maxDuration,
    inputEndpoint,
    endpoint,
    setClusterIds,
    setServices,
    setInstance,
    setTraceId,
    setNamespaces,
    setStartTime,
    setEndTime,
    setMinDuration,
    setMaxDuration,
    setInputEndpoint,
    setEndpoint,
    instanceOption,
    setInstanceOption,
    isFilterDone,
    setIsFilterDone,
    faultTypeList,
    setFaultTypeList,
  }

  return (
    <LogsTraceFilterContext.Provider value={finalValue}>{children}</LogsTraceFilterContext.Provider>
  )
}
