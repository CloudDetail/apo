import React, { useState, useEffect, useRef } from 'react'
import { Flex, Pagination, Table, Button, Modal, Input, Popconfirm } from 'antd'
import { addMutationRule, getMutationList } from 'src/pro/api/alertAnalyze'
import Tag from 'src/core/components/Tag/Tag'
import '../index.css'
import { showToast } from 'src/core/utils/toast'
import { convertTime } from 'src/core/utils/time'

export default function AlertRule({ startTime, endTime, step, setRefresh, refresh }) {
  const columns = [
    {
      title: '突变检测名',
      dataIndex: 'name',
      key: 'name',
      align: 'center',
    },
    {
      title: '组名',
      dataIndex: 'group',
      key: 'group',
      align: 'center',
    },
    {
      title: '检测语句',
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
                okText: '关闭',
                width: '60vw',
                icon: null,
                style: { backgroundColor: '#1F1F1F' },
              })
            }}
          >
            点击查看检测语句
          </Button>
        )
      },
    },
    {
      title: '检测周期(s)',
      dataIndex: 'step',
      key: 'step',
      align: 'center',
    },
    {
      title: '告警等待时间',
      dataIndex: 'for',
      key: 'for',
      align: 'center',
    },
    {
      title: '是否同步到告警规则',
      dataIndex: 'synchronizeToAlertRules',
      key: 'synchronizeToAlertRules',
      align: 'center',
      render: (data) => {
        console.log(data)
        return data ? <p>是</p> : <p>否</p>
      },
    },
    {
      title: '操作',
      dataIndex: 'originItem',
      align: 'center',
      render: (data) => {
        return (
          <Popconfirm
            title="再次执行"
            description="确定再次执行此突变检测吗"
            onConfirm={() => addAgain(data)}
            okText="确认"
            cancelText="取消"
          >
            <Button type="link">再次执行</Button>
          </Popconfirm>
        )
      },
    },
  ]
  const [dataSource, setDataSource] = useState([])
  const [currentPage, setCurrentPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [total, setTotal] = useState(0)
  const [loadingStatus, setLoadingStatus] = useState(false)

  async function getData() {
    const params = {
      currentPage,
      pageSize,
    }
    try {
      setLoadingStatus(true)
      const { pagination, alertRule } = await getMutationList(params)
      setDataSource(handleData(alertRule))
      setTotal(pagination.total)
      setLoadingStatus(false)
    } catch (error) {
      setLoadingStatus(false)
      throw error
    }
  }

  async function addAgain(params) {
    const newParams = {
      ...params,
      mutationCheckPQL: {
        customMetric: params.mutationCheckPQL,
        group: params.group,
      },
      startTime: startTime,
      endTime: endTime,
      step: step,
      synchronizeToAlertRules: false,
    }
    delete newParams.group
    try {
      const result = await addMutationRule(newParams)
      if (result == 'ok') {
        showToast({
          title: '再次执行突变检测成功',
          color: 'success',
        })
        setCurrentPage(1)
        getData()
        setRefresh(!refresh)
      }
    } catch (err) {
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
        synchronizeToAlertRules: item.synchronizeToAlertRules,
        group: item.group,
        originItem: item,
        timestamp: convertTime(item.timestamp, 'yyyy-mm-dd hh:mm:ss'),
      }
    })
    return data
  }

  useEffect(() => {
    getData()
  }, [currentPage, pageSize])

  return (
    <>
      <div className="flex flex-col h-full justify-between relative overflow-y-hidden">
        <div className="flex flex-col" style={{ flex: 1, overflowY: 'auto' }}>
          <div className="text-base px-4">异常检测类型</div>
          <Table
            loading={loadingStatus}
            className="w-full"
            columns={columns}
            dataSource={dataSource}
            pagination={false}
          />
        </div>
        <Pagination
          defaultCurrent={1}
          total={total}
          current={currentPage}
          pageSize={pageSize}
          className="absolute bottom-0 right-0 p-2 w-full flex justify-end"
          pageSizeOptions={[10, 20, 30]}
          onChange={pageChange}
          showQuickJumper={true}
          showSizeChanger={true}
          style={{ marginTop: '20px', zIndex: 10, backgroundColor: '#1F1F1F' }}
        />
      </div>
    </>
  )
}
