import { useMemo } from 'react'
import { useSelector } from 'react-redux'
import StatusInfo from 'src/core/components/StatusInfo'
import BasicTable from 'src/core/components/Table/basicTable'
import TempCell from 'src/core/components/Table/TempCell'
import { selectSecondsTimeRange } from 'src/core/store/reducers/timeRangeReducer'
import { convertTime } from 'src/core/utils/time'

function InstanceTable({ data }) {
  const { startTime, endTime } = useSelector(selectSecondsTimeRange)
  const column = [
    {
      title: '实例名',
      accessor: 'name',
      customWidth: 150,
    },
    {
      title: '平均响应时间',
      accessor: 'latency',
      minWidth: 170,
      Cell: (props) => {
        const { value } = props
        return <TempCell type="latency" data={value} timeRange={{ startTime, endTime }} />
      },
    },
    {
      title: '错误率',
      accessor: 'errorRate',
      minWidth: 170,

      Cell: (props) => {
        const { value } = props
        return <TempCell type="errorRate" timeRange={{ startTime, endTime }} data={value} />
      },
    },
    {
      title: '吞吐量',
      accessor: 'tps',

      minWidth: 170,
      Cell: (props) => {
        const { value } = props
        return <TempCell type="tps" timeRange={{ startTime, endTime }} data={value} />
      },
    },
    {
      title: '日志错误数量',
      accessor: 'logs',

      minWidth: 170,
      Cell: (props) => {
        const { value } = props
        return <TempCell type="logs" timeRange={{ startTime, endTime }} data={value} />
      },
    },
    {
      title: '基础设施状态',
      accessor: 'infrastructureStatus',
      Cell: (props) => {
        const { value, row, column } = props
        const alertReason = row.original?.alertReason?.[column.id]
        return (
          <>
            <StatusInfo status={value} alertReason={alertReason} title={column.title} />
          </>
        )
      },
    },
    {
      title: '网络质量状态',
      accessor: 'netStatus',
      Cell: (props) => {
        const { value, row, column } = props
        const alertReason = row.original?.alertReason?.[column.id]
        return (
          <>
            <StatusInfo status={value} alertReason={alertReason} title={column.title} />
          </>
        )
      },
    },
    {
      title: 'K8s事件状态',
      accessor: 'k8sStatus',
      Cell: (props) => {
        const { value, row, column } = props
        const alertReason = row.original?.alertReason?.[column.id]
        return (
          <>
            <StatusInfo status={value} alertReason={alertReason} title={column.title} />
          </>
        )
      },
    },

    {
      title: '主机节点信息',
      accessor: 'nodeName',
      minWidth: 150,
      Cell: (props) => {
        return (
          <div>
            <div className="flex ">
              <span className="text-gray-400 flex-shrink-0 flex-grow-0">主机名：</span>
              {props.value}
            </div>
            <div className="flex">
              <span className="text-gray-400 flex-shrink-0 flex-grow-0">主机IP：</span>
              {props.row.original.nodeIP}
            </div>
          </div>
        )
      },
    },
    {
      title: '末次部署或重启时间',
      accessor: `timestamp`,
      Cell: (props) => {
        const { value } = props
        return (
          <>
            {value !== null ? (
              convertTime(value, 'yyyy-mm-dd hh:mm:ss')
            ) : (
              <span className="text-slate-400">N/A</span>
            )}
          </>
        )
      },
    },
  ]
  const tableProps = useMemo(() => {
    return {
      columns: column,
      data: data,
      showBorder: false,
      loading: false,
    }
  }, [data])
  return (
    <div className="w-full h-full overflow-auto">
      <div className="w-full h-full text-xs min-w-[1200px]">
        <BasicTable {...tableProps} />
      </div>
    </div>
  )
}
export default InstanceTable
