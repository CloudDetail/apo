import { Card, Tree } from 'antd'
import React, { useEffect, useState } from 'react'
import { useLogsContext } from 'src/contexts/LogsContext'
import { LuDatabase, LuServer } from 'react-icons/lu'
import { ImTable2 } from 'react-icons/im'
const DataSourceTree = () => {
  const { instances, tableInfo, updateTableInfo } = useLogsContext()
  const [treeData, setTreeData] = useState([])
  const [expandedKeys, setExpandedKeys] = useState([])
  const [selectedKeys, setSelectedKeys] = useState([])
  // level:
  // 0: instance:{dataBases:[],instanceName}
  // 1: dataBase:{tables:[],dataBase}
  // 2: table:{cluster,tableName,timeField}
  const treeTitle = (title, icon) => {
    return (
      <div className="flex flex-row">
        <div>{icon}</div>
        <div>{title}</div>
      </div>
    )
  }
  useEffect(() => {
    const expandedKeys = []
    const newTreeData = instances?.map((instance) => {
      const instanceKey = 'instance-' + instance.instanceName
      expandedKeys.push(instanceKey) // 收集instance的key
      return {
        key: instanceKey,
        title: treeTitle(
          instance.instanceName,
          <LuServer className="inline-flex items-center mr-1" />,
        ),
        children: instance.dataBases?.map((dataBase) => {
          const dataBaseKey = 'instance-' + instance.instanceName + '-dataBase-' + dataBase.dataBase
          expandedKeys.push(dataBaseKey) // 收集dataBase的key
          return {
            key: dataBaseKey,
            title: treeTitle(
              dataBase.dataBase,
              <LuDatabase className="inline-flex items-center mr-1" />,
            ),
            children: dataBase.tables?.map((table) => {
              const tableKey = instance.instanceName + dataBase.dataBase + table.tableName
              return {
                key: tableKey,
                title: treeTitle(
                  table.tableName,
                  <ImTable2 className="inline-flex items-center mr-1" />,
                ),
                dataBase: dataBase.dataBase,
                instanceName: instance.instanceName,
                ...table,
              }
            }),
          }
        }),
      }
    })

    setTreeData(newTreeData)
    setExpandedKeys(expandedKeys) // 更新defaultExpandedKeys
  }, [instances])

  const onSelect = (selectedKeys, { selectedNodes }) => {
    if (selectedNodes[0].tableName) {
      updateTableInfo({
        dataBase: selectedNodes[0].dataBase,
        tableName: selectedNodes[0].tableName,
        cluster: selectedNodes[0].cluster,
        timeField: selectedNodes[0].timeField,
        instanceName: selectedNodes[0].instanceName,
      })
    }
  }
  const onExpand = (expandedKeys) => {
    setExpandedKeys(expandedKeys)
  }
  useEffect(() => {
    setSelectedKeys([tableInfo.instanceName + tableInfo.dataBase + tableInfo.tableName])
  }, [tableInfo])
  return (
    <Card
      className="overflow-y-auto h-1/2 w-full overflow-x-hidden"
      title="接入数据库列表"
      classNames={{
        body: 'p-0',
      }}
      style={{ display: 'flex', flexDirection: 'column' }} // 设置 Card 的高度，使用 flexbox
      bodyStyle={{ flexGrow: 1, overflow: 'auto' }}
    >
      <Tree
        selectedKeys={selectedKeys}
        expandedKeys={expandedKeys}
        onSelect={onSelect}
        onExpand={onExpand}
        // onCheck={onCheck}
        treeData={treeData}
        // showIcon
        className="pr-2 h-full"
      />
    </Card>
  )
}

export default DataSourceTree
