import { Button, Card, Popconfirm, Tree } from 'antd'
import React, { useEffect, useState } from 'react'
import { useLogsContext } from 'src/core/contexts/LogsContext'
import { LuDatabase, LuServer } from 'react-icons/lu'
import { ImTable2 } from 'react-icons/im'
import { MdAdd, MdDeleteOutline, MdModeEdit } from 'react-icons/md'
import { deleteLogOtherTableApi, deleteLogRuleApi } from 'core/api/logs'
import { showToast } from 'src/core/utils/toast'
import ConfigTableModal from '../../ConfigTableModal'
import { useTranslation } from 'react-i18next' // 引入i18n

const DataSourceTree = () => {
  const { t } = useTranslation('oss/fullLogs') // 使用i18n
  const { instances, tableInfo, updateTableInfo, getLogTableInfo, updateLoading } = useLogsContext()
  const [treeData, setTreeData] = useState([])
  const [expandedKeys, setExpandedKeys] = useState([])
  const [selectedKeys, setSelectedKeys] = useState([])
  const [modalVisible, setModalVisible] = useState(false)
  const deleteLogRule = (table) => {
    updateLoading(true)
    deleteLogOtherTableApi({
      dataBase: table.dataBase,
      instance: table.instanceName,
      tableName: table.tableName,
    }).then((res) => {
      showToast({
        title: t('fullLogSider.dataSourceTree.deleteLogSuccessToast'),
        color: 'success',
      })
      getLogTableInfo()
    })
  }
  // level:
  // 0: instance:{dataBases:[],instanceName}
  // 1: dataBase:{tables:[],dataBase}
  // 2: table:{cluster,tableName,timeField}
  const titleRender = (nodeData) => {
    return (
      <div className="logRuleItem">
        <div className="flex flex-col">
          <div className="flex flex-row">
            <div>{nodeData.icon}</div>
            <div>{nodeData.title}</div>
          </div>
        </div>
        {nodeData.timeField && (
          <div className="action">
            <Popconfirm
              title={
                <>
                  {t('fullLogSider.dataSourceTree.confirmDeleteLogRulePart1Text')}
                  <span className="font-bold ">{nodeData.tableName}</span>
                  {t('fullLogSider.dataSourceTree.confirmDeleteLogRulePart2Text')}
                </>
              }
              onConfirm={(e) => {
                e.stopPropagation()
                deleteLogRule(nodeData)
              }}
              okText={t('fullLogSider.dataSourceTree.confirmText')}
              cancelText={t('fullLogSider.dataSourceTree.cancelText')}
            >
              <Button
                color="danger"
                variant="filled"
                size="small"
                icon={<MdDeleteOutline />}
                onClick={(e) => {
                  e.stopPropagation()
                }}
              ></Button>
            </Popconfirm>
          </div>
        )}
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
        icon: <LuServer className="inline-flex items-center mr-1" />,
        title: instance.instanceName,
        children: instance.dataBases?.map((dataBase) => {
          const dataBaseKey = 'instance-' + instance.instanceName + '-dataBase-' + dataBase.dataBase
          expandedKeys.push(dataBaseKey) // 收集dataBase的key
          return {
            key: dataBaseKey,
            title: dataBase.dataBase,
            icon: <LuDatabase className="inline-flex items-center mr-1" />,
            children: dataBase.tables?.map((table) => {
              const tableKey = instance.instanceName + dataBase.dataBase + table.tableName
              return {
                key: tableKey,
                title: table.tableName,
                icon: <ImTable2 className="inline-flex items-center mr-1" />,
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

  const onSelect = (selectedKeys, { selected, selectedNodes }) => {
    if (selected && selectedNodes?.[0]?.tableName) {
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
      title={t('fullLogSider.dataSourceTree.ExternalRepositoryText')}
      classNames={{
        body: 'p-0',
      }}
      style={{ display: 'flex', flexDirection: 'column' }} // 设置 Card 的高度，使用 flexbox
      bodyStyle={{ flexGrow: 1, overflow: 'auto' }}
      extra={
        <Button
          type="primary"
          size="small"
          icon={<MdAdd size={20} />}
          onClick={() => setModalVisible(true)}
          className="flex-grow-0 flex-shrink-0"
        >
          {/* <span className="text-xs"></span> */}
        </Button>
      }
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
        titleRender={titleRender}
        blockNode
      />
      <ConfigTableModal modalVisible={modalVisible} closeModal={() => setModalVisible(false)} />
    </Card>
  )
}

export default DataSourceTree
