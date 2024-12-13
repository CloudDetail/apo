import { useMemo } from 'react'
import { useSelector } from 'react-redux'
import DelayLineChart from 'src/core/components/Chart/DelayLineChart'
import ErrorCell from 'src/core/components/ErrorInstance/ErrorCell'
import { ErrorChain } from 'src/core/components/ErrorInstance/ErrorChain'
import BasicTable from 'src/core/components/Table/basicTable'
import Timeline from 'src/core/components/TimeLine'
import { selectSecondsTimeRange } from 'src/core/store/reducers/timeRangeReducer'

function ErrorInstanceTable({ data }) {
  const { startTime, endTime } = useSelector(selectSecondsTimeRange)
  const column = [
    {
      title: '实例名',
      accessor: 'name',
      customWidth: 150,
    },
    {
      title: '当前节点异常',
      accessor: 'propations',
      Cell: (props) => {
        const { value, updateSelectedValue, column } = props
        // updateSelectedValue(column.id, value[0]);
        const update = (item) => {
          updateSelectedValue(column.id, item)
        }

        return <ErrorCell data={value} update={update} />
      },
    },
    {
      title: '错误故障传播链',
      accessor: 'chain',
      dependsOn: 'propations', // 该列依赖于第一列的值
      Cell: (props) => {
        const { value, dependsValue, originalRow } = props
        return (
          <ErrorChain
            data={dependsValue?.customAbbreviation}
            instance={originalRow.name}
            chartId={originalRow.name}
          />
        )
        // return  nodes?.length > 0 ? <Topology canZoom={false} data={{nodes, edges}}  /> : <Empty />
      },
    },

    {
      title: '日志错误数量',
      accessor: 'logs',
      Cell: (props) => {
        const { value } = props
        return (
          <div className="h-52 w-80">
            <DelayLineChart data={value} timeRange={{ startTime, endTime }} type="logs" />
          </div>
        )
      },
    },
    {
      title: '日志信息',
      accessor: 'detail',
      customWidth: 320,
      Cell: (props) => {
        const { value, row } = props
        return (
          <Timeline
            instance={row.original.name}
            nodeName={row.original.nodeName}
            pid={row.original.pid}
            containerId={row.original.containerId}
            type="errorLogs"
            startTime={startTime}
            endTime={endTime}
          />
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
      <div className="w-full h-full text-xs min-w-[1000px]">
        <BasicTable {...tableProps} />
      </div>
    </div>
  )
}
export default ErrorInstanceTable
