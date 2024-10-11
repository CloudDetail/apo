import { Card, Collapse } from 'antd'
import React from 'react'
import IndexCollapse from './component/IndexCollapse'
import './index.css'
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
    },
  ]
  return (
    <>
      <div className="indexList h-full bg-[#141414] ">
        <Collapse items={items} defaultActiveKey={['base']} size="small" />
      </div>
    </>
  )
}
export default IndexList
