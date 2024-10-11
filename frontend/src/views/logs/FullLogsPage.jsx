import React from 'react'
import { LogsProvider } from 'src/contexts/LogsContext'
import FullLogs from './component/FullLogs'

export default function FullLogsPage() {
  return (
    <LogsProvider>
      <FullLogs />
    </LogsProvider>
  )
}
