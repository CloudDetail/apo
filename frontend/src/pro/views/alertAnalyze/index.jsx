import React from 'react'
import { AlertAnalyzeProvider } from 'src/pro/contexts/AlertAnalyzeContext'
import AlertAnalyzeCom from './components'

function AlertAnalyzePage() {
  return (
    <AlertAnalyzeProvider>
      <AlertAnalyzeCom />
    </AlertAnalyzeProvider>
  )
}
export default AlertAnalyzePage
