import React from 'react'
import DescendantTopology from './DescendantTopology'
import { ServiceAnalyzeProvider } from 'src/pro/contexts/ServiceAnalyzeContext'

function ServiceAnalyzePage() {
  return (
    <ServiceAnalyzeProvider>
      <DescendantTopology />
    </ServiceAnalyzeProvider>
  )
}
export default ServiceAnalyzePage
