/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Alert, Card, Segmented, Select, Table, TableProps, Tag, Tree } from 'antd'
import { useEffect, useState } from 'react'
import { AiOutlineInfoCircle } from 'react-icons/ai'
import { getAllDatasourceApi } from 'src/core/api/dataGroup'
import styles from './index.module.scss'
import { DatasourceType } from 'src/core/types/dataGroup'
import DatasourceTag from './component/DatasourceTag'
interface DataType {
  datasource: React.Key
  nested?: string[]
  category?: string
  type: string
}
const viewTypeList: { label: string; value: DatasourceType }[] = [
  { label: '命名空间视图', value: 'namespace' },
  { label: '服务名视图', value: 'service' },
]
const columns = [
  {
    dataIndex: 'datasource',
    title: '服务名',
    // filteredValue: ['营销前置'],
    // filterSearch: true,
    // onFilter: (value, record) =>
    //   record['datasource']
    //     .toString()
    //     .toLowerCase()
    //     .includes((value as string).toLowerCase()),
  },
  {
    dataIndex: 'nested',
    title: '所属命名空间',
    render: (value: string[]) => {
      //           <Tag color="geekblue">geekblue</Tag>
      //   <Tag color="cyan">cyan</Tag>
      return value.map((item) => <Tag color="geekblue">{item}</Tag>)
    },
  },
]
// const tagRender = (props) => {
//   console.log(props)
//   const { label, value, closable, onClose } = props
//   const onPreventMouseDown = (event: React.MouseEvent<HTMLSpanElement>) => {
//     event.preventDefault()
//     event.stopPropagation()
//   }
//   return (
//     <Tag
//       color={label === 'service' ? 'cyan' : 'geekblue'}
//       onMouseDown={onPreventMouseDown}
//       closable={closable}
//       onClose={onClose}
//       style={{ marginInlineEnd: 4 }}
//     >
//       {value}
//     </Tag>
//   )
// }
// const optionRender = (option) => {
//   const { data } = option
//   const onPreventMouseDown = (event: React.MouseEvent<HTMLSpanElement>) => {
//     event.preventDefault()
//     event.stopPropagation()
//   }
//   return (
//     <Tag
//       color={data.type === 'service' ? 'cyan' : 'geekblue'}
//       onMouseDown={onPreventMouseDown}
//       style={{ marginInlineEnd: 4 }}
//     >
//       {data.datasource}
//     </Tag>
//   )
// }
const DatasourceSelector = (props) => {
  const { id, datasourceList = [], onChange } = props
  const [viewType, setViewType] = useState<DatasourceType>('namespace')
  const [checkedTreeKeys, setCheckedTreeKeys] = useState<React.Key[]>([])
  const [namespaceTree, setNameSpaceTree] = useState([])
  const [serviceTableData, setServiceTableData] = useState<DataType[]>([])
  const [serviceSet, setServiceSet] = useState<Set<React.Key>>(new Set())
  const [checkedTableKeys, setCheckedTableKeys] = useState<React.Key[]>([])
  // const [mockValue, setMockValue] = useState([])
  const onCheck = (checkedKeys, { node }) => {
    let result = [...checkedTreeKeys]
    let datasourceListResult = [...datasourceList]
    if (node.checked) {
      result = result.filter((i: React.Key) => i !== node.key)
      datasourceListResult = datasourceList.filter(
        (item) => !(item.datasource === node.datasource && item.type === node.type),
      )
    } else {
      result.push(node.key)
      datasourceListResult.push({
        datasource: node.datasource,
        type: node.type,
        category: node.category,
      })
    }
    setCheckedTreeKeys(result)
    onChange(datasourceListResult)
  }

  const getAllDatasource = () => {
    getAllDatasourceApi().then((res) => {
      const datasourceList = res.namespaceList?.map((namespace) => ({
        title: namespace.datasource,
        key: namespace.datasource,
        ...namespace,
        children: namespace.nested?.map((item) => ({
          title: item,
          key: 'service-' + item,
          type: 'service',
          datasource: item,
        })),
      }))
      setNameSpaceTree(datasourceList)
      setServiceTableData(res.serviceList || [])
      const serviceSet: Set<React.Key> = new Set(
        (res.serviceList || []).map((item: DataType) => item.datasource as React.Key),
      )
      setServiceSet(serviceSet)
    })
  }
  useEffect(() => {
    getAllDatasource()
  }, [])
  const changeCheckedKeys = (datasourceList: DataType[]) => {
    if (viewType === 'namespace') {
      const checkedTreeKeys = datasourceList.map((item) => {
        if (item.type === 'service') {
          return 'service-' + item.datasource
        } else {
          return item.datasource
        }
      })
      setCheckedTreeKeys(checkedTreeKeys)
    } else {
      const result = datasourceList
        .filter((item) => item.type === 'service')
        .map((item) => item.datasource)
      setCheckedTableKeys(result)
    }
  }
  useEffect(() => {
    changeCheckedKeys(datasourceList)
  }, [viewType])

  const rowSelection = {
    onSelect: (record: DataType) => {
      onSelectTableRow(record)
    },
    onSelectAll: (selected, selectedRows, changeRows) => {
      const checkedKeys = selectedRows.map((item) => item.datasource)
      setCheckedTableKeys(checkedKeys)
      if (selected) {
        onChange([...datasourceList, ...changeRows])
      } else {
        onChange(datasourceList.filter((item) => item.type !== 'service'))
      }
    },
    selectedRowKeys: checkedTableKeys,
  }
  const onSelectTableRow = ({ datasource, category }: DataType, event?: any) => {
    if (event) {
      event.stopPropagation()
    }
    if (checkedTableKeys.includes(datasource)) {
      setCheckedTableKeys(checkedTableKeys.filter((i: React.Key) => i !== datasource))
      onChange(
        datasourceList.filter(
          (item) => !(item.datasource === datasource && item.type === 'service'),
        ),
      )
    } else {
      setCheckedTableKeys([...checkedTableKeys, datasource])
      onChange([
        ...datasourceList,
        {
          type: 'service',
          datasource: datasource,
          category: category,
        },
      ])
    }
  }
  const deleteDatasource = (e, datasource: string, type: DatasourceType) => {
    e.preventDefault()
    const result = datasourceList.filter(
      (item) => !(item.type === type && item.datasource === datasource),
    )
    onChange(result)
    changeCheckedKeys(result)
  }

  return (
    <div style={{ maxHeight: '60vh' }} className="flex flex-col" id={id}>
      <div className={styles.tagContainer}>
        {/* <Select
          mode="multiple"
          tagRender={tagRender}
          optionRender={optionRender}
          value={mockValue}
          style={{ width: '100%' }}
          labelInValue
          fieldNames={{ label: 'type', value: 'datasource' }}
          options={[...namespaceTree, ...serviceTableData]}
          onChange={setMockValue}
        /> */}
        {datasourceList.map((item) => (
          <DatasourceTag
            type={item.type}
            datasource={item.datasource}
            closable
            onClose={deleteDatasource}
          />
        ))}
      </div>
      <div className="flex-1 overflow-auto pl-2 py-0 bg-[#141414] flex flex-col">
        <div className="flex-shrink-0 flex-grow-0">
          <Segmented
            options={viewTypeList}
            className="m-2"
            value={viewType}
            onChange={setViewType}
          />
          <Alert
            // closable
            message={
              viewType === 'namespace' ? (
                <ul>
                  <li>
                    整体监控：选中命名空间节点，自动监控该命名空间下的所有服务，服务变化时自动更新。
                  </li>
                  <li>精细监控：可单独选择服务节点，仅监控该服务。</li>
                </ul>
              ) : (
                '仅监控所选服务，独立于命名空间。'
              )
            }
            type="info"
            showIcon
            className="text-xs pr-2"
          />
        </div>
        {viewType === 'namespace' ? (
          <Tree
            checkable
            onCheck={onCheck}
            checkedKeys={checkedTreeKeys}
            treeData={namespaceTree}
            selectable={false}
            blockNode
            className="overflow-auto"
            checkStrictly={true}
          />
        ) : (
          <Table<DataType>
            rowSelection={{ type: 'checkbox', ...rowSelection }}
            columns={columns}
            dataSource={serviceTableData}
            className="overflow-hidden pr-2"
            scroll={{ y: 310 }}
            pagination={false}
            size="small"
            rowKey="datasource"
            onRow={(record) => {
              return {
                onClick: (event) => {
                  onSelectTableRow(record, event)
                },
              }
            }}
          />
        )}
      </div>
    </div>
  )
}
export default DatasourceSelector
