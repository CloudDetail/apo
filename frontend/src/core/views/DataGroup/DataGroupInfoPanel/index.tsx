/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */
import React, { useEffect, useState } from 'react'
import DataGroupTable from './DataGroupTable'
import { getSubGroupsApiV2 } from 'src/core/api/dataGroup'
import DataGroupInfo from './DataGroupInfo'
import { Splitter } from 'antd'
import { DataGroupPermissionInfo } from 'src/core/types/dataGroup'

const DataGroupInfoPanel = ({
  info,
  refreshKey,
  openAddModal,
  openEditModal,
  deleteDataGroup,
  openPermissionModal,
}: {
  info: any
  refreshKey: number
  openAddModal: () => void
  openEditModal: (info: DataGroupPermissionInfo) => void
  openPermissionModal: (info: DataGroupPermissionInfo) => void
  deleteDataGroup: (info: DataGroupPermissionInfo) => void
}) => {
  const [subGroups, setSubGroups] = useState<any[]>([])
  const [datasources, setDatasources] = useState<any[]>([])
  const [alertEventTableHeight, setAlertEventTableHeight] = useState('calc(100vh - 340px)')
  useEffect(() => {
    if (info?.groupId != null) {
      getSubGroupsApiV2(info.groupId).then((res: any) => {
        setSubGroups(res?.subGroups || [])
        setDatasources(res?.datasources || [])
      })
    }
  }, [info, refreshKey])
  const onResize = (sizes) => {
    setAlertEventTableHeight(sizes[1] - 100)
  }
  return (
    <Splitter layout="vertical" onResize={onResize}>
      <Splitter.Panel defaultSize={150}>
        <DataGroupInfo
          info={info}
          datasources={datasources}
          openAddModal={openAddModal}
          handlePermission={() => openPermissionModal(info)}
          openEditModal={openEditModal}
          deleteDataGroup={deleteDataGroup}
        />
      </Splitter.Panel>
      <Splitter.Panel>
        <DataGroupTable
          scrolllHeight={alertEventTableHeight}
          subGroups={subGroups}
          openEditModal={openEditModal}
          deleteDataGroup={deleteDataGroup}
          openPermissionModal={openPermissionModal}
        />
      </Splitter.Panel>
    </Splitter>
  )
}

export default DataGroupInfoPanel
