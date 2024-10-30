import { CCard } from '@coreui/react'
import { Button, Input, Popconfirm, Select, Space } from 'antd'
import React, { useEffect, useMemo, useState } from 'react'
import { RiDeleteBin5Line } from 'react-icons/ri'
import { deleteAlertNotifyApi, getAlertmanagerListApi } from 'src/api/alerts'
import LoadingSpinner from 'src/components/Spinner'
import BasicTable from 'src/components/Table/basicTable'
import { showToast } from 'src/utils/toast'
import { MdAdd, MdOutlineEdit } from 'react-icons/md'
import { useSelector } from 'react-redux'
import ModifyAlertNotifyModal from './modal/ModifyAlertNotifyModal'

export default function AlertsNotify() {
  const [data, setData] = useState([])
  const [loading, setLoading] = useState(false)
  const [pageIndex, setPageIndex] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [total, setTotal] = useState(0)
  const [modalVisible, setModalVisible] = useState(false)
  const [modalInfo, setModalInfo] = useState(null)
  const [searchName, setSearchName] = useState(null)
  const deleteAlertNotify = (row) => {
    deleteAlertNotifyApi({
      name: row.name,
    }).then((res) => {
      showToast({
        title: '删除告警通知成功',
        color: 'success',
      })
      refreshTable()
    })
  }

  const judgmentType = (type) => {
    switch (type) {
      case 'emailConfigs':
        return '邮件'
      case 'webhookConfigs':
        return 'webhook'
      case 'dingTalkConfigs':
        return '钉钉'
      case 'wechatConfigs':
        return '微信'
    }
  }

  const getUrl = (type, row) => {
    switch (type) {
      case 'emailConfigs':
        return row.emailConfigs[0]?.to
      case 'webhookConfigs':
        return row.webhookConfigs[0]?.url
      case 'dingTalkConfigs':
        return row.dingTalkConfigs[0]?.url
      case 'wechatConfigs':
        if (row.wechatConfigs[0]?.api_url) {
          return row.wechatConfigs[0].api_url
        }
      default:
        return 'N/A'
    }
  }

  const typeList = [
    'emailConfigs',
    'webhookConfigs',
    'dingTalkConfigs',
    'wechatConfigs'
  ];




  const column = [
    {
      title: '告警通知规则名',
      accessor: 'name',
      justifyContent: 'left',
      customWidth: '20%',
    },
    {
      title: '通知类型',
      accessor: 'type',
      customWidth: 120,
      Cell: (props) => {
        const row = props.row.original;
        let foundItem = typeList.find(item => Object.hasOwn(row, item));
        foundItem = judgmentType(foundItem)
        return foundItem || null;
      },
    },
    {
      title: '通知邮箱或WebhookUrl',
      accessor: 'to',
      customWidth: '50%',
      Cell: (props) => {
        const row = props.row.original;
        let foundItem = typeList.find(item => Object.hasOwn(row, item));
        foundItem = getUrl(foundItem, row)
        return foundItem
      },
    },
    {
      title: '操作',
      accessor: 'action',
      Cell: (props) => {
        const row = props.row.original
        return (
          <div className="flex">
            <Button
              type="text"
              onClick={() => clickEditRule(row)}
              icon={<MdOutlineEdit className="text-blue-400 hover:text-blue-400" />}
            >
              <span className="text-blue-400 hover:text-blue-400">编辑</span>
            </Button>
            <Popconfirm
              title={
                <>
                  是否确定删除名为“<span className="font-bold ">{row.alert}</span>
                  ”的告警规则
                </>
              }
              onConfirm={() => deleteAlertNotify(row)}
              okText="确定"
              cancelText="取消"
            >
              <Button type="text" icon={<RiDeleteBin5Line />} danger>
                删除
              </Button>
            </Popconfirm>
          </div>
          // <div className=" cursor-pointer">
          //   <AiOutlineDelete color="#97242e" size={18} />
          //   删除
          // </div>
        )
      },
    },
  ]


  const clickAddRule = () => {
    setModalInfo(null)
    setModalVisible(true)
  }
  const clickEditRule = (notifyInfo) => {
    setModalInfo(notifyInfo)
    setModalVisible(true)
  }
  useEffect(() => {
    fetchData()
  }, [pageSize, pageIndex, searchName])
  const fetchData = () => {
    setLoading(true)

    getAlertmanagerListApi({
      currentPage: pageIndex,
      pageSize: pageSize,
      name: searchName,
    })
      .then((res) => {
        setLoading(false)
        setTotal(res.pagination.total)
        setData(res.amConfigReceivers)
      })
      .catch((error) => {
        setLoading(false)
      })
  }
  const handleTableChange = (props) => {
    if (props.pageSize && props.pageIndex) {
      setPageSize(props.pageSize), setPageIndex(props.pageIndex)
    }
  }
  const refreshTable = () => {
    fetchData()
    setPageIndex(1)
  }
  const tableProps = useMemo(() => {
    // 分页处理
    return {
      columns: column,
      data: data,
      onChange: handleTableChange,
      pagination: {
        pageSize: pageSize,
        pageIndex: pageIndex,
        pageCount: Math.ceil(total / pageSize),
      },
      loading: false,
    }
  }, [data, pageIndex, pageSize])
  return (
    <>
      <LoadingSpinner loading={loading} />
      {/* <CToast autohide={false} visible={true} className="align-items-center w-full my-2">
        <div className="d-flex">
          <CToastBody className=" flex flex-row items-center text-xs">
            <IoMdInformationCircleOutline size={20} color="#f7c01a" className="mr-1" />
            配置后预计15s生效，请稍后刷新页面查看最新告警规则。
            仅展示告警规则，如需配置请参考
            <a
              className="underline text-sky-500"
              target="_blank"
              href="https://originx.kindlingx.com/docs/APO%20向导式可观测性中心/配置指南/配置告警规则"
            >
              文档
            </a>
          </CToastBody>
        </div>
      </CToast> */}
      <div className="flex items-center justify-betweeen text-sm p-2 my-2">
        <Space className="flex-grow">
          <Space className="flex-1">
            <span className="text-nowrap">通知规则名：</span>
            <Input
              value={searchName}
              onChange={(e) => {
                setSearchName(e.target.value)
                setPageIndex(1)
              }}
            />
          </Space>
        </Space>

        <Button
          type="primary"
          icon={<MdAdd size={20} />}
          onClick={clickAddRule}
          className="flex-grow-0 flex-shrink-0"
        >
          <span className="text-xs">新增告警通知</span>
        </Button>
      </div>
      <CCard className="text-sm p-2">
        <div
          className="mb-4 h-full p-2 text-xs justify-between"
          style={{ height: 'calc(100vh - 280px)' }}
        >
          <BasicTable {...tableProps} />
        </div>
      </CCard>
      <ModifyAlertNotifyModal
        modalVisible={modalVisible}
        notifyInfo={modalInfo}
        closeModal={() => setModalVisible(false)}
        refresh={refreshTable}
      />
    </>
  )
}
