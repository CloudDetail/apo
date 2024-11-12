import React from 'react'
import { AlertAnalyzeProvider } from 'src/pro/contexts/AlertAnalyzeContext'
import DescendantTopology from './DescendantTopology'

function ServiceAnalyzePage() {
  return (
    <AlertAnalyzeProvider>
      <DescendantTopology />
    </AlertAnalyzeProvider>
  )
}
export default ServiceAnalyzePage
