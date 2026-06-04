import React, { useEffect, useState } from 'react'
// import DataTable from './dataTable';
import './tableStep.scss'
import tooltipPng3 from 'src/core/assets/images/tooltip3.png'
import { Card, Space, Tooltip, Tag, Table } from 'antd'
import { formatUnit } from 'src/core/utils/time'
import { BsQuestionCircle } from 'react-icons/bs'
import { useTranslation } from 'react-i18next'
import { t } from 'i18next'
const i18nKey = 'tableStep'

const staticColumns = [
  {
    title: t(`pro/causeReport:${i18nKey}.label`),
    accessor: 'labels',
    dataIndex: 'labels',
    justifyContent: 'flex-start',
    customWidth: '80%',
    render: (value, record) => {
      return (
        <div>
          {Object.keys(record.labels).map((key, idx) => (
            <Tag key={idx} color="blue">
              {key}:{record.labels[key]}
            </Tag>
          ))}
        </div>
      )
    },
    Cell: (cell: any) => {
      const { original } = cell.cell.row
      return (
        <div>
          {Object.keys(original.labels).map((key, idx) => (
            <Tag key={idx} color="blue">
              {key}:{original.labels[key]}
            </Tag>
          ))}
        </div>
      )
    },
  },
]

interface TData {
  name: string
  baseline: number
  baseAction: string
  current: number
  currentAction: string
  unit: any
  warn: boolean
  labels: {
    [key in string]: any
  }
}
interface Props {
  data: any
}
function TableStep({ data }: Props) {
  const { t } = useTranslation('pro/causeReport')

  const [tableList, setTableList] = useState<any[]>([])
  const requestEndpointTitles = [t('requestEndpoint'), '请求端点', 'Request Endpoint']

  const isRequestEndpointColumn = (title: string) => requestEndpointTitles.includes(title)

  const buildTableColumn = (header: any) => {
    const column: any = {
      title: header.title,
      accessor: header.accessor,
      dataIndex: header.accessor,
    }
    if (isRequestEndpointColumn(header.title)) {
      column.width = '55%'
      column.render = (value: any) => <div className="table_step_request_endpoint">{value}</div>
      column.onCell = () => ({
        className: 'table_step_request_endpoint_cell',
      })
    }
    return column
  }

  const dataHandle = () => {
    let metrics = Object.keys(data.metrics)
    const columnList: any = [...staticColumns]
    const tableList: any[] = []
    if (data.type === 'metrics_baseline') {
      columnList.splice(
        1,
        0,
        ...[
          {
            title: t(`${i18nKey}.historicalBaseline`),
            accessor: 'baseline',
            dataIndex: 'baseline',
            render: (value, record) => {
              if (record.baseAction) {
                return (
                  <Tooltip title={record.baseAction}>
                    <span>{record.unit ? formatUnit(value, record.unit) : value}</span>
                  </Tooltip>
                )
              } else {
                return <span>{record.unit ? formatUnit(value, record.unit) : value}</span>
              }
            },
          },
          {
            title: t(`${i18nKey}.faultInvocation`),
            accessor: 'current',
            dataIndex: 'current',
            render: (value, record) => {
              if (record.currentAction) {
                return (
                  <Tooltip title={record.currentAction}>
                    <span className={record.warn ? 'a_error' : ''}>
                      {record.unit ? formatUnit(value, record.unit) : value}
                    </span>
                  </Tooltip>
                )
              } else {
                return (
                  <span className={record.warn ? 'a_error' : ''}>
                    {record.unit ? formatUnit(value, record.unit) : value}
                  </span>
                )
              }
            },
          },
        ],
      )
    }
    metrics.forEach((metric) => {
      let columns: any = [...columnList]
      let list: TData[] = []
      if (data.type === 'tableMetrics') {
        columns.push({
          title: metric,
          accessor: 'current',
          dataIndex: 'current',
          customWidth: '20%',

          render: (value, record) => {
            if (record.currentAction) {
              return (
                <Tooltip title={record.currentAction}>
                  <span className={record.warn ? 'a_error' : ''}>
                    {record.unit ? formatUnit(value, record.unit) : value}
                  </span>
                </Tooltip>
              )
            } else {
              return (
                <span className={record.warn ? 'a_error' : ''}>
                  {record.unit ? formatUnit(value, record.unit) : value}
                </span>
              )
            }
          },
          Cell: (cell: any) => {
            const { value: text } = cell
            const { original: record } = cell.cell.row
            if (record.currentAction) {
              return (
                <Tooltip title={record.currentAction}>
                  <span className={record.warn ? 'a_error' : ''}>
                    {record.unit ? formatUnit(text, record.unit) : text}
                  </span>
                </Tooltip>
              )
            } else {
              return (
                <span className={record.warn ? 'a_error' : ''}>
                  {record.unit ? formatUnit(text, record.unit) : text}
                </span>
              )
            }
          },
        })
      }
      if (data.type === 'table') {
        columns = []
        data.metrics[metric].header.forEach((header: any) => {
          columns.push(buildTableColumn(header))
          list = data.metrics[metric].rows
        })
      } else {
        data.metrics[metric].forEach((item: any) => {
          list.push({
            name: metric,
            baseline: item.baseline?.value,
            baseAction: item.baseline?.action?.text || '',
            current: item.current.value,
            currentAction: item.current?.action?.text || '',
            unit: item.unit,
            warn: item.is_warn,
            labels: item.labels,
          })
        })
      }

      tableList.push({
        columns: columns,
        tableData: list,
        title: metric,
        hasRequestEndpointColumn: columns.some((column: any) => isRequestEndpointColumn(column.title)),
      })
    })
    setTableList(tableList)
  }

  useEffect(() => {
    dataHandle()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [data])

  return (
    <Card
      title={
        <div className="flex align-items-center w-full">
          {data.report_name === t(`${i18nKey}.externalCallDetails`) ? (
            <Tooltip
              classNames={{
                body: 'w-[600px]',
              }}
              title={
                <div className="operate_tooltip_img relative w-full">
                  <img src={tooltipPng3} className="operate_tooltip_img" />
                </div>
              }
            >
              <Space>
                <BsQuestionCircle className="operate_tooltip_icon" />
                {data.report_name}
              </Space>
            </Tooltip>
          ) : (
            data.report_name
          )}
        </div>
      }
    >
      <div className="table_step_description">
        {tableList.map((table, index) => {
          return tableList?.length > 1 ? (
            <Card
              type="inner"
              title={table.title}
              key={index}
              classNames={{ body: 'p-1' }}
              className="mb-2"
            >
              <Table
                size="small"
                dataSource={table.tableData}
                columns={table.columns}
                tableLayout={table.hasRequestEndpointColumn ? 'fixed' : undefined}
                className="table_step_table"
              />
            </Card>
          ) : (
            <Table
              size="small"
              dataSource={table.tableData}
              columns={table.columns}
              tableLayout={table.hasRequestEndpointColumn ? 'fixed' : undefined}
              className="table_step_table"
            />
          )
        })}
        {data?.conclusion?.length > 0 && (
          <div className="conclusion_div">
            <span>{t(`${i18nKey}.analysisConclusion`)}：</span>
            <span className="conclusion_tag">{data.conclusion}</span>
          </div>
        )}
      </div>
    </Card>
  )
}

export default TableStep
