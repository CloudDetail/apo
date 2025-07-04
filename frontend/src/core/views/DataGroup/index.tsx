/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { useEffect, useState } from 'react'
import { BasicCard } from 'src/core/components/Card/BasicCard'
import DataGroupTree from './DataGroupTree'
import DataGroupTable from './DataGroupTable'
import InfoModal from './InfoModal'
import PermissionModal from './PermissionModal'
import { deleteDataGroupApiV2, getDatasourceByGroupApiV2 } from 'src/core/api/dataGroup'
import { notify } from 'src/core/utils/notify'
import { useTranslation } from 'react-i18next'
import { Splitter } from 'antd'

export default function DataGroupPage() {
  const { t: ct } = useTranslation('common')
  const { t } = useTranslation('core/dataGroup')
  const [dataGroups, setDataGroups] = useState([])
  const [groupInfo, setGroupInfo] = useState(null)
  const [parentGroupInfo, setParentGroupInfo] = useState(null)
  const [infoModalVisible, setInfoModalVisible] = useState(false)
  const [permissionModalVisible, setPermissionModalVisible] = useState(false)
  const [infoGroupId, setInfoGroupId] = useState(null)
  const [refreshKey, setRefreshKey] = useState(0)
  const closeInfoModal = () => {
    setInfoModalVisible(false)
    setGroupInfo(null)
  }
  const closePermissionModal = () => {
    setPermissionModalVisible(false)
    setGroupInfo(null)
  }
  const refresh = () => {
    closeInfoModal()
    closePermissionModal()
    getDataGroups()
    setRefreshKey(refreshKey + 1)
  }
  const getDataGroups = () => {
    getDatasourceByGroupApiV2().then((res) => {
      setDataGroups([res])
      //   let found = false
      //   const current = parentGroupInfo
      //   if (current && current.groupId != null) {
      //     const findNode = (nodes) => {
      //       for (const node of nodes) {
      //         if (node.groupId === current.groupId) return true
      //         if (node.subGroups && findNode(node.subGroups)) return true
      //       }
      //       return false
      //     }
      //     found = findNode([res])
      //   }
      //   if (!found) {
      //     const first = res
      //     if (first) setParentGroupInfo(first)
      //   }
    })
    // setDataGroups([
    //   {
    //     groupId: 'aaaa',
    //     groupName: '全部',
    //     description: '全部数据',
    //     permissionType: 'known',
    //     datasources: [
    //       { id: 'aaaa', name: 'dev-6', type: 'cluster', isChecked: false },
    //       { id: 'bbbb', name: 'request-demo', type: 'service', isChecked: false },
    //     ],
    //     subGroups: [
    //       {
    //         groupId: 'bbbb',
    //         groupName: '营销',
    //         description: '营销部门',
    //         permissionType: 'view',
    //         datasources: [
    //           { id: 'aaaa', name: 'dev-6', type: 'cluster', isChecked: false },
    //           { id: 'bbbb', name: 'request-demo', type: 'service', isChecked: false },
    //         ],
    //         subGroups: [
    //           {
    //             groupId: 'cccc',
    //             groupName: '推荐',
    //             description: '推荐服务',
    //             permissionType: 'view',
    //             datasources: [
    //               { id: 'aaaa', name: 'dev-6', type: 'cluster', isChecked: false },
    //               { id: 'bbbb', name: 'request-demo', type: 'service', isChecked: false },
    //             ],
    //           },
    //         ],
    //       },
    //       {
    //         groupId: 'dddd',
    //         groupName: '安全',
    //         description: '安全部门',
    //         permissionType: 'edit',
    //         datasources: [
    //           { id: 'aaaa', name: 'dev-6', type: 'cluster', isChecked: false },
    //           { id: 'bbbb', name: 'request-demo', type: 'service', isChecked: false },
    //         ],
    //       },
    //     ],
    //   },
    // ])
  }
  useEffect(() => {
    getDataGroups()
  }, [])
  const deleteDataGroup = (groupInfo) => {
    if (groupInfo.subGroups && groupInfo.subGroups.length > 0) {
      notify({
        type: 'error',
        message: t('deleteGroupError'),
      })
      return
    } else {
      deleteDataGroupApiV2(groupInfo.groupId)
        .then((_res) => {
          notify({
            type: 'success',
            message: ct('deleteSuccess'),
          })
          // getDataGroups()
        })
        .finally(() => {
          getDataGroups()
        })
    }
  }
  return (
    <>
      <BasicCard>
        <div className=" w-full h-full">
          <Splitter style={{ height: '100%', boxShadow: '0 0 10px rgba(0, 0, 0, 0.1)' }}>
            <Splitter.Panel defaultSize="25%" className="h-full overflow-hidden w-full">
              <DataGroupTree
                dataGroups={dataGroups}
                setParentGroupInfo={(data) => {
                  setParentGroupInfo(data)
                }}
                openAddModal={(groupId) => {
                  setInfoGroupId(groupId)
                  setInfoModalVisible(true)
                }}
                openEditModal={(record) => {
                  setGroupInfo(record)
                  setInfoModalVisible(true)
                }}
                openPermissionModal={(record) => {
                  setGroupInfo(record)
                  setPermissionModalVisible(true)
                }}
                deleteDataGroup={deleteDataGroup}
              />
            </Splitter.Panel>
            <Splitter.Panel className="ml-2">
              <DataGroupTable
                key={refreshKey}
                parentGroupInfo={parentGroupInfo}
                openAddModal={() => {
                  setInfoGroupId(parentGroupInfo?.groupId)
                  setInfoModalVisible(true)
                }}
                openEditModal={(record) => {
                  setGroupInfo(record)
                  setInfoModalVisible(true)
                }}
                openPermissionModal={(record) => {
                  setGroupInfo(record)
                  setPermissionModalVisible(true)
                }}
                deleteDataGroup={deleteDataGroup}
              />
            </Splitter.Panel>
          </Splitter>
        </div>
      </BasicCard>
      <InfoModal
        open={infoModalVisible}
        closeModal={closeInfoModal}
        groupInfo={groupInfo}
        refresh={refresh}
        groupId={infoGroupId}
      />
      <PermissionModal
        open={permissionModalVisible}
        closeModal={closePermissionModal}
        groupInfo={groupInfo}
        refresh={refresh}
      />
    </>
  )
}
