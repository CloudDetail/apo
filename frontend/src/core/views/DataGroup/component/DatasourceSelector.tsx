/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Card, Tree } from 'antd'
import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import DatasourceTag from './DatasourceTag'
import DatasourceIcon from './DatasourceIcon'
import styles from './index.module.scss'
import { getCheckableDatasourceApi } from 'src/core/api/dataGroup'
import { DatasourceTypes } from 'src/core/types/dataGroup'
interface DataType {
  datasource: React.Key
  nested?: string[]
  category?: string
  type: string
}
type DatasourceGroup = {
  id: string
  name: string
  children?: DatasourceGroup[]
  cluster?: string
  namespace?: string
  // ...其他字段
}

type DatasourceKey = {
  id: string
}

function processTreeAndCollectDatasource(
  tree: DatasourceGroup[],
  datasourceKeys: DatasourceKey[],
): { tree: DatasourceGroup[]; datasourceList: any[] } {
  const datasourceList: any[] = []

  function process(nodes: DatasourceGroup[], path: string[] = []): DatasourceGroup[] {
    console.log(nodes)
    return nodes?.map((node) => {
      let newNode: DatasourceGroup = { ...node }

      newNode.disableCheckbox = !newNode.hasCheckBox
      newNode.path = path
      if (node.children && node.children.length > 0) {
        // 第一级别是 system，不进入路径
        newNode.children = process(node.children, [...path, node.name])
      }
      const match = datasourceKeys?.find((d) => d === node.id)
      if (match) {
        datasourceList.push(newNode)
      }
      return newNode
    })
  }

  const newTree = process(tree)
  console.log(newTree, datasourceList)

  return { tree: newTree, datasourceList }
}

const DatasourceSelector = (props) => {
  const { t } = useTranslation('core/dataGroup')
  const { id, onChange, groupId, datasources, isAdd } = props
  const [dataSourceTree, setDataSourceTree] = useState([])
  const [checkedTreeKeys, setCheckedTreeKeys] = useState<React.Key[]>([])
  const [datasourceList, setDatasourceList] = useState<DataType[]>([])
  const [expandedKeys, setExpandedKeys] = useState<React.Key[]>([])
  const onCheck = (checkedKeys, { checkedNodes }) => {
    setCheckedTreeKeys(checkedKeys.checked)
    setDatasourceList(checkedNodes)
    onChange(checkedKeys.checked)
  }

  const getCheckableDatasource = () => {
    getCheckableDatasourceApi(groupId, isAdd).then((res) => {
      console.log(res)
      const { tree: newTree, datasourceList } = processTreeAndCollectDatasource(
        [res.view],
        res.datasources,
      )
      setDataSourceTree(newTree)
      setDatasourceList(isAdd ? [] : datasourceList)
      setCheckedTreeKeys(isAdd ? [] : res.datasources)
      onChange(isAdd ? [] : res.datasources)
    })
  }
  useEffect(() => {
    getCheckableDatasource()
  }, [groupId])
  const getAllKeys = (nodes) => {
    const keys = []
    const traverse = (nodeList) => {
      nodeList.forEach((node) => {
        keys.push(node.id)
        if (node.children && node.children.length > 0) {
          traverse(node.children)
        }
      })
    }
    traverse(nodes)
    return keys
  }

  useEffect(() => {
    if (dataSourceTree && dataSourceTree.length > 0) {
      const allKeys = getAllKeys(dataSourceTree)
      setExpandedKeys(allKeys)
    }
  }, [dataSourceTree])

  const deleteDatasource = (id: string) => {
    const result = datasourceList.filter((item) => item.id !== id)
    const resultKeys = checkedTreeKeys.filter((item) => item !== id)
    console.log(checkedTreeKeys.filter((item) => item !== id))
    setDatasourceList(result)
    onChange(resultKeys)
    setCheckedTreeKeys(resultKeys)
  }

  return (
    <div style={{ maxHeight: '60vh' }} className="flex w-full" id={id}>
      <Card type="inner" title="可选数据源" className="w-1/2 overflow-hidden" size="small">
        <Tree
          checkable
          onCheck={onCheck}
          checkedKeys={checkedTreeKeys}
          expandedKeys={expandedKeys}
          treeData={dataSourceTree}
          selectable={false}
          icon={({ type }) => <DatasourceIcon type={type} />}
          showIcon
          blockNode
          className={styles.datasource}
          checkStrictly={true}
          fieldNames={{
            key: 'id',
            title: 'name',
          }}
        />
      </Card>
      <Card type="inner" title="已选数据源" className="w-1/2" size="small">
        {DatasourceTypes.map((type) => {
          const result = datasourceList.filter((item) => item.type === type)
          return (
            result.length > 0 && (
              <>
                <div className="font-bold flex items-center">
                  <DatasourceIcon type={type} /> <span className="ml-2">{type}</span>
                  <span className="text-xs text-[var(--ant-color-text-secondary)] ml-2">
                    ({result.length})
                  </span>
                </div>
                {result.map((item) => (
                  <DatasourceTag {...item} closable onRemoveSelection={deleteDatasource} />
                ))}
              </>
            )
          )
        })}
      </Card>
    </div>
  )
}
export default DatasourceSelector
