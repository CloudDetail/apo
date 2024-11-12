import { CToast, CToastBody } from '@coreui/react'
import React from 'react'
import { IoMdInformationCircleOutline } from 'react-icons/io'

function FullTrace() {
  return (
    <div className="text-xs h-full">
      <iframe src={'/jaeger/search'} width="100%" height="100%" frameBorder={0}></iframe>
    </div>
  )
}

export default FullTrace
