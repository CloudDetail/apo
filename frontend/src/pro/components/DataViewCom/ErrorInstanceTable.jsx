import { useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import { useSelector } from 'react-redux'
import DelayLineChart from 'src/core/components/Chart/DelayLineChart'
import ErrorCell from 'src/core/components/ErrorInstance/ErrorCell'
import BasicTable from 'src/core/components/Table/basicTable'
import Timeline from 'src/core/components/TimeLine'
import { selectSecondsTimeRange } from 'src/core/store/reducers/timeRangeReducer'

function ErrorInstanceTable({ data }) {
  const { t } = useTranslation('pro/dataViewCom')

  const { startTime, endTime } = useSelector(selectSecondsTimeRange)
  const column = [
    {
      title: t('instanceName'),
      accessor: 'name',
      customWidth: 150,
    },
    {
      title: t('currentNodeError'),
      accessor: 'propations',
      Cell: (props) => {
        const { value, row } = props

        return <ErrorCell data={value} instance={row.values.name} />
      },
    },

    {
      title: t('logErrorCount'),
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
      title: t('logInfo'),
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
