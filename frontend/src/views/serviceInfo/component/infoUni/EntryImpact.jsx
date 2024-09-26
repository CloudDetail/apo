import { CAccordionBody, CButton, CToast, CToastBody } from '@coreui/react'
import { Tooltip } from 'antd'
import React, { useMemo, useEffect, useState } from 'react'
import { AiOutlineInfoCircle } from 'react-icons/ai'
import { IoMdInformationCircleOutline } from 'react-icons/io'
import { useSelector } from 'react-redux'
import { useNavigate } from 'react-router-dom'
import { useDebounce } from 'react-use'
import { getServiceEntryEndpoints } from 'src/api/serviceInfo'
import LoadingSpinner from 'src/components/Spinner'
import StatusInfo from 'src/components/StatusInfo'
import BasicTable from 'src/components/Table/basicTable'
import TempCell from 'src/components/Table/TempCell'
import { DelaySourceTimeUnit } from 'src/constants'
import { usePropsContext } from 'src/contexts/PropsContext'
import { selectSecondsTimeRange } from 'src/store/reducers/timeRangeReducer'
import { getStep } from 'src/utils/step'
import { convertTime } from 'src/utils/time'
import EndpointTableModal from 'src/views/service/component/EndpointTableModal'
export default function EntryImpact(props) {
  const { handlePanelStatus } = props
  const navigate = useNavigate()
  const [data, setData] = useState([])
  const { serviceName, endpoint } = usePropsContext()
  const [loading, setLoading] = useState(false)
  const { startTime, endTime } = useSelector(selectSecondsTimeRange)
  const [modalVisible, setModalVisible] = useState(false)
  const [modalServiceName, setModalServiceName] = useState()
  const column = [
    {
      title: '入口应用名称',
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

      clickCell: (props) => {
        // const navigate = useNavigate()
        // toServiceInfo()
        const serviceName = props.cell.row.values.serviceName
        const endpoint = props.trs.endpoint
        navigate(
          `/service/info?service-name=${encodeURIComponent(serviceName)}&endpoint=${encodeURIComponent(endpoint)}&breadcrumb-name=${encodeURIComponent(serviceName)}`,
        )
        window.scrollTo(0, 0)
      },
      children: [
        {
          title: '入口服务端点',
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
          accessor: (idx) => `latency`,

          Cell: (props) => {
            // eslint-disable-next-line react/prop-types
            const { value } = props

            return <TempCell type="latency" data={value} timeRange={{ startTime, endTime }} />
          },
        },
        {
          title: '错误率',
          accessor: (idx) => `errorRate`,

          minWidth: 140,
          Cell: (props) => {
            // eslint-disable-next-line react/prop-types
            const { value } = props
            return <TempCell type="errorRate" data={value} timeRange={{ startTime, endTime }} />
          },
        },
        {
          title: '吞吐量',
          accessor: (idx) => `tps`,
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
      accessor: `netStatus`,
      Cell: (/** @type {{ value: any; }} */ props) => {
        // eslint-disable-next-line react/prop-types
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
      accessor: `k8sStatus`,
      Cell: (props) => {
        // eslint-disable-next-line react/prop-types
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
  const getTableData = () => {
    if (startTime && endTime) {
      setLoading(true)
      getServiceEntryEndpoints({
        startTime: startTime,
        endTime: endTime,
        step: getStep(startTime, endTime),
        service: serviceName,
        endpoint: endpoint,
      })
        .then((res) => {
          setData(res.data ?? [])
          handlePanelStatus(res.status)
          //   setPageIndex(1)
          setLoading(false)
        })
        .catch(() => {
          //   setPageIndex(1)
          handlePanelStatus('unknown')
          setData([])
          setLoading(false)
        })
    }
  }
  // useEffect(() => {
  //   getTableData()
  // }, [startTime, endTime, serviceName, endpoint])
  useDebounce(
    () => {
      getTableData()
    },
    300, // 延迟时间 300ms
    [startTime, endTime, serviceName, endpoint],
  )
  const tableProps = useMemo(() => {
    return {
      columns: column,
      data: data,
      showBorder: false,
      loading: false,
    }
  }, [data])
  return (
    <CAccordionBody className="text-xs relative">
      <LoadingSpinner loading={loading} />
      <CToast autohide={false} visible={true} className="align-items-center w-full mb-2">
        <div className="d-flex">
          <CToastBody className=" flex flex-row items-center text-xs">
            <IoMdInformationCircleOutline size={20} color="#f7c01a" className="mr-1" />
            可能受该接口影响的所有服务入口分析，其中“服务入口”是指业务被访问时调用的第一个服务端点，是调用链路中的最上游。
          </CToastBody>
        </div>
      </CToast>
      <BasicTable {...tableProps} />
      <EndpointTableModal
        visible={modalVisible}
        serviceName={modalServiceName}
        timeRange={{ startTime, endTime }}
        closeModal={() => setModalVisible(false)}
      />
    </CAccordionBody>
  )
}
