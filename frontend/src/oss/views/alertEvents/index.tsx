/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Button, Modal, Tag as AntdTag, Tooltip, Statistic, Checkbox, Image, Card } from 'antd'
import { useEffect, useMemo, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useSelector } from 'react-redux'
import { getAlertEventsApi } from 'src/core/api/alerts'
import BasicTable from 'src/core/components/Table/basicTable'
import { convertUTCToLocal } from 'src/core/utils/time'
import WorkflowsIframe from '../workflows/workflowsIframe'
import Tag from 'src/core/components/Tag/Tag'
import PieChart from './PieChart'
import CountUp from 'react-countup'
import filterSvg from 'core/assets/images/filter.svg'
import ReactJson from 'react-json-view'
import { useDebounce } from 'react-use'
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
      label: <Tag type={'error'}>{t('firing')}</Tag>,
      value: 'firing',
    },
    {
      label: <Tag type={'success'}>{t('resolved')}</Tag>,
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
      {/* Todo: need to be translated */}
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

  const chartData = [
    { name: t('firing'), value: firingCounts, type: 'error' },
    { name: t('resolved'), value: resolvedCounts, type: 'success' },
  ]
  return (
    <div className="flex pb-2 h-full flex-1  ">
      <div className="w-full ml-1 rounded-xl flex h-full bg-[#141414] p-0">
        <div className="h-full shrink-0 pl-4 flex">
          {chartData.map((item) => (
            <div className="w-[100px] h-full block items-center">
              <Statistic
                className="h-full flex flex-col justify-center"
                title={<Tag type={item.type}>{item.name}</Tag>}
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
  return (
    <div className=" pb-2 h-full  shrink-0 w-1/2 mr-3">
      <div className="w-full rounded-xl flex h-full bg-[#141414] p-2 ">
        <div className="ml-3 mr-7">
          <Image src={filterSvg} width={50} height={'100%'} preview={false} />
        </div>
        {alertCheck && (
          <div className="flex flex-col h-full justify-center">
            <Statistic
              className=" flex flex-col justify-center"
              title={<span className="text-white">{t('rate')}</span>}
              value={firingCounts === 0 ? 0 : (invalidCounts / firingCounts) * 100}
              precision={2}
              suffix="%"
              formatter={formatter}
            />
            <span className="text-gray-400 text-xs">
              {t('In')}
              <span className="mx-1">
                <Tag type={'error'}>{firingCounts}</Tag>
              </span>
              {t('alerts, AI identified')}{' '}
              <span className="mx-1">
                <Tag type={'warning'}>{invalidCounts}</Tag>
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
  const [workflowId, setWorkflowId] = useState(null)
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

  const getAlertEvents = () => {
    // 🔁 每次都清掉旧定时器
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
      setWorkflowId(res.alertEventAnalyzeWorkflowId)
      setAlertCheckId(res.alertCheckId)

      setInvalidCounts(res?.counts['firing-invalid'])
      setFiringCounts(res?.counts?.firing)
      setResolvedCounts(res?.counts?.resolved)

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
        getAlertEvents()
      }
    },
    300,
    [pagination.pageIndex, pagination.pageSize, startTime, endTime, statusFilter, validFilter],
  )

  function openWorkflowModal(workflowParams) {
    let result = '/dify/app/' + workflowId + '/run-once?'
    const params = Object.entries(workflowParams)
      .map(([key, value]) => `${key}=${encodeURIComponent(value)}`)
      .join('&')
    setWorkflowUrl(result + params)
    setModalOpen(true)
    // buildParams('workflowParams', workflowParams)
    // return paramsArray.join('&')
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
  const columns = [
    {
      title: t('alertName'),
      accessor: 'name',
      justifyContent: 'left',
      minWidth: 150,
    },
    {
      title: t('alertDetail'),
      accessor: 'tags',
      justifyContent: 'left',
      Cell: ({ value, row }) => {
        const [visible, setVisible] = useState(false)

        const { detail } = row.original
        return (
          <div className="overflow-hidden">
            {Object.entries(value || {}).map(([key, tagValue]) => (
              <AntdTag className="text-pretty mb-1 break-all">
                {key} = {tagValue}
              </AntdTag>
            ))}

            {isJSONString(detail) && (
              <Button
                color="primary"
                variant="text"
                size="small"
                onClick={() => setVisible(!visible)}
              >
                {visible ? t('collapse') : t('expand')}
              </Button>
            )}

            {visible && (
              <ReactJson
                src={JSON.parse(detail || '')}
                theme="brewer"
                collapsed={false}
                displayDataTypes={false}
                style={{ width: '100%' }}
                enableClipboard={false}
              />
            )}
          </div>
        )
      },
    },

    {
      title: t('lastAlertTime'),
      accessor: 'updateTime',
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
      title: t('status'),
      accessor: 'status',
      customWidth: 120,
      Cell: ({ value, row }) => {
        const result = convertUTCToLocal(row.original.endTime)
        return (
          <div className="text-center">
            <Tag type={value === 'firing' ? 'error' : 'success'}>{t(value)}</Tag>
            {value === 'resolved' && (
              <span className="text-[10px] block text-gray-400">
                {t('resolvedOn')} {result}
              </span>
            )}
          </div>
        )
      },
    },
    {
      title: t('isValid'),
      accessor: 'isValid',
      customWidth: 160,
      Cell: (props) => {
        const { value, row } = props
        return !alertCheckId ? (
          workflowMissToast('alertCheckId')
        ) : ['unknown', 'skipped'].includes(value) ||
          (value === 'failed' && !row.original.workflowRunId) ? (
          <span className="text-gray-400 text-xs text-wrap [word-break:auto-phrase] text-center">
            {t(value)}
          </span>
        ) : (
          <Button
            type="link"
            className="text-xs text-wrap [word-break:auto-phrase] "
            size="small"
            onClick={() => {
              openResultModal(row.original.workflowRunId)
            }}
          >
            {t(value === 'failed' ? 'failedTo' : value)}
          </Button>
        )
      },
    },
    {
      title: <>{t('cause')}</>,
      accessor: 'cause',
      customWidth: 160,
      Cell: (props) => {
        const { workflowParams } = props.row.original
        return !workflowId ? (
          workflowMissToast('workflowId')
        ) : (
          <Button
            type="link"
            className="text-xs"
            size="small"
            onClick={() => {
              openWorkflowModal(workflowParams)
            }}
          >
            {t('viewWorkflow')}
          </Button>
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
          {workflowUrl && <WorkflowsIframe src={workflowUrl} />}
        </Modal>
      </div>
    </>
  )
}
export default AlertEventsPage
