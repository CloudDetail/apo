import { Button, Flex, Modal, Popconfirm, Table } from 'antd'
import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { MdOutlineEdit } from 'react-icons/md'
import { RiDeleteBin5Line } from 'react-icons/ri'
import { useNavigate } from 'react-router-dom'
import { deleteClusterIntegrationApi, getIntegrationClusterListApi } from 'src/core/api/integration'
import { showToast } from 'src/core/utils/toast'
import InstallCmd from './Integration/InstallCmd'
import { GoCommandPalette } from 'react-icons/go'
const ClusterTable = () => {
  const { t } = useTranslation('core/dataIntegration')
  const { t: ct } = useTranslation('common')
  const navigate = useNavigate()
  const [data, setData] = useState([])
  const [modalOpen, setModalOpen] = useState(false)
  const [clusterId, setClusterId] = useState(null)
  const deleteClusterIntegration = (id: string) => {
    deleteClusterIntegrationApi(id).then((res) => {
      showToast({
        color: 'success',
        title: ct('deleteSuccess'),
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
      width: '30%',
      render: (_, record) => {
        return (
          <Flex align="center">
            <Button
              type="text"
              onClick={() => {
                // setInfoModalVisible(true)
                // setGroupInfo(record)
                toSettingPage(record.id)
              }}
              icon={<MdOutlineEdit className="text-blue-400 hover:text-blue-400" />}
            >
              <span className="text-blue-400 hover:text-blue-400">{ct('edit')}</span>
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
                setClusterId(record.id)
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
      setData(res.clusters || [])
    })
  }
  const toSettingPage = (clusterId?: string) => {
    let url = '/integration/data/settings'
    if (clusterId) {
      url += '?clusterId=' + encodeURIComponent(clusterId)
    }
    navigate(url)
  }
  useEffect(() => {
    getData()
  }, [])
  return (
    <div className="flex flex-col">
      <div className="flex items-center justify-between">
        <div>{/* //serach */}</div>
        <Button type="primary" onClick={() => toSettingPage()}>
          {ct('add')}
        </Button>
      </div>

      <Table columns={columns} dataSource={data} />
      <Modal
        open={modalOpen}
        footer={null}
        onCancel={() => {
          setModalOpen(false)
          setClusterId(null)
        }}
      >
        <InstallCmd clusterId={clusterId} />
      </Modal>
    </div>
  )
}
export default ClusterTable
