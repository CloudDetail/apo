import React, { useEffect, useRef } from 'react'
import IframeDashboard from 'src/core/components/Dashboard/IframeDashboard'

function SystemDashboard() {
  return (
    <div className="text-xs" style={{ height: 'calc(100vh - 120px)' }}>
      <IframeDashboard src={'grafana/d/k8s_views_global/e99b86-e7bea4-e680bb-e8a788?orgId=1'} />
    </div>
  )
}
export default SystemDashboard
