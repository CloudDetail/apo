import { useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import { useSelector } from 'react-redux'
import StatusInfo from 'src/core/components/StatusInfo'
import BasicTable from 'src/core/components/Table/basicTable'
import TempCell from 'src/core/components/Table/TempCell'
import { selectSecondsTimeRange } from 'src/core/store/reducers/timeRangeReducer'
import { convertTime } from 'src/core/utils/time'

function InstanceTable({ data }) {
  const { t } = useTranslation('pro/dataViewCom')

  const { startTime, endTime } = useSelector(selectSecondsTimeRange)
  const column = [
    {
      title: t('instanceName'),
      accessor: 'name',
      customWidth: 150,
    },
    {
      title: t('averageResponseTime'),
      accessor: 'latency',
      minWidth: 170,
      Cell: (props) => {
        const { value } = props
        return <TempCell type="latency" data={value} timeRange={{ startTime, endTime }} />
      },
    },
    {
      title: t('errorRate'),
      accessor: 'errorRate',
      minWidth: 170,
      Cell: (props) => {
        const { value } = props
        return <TempCell type="errorRate" timeRange={{ startTime, endTime }} data={value} />
      },
    },
    {
      title: t('throughput'),
      accessor: 'tps',
      minWidth: 170,
      Cell: (props) => {
        const { value } = props
        return <TempCell type="tps" timeRange={{ startTime, endTime }} data={value} />
      },
    },
    {
      title: t('logErrorCount'),
      accessor: 'logs',
      minWidth: 170,
      Cell: (props) => {
        const { value } = props
        return <TempCell type="logs" timeRange={{ startTime, endTime }} data={value} />
      },
    },
    {
      title: t('infrastructureStatus'),
      accessor: 'infrastructureStatus',
      Cell: (props) => {
        const { value, row, column } = props
        const alertReason = row.original?.alertReason?.[column.id]
        return <StatusInfo status={value} alertReason={alertReason} title={column.title} />
      },
    },
    {
      title: t('networkQualityStatus'),
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
      title: t('k8sEventStatus'),
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
      title: t('hostNodeInfo'),
      accessor: 'nodeName',
      minWidth: 150,
      Cell: (props) => {
        return (
          <div>
            <div className="flex ">
              <span className="text-gray-400 flex-shrink-0 flex-grow-0">{t('hostName')}</span>
              {props.value}
            </div>
            <div className="flex">
              <span className="text-gray-400 flex-shrink-0 flex-grow-0">{t('hostIP')}</span>
              {props.row.original.nodeIP}
            </div>
          </div>
        )
      },
    },
    {
      title: t('lastDeploymentOrRestartTime'),
      accessor: `timestamp`,
      Cell: (props) => {
        const { value } = props
        return (
          <>
            {value !== null ? (
              convertTime(value, 'yyyy-mm-dd hh:mm:ss')
            ) : (
              <span className="text-slate-400">{t('na')}</span>
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
