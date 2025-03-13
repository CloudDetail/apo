/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { createContext, useContextSelector } from '@fluentui/react-context-selector'
import { ReactNode, useEffect, useMemo, useState } from 'react'

interface ChartsContextType {}
const ChartsContext = createContext<ChartsContextType>({} as ChartsContextType)

export const useChartsContext = <T,>(selector: (context: any) => T): T =>
  useContextSelector(ChartsContext, selector)
export const ChartsProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const [chartsData, setChartsData] = useState(null)
  const [chartsLoading, setChartsLoading] = useState(true)

  const finalValue = {
    chartsData,
    setChartsData,
    chartsLoading,
    setChartsLoading,
  }
  return <ChartsContext.Provider value={finalValue}>{children}</ChartsContext.Provider>
}
