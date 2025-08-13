import { Button, Table, Typography } from 'antd'
import ServiceRuleList from './ServiceMatcher'
import { useState } from 'react'
import { TbChevronDown } from 'react-icons/tb'
import { useTranslation } from 'react-i18next'
import { useDebounce } from 'react-use'
import { getServiceWithIstanceApi } from 'src/pro/api/dataplane'
import InstanceTag from './ServiceMatcher/InstanceTag'

const ServiceMatcherPage = ({
  timeRange,
  refreshTimeRange,
}: {
  timeRange: any
  refreshTimeRange: any
}) => {
  const [loading, setLoading] = useState(false)
  const [expandedRowKeys, setExpandedRowKeys] = useState<React.Key[]>([])
  const { t } = useTranslation('pro/dataplane')
  const { startTime, endTime, rangeType } = timeRange
  const [currentPage, setCurrentPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [total, setTotal] = useState(0)
  const columns = [
    {
      title: t('serviceNameColumn'),
      dataIndex: 'name',
      key: 'name',
      width: 220,
    },
    {
      title: t('matchingInstances'),
      dataIndex: 'instances',
      key: 'instances',
      render: (text: any) => {
        return (
          <Typography.Paragraph
            className="m-0 items-center flex flex-wrap w-full"
            ellipsis={{
              expandable: 'collapsible',
              rows: 2,
            }}
            onClick={(e) => {
              e.preventDefault()
            }}
          >
            {text?.map((item: any) => (
              <InstanceTag
                instanceName={item.instanceName}
                source={item.source?.type}
                key={item.instanceName}
              />
            ))}
          </Typography.Paragraph>
        )
      },
    },
    {
      title: t('actions'),
      dataIndex: 'action',
      key: 'action',
      width: 200,
      render: (text: string, record: any) => {
        return (
          <Button
            type="link"
            onClick={() => {
              setExpandedRowKeys([...expandedRowKeys, record.serviceName])
            }}
          >
            {t('matchOtherInstances')}
          </Button>
        )
      },
    },
  ]
  const [data, setData] = useState([])
  const getServiceWithIstance = () => {
    setLoading(true)
    getServiceWithIstanceApi({
      startTime,
      endTime,
      currentPage,
      pageSize,
    })
      .then((res) => {
        const responseData = res.data || res
        const totalPages = Math.ceil(responseData.total / pageSize)
        if (currentPage > totalPages && totalPages > 0) {
          setCurrentPage(1)
          return
        }
        setData(responseData.services)
        setTotal(responseData.total)
      })
      .catch((err) => {
        console.log(err)
      })
      .finally(() => {
        setLoading(false)
      })
  }
  useDebounce(
    () => {
      if (startTime && endTime) {
        getServiceWithIstance()
      }
    },
    300,
    [startTime, endTime, currentPage, pageSize],
  )
  const refresh = () => {
    if (rangeType) {
      refreshTimeRange?.()
    } else {
      getServiceWithIstance()
    }
  }
  return (
    <div>
      <Table
        rowKey={(record) => record.id}
        columns={columns}
        size="small"
        dataSource={data}
        loading={loading}
        pagination={{
          current: currentPage,
          pageSize: pageSize,
          total: total,
          onChange: (page, pageSize) => {
            setCurrentPage(page)
            setPageSize(pageSize)
          },
        }}
        scroll={{ y: 'calc(100vh - 300px)' }}
        expandable={{
          expandedRowKeys: expandedRowKeys,
          onExpandedRowsChange: (keys) => {
            setExpandedRowKeys([...keys])
          },
          expandedRowRender: (record) => (
            <ServiceRuleList serviceName={record.name} onRefresh={refresh} timeRange={timeRange} />
          ),
          rowExpandable: () => true,
          expandRowByClick: true,
          showExpandColumn: true,
          expandIcon: ({ record }) => (
            <Button
              type="link"
              onClick={(e) => {
                e.stopPropagation()
                const key = record.id
                const isExpanded = expandedRowKeys.includes(key)
                const newKeys = isExpanded
                  ? expandedRowKeys.filter((k) => k !== key)
                  : [...expandedRowKeys, key]
                setExpandedRowKeys(newKeys)
              }}
              icon={<TbChevronDown size={15} />}
            ></Button>
          ),
        }}
      />
    </div>
  )
}

export default ServiceMatcherPage
