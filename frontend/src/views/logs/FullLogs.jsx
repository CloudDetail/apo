import React, { useState } from 'react'
import HighLightCode from '../serviceInfo/component/infoUni/HighLightCode'
import LogsTraceFilter from 'src/components/Filter/LogsTraceFilter'
function FullLogs(props) {
  const { logsList } = props
  return (
    <div className="h-full flex flex-col overflow-hidden">
      <div className="flex-grow-0 flex-shrink-0">
        <LogsTraceFilter type="logs"/>
      </div>
      <div className="flex-grow flex-shrink overflow-hidden flex-column-tab ">
        <HighLightCode timestamp={1722254280000000} />
        {/* <HighLightCode timestamp={logsList[0]} /> */}
      </div>
    </div>
  )
}
export default FullLogs
