import React, { useEffect, useRef } from 'react'
import IframeDashboard from 'src/components/Dashboard/IframeDashboard'

function ApplicationDashboard() {
  return (
    <div className="text-xs" style={{height: 'calc(100vh - 120px)'}}>
      <IframeDashboard src={'grafana/d/b0102ebf-9e5e-4f21-80aa-9c2565cd3dcb/originx-polaris-metrics-service-level?orgId=1'}/>
    </div>
  )
}

export default ApplicationDashboard
