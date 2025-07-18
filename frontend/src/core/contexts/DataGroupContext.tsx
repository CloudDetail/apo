/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { createContext, useContextSelector } from '@fluentui/react-context-selector'
import { ReactNode, useMemo, useState } from 'react'
import { useUserContext } from './UserContext'
import { getDatasourceByGroupApiV2 } from '../api/dataGroup'
import { DataGroupItem } from '../types/dataGroup'

interface DataGroupContextType {}
const DataGroupContext = createContext<DataGroupContextType>({} as DataGroupContextType)

export const useDataGroupContext = <T,>(selector: (context: any) => T): T =>
  useContextSelector(DataGroupContext, selector)
export const DataGroupProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const { user } = useUserContext()
  const [originalDataGroup, setOriginalDataGroup] = useState<DataGroupItem[]>([])
  const getDataGroup = async () => {
    try {
      const res = await getDatasourceByGroupApiV2()
      const data = Array.isArray(res) ? res : res ? [res] : []
      setOriginalDataGroup((prev) => {
        return JSON.stringify(prev) === JSON.stringify(data) ? prev : data
      })
    } catch (error) {
      console.error(error)
    }
  }
  const { flattenedAvailableNodes, availableNodeIds, allNodeIds, dataGroup } = useMemo(() => {
    const availableNodes: any[] = []
    const allIds: number[] = []
    const processTree = (nodes: any[]): any[] => {
      return nodes.map((node) => {
        const disabled = node.permissionType === 'known'
        allIds.push(Number(node.groupId))
        if (!disabled) {
          availableNodes.push(node)
        }
        let children: any[] = []
        if (node.subGroups?.length) {
          children = processTree(node.subGroups)
        }
        return {
          ...node,
          disabled,
          subGroups: children,
        }
      })
    }

    const dataGroup = processTree(originalDataGroup)

    return {
      flattenedAvailableNodes: availableNodes,
      availableNodeIds: new Set(availableNodes.map((node) => Number(node.groupId))),
      allNodeIds: allIds,
      dataGroup,
    }
  }, [originalDataGroup])

  const finalValue = {
    dataGroup,
    flattenedAvailableNodes,
    availableNodeIds,
    allNodeIds,
    getDataGroup,
  }
  return <DataGroupContext.Provider value={finalValue}>{children}</DataGroupContext.Provider>
}
