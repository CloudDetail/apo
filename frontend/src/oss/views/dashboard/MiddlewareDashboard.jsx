/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react'
import IframeDashboard from 'src/core/components/Dashboard/IframeDashboard'
import { useTranslation } from 'react-i18next'
import { Alert } from 'antd'

function MiddlewareDashboard() {
  const { t, i18n } = useTranslation('oss/middleware')

  return (
    <div className="text-xs" style={{ height: 'calc(100vh - 160px)' }}>
      <div className="flex">
        <Alert
          type="warning"
          showIcon
          className=" flex flex-row items-center text-xs w-auto mb-1"
          message={
            <>
              {' '}
              <span className="text-nowrap">{t('dashboard.toastMessage')}</span>
              <a
                className="underline text-sky-500 "
                target="_blank"
                href={
                  i18n.language === 'zh'
                    ? 'https://kindlingx.com/docs/APO%20向导式可观测性中心/配置指南/监控中间件/'
                    : 'https://docs.autopilotobservability.com/category/metrics-monitoring'
                }
              >
                {t('dashboard.documentation')}
              </a>
            </>
          }
        >
          {/* <IoMdInformationCircleOutline size={20} color="#f7c01a" className="mr-1" /> */}
        </Alert>
      </div>
      <IframeDashboard dashboardKey="middleware" />
    </div>
  )
}
export default MiddlewareDashboard
