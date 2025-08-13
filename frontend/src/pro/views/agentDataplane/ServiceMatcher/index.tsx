import { Button, Space, Table, Typography } from 'antd'
import { useState } from 'react'
import ConfigModal from './ConfigModal'
import { deleteServiceNameRuleApi, listServiceNameRuleApi } from 'src/pro/api/dataplane'
import { useDebounce } from 'react-use'
import InstanceTag from './InstanceTag'
import { useTranslation } from 'react-i18next'
import { notify } from 'src/core/utils/notify'

const ServiceRuleList = ({
  serviceName,
  onRefresh,
  timeRange,
}: {
  serviceName: string
  onRefresh: () => void
  timeRange: any
}) => {
  const [data, setData] = useState([])
  const [open, setOpen] = useState(false)
  const [info, setInfo] = useState(null)
  const { t } = useTranslation('pro/dataplane')
  const [loading, setLoading] = useState(true)
  const openModal = (info: any) => {
    setOpen(true)
    setInfo(info)
  }
  const deleteRule = (id: string) => {
    deleteServiceNameRuleApi({ ruleId: id }).then(() => {
      notify({
        type: 'success',
        message: t('deleteSuccess'),
      })
      onRefresh()
    })
  }
  const columns = [
    {
      title: 'id',
      dataIndex: 'id',
      hidden: true,
    },
    {
      title: t('conditions'),
      dataIndex: 'conditions',
      width: '30%',
      render: (text: any) => {
        return (
          <div className="flex flex-wrap gap-1">
            {text?.map((item: any, index: number) => {
              return (
                <>
                  {index > 0 && <span className="text-[var(--ant-blue)] mx-1"> {t('and')}</span>}
                  <span className="mx-1  text-[var(--ant-color-text-base)]">{item.key}</span>
                  <span className="text-[var(--ant-blue)]">{t(item.matchType)} </span>
                  <span className="text-[var(--ant-lime-7)] mx-1">"{item.value}"</span>
                </>
              )
            })}
          </div>
        )
      },
    },
    {
      title: t('instances'),
      dataIndex: 'instances',
      render: (text: any) => {
        return (
          <Typography.Paragraph
            className="m-0 items-center flex flex-wrap w-full"
            ellipsis={{
              rows: 2,
              expandable: 'collapsible',
            }}
          >
            {text?.map((item: any) => <InstanceTag instanceName={item} source="rule" />)}
          </Typography.Paragraph>
        )
      },
    },
    {
      title: t('actions'),
      dataIndex: 'action',
      width: '200px',
      render: (text: string, record: any) => {
        return (
          <Space>
            <Button
              type="link"
              onClick={() => {
                deleteRule(record.id)
              }}
            >
              {t('delete')}
            </Button>
            <Button type="link" onClick={() => openModal(record)}>
              {t('edit')}
            </Button>
          </Space>
        )
      },
    },
  ]
  const getServiceRuleList = () => {
    setLoading(true)
    listServiceNameRuleApi({
      serviceName,
      startTime: timeRange?.startTime,
      endTime: timeRange?.endTime,
    })
      .then((res: any) => {
        setData(res.data?.rules || [])
      })
      .finally(() => {
        setLoading(false)
      })
  }
  useDebounce(
    () => {
      if (serviceName && timeRange?.startTime && timeRange?.endTime) {
        console.log('endTime', timeRange.endTime)
        getServiceRuleList()
      }
    },
    300,
    [serviceName, timeRange],
  )
  const handleRefresh = () => {
    if (timeRange?.rangeType) {
      onRefresh()
    } else {
      getServiceRuleList()
    }
  }
  return (
    <div className="px-4 max-h-[300px] overflow-y-auto">
      <div className="flex justify-between">
        <div className="text-sm font-bold flex-1">
          {t('serviceMatchInstanceRules', { serviceName })}
        </div>
        <div className="">
          <Button type="primary" onClick={() => openModal(null)}>
            {t('addRule')}
          </Button>
        </div>
      </div>
      <Table
        loading={loading}
        columns={columns}
        dataSource={data}
        size="small"
        pagination={false}
        scroll={{ y: 300 }}
      />
      <ConfigModal
        serviceName={serviceName}
        open={open}
        onCancel={() => setOpen(false)}
        info={info}
        onRefresh={handleRefresh}
      />
    </div>
  )
}

export default ServiceRuleList
