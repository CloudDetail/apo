import { CToast, CToastBody } from '@coreui/react'
import React from 'react'
import { IoMdInformationCircleOutline } from 'react-icons/io'
import IframeDashboard from 'src/components/Dashboard/IframeDashboard'

function MiddlewareDashboard() {
  return (
    <div className="text-xs" style={{ height: 'calc(100vh - 160px)' }}>
      <CToast autohide={false} visible={true} className="align-items-center w-full mb-2">
        <div className="d-flex">
          <CToastBody className=" flex flex-row items-center text-xs">
            <IoMdInformationCircleOutline size={20} color="#f7c01a" className="mr-1" />
            采集中间件指标的配置方式请参考
            <a
              className="underline text-sky-500"
              target="_blank"
              href="https://originx.kindlingx.com/docs/APO%20向导式可观测性中心/配置指南/监控中间件"
            >
              文档
            </a>
          </CToastBody>
        </div>
      </CToast>
      <IframeDashboard src={'grafana/dashboards/f/edwu5b9rkv94wb'} />
    </div>
  )
}
export default MiddlewareDashboard
