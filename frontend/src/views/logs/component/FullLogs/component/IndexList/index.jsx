import { Card, Collapse } from 'antd'
import React from 'react'
import IndexCollapse from './component/IndexCollapse'
import './index.css'
import FullTextSearch from '../SerarchBar/RawLogQuery/FullTextSearch'
const IndexList = () => {
  const IndexType = {
    base: '基础字段',
    log: '日志字段',
  }
  const items = [
    {
      key: 'base',
      label: '基础字段',
      children: <IndexCollapse type="base" />,
    },
    {
      key: 'log',
      label: '日志字段',
      children: <IndexCollapse type="log" />,

      style: {
        maxHeight: '50%',
        overflow: 'hidden',
      },
    },
  ]
  return (
    <div className="flex flex-col h-full">
      <div className="flex-shrink-0 py-2">
        <FullTextSearch />
      </div>
      <Collapse
        items={items}
        defaultActiveKey={['base', 'log']}
        size="small"
        className=" flex-1 indexList h-full overflow-hidden mb-2"
      />
    </div>
  )
}
export default IndexList
