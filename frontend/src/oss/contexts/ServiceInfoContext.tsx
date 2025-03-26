/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { createContext, useContextSelector } from '@fluentui/react-context-selector'
import { ReactNode, useEffect, useState } from 'react'
type Status = 'unknown' | 'normal' | 'critical' | 'loading'
type PanelKey =
  | 'logs'
  | 'trace'
  | 'impact'
  | 'alert'
  | 'instance'
  | 'k8s'
  | 'error'
  | 'sql'
  | 'metrics'
  | 'polarisMetrics'

interface ServiceInfoContextType {}
const ServiceInfoContext = createContext<ServiceInfoContextType>({} as ServiceInfoContextType)

export const useServiceInfoContext = <T,>(selector: (context: any) => T): T =>
  useContextSelector(ServiceInfoContext, selector)
export const ServiceInfoProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const [statusPanels, setStatusPanels] = useState({
    instance: 'loading',
    k8s: 'loading',
    error: 'loading',
    alert: 'loading',
    impact: 'loading',
  })
  const [dashboardVariable, setDashboardVariable] = useState(null)
  const [activeTabKey, setActiveTabKey] = useState(['polarisMetrics'])

  const openTab = (key: PanelKey) => {
    setActiveTabKey((prev) => (prev.includes(key) ? prev : [...prev, key]))
  }

  const setPanelsStatus = (key: PanelKey, status: Status) => {
    setStatusPanels((pre) => ({
      ...pre,
      [key]: status,
    }))
  }
  useEffect(() => {}, [])
  const finalValue = {
    statusPanels,
    setPanelsStatus,
    dashboardVariable,
    setDashboardVariable,
    openTab,
    activeTabKey,
    setActiveTabKey,
  }
  return <ServiceInfoContext.Provider value={finalValue}>{children}</ServiceInfoContext.Provider>
}
