/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState } from 'react'
import { AppContent, AppSidebar, AppFooter, AppHeader } from '../components/index'
import { MenuUnfoldOutlined, MenuFoldOutlined } from '@ant-design/icons';
import { Button, Layout } from 'antd'
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
        trigger={null}
        collapsed={collapsed}
        collapsedWidth={70}
        onCollapse={(value) => setCollapsed(value)}
        style={{
          position: 'fixed',
          overflowX: 'hidden',
          overflowY: 'auto',
          transition: 'all 0.3s',
          zIndex: 999,
          height: '100vh',
          borderRight: '1px solid var(--ant-color-border-secondary)'
        }}
        className={`custom-scrollbar ${collapsed ? 'siderCollapsed' : ''}`}
        width={200}
      >
        <div className="h-[60px] flex w-full overflow-hidden items-center justify-center gap-1">
          <CImage
            src={logo}
            className="w-[32px] sidebar-brand-narrow flex-shrink-0"
            alt="CoreuiVue"
          />
          {!collapsed && <span className="flex-shrink-0 text-lg">{t('apoTitle')}</span>}
        </div>
        <AppSidebar />
        {/* <Button
          type="text"
          icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
          onClick={() => setCollapsed(!collapsed)}
          style={{
            fontSize: '16px',
            width: 45,
            height: 45,
            position: 'absolute',
            margin: '4px',
            bottom: 0
          }}
        ></Button> */}
      </Sider>
      <Layout
        style={{
          marginLeft: collapsed ? '70px' : '200px',
          transition: 'margin-left 0.3s ease-in-out'
        }}
      >
        <AppHeader />
        <div className="body flex-grow-1">
          <AppContent />
        </div>
      </Layout>
    </Layout>
  )
}

export default DefaultLayout
