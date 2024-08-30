import { CAccordionBody } from '@coreui/react'
import React, { useEffect, useRef, useState } from 'react'
import { useSelector } from 'react-redux'
import IframeDashboard from 'src/components/Dashboard/IframeDashboard'
import { selectProcessedTimeRange, timeRangeList } from 'src/store/reducers/timeRangeReducer'

function MetricsDashboard({ variable }) {
  const [src, setSrc] = useState()

  useEffect(() => {
    let src = `grafana/d/adst2iva9181se/e59fba-e7a180-e8aebe-e696bd-e68385-e586b5?orgId=1`

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
      {/* {src && (
        <IframeDashboard
          src={src}
        />
      )} */}
    </CAccordionBody>
  )
}

export default MetricsDashboard
