import { Card, Tree } from 'antd'
import React, { useEffect, useState } from 'react'
import { useLogsContext } from 'src/contexts/LogsContext'

const LogRuleList = () => {
  const { logRules, tableInfo, updateTableInfo } = useLogsContext()
  const [treeData, setTreeData] = useState([])

  const [selectedKeys, setSelectedKeys] = useState([])

  const menuLabel = (parseName, parseInfo) => {
    return (
      <div className="flex flex-col">
        <div>{parseName}</div>
        <div className="text-xs text-gray-400">{parseInfo}</div>
      </div>
    )
  }
  useEffect(() => {
    setTreeData(
      logRules.map((rule, index) => ({
        key: rule.dataBase + rule.tableName,
        title: menuLabel(rule.parseName, rule.parseInfo),
        ...rule,
      })),
    )
  }, [logRules])
  const onSelect = (selectedKeys, { selectedNodes }) => {
    updateTableInfo({
      dataBase: selectedNodes[0].dataBase,
      tableName: selectedNodes[0].tableName,
      cluster: '',
      parseName: selectedNodes[0].parseName,
    })
  }
  useEffect(() => {
    setSelectedKeys([tableInfo.dataBase + tableInfo.tableName])
  }, [tableInfo])
  return (
    <Card
      className="overflow-y-auto h-1/2 w-full overflow-x-hidden"
      title="日志规则列表"
      classNames={{
        body: 'p-0 pr-2',
      }}
    >
      <Tree
        selectedKeys={selectedKeys}
        onSelect={onSelect}
        // onCheck={onCheck}
        treeData={treeData}
        className="pr-3"
      />
    </Card>
  )
}

export default LogRuleList
