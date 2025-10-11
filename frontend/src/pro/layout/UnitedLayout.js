import React from 'react'
import { AppContent, AppSidebar, AppHeader, AppBreadcrumb } from 'src/core/components/index'
import { Layout, Alert, Button } from 'antd'
import Sider from 'antd/es/layout/Sider'
import 'src/core/layout/index.css'
import { useLicense } from 'src/pro/contexts/LicenseContext'
import { useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'

const UnitedLayout = () => {
  const { licenseInfo } = useLicense()
  const navigate = useNavigate()
  const { t } = useTranslation('common')

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
    <Layout style={{ height: '100vh', overflow: 'hidden' }} className='flex flex-col '>
      {getLicenseAlert()}
      <AppHeader type="united" />
      <Layout className='flex-1 h-0'>
        {/* <Sider collapsed collapsedWidth={70}></Sider> */}
        <Sider
          trigger={null}
          collapsible
          collapsed={true}
          collapsedWidth={70}
          style={{
            overflow: 'auto',
            transition: 'all 0.3s',
            zIndex: 999,
          }}
          width={250}
          className={'siderCollapsed'}
        >
          {/* <div className="h-[60px] flex w-full overflow-hidden items-center">
            <CImage
              src={logo}
              className="w-[42px] sidebar-brand-narrow flex-shrink-0 m-3"
              alt="CoreuiVue"
            />
            <span className="flex-shrink-0 text-lg">向导式可观测平台</span>
          </div> */}
          <AppSidebar />
        </Sider>
        <Layout>
          <div className="body flex-grow-1 h-full flex flex-col">
            <div className="px-4 py-0 text-sm flex-0">
              <AppBreadcrumb />
            </div>
            <AppContent />
          </div>
        </Layout>
      </Layout>
    </Layout>
  )
}

export default UnitedLayout
