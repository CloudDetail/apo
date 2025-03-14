/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState } from 'react'
import { AppContent, AppSidebar, AppFooter, AppHeader } from '../components/index'
import { Layout } from 'antd'
import Sider from 'antd/es/layout/Sider'
import { Header } from 'antd/es/layout/layout'
import { CImage } from '@coreui/react'
import logo from 'src/core/assets/brand/logo.svg'
import './index.css'
import { useTranslation } from 'react-i18next'

const DefaultLayout = () => {
  const { t } = useTranslation()
  const [collapsed, setCollapsed] = useState(true)

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider
        collapsible
        collapsed={collapsed}
        collapsedWidth={70}
        onCollapse={(value) => setCollapsed(value)}
        style={{
          overflowX: 'hidden',
          overflowY: 'auto',
          transition: 'all 0.3s',
          zIndex: 999,
          height: '100vh',
          borderRight: '1px solid #424242'
        }}
        className='custom-scrollbar'
        width={250}
      >
        <div className="h-[60px] flex w-full overflow-hidden items-center">
          <CImage
            src={logo}
            className="w-[42px] sidebar-brand-narrow flex-shrink-0 m-3"
            alt="CoreuiVue"
          />
          <span className="flex-shrink-0 text-lg">{t('apoTitle')}</span>
        </div>
        <AppSidebar collapsed={collapsed}/>
      </Sider>
      <Layout>
        <AppHeader />
        <div className="body flex-grow-1">
          <AppContent />
        </div>
      </Layout>
    </Layout>
  )
}

export default DefaultLayout
