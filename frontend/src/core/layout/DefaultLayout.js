import React, { useState } from 'react'
import { AppContent, AppSidebar, AppFooter, AppHeader } from '../components/index'
import { Layout } from 'antd'
import Sider from 'antd/es/layout/Sider'
import { Header } from 'antd/es/layout/layout'
import { CImage } from '@coreui/react'
import logo from 'src/core/assets/brand/logo.svg'
import './index.css'
const DefaultLayout = () => {
  const [collapsed, setCollapsed] = useState(true)
  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider collapsed collapsedWidth={70}></Sider>
      <Sider
        trigger={null}
        collapsible
        collapsed={collapsed}
        onMouseEnter={() => setCollapsed(false)}
        onMouseLeave={() => setCollapsed(true)}
        collapsedWidth={70}
        style={{
          overflow: 'auto',
          transition: 'all 0.3s',
          position: 'fixed',
          zIndex: 999,
          height: '100vh',
        }}
        width={250}
        className={collapsed ? 'siderCollapsed border-end' : 'border-end'}
      >
        <div className="h-[60px] flex w-full overflow-hidden items-center">
          <CImage
            src={logo}
            className="w-[42px] sidebar-brand-narrow flex-shrink-0 m-3"
            alt="CoreuiVue"
          />
          <span className="flex-shrink-0 text-lg">向导式可观测平台</span>
        </div>
        <AppSidebar collapsed={collapsed} />
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
