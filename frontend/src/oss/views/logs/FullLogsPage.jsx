import React from 'react'
import { LogsProvider } from 'src/core/contexts/LogsContext'
import FullLogs from './component/FullLogs'

export default function FullLogsPage() {
  return (
    <LogsProvider>
      <FullLogs />
    </LogsProvider>
  )
}
