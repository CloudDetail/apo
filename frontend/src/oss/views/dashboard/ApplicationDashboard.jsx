/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useEffect, useRef } from 'react'
import IframeDashboard from 'src/core/components/Dashboard/IframeDashboard'

function ApplicationDashboard() {
  return (
    <div className="text-xs" style={{ height: 'calc(100vh - 120px)' }}>
      <IframeDashboard
        dashboardKey="application"
      />
    </div>
  )
}

export default ApplicationDashboard
