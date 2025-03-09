/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useEffect, useRef } from 'react'
import IframeDashboard from 'src/core/components/Dashboard/IframeDashboard'


function SystemDashboard() {
  return (
    <div className="text-xs" style={{ height: 'calc(100vh - 120px)' }}>
      <IframeDashboard dashboardKey="system"  />
    </div>
  )
}
export default SystemDashboard
