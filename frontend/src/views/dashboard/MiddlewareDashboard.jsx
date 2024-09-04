import React, { useEffect, useRef } from 'react'
import IframeDashboard from 'src/components/Dashboard/IframeDashboard'

function MiddlewareDashboard() {
  return (
    <div className="text-xs" style={{ height: 'calc(100vh - 120px)' }}>
      <IframeDashboard src={'grafana/dashboards/f/edwu5b9rkv94wb'} />
    </div>
  )
}
export default MiddlewareDashboard
