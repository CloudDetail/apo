/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Button, Flex, Modal, Popconfirm, Table } from 'antd'
import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { MdOutlineEdit } from 'react-icons/md'
import { RiDeleteBin5Line } from 'react-icons/ri'
import { useLocation, useNavigate } from 'react-router-dom'
import { deleteClusterIntegrationApi, getIntegrationClusterListApi } from 'src/core/api/integration'
import { notify } from 'src/core/utils/notify'
import { GoCommandPalette } from 'react-icons/go'
import { BasicCard } from 'src/core/components/Card/BasicCard'
import InstallCmd from './Integration/InstallCmd'
import BotManagement from './BotManagement'
const ClusterTable = () => {
  const { t } = useTranslation('core/dataIntegration')
  const { t: ct } = useTranslation('common')
  const navigate = useNavigate()
  const [data, setData] = useState([])
  const [modalOpen, setModalOpen] = useState(false)
  const [clusterInfo, setClusterInfo] = useState(null)
  const deleteClusterIntegration = (id: string) => {
    deleteClusterIntegrationApi(id).then(() => {
      notify({
        type: 'success',
        message: ct('deleteSuccess'),
      })
      getData()
    })
  }
  const { pathname } = useLocation()
  const isMinimal = pathname === '/probe-management'
  const columns = [
    {
      dataIndex: 'id',
      hidden: true,
    },
    {
      dataIndex: 'name',
      title: t('clusterName'),
      width: '20%',
    },
    {
      dataIndex: 'clusterType',
      title: t('clusterType'),
      width: '15%',
    },
    {
      dataIndex: 'operation',
      title: ct('operation'),
      width: '40%',
      render: (_, record) => {
        return (
          <Flex align="center">
            <Button
              type="text"
              onClick={() => {
                // setInfoModalVisible(true)
                // setGroupInfo(record)
                toSettingPage(record.id, record.clusterType)
              }}
              icon={
                <MdOutlineEdit className="!text-[var(--ant-color-primary-text)] !hover:text-[var(--ant-color-primary-text-active)]" />
              }
            >
              <span className="text-[var(--ant-color-primary-text)] hover:text-[var(--ant-color-primary-text-active)]">
                {ct('edit')}
              </span>
            </Button>
            <Popconfirm
              title={t('confirmDelete', {
                name: record.name,
              })}
              onConfirm={() => deleteClusterIntegration(record.id)}
              okText={ct('confirm')}
              cancelText={ct('cancel')}
            >
              <Button type="text" icon={<RiDeleteBin5Line />} danger>
                {ct('delete')}
              </Button>
            </Popconfirm>
            <Button
              color="primary"
              variant="outlined"
              icon={<GoCommandPalette />}
              onClick={() => {
                setModalOpen(true)
                setClusterInfo(record)
              }}
            >
              {t('installCmdTitle')}
            </Button>
          </Flex>
        )
      },
    },
  ]
  const getData = () => {
    getIntegrationClusterListApi().then((res) => {
      const clusters = (res as any)?.clusters ?? []
      setData(clusters)
    })
  }
  const toSettingPage = (clusterId?: string, clusterType?: 'k8s' | 'vm') => {
    let url = 'settings'

    if (clusterId && clusterType) {
      url +=
        '?clusterId=' +
        encodeURIComponent(clusterId) +
        '&clusterType=' +
        encodeURIComponent(clusterType)
    }
    navigate(url)
  }
  useEffect(() => {
    getData()
  }, [])
  return (
    <>
      {/* 主要内容区域 */}
      {import.meta.env.VITE_APP_CODE_VERSION === 'EE' ? (
        <div className="page-container">
          <div className="content-grid">
            {/* 左侧部分 - Probes Management */}
            <div className="content-section">
              <div className="section-header">
                <div className="section-title-wrapper">
                  <h2 className="section-title">Probes Management</h2>
                  <p className="section-description text-[var(--ant-color-text-tertiary)]">
                    Manage Data Collection Probes
                  </p>
                </div>
                <Button type="primary" onClick={() => toSettingPage()}>
                  {ct('add')}
                </Button>
              </div>

              <BasicCard>
                <BasicCard.Table>
                  <Table
                    columns={columns}
                    dataSource={data}
                    scroll={{ y: 'calc(100vh - 200px)', x: 'max-content' }}
                    size="small"
                    pagination={false}
                  />
                </BasicCard.Table>
              </BasicCard>
            </div>

            {/* 右侧部分 - Bots Management */}
            <BotManagement />
          </div>
        </div>
      ) : (
        <>
          <BasicCard>
            <BasicCard.Header>
              <div className="w-full flex items-center justify-between mt-2">
                <div>{/* //serach */}</div>
                <Button type="primary" onClick={() => toSettingPage()}>
                  {ct('add')}
                </Button>
              </div>
            </BasicCard.Header>
            <BasicCard.Table>
              <Table columns={columns} dataSource={data} size="small" pagination={false} />
            </BasicCard.Table>
          </BasicCard>
        </>
      )}
      <Modal
        open={modalOpen}
        footer={null}
        onCancel={() => {
          setModalOpen(false)
          setClusterInfo(null)
        }}
        // title={t('installCmdTitle')}
        width={800}
        styles={{ body: { height: '70vh', overflowY: 'hidden', overflowX: 'hidden' } }}
      >
        <InstallCmd
          clusterId={clusterInfo?.id}
          clusterType={clusterInfo?.clusterType}
          apoCollector={clusterInfo?.apoCollector}
          isMinimal={isMinimal}
        />
      </Modal>
    </>
  )
}
export default ClusterTable
