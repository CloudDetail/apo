import { CCard } from '@coreui/react'
import React, { useEffect, useState } from 'react'
import ReactJson from 'react-json-view'
import { getAlertRulesApi } from 'src/api/alerts'

export default function AlertsPage() {
  const [data, setData] = useState()
  const getAlertsRule = () => {
    getAlertRulesApi().then((res) => {
      setData(res.data)
    })
  }
  useEffect(() => {
    getAlertsRule()
  }, [])
  return (
    <CCard className="text-sm p-2">
      {/* <div
        className="whitespace-pre-wrap font-mono leading-normal overflow-x-auto p-4 w-full"
        style={{ fontFamily: '"Courier New", Courier, monospace' }}
      >
        {JSON.stringify(data, null, 2)}
      </div> */}
      <ReactJson src={data} theme="brewer" collapsed={false} displayDataTypes={false} />
    </CCard>
  )
}
