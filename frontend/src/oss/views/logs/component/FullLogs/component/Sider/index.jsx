/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react'
import LogRuleList from './LogRuleList'
import DataSourceTree from './DataSourceTree'

const FullLogSider = () => {
  return (
    <div className="flex flex-col h-full">
      <div className="flex-1">
        <LogRuleList />
      </div>
      <div className="flex-1">
        <DataSourceTree />
      </div>
    </div>
  )
}

export default FullLogSider
