import { Popover, Table } from 'antd'
import { useTranslation } from 'react-i18next'
import DelayLineChart from 'src/core/components/Chart/DelayLineChart'
const type = ['errorRate', 'latency']
const AbnormalityDetection = ({ args }: { args: any }) => {
  const { columns, data } = args
  const { t } = useTranslation('pro/agent')
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
              <Popover
                content={
                  <div className="text-center">
                    <span>{t(record.type)}</span>
                    {type.map((key) => {
                      return (
                        item[key] && (
                          <div
                            key={key}
                            onClickCapture={(e) => e.stopPropagation()}
                            className="h-52 w-64"
                          >
                            <DelayLineChart
                              data={item[key]?.chartData}
                              type={key}
                              allowTimeBrush={false}
                              needFillZero={false}
                            />
                          </div>
                        )
                      )
                    })}
                  </div>
                }
                key={item.service}
              >
                <span>
                  <span className="underline cursor-pointer text-[var(--ant-color-primary)]">
                    {item.service}
                  </span>{' '}
                  {index === value.length - 1 ? '' : ','}
                </span>
              </Popover>
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
    </>
  )
}

export default AbnormalityDetection
