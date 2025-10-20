/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react'
import { AppContent, AppSidebar, AppHeader } from '../components/index'
import { Layout } from 'antd'
import Sider from 'antd/es/layout/Sider'
import { CImage } from '@coreui/react'
import './index.css'
import { useDeployMode } from '../contexts/DeployModeContext'

const DefaultLayout = () => {
  const { getLogo } = useDeployMode()

  return (
    <Layout style={{ height: '100vh', overflow: 'hidden' }} className='flex flex-col '>
      <Sider
        collapsible
        trigger={null}
        collapsed={true}
        collapsedWidth={70}
        style={{
          position: 'fixed',
          overflowX: 'hidden',
          overflowY: 'auto',
          transition: 'all 0.3s',
          zIndex: 999,
          height: '100vh',
          borderRight: '1px solid var(--ant-color-border-secondary)',
        }}
        className={'custom-scrollbar siderCollapsed '}
        width={200}
      >
        <div className='h-full flex flex-col overflow-hidden'>

          <div
            className="h-[55px] flex w-full overflow-hidden items-center justify-center p-2 flex-0"
            style={{
              backgroundColor: 'var(--color-sider)',
            }}
          >
            <CImage
              src={getLogo()}
              className="w-[36px] sidebar-brand-narrow flex-shrink-0"
              alt="Logo"
            />
            {/* <span className="flex-shrink-0 text-lg">{t('apoTitle')}</span> */}
          </div>
          <AppSidebar />
        </div>
      </Sider>
      <Layout
        className='flex flex-col'
        style={{
          marginLeft: '70px',
          transition: 'margin-left 0.3s ease-in-out',
        }}
      >
        <AppHeader />
        <div className="body flex-1 h-0  flex flex-col">
          <AppContent />
        </div>
      </Layout>
    </Layout>
  )
}

export default DefaultLayout
