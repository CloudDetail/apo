import { CButton, CCard, CToast, CToastBody } from '@coreui/react'
import React, { useEffect, useMemo, useState } from 'react'
import { IoMdInformationCircleOutline } from 'react-icons/io'
import { getAlertRulesApi, reloadAlertRulesApi } from 'src/api/alerts'
import LoadingSpinner from 'src/components/Spinner'
import BasicTable from 'src/components/Table/basicTable'
import Tag from 'src/components/Tag/Tag'
import { formatTime } from 'src/utils/time'
import { showToast } from 'src/utils/toast'

export default function AlertsPage() {
  const [data, setData] = useState([])
  const [loading, setLoading] = useState(false)
  const getAlertsRule = () => {
    setLoading(true)
    getAlertRulesApi({
      type: 'alert',
      exclude_alerts: true,
    })
      .then((res) => {
        setLoading(false)
        setData(res.data.groups)
      })
      .catch(() => {
        setData([])
        setLoading(false)
      })
      .finally(() => {
        setLoading(false)
      })
  }
  const reloadRule = () => {
    setLoading(true)
    reloadAlertRulesApi()
      .then((res) => {
        showToast({
          title: '重载规则成功',
          color: 'success',
        })
        getAlertsRule()
        // setLoading(false)
      })
      .catch(() => {
        setLoading(false)
      })
  }
  const column = [
    {
      title: '组名',
      accessor: 'name',
      customWidth: 120,
      justifyContent: 'left',
    },
    // {
    //   title: 'Type',
    //   accessor: 'type',
    // },
    {
      title: 'Rules',
      accessor: 'rules',
      justifyContent: 'left',
      hide: true,
      isNested: true,
      children: [
        {
          title: '规则名',
          accessor: 'name',
          justifyContent: 'left',
          customWidth: 200,
        },
        // {
        //   title: '规则类型',
        //   accessor: 'type',
        //   justifyContent: 'left',
        //   customWidth: 100,
        //   Cell: ({ value }) => {
        //     return <Tag type={value === 'alerting' ? 'warning' : 'primary'}>{value}</Tag>
        //   },
        // },
        {
          title: '查询语句',
          accessor: 'query',
          justifyContent: 'left',
        },

        {
          title: '持续时间',
          accessor: 'duration',
          customWidth: 100,
          Cell: ({ value }) => {
            return formatTime(value * 1000)
          },
        },
        {
          title: '告警状态',
          accessor: 'state',
          customWidth: 150,
          Cell: ({ value }) => {
            // 告警中 准备告警 正常
            const stateMap = {
              firing: {
                type: 'error',
                context: '告警中',
              },
              pending: {
                type: 'warning',
                context: '准备告警',
              },
              inactive: {
                type: 'success',
                context: '正常',
              },
            }
            return <Tag type={stateMap[value].type}>{stateMap[value].context}</Tag>
          },
        },

        // {
        //   title: 'Action',
        //   accessor: 'action',
        //   customWidth: 100,
        //   Cell: () => {
        //     return (
        //       <div className=" cursor-pointer">
        //         <AiFillEdit color="#4566d6" size={18} />
        //       </div>
        //     )
        //   },
        // },
        //
      ],
    },
  ]
  useEffect(() => {
    getAlertsRule()
  }, [])
  const tableProps = useMemo(() => {
    // const paginatedData = data.slice((pageIndex - 1) * pageSize, pageIndex * pageSize)
    return {
      columns: column,
      data: data,

      loading: false,
    }
  }, [data])
  return (
    <>
      <LoadingSpinner loading={loading} />
      <CToast autohide={false} visible={true} className="align-items-center w-full my-2">
        <div className="d-flex">
          <CToastBody className=" flex flex-row items-center text-xs">
            <IoMdInformationCircleOutline size={20} color="#f7c01a" className="mr-1" />
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
      </CToast>
      <CCard className="text-sm p-2">
        <div
          className="mb-4 h-full p-2 text-xs justify-between"
          style={{ height: 'calc(100vh - 210px)' }}
        >
          <div className="text-right">
            <CButton color="primary" size="sm" onClick={reloadRule}>
              重载规则
            </CButton>
          </div>
          <BasicTable {...tableProps} />
        </div>
      </CCard>
    </>
  )
}
