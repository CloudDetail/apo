/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */
import { LogsTraceFilterProvider } from 'src/oss/contexts/LogsTraceFilterContext'
import FaultSiteTrace from './FaultSiteTrace'

function FaultSiteTracePage() {
  return (
    <LogsTraceFilterProvider>
      <FaultSiteTrace />
    </LogsTraceFilterProvider>
  )
}

export default FaultSiteTracePage
