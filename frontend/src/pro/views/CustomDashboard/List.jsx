/**
 * Copyright 2026 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useEffect, useState } from 'react'
import { Button, Input, Modal, Space, Table, Tag, Tooltip } from 'antd'
import { LuCopy, LuEye, LuPencil, LuPlus, LuTrash2 } from 'react-icons/lu'
import { useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useUserContext } from 'src/core/contexts/UserContext'
import {
  copyCustomDashboard,
  createCustomDashboard,
  deleteCustomDashboard,
  listCustomDashboards,
} from 'src/pro/api/customDashboard'
import {
  createRequestForTemplate,
  dashboardActions,
  formatDashboardTimestamp,
  runDashboardMutationWithPermissionRefresh,
  visibilityLabel,
} from './behavior.mjs'
import './index.scss'

const visibilityColor = { everyone: 'green', team: 'blue', personal: 'default' }
const warnPermissionRefreshFailure = (error) => {
  console.warn('[custom-dashboard] Failed to refresh permission menu after dashboard mutation.', error)
}

const CustomDashboardList = () => {
  const navigate = useNavigate()
  const { t } = useTranslation('pro/customDashboard')
  const { getUserPermission } = useUserContext()
  const [loading, setLoading] = useState(false)
  const [keyword, setKeyword] = useState('')
  const [dashboards, setDashboards] = useState([])
  const [canCreate, setCanCreate] = useState(false)
  const [creating, setCreating] = useState(false)

  const loadData = async () => {
    setLoading(true)
    try {
      const list = await listCustomDashboards({ keyword })
      setDashboards(list?.dashboards || [])
      setCanCreate(Boolean(list?.canCreate))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadData()
  }, [])

  const createDashboard = async () => {
    setCreating(true)
    try {
      const dashboard = await runDashboardMutationWithPermissionRefresh(
        () => createCustomDashboard(createRequestForTemplate('blank')),
        getUserPermission,
        warnPermissionRefreshFailure,
      )
      navigate(`/custom-dashboard/${dashboard.id}/edit`)
    } finally {
      setCreating(false)
    }
  }

  const copyDashboard = async (id) => {
    const dashboard = await runDashboardMutationWithPermissionRefresh(
      () => copyCustomDashboard(id),
      getUserPermission,
      warnPermissionRefreshFailure,
    )
    navigate(`/custom-dashboard/${dashboard.id}/edit`)
  }

  const removeDashboard = async (id) => {
    Modal.confirm({
      title: t('deleteDashboard'),
      content: t('deleteDashboardConfirm'),
      okButtonProps: { danger: true },
      onOk: async () => {
        await runDashboardMutationWithPermissionRefresh(
          () => deleteCustomDashboard(id),
          getUserPermission,
          warnPermissionRefreshFailure,
        )
        await loadData()
      },
    })
  }

  return (
    <div className="custom-dashboard-page">
      <div className="custom-dashboard-toolbar">
        <Space>
          <Input.Search
            allowClear
            value={keyword}
            onChange={(event) => setKeyword(event.target.value)}
            onSearch={loadData}
            style={{ width: 280 }}
          />
        </Space>
        <Space>
          {canCreate && (
            <Button type="primary" icon={<LuPlus />} loading={creating} onClick={createDashboard}>
              {t('new')}
            </Button>
          )}
        </Space>
      </div>
      <Table
        rowKey="id"
        loading={loading}
        dataSource={dashboards}
        columns={[
          {
            title: t('name'),
            dataIndex: 'name',
            render: (text, record) => (
              <Button type="link" onClick={() => navigate(`/custom-dashboard/${record.id}`)} style={{ padding: 0 }}>
                {text}
              </Button>
            ),
          },
          {
            title: t('description'),
            dataIndex: 'description',
            ellipsis: true,
            render: (text) => text ? (
              <Tooltip title={text}>
                <span>{text}</span>
              </Tooltip>
            ) : '-',
          },
          { title: t('creator'), dataIndex: 'creator', width: 140 },
          {
            title: t('visibility'),
            dataIndex: 'visibility',
            width: 120,
            render: (value) => <Tag color={visibilityColor[value]}>{visibilityLabel(t, value)}</Tag>,
          },
          {
            title: t('updated'),
            dataIndex: 'updatedAt',
            width: 170,
            render: formatDashboardTimestamp,
          },
          {
            title: t('actions'),
            width: 190,
            render: (_, record) => (
              <Space size={4}>
                <Button icon={<LuEye />} type="text" onClick={() => navigate(`/custom-dashboard/${record.id}`)} />
                {dashboardActions(record, canCreate).edit && <Button icon={<LuPencil />} type="text" onClick={() => navigate(`/custom-dashboard/${record.id}/edit`)} />}
                {dashboardActions(record, canCreate).copy && <Button icon={<LuCopy />} type="text" onClick={() => copyDashboard(record.id)} />}
                {dashboardActions(record, canCreate).delete && <Button icon={<LuTrash2 />} danger type="text" onClick={() => removeDashboard(record.id)} />}
              </Space>
            ),
          },
        ]}
      />
    </div>
  )
}

export default CustomDashboardList
