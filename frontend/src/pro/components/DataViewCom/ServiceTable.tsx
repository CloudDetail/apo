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

export default function ServiceTable({ data }) {
  const { startTime, endTime } = useSelector(selectSecondsTimeRange)
  const column = [
    {
      title: '应用名称',
      accessor: 'serviceName',
      customWidth: 150,
    },
    {
      title: '命名空间',
      accessor: 'namespaces',
      customWidth: 120,
      Cell: (props) => {
        return (props.value ?? []).length > 0 ? (
          props.value.join()
        ) : (
          <span className="text-slate-400">N/A</span>
        )
      },
    },
    {
      title: '应用详情',
      accessor: 'serviceDetails',
      hide: true,
      isNested: true,
      customWidth: '55%',
      // clickCell: (props) => {
      //   // const navigate = useNavigate()
      //   // toServiceInfo()
      //   const serviceName = props.cell.row.values.serviceName
      //   const endpoint = props.trs.endpoint
      //   // TODO encode
      //   navigate(
      //     `/service/info?service-name=${encodeURIComponent(serviceName)}&endpoint=${encodeURIComponent(endpoint)}&breadcrumb-name=${encodeURIComponent(serviceName)}`,
      //   )
      // },

      children: [
        {
          title: '服务端点',
          accessor: 'endpoint',
          canExpand: false,
        },
        {
          title: (
            <Tooltip
              title={
                <div>
                  <div>自身：服务自身延时占比50%以上</div>
                  <div>依赖：请求下游服务延时占比50%以上</div>
                  <div>未知：未找到相关指标</div>
                </div>
              }
            >
              <div className="flex flex-row justify-center items-center">
                <div>
                  <div className="text-center flex flex-row ">延时主要来源</div>
                  <div className="block text-[10px]">(自身/依赖/未知)</div>
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
          title: '平均响应时间',
          minWidth: 140,
          accessor: `latency`,

          Cell: (props) => {
            // eslint-disable-next-line react/prop-types
            const { value } = props

            return (
              <TempCell type="latency" data={value} timeRange={{ startTime, endTime }} />
              // <div display="flex" sx={{ alignItems: 'center' }} color="white">
              //   <div sx={{ flex: 1, mr: 1 }} color="white">
              //     {/* eslint-disable-next-line react/prop-types */}
              //     {value.value}
              //   </div>

              //   <div display="flex" sx={{ alignItems: 'center', height: 30 }}>
              //     {' '}
              //     <AreaChart color="rgba(76, 192, 192, 1)" />
              //   </div>
              // </div>
            )
          },
        },
        {
          title: '错误率',
          accessor: `errorRate`,

          minWidth: 140,
          Cell: (props) => {
            // eslint-disable-next-line react/prop-types
            const { value } = props
            return <TempCell type="errorRate" data={value} timeRange={{ startTime, endTime }} />
          },
        },
        {
          title: '吞吐量',
          accessor: `tps`,
          minWidth: 140,

          Cell: (props) => {
            // eslint-disable-next-line react/prop-types
            const { value } = props
            return <TempCell type="tps" data={value} timeRange={{ startTime, endTime }} />
          },
        },
      ],
    },
    {
      title: '日志错误数',
      accessor: `logs`,
      minWidth: 140,
      Cell: (props) => {
        // eslint-disable-next-line react/prop-types
        const { value } = props
        return <TempCell type="logs" data={value} timeRange={{ startTime, endTime }} />
      },
    },
    {
      title: '基础设施状态',
      accessor: `infrastructureStatus`,
      Cell: (props) => {
        // eslint-disable-next-line react/prop-types
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
      title: '网络质量状态',
      accessor: `netStatus`,
      Cell: (/** @type {{ value: any; }} */ props) => {
        // eslint-disable-next-line react/prop-types
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
      title: 'K8s事件状态',
      accessor: `k8sStatus`,
      Cell: (props) => {
        // eslint-disable-next-line react/prop-types
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
      title: '末次部署或重启时间',
      accessor: `timestamp`,
      minWidth: 90,
      Cell: (props) => {
        const { value } = props
        return (
          <>
            {value !== null ? (
              convertTime(value, 'yyyy-mm-dd hh:mm:ss')
            ) : (
              <span className="text-slate-400">N/A</span>
            )}{' '}
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
