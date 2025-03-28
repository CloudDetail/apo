/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { CAccordionBody } from '@coreui/react'
import React, { useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useSelector } from 'react-redux'
import IframeDashboard from 'src/core/components/Dashboard/IframeDashboard'
import { selectProcessedTimeRange, timeRangeList } from 'src/core/store/reducers/timeRangeReducer'

function MetricsDashboard({ variable }) {
  const [src, setSrc] = useState("")
  const { i18n } = useTranslation()

  useEffect(() => {
    let src = i18n.language === 'zh'
      ? `grafana/d/adst2iva9181se/e59fba-e7a180-e8aebe-e696bd-e68385-e586b5?orgId=1`
      : `grafana/d/bba60ba1600c34/e59fba-e7a180-e8aebe-e696bd-e68385-e586b5?orgId=1`

    variable?.namespaceList?.forEach((item) => {
      src += `&var-namespace=${encodeURIComponent(item)}`
    })

    if (variable?.podList.length > 0) {
      variable?.podList?.forEach((item) => {
        src += `&var-pod=${encodeURIComponent(item)}`
      })
    } else {
      src += `&var-pod=ALL`
    }
    if (variable?.service) {
      src += `&var-service_name=${encodeURIComponent(variable?.service)}`
    }
    setSrc(src)
  }, [variable])
  return (
    <CAccordionBody className="text-xs h-[800px]">
      {src && <IframeDashboard srcProp={src} />}
    </CAccordionBody>
  )
}

export default MetricsDashboard
