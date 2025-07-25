import { Modal, Table } from 'antd'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import DelayLineChart from 'src/core/components/Chart/DelayLineChart'

const AbnormalityDetection = ({ args }: { args: any }) => {
  const [open, setOpen] = useState(false)
  const { columns, data } = args
  const { t } = useTranslation('pro/agent')
  const [selectedServiceInfo, setSelectedServiceInfo] = useState(null)
  const defaultColumns = [
    {
      title: 'type',
      dataIndex: 'type',
    },
    {
      title: '服务列表',
      dataIndex: 'services',
      render: (value, record) => {
        return (
          <div>
            {value?.map((item, index) => (
              <span
                key={item.service}
                onClick={() => {
                  setOpen(true)
                  setSelectedServiceInfo({
                    service: item.service,
                    data: {
                      latency: item.latency,
                      errorRate: item.errorRate,
                    },
                  })
                }}
              >
                <span className="underline cursor-pointer text-[var(--ant-color-primary)]">
                  {item.service}
                </span>{' '}
                {index === value.length - 1 ? '' : ','}
              </span>
            ))}
          </div>
        )
      },
    },
  ]
  return (
    <>
      <Table
        columns={defaultColumns.filter((item) => columns?.includes(item.dataIndex))}
        dataSource={data}
        size="small"
      />

      <Modal
        title={selectedServiceInfo?.service}
        open={open}
        onCancel={() => setOpen(false)}
        width={'60%'}
      >
        <div className="flex items-center gap-2 h-60">
          {selectedServiceInfo?.data &&
            Object.entries(selectedServiceInfo?.data)?.map(([key, value]) => {
              return (
                <div key={key} className="flex-1 h-52 w-80">
                  <div className="text-lg font-bold">{t(key)}</div>
                  {/* <Card type="inner" title={t(key)} className="h-52 w-80"> */}
                  <DelayLineChart
                    data={value?.chartData}
                    type={key}
                    allowTimeBrush={false}
                    needFillZero={false}
                  />
                  {/* </Card> */}
                </div>
              )
            })}
        </div>
      </Modal>
    </>
  )
}

export default AbnormalityDetection
