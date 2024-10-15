import React from 'react'
import LogRuleList from './LogRuleList'
import DataSourceTree from './DataSourceTree'
import Sider from 'antd/es/layout/Sider'
import { Card } from 'antd'

const FullLogSider = () => {
  return (
    <>
      <LogRuleList />
      <DataSourceTree />
    </>
  )
}

export default FullLogSider
