/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useEffect, useMemo } from 'react'
import { Menu, Layout, Card, Typography, theme } from 'antd'
import { TeamOutlined, SafetyCertificateOutlined } from '@ant-design/icons'
import LoadingSpinner from 'src/core/components/Spinner'
import { useTranslation } from 'react-i18next'
import { useMenuPermission } from './useMenuPermission'
import PermissionTree from 'src/core/components/PermissionTree'

function MenuManagePage() {
  const {
    roleList,
    selectedRole,
    loading,
    updateLoading,
    fetchRoles,
    handleRoleSelect,
    handleSavePermissions,
  } = useMenuPermission()
  const { t } = useTranslation('core/menuManage')
  const { useToken } = theme
  const { token } = useToken()

  const menuItems = useMemo(() => {
    return roleList.map((role) => ({
      key: role.roleId,
      label: role.roleName,
      icon: <TeamOutlined className="ml-6 mr-1" />,
    }))
  }, [roleList])

  const onSelect = ({ key }: { key: string }) => {
    handleRoleSelect(key)
  }

  useEffect(() => {
    fetchRoles()
  }, [])

  return (
    <Layout style={{ height: '100%', overflow: 'hidden' }}>
      <LoadingSpinner loading={loading || updateLoading} />
      <Layout.Content className="p-0 flex gap-0 h-full">
        <Card
          className="w-48"
          style={{ width: '20%', borderTopRightRadius: '0px', borderBottomRightRadius: '0px' }}
          styles={{
            body: {
              height: '100%',
              padding: '0px',
              paddingBlockStart: '2px',
            },
            header: {
              backgroundColor: token.colorBgContainer,
            },
          }}
          title={
            <Typography.Title level={5} className="mb-0 flex items-center">
              <TeamOutlined className="mr-2 text-[var(--ant-color-primary)]" />
              {t('index.roleList')}
            </Typography.Title>
          }
          // bordered={false}
        >
          <Menu
            selectedKeys={[selectedRole?.roleId?.toString()]}
            mode="vertical"
            items={menuItems}
            // className="border-none"
            onSelect={onSelect}
            style={{
              height: 'calc(100vh - 200px)',
              border: 'none',
              overflowY: 'auto',
              backgroundColor: 'var(--ant-color-bg-container)',
            }}
          />
        </Card>

        <Card
          className="flex-1 shadow-md flex flex-col"
          style={{ borderTopLeftRadius: '0px', borderBottomLeftRadius: '0px' }}
          styles={{
            body: {
              padding: '0px',
              paddingBlockStart: '2px',
              flex: 1,
              height: 0,
              display: 'flex',
              flexDirection: 'column',
              overflow: 'hidden',
            },
            header: {
              backgroundColor: token.colorBgContainer,
            },
          }}
          title={
            <Typography.Title level={5} className="mb-0 flex items-center">
              <SafetyCertificateOutlined className="mr-2 text-[var(--ant-color-primary)]" />
              {selectedRole
                ? `${t('index.menuPermissionSetting')} - ${selectedRole.roleName}`
                : t('index.selectRole')}
            </Typography.Title>
          }
          // bordered={false}
        >
          {selectedRole ? (
            <PermissionTree
              subjectId={selectedRole.roleId}
              subjectType="role"
              onSave={handleSavePermissions}
              className="permission-tree"
              style={{ height: '100%', border: 'none' }}
              actionStyle={{ paddingInlineEnd: '32px', justifyContent: 'flex-end' }}
            />
          ) : (
            <div className="text-center text-[var(--ant-color-text-secondary)] py-8">
              {t('index.selectRole')}
            </div>
          )}
        </Card>
      </Layout.Content>
    </Layout>
  )
}

export default MenuManagePage
