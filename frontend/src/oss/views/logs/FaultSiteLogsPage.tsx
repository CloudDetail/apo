/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */
import { LogsTraceFilterProvider } from 'src/oss/contexts/LogsTraceFilterContext'
import FaultSiteLogs from './FaultSiteLogs'

function FaultSiteLogsPage() {
  return (
    <LogsTraceFilterProvider>
      <FaultSiteLogs />
    </LogsTraceFilterProvider>
  )
}

export default FaultSiteLogsPage
