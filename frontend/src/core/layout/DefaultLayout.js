/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react'
import { AppContent, AppSidebar, AppHeader } from '../components/index'
import { Alert, Button, Layout } from 'antd'
import Sider from 'antd/es/layout/Sider'
import { CImage } from '@coreui/react'
import logo from 'src/core/assets/brand/logo.svg'
import './index.css'
import { useTranslation } from 'react-i18next'
import { useLicense } from 'src/pro/contexts/LicenseContext'
import { useNavigate } from 'react-router-dom'

const DefaultLayout = () => {
  const { t } = useTranslation()
  const { licenseInfo } = useLicense()
  const navigate = useNavigate()
  // 获取 license 状态信息
  const getLicenseAlert = () => {
    const { active } = licenseInfo

    // active: 0-未激活, 1-已激活, 2-已过期, -1-无效
    if (active === 1) {
      return null // 已激活，不显示提示
    }

    switch (active) {
      case 0:
        return (
          <Alert
            message={t('common:licenseStatus.inactive')}
            type="warning"
            banner
            closable
            action={
              <Button
                size="small"
                type="link"
                onClick={() => navigate('/system/platform-manage')}
              >
                {t('common:licenseStatus.goToManage')}
              </Button>
            }
            style={{ marginBottom: 0, padding: '16px 32px' }}
          />
        )
      case 2:
        return (
          <Alert
            message={t('common:licenseStatus.expired')}
            type="error"
            banner
            closable
            action={
              <Button
                size="small"
                type="link"
                onClick={() => navigate('/system/platform-manage')}
              >
                {t('common:licenseStatus.goToManage')}
              </Button>
            }
            style={{ marginBottom: 0, padding: '16px 32px' }}
          />
        )
      case -1:
        return (
          <Alert
            message={t('common:licenseStatus.invalid')}
            type="error"
            banner
            closable
            action={
              <Button
                size="small"
                type="link"
                onClick={() => navigate('/system/platform-manage')}
              >
                {t('common:licenseStatus.goToManage')}
              </Button>
            }
            style={{ marginBottom: 0, padding: '16px 32px' }}
          />
        )
      default:
        return null
    }
  }
  return (
    <Layout style={{ height: '100vh', overflow: 'hidden', flexDirection: 'column' }} className='flex flex-col '>
      {getLicenseAlert()}
      <div className='flex flex-row flex-1 h-full overflow-hidden'>
        <Sider
          collapsible
          trigger={null}
          collapsed={true}
          collapsedWidth={70}
          style={{
            // position: 'fixed',
            overflowX: 'hidden',
            overflowY: 'auto',
            transition: 'all 0.3s',
            zIndex: 999,
            // height: '100vh',
            borderRight: '1px solid var(--ant-color-border-secondary)',

          }}
          className={'siderCollapsed'}
          width={200}
        >
          <div className='flex flex-col flex-1 h-full'>
            <div
              className="h-[55px] flex w-full overflow-hidden items-center justify-center p-2 flex-0"
              style={{
                position: 'sticky',
                top: 0,
                zIndex: 1,
                backgroundColor: 'var(--color-sider)',
              }}
            >
              <CImage
                src={logo}
                className="w-[36px] sidebar-brand-narrow flex-shrink-0"
                alt="CoreuiVue"
              />
              {/* <span className="flex-shrink-0 text-lg">{t('apoTitle')}</span> */}
            </div>
            <AppSidebar />
          </div>

        </Sider>
        <Layout
          className='flex flex-col flex-1 overflow-hidden'
          style={{
            // marginLeft: '70px',
            transition: 'margin-left 0.3s ease-in-out',
          }}
        >
          <AppHeader />
          <div className="body flex-1 h-0 overflow-hidden flex flex-col">
            <AppContent />
          </div>
        </Layout>
      </div>
    </Layout>
  )
}

export default DefaultLayout
