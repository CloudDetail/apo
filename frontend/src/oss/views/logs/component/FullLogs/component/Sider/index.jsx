/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

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
