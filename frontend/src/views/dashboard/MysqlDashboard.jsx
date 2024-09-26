import React, { useEffect, useRef } from 'react'
import IframeDashboard from 'src/components/Dashboard/IframeDashboard'

function MysqlDashboard() {
  return (
    <div className="text-xs" style={{ height: 'calc(100vh - 120px)' }}>
      <IframeDashboard src={'grafana/d/0D6dTg3Zk/mysql-e68c87-e6a087?orgId=1'} />
    </div>
  )
}
export default MysqlDashboard
