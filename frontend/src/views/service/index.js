import React, { useMemo, useEffect, useState } from 'react'
import BasicTable from 'src/components/Table/basicTable'
import { CButton, CCard, CFormInput, CToast, CToastBody, CToastClose } from '@coreui/react'
import { useNavigate } from 'react-router-dom'
import TempCell from 'src/components/Table/TempCell'
import StatusInfo from 'src/components/StatusInfo'
import { IoMdInformationCircleOutline } from 'react-icons/io'
import ThresholdCofigModal from './component/ThresholdCofigModal'
import { getServicesAlertApi, getServicesEndpointsApi } from 'src/api/service'
import { useSelector } from 'react-redux'
import { selectSecondsTimeRange, toNearestSecond } from 'src/store/reducers/timeRangeReducer'
import { getStep } from 'src/utils/step'
import { DelaySourceTimeUnit } from 'src/constants'
import { convertTime } from 'src/utils/time'
import EndpointTableModal from './component/EndpointTableModal'
import LoadingSpinner from 'src/components/Spinner'
export default function ServiceView() {
  const navigate = useNavigate()
  const [data, setData] = useState([])
  const [serachServiceName, setSerachServiceName] = useState()
  const [loading, setLoading] = useState(false)
  const [modalVisible, setModalVisible] = useState(false)
  const [modalServiceName, setModalServiceName] = useState()
  const [requestTimeRange, setRequestTimeRange] = useState({
    startTime: null,
    endTime: null,
  })
  const [pageIndex, setPageIndex] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const column = [
    {
      title: '应用名称',
      accessor: 'serviceName',
      customWidth: 150,
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
        // TODO encode
        navigate(
          `/service/info?service-name=${encodeURIComponent(serviceName)}&endpoint=${encodeURIComponent(endpoint)}&breadcrumb-name=${encodeURIComponent(serviceName)}`,
        )
      },
      showMore: (original) => {
        const clickMore = () => {
          setModalVisible(true)
          setModalServiceName(original.serviceName)
        }
        return (
          original.endpointCount > 3 && (
            <CButton color="info" variant="ghost" size="sm" onClick={clickMore}>
              更多
            </CButton>
          )
        )
        // return
      },

      children: [
        {
          title: '服务端点',
          accessor: 'endpoint',
          canExpand: false,
        },
        {
          title: (
            <div className="text-center">
              延时主要来源<div className="block">(自身/依赖)</div>
            </div>
          ),
          accessor: 'delaySource',
          canExpand: false,
          customWidth: 100,
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

            return (
              <TempCell type="latency" data={value} timeRange={requestTimeRange} />
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
          accessor: (idx) => `errorRate`,

          minWidth: 140,
          Cell: (props) => {
            // eslint-disable-next-line react/prop-types
            const { value } = props
            return <TempCell type="errorRate" data={value} timeRange={requestTimeRange} />
          },
        },
        {
          title: '吞吐量',
          accessor: (idx) => `tps`,
          minWidth: 140,

          Cell: (props) => {
            // eslint-disable-next-line react/prop-types
            const { value } = props
            return <TempCell type="tps" data={value} timeRange={requestTimeRange} />
          },
        },
      ],
    },
    {
      title: '应用基础信息告警',
      accessor: 'serviceInfo',
      hide: true,
      isNested: true,
      // customWidth: '55%',
      api: async (props) => {
        const { serviceName } = props
        try {
          const result = await getServicesAlert([serviceName])
          return { data: result, error: null }
        } catch (error) {
          console.error('Error calling getServicesAlert:', error)
          return { data: [], error: error }
        }
      },
      children: [
        {
          title: '日志错误数',
          accessor: `logs`,
          minWidth: 140,
          Cell: (props) => {
            // eslint-disable-next-line react/prop-types
            const { value } = props
            return <TempCell type="logs" data={value} timeRange={requestTimeRange} />
          },
        },
        {
          title: '基础设施状态',
          accessor: `infrastructureStatus`,
          Cell: (props) => {
            // eslint-disable-next-line react/prop-types
            const { value } = props
            return (
              <>
                <StatusInfo status={value} />
              </>
            )
          },
        },
        {
          title: '网络质量状态',
          accessor: `netStatus`,
          Cell: (/** @type {{ value: any; }} */ props) => {
            // eslint-disable-next-line react/prop-types
            const { value } = props
            return (
              <>
                <StatusInfo status={value} />
              </>
            )
          },
        },
        {
          title: 'K8s事件状态',
          accessor: `k8sStatus`,
          Cell: (props) => {
            // eslint-disable-next-line react/prop-types
            const { value } = props
            return (
              <>
                <StatusInfo status={value} />
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
            return <>{value !== null ? convertTime(value, 'yyyy-mm-dd hh:mm:ss') : 'N/A'} </>
          },
        },
      ],
    },
  ]
  const { startTime, endTime } = useSelector(selectSecondsTimeRange)

  useEffect(() => {
    setLoading(true)
    // 记录请求的时间范围，以便后续趋势图补0
    setRequestTimeRange({
      startTime: startTime,
      endTime: endTime,
    })
    getServicesEndpointsApi({
      startTime: startTime,
      endTime: endTime,
      step: getStep(startTime, endTime),
      serviceName: serachServiceName,
      sortRule: 1,
    })
      .then((res) => {
        if (res && res?.length > 0) {
          setData(res)
        } else {
          setData([])
        }
        setPageIndex(1)
        setLoading(false)
      })
      .catch(() => {
        setPageIndex(1)
        setData([])
        setLoading(false)
      })
  }, [startTime, endTime, serachServiceName])

  const handleKeyDown = (event) => {
    if (event.key === 'Enter') {
      setSerachServiceName(event.target.value)
      // 在这里添加回车事件的处理逻辑
      // 例如，你可以在这里提交表单、搜索数据等
    }
  }
  const getServicesAlert = (serviceNames, returnData) => {
    return getServicesAlertApi({
      startTime: startTime,
      endTime: endTime,
      step: getStep(startTime, endTime),
      serviceNames: serviceNames,
      returnData: returnData,
    })
      .then((res) => {
        if (res && res?.length > 0) {
          return res
        } else {
          return []
        }
        // setLoading(false)
      })
      .catch(() => {
        return []
        // setLoading(false)
      })
  }

  const handleTableChange = (props) => {
    if (props.pageSize && props.pageIndex) {
      setPageSize(props.pageSize), setPageIndex(props.pageIndex)
    }
  }
  const tableProps = useMemo(() => {
    const paginatedData = data.slice((pageIndex - 1) * pageSize, pageIndex * pageSize)
    return {
      columns: column,
      data: paginatedData,
      showBorder: true,
      loading: false,
      onChange: handleTableChange,
      pagination: {
        pageSize: pageSize,
        pageIndex: pageIndex,
        pageCount: Math.ceil(data.length / 10),
      },
    }
  }, [data, pageIndex, pageSize])
  return (
    <div style={{ width: '100%', overflow: 'hidden' }}>
      <LoadingSpinner loading={loading} />
      <CToast autohide={false} visible={true} className="align-items-center w-full my-2">
        <div className="d-flex">
          <CToastBody className=" flex flex-row items-center text-xs">
            <IoMdInformationCircleOutline size={20} color="#f7c01a" className="mr-1" />
            请根据本页提供的各种提示信息，选择您最怀疑的服务端点，选择服务端点之后可以查看具体的相关指标、日志和事件信息、Trace信息
          </CToastBody>
        </div>
      </CToast>
      <div className="p-2 my-2 flex flex-row justify-between">
        <div className="flex flex-row items-center mr-5 text-sm">
          <span className="text-nowrap">应用名称：</span>
          <CFormInput placeholder="检索" size="sm" onKeyDown={handleKeyDown} />
        </div>

        <div>{/* <ThresholdCofigModal /> */}</div>
      </div>

      <CCard style={{ height: 'calc(100vh - 220px)' }}>
        <div className="mb-4 h-full p-2 text-xs justify-between">
          <BasicTable {...tableProps} />
        </div>
        <EndpointTableModal
          visible={modalVisible}
          serviceName={modalServiceName}
          timeRange={requestTimeRange}
          closeModal={() => setModalVisible(false)}
        />
      </CCard>
    </div>
  )
}
