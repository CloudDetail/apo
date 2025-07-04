/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { createContext, useContextSelector } from '@fluentui/react-context-selector'
import { ReactNode, useEffect, useState } from 'react'
import { useUserContext } from './UserContext'
import { getUserGroupApi } from '../api/dataGroup'
import { DataGroupItem } from '../types/dataGroup'

interface DataGroupContextType {}
const DataGroupContext = createContext<DataGroupContextType>({} as DataGroupContextType)

export const useDataGroupContext = <T,>(selector: (context: any) => T): T =>
  useContextSelector(DataGroupContext, selector)
export const DataGroupProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const { user } = useUserContext()
  const [dataGroup, setDataGroup] = useState<DataGroupItem[]>([])
  const getDataGroup = async () => {
    try {
      const res = await getUserGroupApi(user.userId, 'apm')
      setDataGroup(res)
      // setDataGroup([
      //   {
      //     groupId: '692489354112',
      //     groupName: '测试数据组1',
      //     description: '',
      //     subGroups: [
      //       {
      //         groupId: '11',
      //         groupName: '测试数据组1-1',
      //         description: '',
      //         subGroups: [
      //           {
      //             groupId: '111',
      //             groupName: '测试数据组1-1-1',
      //             description: '',
      //             subGroups: null,
      //           },
      //         ],
      //       },
      //     ],
      //   },
      //   {
      //     groupId: '770368178368',
      //     groupName: 'train-ticket',
      //     description: '',
      //     subGroups: [
      //       {
      //         groupId: '21',
      //         groupName: '测试数据组2-1',
      //         description: '',
      //         disabled: true,
      //         subGroups: [
      //           {
      //             groupId: '211',
      //             groupName: '测试数据组2-1-1',
      //             description: '',
      //             subGroups: null,
      //           },
      //         ],
      //       },
      //     ],
      //   },
      // ])
      console.log(res)
    } catch (error) {
      console.error(error)
    }
  }
  useEffect(() => {
    if (user.userId) {
      getDataGroup()
    }
  }, [user.userId])
  const finalValue = {
    dataGroup,
  }
  return <DataGroupContext.Provider value={finalValue}>{children}</DataGroupContext.Provider>
}
