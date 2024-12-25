import React, { useMemo, useState, useEffect } from 'react'
import StatusInfo from 'src/core/components/StatusInfo'
import BasicTable from 'src/core/components/Table/basicTable'
import { useNavigate } from 'react-router-dom'
import { convertTime } from 'src/core/utils/time'
import { getServiceDsecendantRelevanceApi } from 'core/api/serviceInfo'
import { useDispatch } from 'react-redux'
import { getStep } from 'src/core/utils/step'
import { DelaySourceTimeUnit } from 'src/constants'
import { Tooltip } from 'antd'
import { AiOutlineInfoCircle } from 'react-icons/ai'
import { useDebounce } from 'react-use'
import { showToast } from 'src/core/utils/toast'
import { useTranslation } from 'react-i18next'

function DependentTable(props) {
  const { serviceName, endpoint, startTime, endTime, storeDisplayData = false } = props
  const navigate = useNavigate()
  const [data, setData] = useState([])
  const [loading, setLoading] = useState(false)
  const dispatch = useDispatch()
  const { t } = useTranslation('oss/serviceInfo')

  const setDisplayData = (value) => {
    dispatch({ type: 'setDisplayData', payload: value })
  }
  const getTableData = () => {
    if (startTime && endTime) {
      setLoading(true)

      getServiceDsecendantRelevanceApi({
        startTime: startTime,
        endTime: endTime,
        service: serviceName,
        endpoint: endpoint,
        step: getStep(startTime, endTime),
      })
        .then((res) => {
          setData(res ?? [])
          setLoading(false)
          if (storeDisplayData) setDisplayData((res ?? []).slice(0, 5))
        })
        .catch((error) => {
          setData([])
          setLoading(false)
        })
    }
  }
  useEffect(() => {
    return () => {
      if (storeDisplayData) setDisplayData(null)
    }
  }, [])
  //防抖避免跳转使用旧时间
  useDebounce(
    () => {
      if (serviceName && endpoint) getTableData()
    },
    300, // 延迟时间 300ms
    [startTime, endTime, serviceName, endpoint],
  )
  const columns = useMemo(
    () => [
      {
        title: t('dependent.dependentTable.serviceName'),
        accessor: 'serviceName',
        customWidth: 150,
      },
      {
        title: t('dependent.dependentTable.endpoint'),
        accessor: 'endpoint',
      },
      {
        title: (
          <Tooltip
            title={
              <div>
                <div className="text-[#D3D3D3]">
                  {t('dependent.dependentTable.delaySource.title1')}
                </div>
                <div className="text-[#D3D3D3] mt-2">
                  {t('dependent.dependentTable.delaySource.title2')}
                </div>
                <div className="text-[#D3D3D3] mt-2">
                  {t('dependent.dependentTable.delaySource.title3')}
                </div>
              </div>
            }
          >
            <div className="flex flex-row justify-center items-center">
              <div>
                <div className="text-center flex flex-row ">
                  {t('dependent.dependentTable.delaySource.title4')}
                </div>
                <div className="block text-[10px]">
                  {t('dependent.dependentTable.delaySource.title5')}
                </div>
              </div>
              <AiOutlineInfoCircle size={16} className="ml-1" />
            </div>
          </Tooltip>
        ),
        accessor: 'delaySource',
        canExpand: false,
        customWidth: 112,
        Cell: ({ value }) => {
          return <>{DelaySourceTimeUnit[value]}</>
        },
      },
      {
        title: t('dependent.dependentTable.REDStatus'),
        accessor: `REDStatus`,
        Cell: (props) => {
          const { value, row, column } = props
          const alertReason = row.original?.alertReason?.[column.id]
          return (
            <>
              <StatusInfo status={value} alertReason={alertReason} title={column.title} />
            </>
          )
        },
      },
      {
        title: t('dependent.dependentTable.logsStatus'),
        accessor: `logsStatus`,
        Cell: (props) => {
          const { value, row, column } = props
          const alertReason = row.original?.alertReason?.[column.id]
          return (
            <>
              <StatusInfo status={value} alertReason={alertReason} title={column.title} />
            </>
          )
        },
      },
      {
        title: t('dependent.dependentTable.infrastructureStatus'),
        accessor: `infrastructureStatus`,
        Cell: (props) => {
          const { value, row, column } = props
          const alertReason = row.original?.alertReason?.[column.id]
          return (
            <>
              <StatusInfo status={value} alertReason={alertReason} title={column.title} />
            </>
          )
        },
      },
      {
        title: t('dependent.dependentTable.netStatus'),
        accessor: `netStatus`,
        Cell: (props) => {
          const { value, row, column } = props
          const alertReason = row.original?.alertReason?.[column.id]
          return (
            <>
              <StatusInfo status={value} alertReason={alertReason} title={column.title} />
            </>
          )
        },
      },
      {
        title: t('dependent.dependentTable.k8sStatus'),
        accessor: `k8sStatus`,
        Cell: (props) => {
          const { value, row, column } = props
          const alertReason = row.original?.alertReason?.[column.id]
          return (
            <>
              <StatusInfo status={value} alertReason={alertReason} title={column.title} />
            </>
          )
        },
      },
      {
        title: t('dependent.dependentTable.timestamp'),
        accessor: `timestamp`,
        Cell: (props) => {
          const { value } = props
          return (
            <>
              {value !== null ? (
                convertTime(value, 'yyyy-mm-dd hh:mm:ss')
              ) : (
                <span className="text-slate-400">N/A</span>
              )}{' '}
            </>
          )
        },
      },
    ],
    [t],
  )

  const toServiceInfoPage = (props) => {
    if (props.isTraced) {
      navigate(
        `/service/info?service-name=${encodeURIComponent(props.serviceName)}&endpoint=${encodeURIComponent(props.endpoint)}&breadcrumb-name=${encodeURIComponent(props.serviceName)}`,
      )
    } else {
      showToast({ title: t('dependent.dependentTable.unmonitoredService'), color: 'info' })
    }
  }

  const tableProps = useMemo(() => {
    return {
      columns: columns,
      data: data ?? [],
      loading: loading,
      clickRow: toServiceInfoPage,
    }
  }, [columns, data, startTime, endTime, loading])

  return <>{data && <BasicTable {...tableProps} />}</>
}

export default DependentTable
