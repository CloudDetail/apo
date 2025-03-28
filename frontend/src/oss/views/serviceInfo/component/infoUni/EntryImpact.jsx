/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Tooltip } from 'antd'
import React, { useMemo, useState } from 'react'
import { AiOutlineInfoCircle } from 'react-icons/ai'
import { useSelector } from 'react-redux'
import { useNavigate } from 'react-router-dom'
import { useDebounce } from 'react-use'
import { getServiceEntryEndpoints } from 'core/api/serviceInfo'
import StatusInfo from 'src/core/components/StatusInfo'
import BasicTable from 'src/core/components/Table/basicTable'
import TempCell from 'src/core/components/Table/TempCell'
import { DelaySourceTimeUnit } from 'src/constants'
import { usePropsContext } from 'src/core/contexts/PropsContext'
import { selectSecondsTimeRange } from 'src/core/store/reducers/timeRangeReducer'
import { getStep } from 'src/core/utils/step'
import { convertTime } from 'src/core/utils/time'
import { useTranslation } from 'react-i18next'
import { useServiceInfoContext } from 'src/oss/contexts/ServiceInfoContext'

export default function EntryImpact(props) {
  const setPanelsStatus = useServiceInfoContext((ctx) => ctx.setPanelsStatus)
  const openTab = useServiceInfoContext((ctx) => ctx.openTab)

  const navigate = useNavigate()
  const [data, setData] = useState([])
  const { serviceName, endpoint } = usePropsContext()
  const [loading, setLoading] = useState(false)
  const { startTime, endTime } = useSelector(selectSecondsTimeRange)
  const { t } = useTranslation('oss/serviceInfo')

  const columns = useMemo(
    () => [
      {
        title: t('entryImpact.entryAppName'),
        accessor: 'serviceName',
        customWidth: 150,
      },
      {
        title: t('entryImpact.namespace'),
        accessor: 'namespaces',
        customWidth: 120,
        Cell: (props) => {
          return (props.value ?? []).length > 0 ? (
            props.value.join()
          ) : (
            <span className="text-slate-400">N/A</span>
          )
        },
      },
      {
        title: t('entryImpact.appDetails'),
        accessor: 'serviceDetails',
        hide: true,
        isNested: true,
        customWidth: '55%',
        clickCell: (props) => {
          const serviceName = props.cell.row.values.serviceName
          const endpoint = props.trs.endpoint
          navigate(
            `/service/info?service-name=${encodeURIComponent(serviceName)}&endpoint=${encodeURIComponent(endpoint)}&breadcrumb-name=${encodeURIComponent(serviceName)}`,
          )
          window.scrollTo(0, 0)
        },
        children: [
          {
            title: t('entryImpact.entryServiceEndpoint'),
            accessor: 'endpoint',
            canExpand: false,
          },
          {
            title: (
              <Tooltip
                title={
                  <div>
                    <div className="text-[#D3D3D3]">{t('entryImpact.delaySource.title1')}</div>
                    <div className="text-[#D3D3D3] mt-2">{t('entryImpact.delaySource.title2')}</div>
                    <div className="text-[#D3D3D3] mt-2">{t('entryImpact.delaySource.title3')}</div>
                  </div>
                }
              >
                <div className="flex flex-row justify-center items-center">
                  <div>
                    <div className="text-center flex flex-row ">
                      {t('entryImpact.delaySource.title4')}
                    </div>
                    <div className="block text-[10px]">{t('entryImpact.delaySource.title5')}</div>
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
            title: t('entryImpact.avgResponseTime'),
            minWidth: 150,
            accessor: (idx) => `latency`,
            Cell: (props) => {
              const { value } = props
              return <TempCell type="latency" data={value} timeRange={{ startTime, endTime }} />
            },
          },
          {
            title: t('entryImpact.errorRate'),
            accessor: (idx) => `errorRate`,
            minWidth: 150,
            Cell: (props) => {
              const { value } = props
              return <TempCell type="errorRate" data={value} timeRange={{ startTime, endTime }} />
            },
          },
          {
            title: t('entryImpact.throughput'),
            accessor: (idx) => `tps`,
            minWidth: 150,
            Cell: (props) => {
              const { value } = props
              return <TempCell type="tps" data={value} timeRange={{ startTime, endTime }} />
            },
          },
        ],
      },
      {
        title: t('entryImpact.logErrorCount'),
        accessor: `logs`,
        minWidth: 150,
        Cell: (props) => {
          const { value } = props
          return <TempCell type="logs" data={value} timeRange={{ startTime, endTime }} />
        },
      },
      {
        title: t('entryImpact.infrastructureStatus'),
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
        title: t('entryImpact.networkQualityStatus'),
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
        title: t('entryImpact.k8sEventStatus'),
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
        title: t('entryImpact.lastDeploymentOrRestartTime'),
        accessor: `timestamp`,
        minWidth: 90,
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
    [t, startTime, endTime, navigate],
  )

  const getTableData = () => {
    if (startTime && endTime) {
      setLoading(true)
      getServiceEntryEndpoints({
        startTime: startTime,
        endTime: endTime,
        step: getStep(startTime, endTime),
        service: serviceName,
        endpoint: endpoint,
      })
        .then((res) => {
          setData(res.data ?? [])
          if (res?.status === 'critical') openTab('impact')
          setPanelsStatus('impact', res.status)
          //   setPageIndex(1)
          setLoading(false)
        })
        .catch(() => {
          //   setPageIndex(1)
          setPanelsStatus('impact', 'unknown')
          setData([])
          setLoading(false)
        })
    }
  }
  // useEffect(() => {
  //   getTableData()
  // }, [startTime, endTime, serviceName, endpoint])
  useDebounce(
    () => {
      getTableData()
    },
    300, // 延迟时间 300ms
    [startTime, endTime, serviceName, endpoint],
  )
  const tableProps = useMemo(() => {
    return {
      columns: columns,
      data: data,
      showBorder: false,
      loading: false,
      pagination: {
        pageSize: 5,
        pageIndex: 1,
        total: data.length,
      },
      scrollY: 300,
    }
  }, [columns, data])
  return (
    <div className="text-xs relative">
      <BasicTable {...tableProps} />
    </div>
  )
}
