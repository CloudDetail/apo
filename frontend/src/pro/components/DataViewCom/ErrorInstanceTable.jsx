import { useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import { useSelector } from 'react-redux'
import DelayLineChart from 'src/core/components/Chart/DelayLineChart'
import ErrorCell from 'src/core/components/ErrorInstance/ErrorCell'
import BasicTable from 'src/core/components/Table/basicTable'
import Timeline from 'src/core/components/TimeLine'
import { selectSecondsTimeRange } from 'src/core/store/reducers/timeRangeReducer'

function ErrorInstanceTable({ data, service, endpoint }) {
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
      customWidth: 220,
      Cell: (props) => {
        const { value } = props
        return (
          <div className=" w-full h-[200px] relative">
            <DelayLineChart data={value} timeRange={{ startTime, endTime }} type="logs" />
          </div>
        )
      },
    },
    {
      title: t('logInfo'),
      accessor: 'detail',
      customWidth: 280,
      Cell: (props) => {
        const { value, row } = props
        return (
          <Timeline
            instance={row.original.name}
            nodeName={row.original.nodeName}
            pid={row.original.pid}
            containerId={row.original.containerId}
            type="errorLogs"
            service={service}
            endpoint={endpoint}
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
      scrollY: 450,
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
export default ErrorInstanceTable
