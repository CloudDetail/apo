import { CauseReportProvider } from 'src/pro/contexts/CauseReportContext'
import React from 'react'
import CauseReportCom from './CauseReportCom'
import './index.scss'
export default function CauseReport() {
  return (
    <CauseReportProvider>
      <CauseReportCom />
    </CauseReportProvider>
  )
}
