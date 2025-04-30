import React, { useState, useEffect } from 'react'
import { Table, Pagination, Button, Modal, Input, Popconfirm } from 'antd'
import { getDetectionRecordApi, addMutationRule } from 'src/pro/api/alertAnalyze'
import { convertTime } from 'src/core/utils/time'
import { showToast } from 'src/core/utils/toast'
import '../index.css'
import { useTranslation } from 'react-i18next'

const i18nKey = 'detectionRecord'
export default function DetectionRecord({ refresh }) {
  const { t } = useTranslation('pro/alertAnalyze')

  const [dataSource, setDataSource] = useState([])
  const [currentPage, setCurrentPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [total, setTotal] = useState(0)
  const [loadingStatus, setLoadingStatus] = useState(false)
  const columns = [
    {
      title: t(`${i18nKey}.mutationDetectionName`),
      dataIndex: 'name',
      key: 'name',
      align: 'center',
    },
    {
      title: t(`${i18nKey}.groupName`),
      dataIndex: 'group',
      key: 'group',
      align: 'center',
    },
    {
      title: t(`${i18nKey}.detectionStatement`),
      dataIndex: 'mutationCheckPQL',
      key: 'mutationCheckPQL',
      align: 'center',
      render: (_, { mutationCheckPQL }) => {
        return (
          <Button
            type="link"
            onClick={() => {
              Modal.info({
                centered: true,
                content: (
                  <Input.TextArea
                    value={mutationCheckPQL}
                    readOnly
                    autoSize={{ minRows: 3, maxRows: 10 }}
                    style={{ backgroundColor: '#1F1F1F', color: '#FFFFFF' }}
                  />
                ),
                closable: false,
                okText: t(`${i18nKey}.close`),
                width: '60vw',
                icon: null,
                style: { backgroundColor: '#1F1F1F' },
              })
            }}
          >
            {t(`${i18nKey}.viewStatement`)}
          </Button>
        )
      },
    },
    {
      title: t(`${i18nKey}.queryStartTime`),
      dataIndex: 'startTime',
      key: 'startTime',
      align: 'center',
    },
    {
      title: t(`${i18nKey}.queryEndTime`),
      dataIndex: 'endTime',
      key: 'endTime',
      align: 'center',
    },
    {
      title: t(`${i18nKey}.detectionCycle`),
      dataIndex: 'step',
      key: 'step',
      align: 'center',
    },
    {
      title: t(`${i18nKey}.alertWaitTime`),
      dataIndex: 'for',
      key: 'for',
      align: 'center',
    },
    {
      title: t(`${i18nKey}.executionTime`),
      dataIndex: 'timestamp',
      key: 'timestamp',
      align: 'center',
    },
  ]

  async function getData() {
    const params = {
      currentPage,
      pageSize,
    }
    try {
      setLoadingStatus(true)
      const { pagination, mutatedRule } = await getDetectionRecordApi(params)
      setDataSource(handleData(mutatedRule))
      setLoadingStatus(false)
      setTotal(pagination.total)
    } catch (err) {
      setLoadingStatus(false)
      console.log(err)
    }
  }

  function pageChange(page, newPageSize) {
    if (pageSize !== newPageSize) {
      setCurrentPage(1) // 当页数变化时跳转到第一页
      setPageSize(newPageSize) // 更新新的每页条数
    } else {
      setCurrentPage(page) // 如果页数未变，直接跳转到目标页
    }
  }

  function handleData(data) {
    data = data.map((item) => {
      return {
        name: item.name,
        mutationCheckPQL: item.mutationCheckPQL,
        for: item.for,
        startTime: convertTime(item.startTime, 'yyyy-mm-dd hh:mm:ss'),
        endTime: convertTime(item.endTime, 'yyyy-mm-dd hh:mm:ss'),
        step: convertTime(item.step, 's'),
        synchronize_to_alert_rules: item.synchronize_to_alert_rules,
        group: item.group,
        originItem: item,
        timestamp: convertTime(item.timestamp, 'yyyy-mm-dd hh:mm:ss'),
      }
    })
    return data
  }
  useEffect(() => {
    getData()
  }, [currentPage, pageSize, refresh])
  useEffect(() => {
    setCurrentPage(1)
  }, [refresh])
  return (
    <>
      <div className="flex flex-col h-full relative overflow-y-hidden">
        <div className="flex flex-col" style={{ flex: 1, overflowY: 'auto' }}>
          <div className="text-base px-4">{t(`${i18nKey}.anomalyDetectionRecord`)}</div>
          <Table
            columns={columns}
            dataSource={dataSource}
            className="w-full"
            pagination={false}
            loading={loadingStatus}
          />
        </div>
        <div className="absolute bottom-0 right-0 w-full p-2 bg-[#1F1F1F] flex justify-end">
          <Pagination
            defaultCurrent={1}
            total={total}
            current={currentPage}
            pageSize={pageSize}
            pageSizeOptions={[10, 20, 30]}
            onChange={pageChange}
            showQuickJumper={true}
            showSizeChanger={true}
            style={{ zIndex: 10, backgroundColor: '#1F1F1F' }}
          />
        </div>
      </div>
    </>
  )
}
