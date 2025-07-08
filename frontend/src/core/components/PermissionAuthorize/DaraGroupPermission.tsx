/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { useEffect, useState } from 'react'
import { getDatasourceByGroupApiV2 } from 'src/core/api/dataGroup'
import { DataGroupItem } from 'src/core/types/dataGroup'
import { useTranslation } from 'react-i18next'
import Search from 'antd/es/transfer/search'
import { Tree } from 'antd'

interface DataGroupPermissionProps {
  id: string
  dataGroupList: any[]
  onChange: any
  type: 'team' | 'user'
  permissionSourceTeam: any[]
}

const DataGroupPermission = (props: DataGroupPermissionProps) => {
  const { t } = useTranslation('core/permission')
  const { id, dataGroupList = [], onChange, type, permissionSourceTeam } = props
  const [checkedKeys, setCheckedKeys] = useState([])
  const [treeData, setTreeData] = useState<DataGroupItem[]>([])
  const [expandedKeys, setExpandedKeys] = useState([])

  const getDataGroups = () => {
    getDatasourceByGroupApiV2().then((res) => {
      setTreeData([res])
      const allKeys = getAllKeys([res])
      setExpandedKeys(allKeys)
      setCheckedKeys(dataGroupList)
    })
    // getDataGroupsApi({
    //   currentPage: 1,
    //   pageSize: 1000,
    // }).then((res) => {
    //   setData(res.dataGroupList)
    // })
  }
  // 递归获取所有节点的key，用于展开所有节点
  const getAllKeys = (nodes) => {
    const keys = []
    const traverse = (nodeList) => {
      nodeList.forEach((node) => {
        keys.push(node.groupId)
        if (node.subGroups && node.subGroups.length > 0) {
          traverse(node.subGroups)
        }
      })
    }
    traverse(nodes)
    return keys
  }

  useEffect(() => {
    getDataGroups()
  }, [])

  const deleteDataGroup = (e, groupId: string) => {
    e.preventDefault()
    const result = dataGroupList.filter((item) => item.groupId !== groupId)
    onChange(result)
  }
  return (
    <div style={{ maxHeight: '40vh' }} className="flex flex-col overflow-auto w-full" id={id}>
      <Search
        style={{ marginBottom: 8 }}
        className="h-[42px]"
        placeholder="Search"
        onChange={onChange}
      />
      {treeData && treeData.length > 0 && (
        <Tree
          checkable
          selectable={false}
          checkedKeys={checkedKeys}
          expandedKeys={expandedKeys}
          onExpand={setExpandedKeys}
          onCheck={(checkedKeys, { node }) => {
            setCheckedKeys(checkedKeys)
            onChange(checkedKeys)
          }}
          treeData={treeData}
          style={{ width: '100%' }}
          className="pr-3 h-full w-full"
          fieldNames={{
            title: 'groupName',
            key: 'groupId',
            children: 'subGroups',
          }}
          blockNode
        />
      )}
    </div>
  )
}
export default DataGroupPermission
