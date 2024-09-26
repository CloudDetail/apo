import { CCard, CToast, CToastBody } from '@coreui/react'
import { Button, Popconfirm } from 'antd'
import React, { useEffect, useMemo, useState } from 'react'
import { RiDeleteBin5Line } from 'react-icons/ri'
import { IoMdInformationCircleOutline } from 'react-icons/io'
import { deleteRuleApi, getAlertRulesApi } from 'src/api/alerts'
import LoadingSpinner from 'src/components/Spinner'
import BasicTable from 'src/components/Table/basicTable'
import { showToast } from 'src/utils/toast'
import { MdAdd, MdOutlineEdit } from 'react-icons/md'
import ModifyAlertRuleModal from './modal/ModifyAlertRuleModal'

export default function AlertsPage() {
  const [data, setData] = useState([])
  const [loading, setLoading] = useState(false)
  const [pageIndex, setPageIndex] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [total, setTotal] = useState(0)
  const [modalVisible, setModalVisible] = useState(false)
  const [modalInfo, setModalInfo] = useState(null)
  const getAlertsRule = () => {
    setLoading(true)
    getAlertRulesApi({
      currentPage: 1,
      pageSize: 10000,
    })
      .then((res) => {
        setLoading(false)
        setData(res.alertRules)
        setTotal(res.pagination.total)
      })
      .catch(() => {
        setData([])
        setLoading(false)
      })
      .finally(() => {
        setLoading(false)
      })
  }
  const deleteAlertRule = (rule) => {
    deleteRuleApi({
      group: rule.group,
      alert: rule.alert,
    }).then((res) => {
      showToast({
        title: '删除告警规则成功',
        color: 'success',
      })
      refreshTable()
    })
  }
  const column = [
    {
      title: '组名',
      accessor: 'group',
      customWidth: 120,
      justifyContent: 'left',
    },
    {
      title: '名称',
      accessor: 'alert',
      justifyContent: 'left',
      customWidth: 300,
    },

    {
      title: '持续时间',
      accessor: 'for',
      customWidth: 100,
    },
    {
      title: '查询语句',
      accessor: 'expr',
      justifyContent: 'left',
      Cell: ({ value }) => {
        return <span className="text-gray-400">{value}</span>
      },
    },

    // {
    //   title: '告警级别',
    //   accessor: 'labels',
    //   customWidth: 150,
    //   Cell: ({ value }) => {
    //     // 告警中 准备告警 正常
    //     const stateMap = {
    //       warning: {
    //         type: 'error',
    //         context: '告警中',
    //       },
    //       pending: {
    //         type: 'warning',
    //         context: '准备告警',
    //       },
    //       info: {
    //         type: 'success',
    //         context: '正常',
    //       },
    //     }
    //     return <Tag type={stateMap[value.severity].type}>{stateMap[value.severity].context}</Tag>
    //   },
    // },
    {
      title: '操作',
      accessor: 'action',
      customWidth: 200,
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
              onConfirm={() => deleteAlertRule(row)}
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
  const clickEditRule = (ruleInfo) => {
    setModalInfo(ruleInfo)
    setModalVisible(true)
  }
  useEffect(() => {
    getAlertsRule()
  }, [])
  const handleTableChange = (props) => {
    if (props.pageSize && props.pageIndex) {
      setPageSize(props.pageSize), setPageIndex(props.pageIndex)
    }
  }
  const refreshTable = () => {
    getAlertsRule()
    setPageIndex(1)
  }
  const tableProps = useMemo(() => {
    const paginatedData = data.slice((pageIndex - 1) * pageSize, pageIndex * pageSize)
    return {
      columns: column,
      data: paginatedData,
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
      <CCard className="text-sm p-2">
        <div
          className="mb-4 h-full p-2 text-xs justify-between"
          style={{ height: 'calc(100vh - 210px)' }}
        >
          <div className="text-right">
            <Button type="primary" icon={<MdAdd size={20} />} onClick={clickAddRule}>
              <span className="text-xs">新增告警规则</span>
            </Button>
            {/* <CButton color="primary" size="sm" onClick={reloadRule}>
              重载规则
            </CButton> */}
          </div>
          <BasicTable {...tableProps} />
        </div>
      </CCard>
      <ModifyAlertRuleModal
        modalVisible={modalVisible}
        ruleInfo={modalInfo}
        closeModal={() => setModalVisible(false)}
        refresh={refreshTable}
      />
    </>
  )
}
