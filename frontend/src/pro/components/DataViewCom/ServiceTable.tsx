import { useNavigate } from 'react-router-dom'
import { CButton } from '@coreui/react'
import { Tooltip } from 'antd'
import { AiOutlineInfoCircle } from 'react-icons/ai'
import { DelaySourceTimeUnit } from 'src/constants'
import TempCell from 'src/core/components/Table/TempCell'
import BasicTable from 'src/core/components/Table/basicTable'
import { convertTime } from 'src/core/utils/time'
import StatusInfo from 'src/core/components/StatusInfo'
import { useSelector } from 'react-redux'
import { selectSecondsTimeRange } from 'src/core/store/reducers/timeRangeReducer'
import { useMemo } from 'react'
import { useTranslation } from 'react-i18next'

export default function ServiceTable({ data }) {
  const { t } = useTranslation('pro/dataViewCom')

  const { startTime, endTime } = useSelector(selectSecondsTimeRange)
  const column = [
    {
      title: t('applicationName'),
      accessor: 'serviceName',
      customWidth: 150,
    },
    {
      title: t('namespace'),
      accessor: 'namespaces',
      customWidth: 120,
      Cell: (props) => {
        return (props.value ?? []).length > 0 ? (
          props.value.join()
        ) : (
          <span className="text-slate-400">{t('na')}</span>
        )
      },
    },
    {
      title: t('applicationDetails'),
      accessor: 'serviceDetails',
      hide: true,
      isNested: true,
      customWidth: '55%',
      children: [
        {
          title: t('serviceEndpoint'),
          accessor: 'endpoint',
          canExpand: false,
        },
        {
          title: (
            <Tooltip
              title={
                <div>
                  <div>{t('delaySourceTooltipSelf')}</div>
                  <div>{t('delaySourceTooltipDependency')}</div>
                  <div>{t('delaySourceTooltipUnknown')}</div>
                </div>
              }
            >
              <div className="flex flex-row justify-center items-center">
                <div>
                  <div className="text-center flex flex-row ">{t('delaySource')}</div>
                  <div className="block text-[10px]">{t('delaySourceType')}</div>
                </div>
                <AiOutlineInfoCircle size={16} className="ml-1" />
              </div>
            </Tooltip>
          ),
          accessor: 'delaySource',
          canExpand: false,
          customWidth: 112,
          Cell: (props) => {
            const { value } = props
            return <>{DelaySourceTimeUnit[value]}</>
          },
        },
        {
          title: t('averageResponseTime'),
          minWidth: 140,
          accessor: `latency`,
          Cell: (props) => {
            const { value } = props
            return <TempCell type="latency" data={value} timeRange={{ startTime, endTime }} />
          },
        },
        {
          title: t('errorRate'),
          accessor: `errorRate`,
          minWidth: 140,
          Cell: (props) => {
            const { value } = props
            return <TempCell type="errorRate" data={value} timeRange={{ startTime, endTime }} />
          },
        },
        {
          title: t('throughput'),
          accessor: `tps`,
          minWidth: 140,
          Cell: (props) => {
            const { value } = props
            return <TempCell type="tps" data={value} timeRange={{ startTime, endTime }} />
          },
        },
      ],
    },
    {
      title: t('logErrorCount'),
      accessor: `logs`,
      minWidth: 140,
      Cell: (props) => {
        const { value } = props
        return <TempCell type="logs" data={value} timeRange={{ startTime, endTime }} />
      },
    },
    {
      title: t('infrastructureStatus'),
      accessor: `infrastructureStatus`,
      Cell: (props) => {
        const { value, trs, column } = props
        const alertReason = trs?.alertReason?.[column.accessor]
        return (
          <>
            <StatusInfo status={value} alertReason={alertReason} title={column.title} />
          </>
        )
      },
    },
    {
      title: t('networkQualityStatus'),
      accessor: `netStatus`,
      Cell: (props) => {
        const { value, trs, column } = props
        const alertReason = trs?.alertReason?.[column.accessor]
        return (
          <>
            <StatusInfo status={value} alertReason={alertReason} title={column.title} />
          </>
        )
      },
    },
    {
      title: t('k8sEventStatus'),
      accessor: `k8sStatus`,
      Cell: (props) => {
        const { value, trs, column } = props
        const alertReason = trs?.alertReason?.[column.accessor]
        return (
          <>
            <StatusInfo status={value} alertReason={alertReason} title={column.title} />
          </>
        )
      },
    },
    {
      title: t('lastDeploymentOrRestartTime'),
      accessor: `timestamp`,
      minWidth: 90,
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
      <div className="w-full h-full text-xs min-w-[1400px]">
        <BasicTable {...tableProps} />
      </div>
    </div>
  )
}
