/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Button, Modal, Tag as AntdTag, Tooltip, Statistic, Checkbox, Image, Card } from 'antd'
import { useEffect, useMemo, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useSelector } from 'react-redux'
import { useNavigate } from 'react-router-dom'
import { getAlertEventsApi } from 'src/core/api/alerts'
import BasicTable from 'src/core/components/Table/basicTable'
import { convertUTCToBeijing } from 'src/core/utils/time'
import WorkflowsIframe from '../workflows/workflowsIframe'
import Tag from 'src/core/components/Tag/Tag'
import PieChart from './PieChart'
import CountUp from 'react-countup'
import filterSvg from 'core/assets/images/filter.svg'
import ReactJson from 'react-json-view'
function isJSONString(str) {
  try {
    JSON.parse(str)
    return true
  } catch (e) {
    return false
  }
}
const Filter = () => {
  const { t } = useTranslation('oss/alertEvents')

  const statusOptions = [
    { label: <Tag type={'error'}>{t('firing')}</Tag>, value: 'firing' },
    { label: <Tag type={'success'}>{t('resolved')}</Tag>, value: 'resolved' },
  ]
  const validOptions = [
    { label: t('valid'), value: 'valid' },
    { label: t('invalid'), value: 'invalid' },
    { label: t('other'), value: 'other' },
  ]
  return (
    <div className="flex pb-2 ">
      <div>
        告警状态：
        <Checkbox.Group
          // onChange={onChangeTypeList}
          options={statusOptions}
        ></Checkbox.Group>
      </div>
      <div>
        告警有效性：
        <Checkbox.Group
          // onChange={onChangeTypeList}
          options={validOptions}
        ></Checkbox.Group>
      </div>
    </div>
  )
}
const formatter = (value) => <CountUp end={value as number} separator="," />
const StatusPanel = () => {
  const { t } = useTranslation('oss/alertEvents')

  const chartData = [
    { name: t('firing'), value: 1048, type: 'error' },
    { name: t('resolved'), value: 735, type: 'success' },
  ]
  return (
    <div className="flex pb-2 h-full flex-1  ">
      <div className="w-full ml-1 rounded-xl flex h-full bg-[#141414] p-2">
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
const ExtraPanel = () => {
  return (
    <div className=" pb-2 h-full  shrink-0 w-1/2 mr-3">
      <div className="w-full rounded-xl flex h-full bg-[#141414] p-2 ">
        <div className="ml-3 mr-7">
          <Image src={filterSvg} width={50} height={'100%'} preview={false} />
        </div>
        <div className="flex flex-col h-full justify-center">
          <Statistic
            className=" flex flex-col justify-center"
            title={<span className="text-white">{'告警降噪率'}</span>}
            value={30}
            precision={2}
            suffix="%"
            formatter={formatter}
          />
          {/* <span className="text-gray-400 text-xs">AI辅助识别无效告警，助力聚焦核心问题</span> */}
          <span className="text-gray-400 text-xs">
            在
            <span className="mx-1">
              <Tag type={'error'}>1048</Tag>
            </span>
            条告警中，AI辅助识别{' '}
            <span className="mx-1">
              <Tag type={'warning'}>839</Tag>
            </span>
            条为无效告警，主动降噪
          </span>
        </div>
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
  const timerRef = useRef(null)
  const workflowMissToast = (type: 'alertCheckId' | 'workflowId') => {
    return (
      <Tooltip title={type === 'alertCheckId' ? t('missToast1') : t('missToast2')}>
        <span className="text-gray-400 text-xs">{t('workflowMiss')}</span>
      </Tooltip>
    )
  }
  const getAlertEvents = () => {
    getAlertEventsApi({
      startTime,
      endTime,
      pagination: {
        currentPage: pagination.pageIndex,
        pageSize: pagination.pageSize,
      },
    }).then((res) => {
      const totalPages = Math.ceil(res.pagination.total / pagination.pageSize)
      if (pagination.pageIndex > totalPages && totalPages > 0) {
        setPagination({ ...pagination, pageIndex: totalPages })
        return
      }
      setAlertEvents(res?.events || [])
      setPagination({
        ...pagination,
        total: res?.pagination.total || 0,
      })
      setWorkflowId(res.alertEventAnalyzeWorkflowId)
      setAlertCheckId(res.alertCheckId)
      if (timerRef.current) {
        clearTimeout(timerRef.current)
      }

      timerRef.current = setTimeout(
        () => {
          getAlertEvents()
        },
        5 * 60 * 1000,
      )
    })
  }

  useEffect(() => {
    if (startTime && endTime) {
      getAlertEvents()
    }
    return () => {
      if (timerRef.current) {
        clearTimeout(timerRef.current)
      }
    }
  }, [pagination.pageIndex, pagination.pageSize, startTime, endTime])
  function openWorkflowModal(workflowParams) {
    let result = '/dify/app/' + workflowId + '/run?'
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
            {isJSONString(detail) && (
              <Button
                color="primary"
                variant="text"
                size="small"
                onClick={() => setVisible(!visible)}
              >
                {visible ? '收起' : '更多'}
              </Button>
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
        const result = convertUTCToBeijing(value)
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
        const result = convertUTCToBeijing(row.original.endTime)
        return (
          <div className="text-center">
            <Tag type={value === 'firing' ? 'error' : 'success'}>{t(value)}</Tag>
            {value === 'resolved' && (
              <span className="text-[10px] block text-gray-400">解决于{result}</span>
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
  const updatePagination = (pagination) => setPagination({ ...pagination, ...pagination })
  const changePagination = (page, pageSize) => {
    updatePagination({
      pageSize: pageSize,
      pageIndex: page,
      total: pagination.total,
    })
  }
  const tableProps = useMemo(() => {
    return {
      columns: columns,
      data: alertEvents,
      showBorder: false,
      loading: false,
    }
  }, [alertEvents, pagination.pageIndex, pagination.pageSize, pagination.total])
  const chartHeight = 150
  const headHeight =
    (import.meta.env.VITE_APP_CODE_VERSION === 'CE' ? 60 : 100) + chartHeight + 'px'
  return (
    <>
      <div className=" overflow-hidden">
        <div style={{ height: chartHeight }} className="flex">
          <ExtraPanel />
          <StatusPanel />
        </div>
        <Card
          style={{
            height: 'calc(100vh - ' + headHeight + ')',
          }}
        >
          <Filter />

          <BasicTable {...tableProps} />
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
