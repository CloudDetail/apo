/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useSelector } from 'react-redux'
import IframeDashboard from 'src/core/components/Dashboard/IframeDashboard'
import { useServiceInfoContext } from 'src/oss/contexts/ServiceInfoContext'

function MetricsDashboard() {
  const dashboardVariable = useServiceInfoContext((ctx) => ctx.dashboardVariable)
  const [src, setSrc] = useState()
  const { i18n } = useTranslation()
  const { theme } = useSelector((state) => state.settingReducer)
  useEffect(() => {
    let src =
      i18n.language === 'zh'
        ? `grafana/d/adst2iva9181se/e59fba-e7a180-e8aebe-e696bd-e68385-e586b5?orgId=1`
        : `grafana/d/bba60ba1600c34/e59fba-e7a180-e8aebe-e696bd-e68385-e586b5?orgId=1`

    dashboardVariable?.namespaceList?.forEach((item) => {
      src += `&var-namespace=${encodeURIComponent(item)}`
    })

    if (dashboardVariable?.podList.length > 0) {
      dashboardVariable?.podList?.forEach((item) => {
        src += `&var-pod=${encodeURIComponent(item)}`
      })
    } else {
      src += `&var-pod=ALL`
    }
    if (dashboardVariable?.service) {
      src += `&var-service_name=${encodeURIComponent(dashboardVariable?.service)}`
    }
    src += `&theme=${theme}`
    setSrc(src)
  }, [dashboardVariable, theme])
  return <div className="text-xs h-[800px]">{src && <IframeDashboard srcProp={src} />}</div>
}

export default MetricsDashboard
