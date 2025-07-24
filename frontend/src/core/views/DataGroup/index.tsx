/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useCallback, useEffect, useState } from 'react'
import { BasicCard } from 'src/core/components/Card/BasicCard'
import DataGroupTree from './DataGroupTree'
import InfoModal from './InfoModal'
import PermissionModal from './PermissionModal'
import { deleteDataGroupApiV2 } from 'src/core/api/dataGroup'
import { notify } from 'src/core/utils/notify'
import { useTranslation } from 'react-i18next'
import { Splitter } from 'antd'
import DataGroupInfoPanel from './DataGroupInfoPanel'
import { useDataGroupContext } from 'src/core/contexts/DataGroupContext'

interface DataGroupInfo {
  groupId: number
  groupName: string
  description: string
  permissionType: 'known' | 'view' | 'edit'

  subGroups?: DataGroupInfo[]
}

export default function DataGroupPage() {
  const { t: ct } = useTranslation('common')
  const { t } = useTranslation('core/dataGroup')
  const [groupInfo, setGroupInfo] = useState<DataGroupInfo | null>(null)
  const [parentGroupInfo, setParentGroupInfo] = useState<DataGroupInfo | null>(null)
  const [infoModalVisible, setInfoModalVisible] = useState<boolean>(false)
  const [permissionModalVisible, setPermissionModalVisible] = useState<boolean>(false)
  const [infoGroupId, setInfoGroupId] = useState<number | null>(null)
  const [refreshKey, setRefreshKey] = useState<number>(0)
  const dataGroup = useDataGroupContext((ctx) => ctx.dataGroup)
  const getDataGroup = useDataGroupContext((ctx) => ctx.getDataGroup)
  const closeInfoModal = useCallback(() => {
    setInfoModalVisible(false)
    setGroupInfo(null)
  }, [])

  const closePermissionModal = useCallback(() => {
    setPermissionModalVisible(false)
    setGroupInfo(null)
  }, [])

  const refresh = useCallback(() => {
    closeInfoModal()
    closePermissionModal()
    getDataGroup()
    setRefreshKey((prev) => prev + 1)
  }, [closeInfoModal, closePermissionModal, getDataGroup])

  useEffect(() => {
    getDataGroup()
  }, [])

  const deleteDataGroup = useCallback(
    (groupInfo: DataGroupInfo) => {
      if (groupInfo.subGroups && groupInfo.subGroups.length > 0) {
        notify({
          type: 'error',
          message: t('deleteGroupError'),
        })
        return
      } else {
        deleteDataGroupApiV2(groupInfo.groupId)
          .then(() => {
            notify({
              type: 'success',
              message: ct('deleteSuccess'),
            })
            // getDataGroup()
          })
          .finally(() => {
            getDataGroup()
          })
      }
    },
    [t, ct, getDataGroup],
  )

  const openAddModal = useCallback((groupId: number) => {
    setInfoGroupId(groupId)
    setInfoModalVisible(true)
  }, [])

  const openEditModal = useCallback((record: DataGroupInfo) => {
    setGroupInfo(record)
    setInfoModalVisible(true)
  }, [])

  const openPermissionModal = useCallback((record: DataGroupInfo) => {
    setGroupInfo(record)
    setPermissionModalVisible(true)
  }, [])

  const openAddModalForTable = useCallback(() => {
    setInfoGroupId(parentGroupInfo?.groupId)
    setInfoModalVisible(true)
  }, [parentGroupInfo])

  return (
    <>
      <BasicCard>
        <div className="w-full h-full">
          <Splitter style={{ height: '100%' }}>
            <Splitter.Panel defaultSize="25%" className="h-full overflow-hidden w-full">
              <DataGroupTree
                dataGroups={dataGroup}
                setParentGroupInfo={(data) => {
                  setParentGroupInfo(data)
                }}
                openAddModal={openAddModal}
                openEditModal={openEditModal}
                openPermissionModal={openPermissionModal}
                deleteDataGroup={deleteDataGroup}
              />
            </Splitter.Panel>
            <Splitter.Panel className="ml-2">
              <DataGroupInfoPanel
                info={parentGroupInfo}
                openAddModal={openAddModalForTable}
                openEditModal={openEditModal}
                openPermissionModal={openPermissionModal}
                deleteDataGroup={deleteDataGroup}
                refreshKey={refreshKey}
              />
            </Splitter.Panel>
          </Splitter>
        </div>
      </BasicCard>
      <InfoModal
        open={infoModalVisible}
        closeModal={closeInfoModal}
        groupInfo={groupInfo as any}
        refresh={refresh}
        groupId={infoGroupId}
      />
      <PermissionModal
        open={permissionModalVisible}
        closeModal={closePermissionModal}
        groupInfo={groupInfo as any}
        refresh={refresh}
      />
    </>
  )
}
