/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Button, Flex, Modal, Popconfirm, Table } from 'antd'
import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { MdOutlineEdit } from 'react-icons/md'
import { RiDeleteBin5Line } from 'react-icons/ri'
import { useNavigate } from 'react-router-dom'
import { deleteClusterIntegrationApi, getIntegrationClusterListApi } from 'src/core/api/integration'
import { notify } from 'src/core/utils/notify'
import InstallCmd from './Integration/InstallCmd'
import { GoCommandPalette } from 'react-icons/go'
import { BasicCard } from 'src/core/components/Card/BasicCard'
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
  const columns = [
    {
      dataIndex: 'id',
      hidden: true,
    },
    {
      dataIndex: 'name',
      title: t('clusterName'),
      width: '30%',
    },
    {
      dataIndex: 'clusterType',
      title: t('clusterType'),
      width: '30%',
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
        <Table columns={columns} dataSource={data} scroll={{ y: 'calc(100vh - 265px)' }} />
      </BasicCard.Table>

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
        />
      </Modal>
    </BasicCard>
  )
}
export default ClusterTable
