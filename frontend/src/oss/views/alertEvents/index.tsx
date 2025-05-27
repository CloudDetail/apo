/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Button, Modal, Tooltip, Statistic, Checkbox, Image, Card, Tag, theme, Result } from 'antd'
import { useEffect, useMemo, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useSelector } from 'react-redux'
import { getAlertEventsApi, getAlertWorkflowIdApi } from 'src/core/api/alerts'
import BasicTable from 'src/core/components/Table/basicTable'
import { convertUTCToLocal } from 'src/core/utils/time'
import WorkflowsIframe from '../workflows/workflowsIframe'
import PieChart from './PieChart'
import CountUp from 'react-countup'
import filterSvg from 'core/assets/images/filter.svg'
import { useDebounce } from 'react-use'
import { AlertDeration, ALertIsValid, AlertStatus, AlertTags } from './components/AlertInfoCom'
import { useNavigate } from 'react-router-dom'
import LoadingSpinner from 'src/core/components/Spinner'
function isJSONString(str) {
  try {
    JSON.parse(str)
    return true
  } catch (e) {
    return false
  }
}

const Filter = ({ onStatusFilterChange, onValidFilterChange }) => {
  const { t } = useTranslation('oss/alertEvents')

  const statusOptions = [
    {
      label: <Tag color={'error'}>{t('firing')}</Tag>,
      value: 'firing',
    },
    {
      label: <Tag color={'success'}>{t('resolved')}</Tag>,
      value: 'resolved',
    },
  ]
  const validOptions = [
    { label: t('valid'), value: 'valid' },
    { label: t('invalid'), value: 'invalid' },
    { label: t('other'), value: 'other' },
  ]
  return (
    <div className="flex pb-2 ">
      <div>
        {t('alertStatus')}:{' '}
        <Checkbox.Group
          onChange={onStatusFilterChange}
          options={statusOptions}
          defaultValue={['firing']}
        ></Checkbox.Group>
      </div>
      <div>
        {t('alertValidity')}:{' '}
        <Checkbox.Group
          onChange={onValidFilterChange}
          options={validOptions}
          defaultValue={['valid', 'other']}
        ></Checkbox.Group>
      </div>
    </div>
  )
}
const formatter = (value) => <CountUp end={value as number} separator="," />

// Current info right panel
const StatusPanel = ({ firingCounts, resolvedCounts }) => {
  const { t } = useTranslation('oss/alertEvents')
  const { useToken } = theme
  const { token } = useToken()

  const chartData = [
    { name: t('firing'), value: firingCounts, type: 'error' },
    { name: t('resolved'), value: resolvedCounts, type: 'success' },
  ]
  return (
    <div className="flex pb-2 h-full flex-1  ">
      <div
        className="w-full ml-1 rounded-xl flex h-full p-0"
        style={{ backgroundColor: token.colorBgContainer }}
      >
        <div className="h-full shrink-0 pl-4 flex">
          {chartData.map((item) => (
            <div className="w-[100px] h-full block items-center">
              <Statistic
                className="h-full flex flex-col justify-center"
                title={<Tag color={item.type}>{item.name}</Tag>}
                value={item.value}
                formatter={formatter}
              />
            </div>
          ))}
          {/* <div className="">
            <Statistic
              className="h-full flex flex-col justify-center"
              title={<span className="text-white">{'告警降噪率'}</span>}
              value={30}
              precision={2}
              suffix="%"
              formatter={formatter}
            />
          </div> */}
        </div>
        <div className="grow">
          <PieChart data={chartData} />
        </div>
      </div>
    </div>
  )
}

// Current info left panel
const ExtraPanel = ({ firingCounts, invalidCounts, alertCheck }) => {
  const { t } = useTranslation('oss/alertEvents')
  const { useToken } = theme
  const { token } = useToken()
  return (
    <div className=" pb-2 h-full  shrink-0 w-1/2 mr-3">
      <div
        className="w-full rounded-xl flex h-full p-2 "
        style={{ backgroundColor: token.colorBgContainer }}
      >
        <div className="ml-3 mr-7">
          <Image src={filterSvg} width={50} height={'100%'} preview={false} />
        </div>
        {alertCheck && (
          <div className="flex flex-col h-full justify-center">
            <Statistic
              className=" flex flex-col justify-center"
              title={<span>{t('rate')}</span>}
              value={firingCounts === 0 ? 0 : (invalidCounts / firingCounts) * 100}
              precision={2}
              suffix="%"
              formatter={formatter}
            />
            <span className="text-gray-400 text-xs" style={{ color: token.colorTextSecondary }}>
              {t('In')}
              <span className="mx-1">
                <Tag color={'error'}>{firingCounts}</Tag>
              </span>
              {t('alerts, AI identified')}{' '}
              <span className="mx-1">
                <Tag color={'warning'}>{invalidCounts}</Tag>
              </span>
              {t('invalid alerts for auto suppression')}
            </span>
          </div>
        )}
        {!alertCheck && (
          <div className="flex flex-col h-full justify-center gap-4">
            <span className="text-white">{t('rate')}</span>
            <span className="text-white">{t('noAlertCheckId')}</span>
          </div>
        )}
      </div>
    </div>
  )
}

const AlertEventsPage = () => {
  const { t } = useTranslation('oss/alertEvents')
  const { t: ct } = useTranslation('common')
  const navigate = useNavigate()
  const [pagination, setPagination] = useState({
    pageIndex: 1,
    pageSize: 10,
    total: 0,
  })
  const [alertEvents, setAlertEvents] = useState([])
  const { groupLabel } = useSelector((state) => state.groupLabelReducer)
  const { startTime, endTime } = useSelector((state) => state.timeRange)
  const [modalOpen, setModalOpen] = useState(false)
  const [workflowUrl, setWorkflowUrl] = useState(null)
  const [alertCheckId, setAlertCheckId] = useState(null)
  const [invalidCounts, setInvalidCounts] = useState(0)
  const [firingCounts, setFiringCounts] = useState(0)
  const [resolvedCounts, setResolvedCounts] = useState(0)
  const [statusFilter, setStatusFilter] = useState(['firing'])
  const [validFilter, setValidFilter] = useState(['valid', 'other'])
  const timerRef = useRef(null)
  const workflowMissToast = (type: 'alertCheckId' | 'workflowId') => {
    return (
      <Tooltip title={type === 'alertCheckId' ? t('missToast1') : t('missToast2')}>
        <span className="text-gray-400 text-xs">{t('workflowMiss')}</span>
      </Tooltip>
    )
  }
  const getAlertEventsRef = useRef<() => void>(() => {})
  const [loading, setLoading] = useState(true)
  const getAlertEvents = () => {
    if (timerRef.current) {
      clearTimeout(timerRef.current)
      timerRef.current = null
    }
    const validFilterReady = validFilter.includes('other')
      ? [...validFilter.filter((f) => f !== 'other'), 'skipped', 'failed', 'unknown']
      : validFilter

    getAlertEventsApi({
      startTime,
      endTime,
      pagination: {
        currentPage: pagination.pageIndex,
        pageSize: pagination.pageSize,
      },
      filter: {
        status: statusFilter,
        validity: validFilterReady,
      },
    }).then((res) => {
      const totalPages = Math.ceil(res.pagination.total / pagination.pageSize)
      if (pagination.pageIndex > totalPages && totalPages > 0) {
        setPagination({ ...pagination, pageIndex: totalPages })
        return
      }

      setAlertEvents(res?.events || [])
      setPagination({ ...pagination, total: res?.pagination.total || 0 })
      // setWorkflowId(res.alertEventAnalyzeWorkflowId)
      setAlertCheckId(res.alertCheckId)

      setInvalidCounts(res?.counts['firing-invalid'])
      setFiringCounts(res?.counts?.firing)
      setResolvedCounts(res?.counts?.resolved)
      setLoading(false)
      timerRef.current = setTimeout(
        () => {
          getAlertEventsRef.current()
        },
        5 * 60 * 1000,
      )
    })
  }
  useDebounce(
    () => {
      if (startTime && endTime) {
        setLoading(true)

        getAlertEvents()
      }
    },
    300,
    [pagination.pageIndex, pagination.pageSize, startTime, endTime, statusFilter, validFilter],
  )

  async function openWorkflowModal(workflowParams, group, name) {
    try {
      setLoading(true)
      setModalOpen(true)
      const workflowId = await getWorkflowId(group, name)
      if (!workflowId) {
        throw new Error()
      }
      let result = '/dify/app/' + workflowId + '/run-once?'
      const params = Object.entries(workflowParams)
        .map(([key, value]) => `${key}=${encodeURIComponent(value)}`)
        .join('&')
      setWorkflowUrl(result + params)
    } catch {
      setLoading(false)
      return
    } finally {
      setLoading(false)
    }
  }
  function openResultModal(workflowRunId) {
    let result = '/dify/app/' + alertCheckId + '/logs/' + workflowRunId
    setWorkflowUrl(result)
    setModalOpen(true)
  }
  const closeModal = () => {
    setWorkflowUrl(null)
    setModalOpen(false)
  }
  async function getWorkflowId(alertGroup, alertName) {
    try {
      const res = await getAlertWorkflowIdApi({ alertGroup, alertName })
      return res?.workflowId
    } catch (error) {
      console.error('获取 workflowId 失败:', error)
      return null
    }
  }
  const columns = [
    {
      title: t('alertName'),
      accessor: 'name',
      justifyContent: 'left',
      minWidth: 150,
      Cell: ({ value, row }) => {
        const level = row.original.severity
        return (
          <span className="text-sm break-words">
            <span className="align-middle inline-block">
              <AlertLevel level={level} />
            </span>
            {value}
          </span>
        )
      },
    },
    {
      title: t('alertDetail'),
      accessor: 'tags',
      justifyContent: 'left',
      Cell: ({ value, row }) => {
        return <AlertTags tags={value} detail={row.original.detail} />
      },
    },

    {
      title: t('createTime'),
      accessor: 'createTime',
      customWidth: 100,
      Cell: ({ value }) => {
        const result = convertUTCToLocal(value)
        return (
          <div>
            <div>{result.split(' ')[0]}</div>
            <div>{result.split(' ')[1]}</div>
          </div>
        )
      },
    },
    {
      title: t('duration'),
      accessor: 'duration',
      customWidth: 100,
      Cell: ({ value, row }) => {
        const updateTime = convertUTCToLocal(row.original.updateTime)
        return <AlertDeration duration={value} updateTime={updateTime} />
      },
    },
    {
      title: t('status'),
      accessor: 'status',
      customWidth: 100,
      Cell: ({ value, row }) => {
        const resolvedTime = convertUTCToLocal(row.original.endTime)
        return <AlertStatus status={value} resolvedTime={resolvedTime} />
      },
    },
    {
      title: t('isValid'),
      accessor: 'validity',
      customWidth: 210,
      Cell: (props) => {
        const { value, row } = props
        const checkTime = convertUTCToLocal(row.original.lastCheckAt)

        return (
          <ALertIsValid
            isValid={value}
            alertCheckId={alertCheckId}
            checkTime={checkTime}
            openResultModal={() => openResultModal(row.original.workflowRunId)}
            workflowRunId={row.original.workflowRunId}
          />
        )
      },
    },
    {
      title: ct('operation'),
      accessor: 'operation',
      customWidth: 230,
      Cell: (props) => {
        const { workflowParams, group, name, alertId, id } = props.row.original
        return (
          <div className="flex flex-col">
            <Button
              color="primary"
              variant="outlined"
              className="text-xs my-2"
              size="small"
              onClick={async () => {
                await openWorkflowModal(workflowParams, group, name)
              }}
            >
              {t('cause')}
            </Button>
            <Button
              color="primary"
              variant="outlined"
              className="text-xs"
              size="small"
              onClick={() => {
                navigate(
                  `/alerts/events/detail/${encodeURIComponent(alertId)}/${encodeURIComponent(id)}`,
                )
              }}
            >
              {t('viewDetail')}
            </Button>
          </div>
        )
      },
    },
  ]
  const updatePagination = (newPagination) => setPagination({ ...pagination, ...newPagination })
  const changePagination = (pageIndex, pageSize) => {
    updatePagination({
      pageSize: pageSize,
      pageIndex: pageIndex,
      // total: pagination.total,
    })
  }
  const tableProps = useMemo(() => {
    return {
      columns: columns,
      data: alertEvents,
      showBorder: false,
      loading: false,
      pagination: {
        pageSize: pagination.pageSize,
        pageIndex: pagination.pageIndex,
        total: pagination.total,
      },
      onChange: changePagination,
    }
  }, [alertEvents, pagination.pageIndex, pagination.pageSize, pagination.total])
  const chartHeight = 150
  const headHeight =
    (import.meta.env.VITE_APP_CODE_VERSION === 'CE' ? 60 : 100) + chartHeight + 'px'
  getAlertEventsRef.current = getAlertEvents

  useEffect(() => {
    return () => {
      if (timerRef.current) {
        clearTimeout(timerRef.current)
      }
    }
  }, [])

  return (
    <>
      <div className="overflow-hidden h-full flex flex-col">
        <div style={{ height: chartHeight }} className="flex">
          <ExtraPanel
            invalidCounts={invalidCounts}
            firingCounts={firingCounts}
            alertCheck={alertCheckId}
          />
          <StatusPanel firingCounts={firingCounts} resolvedCounts={resolvedCounts} />
        </div>
        <Card
          style={{
            height: `calc(100vh - ${headHeight})`,
            display: 'flex',
            flexDirection: 'column',
          }}
          styles={{
            body: {
              flex: 1,
              overflow: 'auto',
              padding: '16px',
              display: 'flex',
              flexDirection: 'column',
            },
          }}
        >
          <Filter
            onStatusFilterChange={(checkedValues) => {
              setStatusFilter(checkedValues)
            }}
            onValidFilterChange={(checkedValues) => {
              setValidFilter(checkedValues)
            }}
          />
          <div className="flex-1 overflow-hidden">
            <LoadingSpinner loading={loading} />
            <BasicTable {...tableProps} />
          </div>
        </Card>
        <Modal
          open={modalOpen}
          title={t('workflowsModal')}
          onCancel={closeModal}
          destroyOnClose
          centered
          footer={() => <></>}
          maskClosable={false}
          width={'80vw'}
          styles={{ body: { height: '80vh', overflowY: 'hidden', overflowX: 'hidden' } }}
        >
          <LoadingSpinner loading={loading} />
          {!loading && !workflowUrl && (
            <Result
              status="error"
              title={t('missToast2')}
              className="h-full flex flex-col items-center justify-center w-full"
            />
          )}
          {workflowUrl && <WorkflowsIframe src={workflowUrl} />}
        </Modal>
      </div>
    </>
  )
}
export default AlertEventsPage
