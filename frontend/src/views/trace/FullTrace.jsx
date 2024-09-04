import React, { useEffect, useRef } from 'react'
import IframeDashboard from 'src/components/Dashboard/IframeDashboard'

function FullTrace() {
  return (
    <div className="text-xs h-full">
      <iframe src={'/jaeger/search'} width="100%" height="100%" frameBorder={0}></iframe>
    </div>
  )
}

export default FullTrace
