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
