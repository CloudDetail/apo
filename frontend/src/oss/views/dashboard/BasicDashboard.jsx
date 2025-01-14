/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useEffect, useRef } from 'react'
import IframeDashboard from 'src/core/components/Dashboard/IframeDashboard'

function BasicDashboard() {
  return (
    <div className="text-xs" style={{ height: 'calc(100vh - 120px)' }}>
      <IframeDashboard
        src={'grafana/d/adst2iva9181se/e59fba-e7a180-e8aebe-e696bd-e68385-e586b5?orgId=1'}
      />
    </div>
  )
}

export default BasicDashboard
