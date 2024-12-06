import React, { useMemo, useEffect, useState } from 'react'
import BasicTable from 'src/core/components/Table/basicTable'
import { CButton, CCard, CFormInput, CToast, CToastBody, CToastClose } from '@coreui/react'
import { useNavigate } from 'react-router-dom'
import TempCell from 'src/core/components/Table/TempCell'
import StatusInfo from 'src/core/components/StatusInfo'
import { IoMdInformationCircleOutline } from 'react-icons/io'
import ThresholdCofigModal from './component/ThresholdCofigModal'
import {
  getServicesAlertApi,
  getServicesEndpointsApi,
  getServiceListApi,
  getServiceEndpointNameApi,
  getNamespacesApi
} from 'core/api/service'
import { useSelector } from 'react-redux'
import { selectSecondsTimeRange, toNearestSecond } from 'src/core/store/reducers/timeRangeReducer'
import { getStep } from 'src/core/utils/step'
import { DelaySourceTimeUnit } from 'src/constants'
import { convertTime } from 'src/core/utils/time'
import EndpointTableModal from './component/EndpointTableModal'
import LoadingSpinner from 'src/core/components/Spinner'
import { Input, Tooltip, Select } from 'antd'
import { AiOutlineInfoCircle } from 'react-icons/ai'
import { useDebounce } from 'react-use'
export default function ServiceView() {
  const navigate = useNavigate()
  const [data, setData] = useState([])
  const [serachServiceName, setSerachServiceName] = useState()
  const [serachEndpointName, setSerachEndpointName] = useState()
  const [serachNamespace, setSerachNamespace] = useState()
  const [loading, setLoading] = useState(true)
  const [modalVisible, setModalVisible] = useState(false)
  const [modalServiceName, setModalServiceName] = useState()
  const [requestTimeRange, setRequestTimeRange] = useState({
    startTime: null,
    endTime: null,
  })
  const [pageIndex, setPageIndex] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [serviceNameOptions, setServiceNameOptions] = useState([])
  const [endpointNameOptions, setEndpointNameOptions] = useState([])
  const [namespaceOptions, setNamespaceOptions] = useState([])
  const { startTime, endTime } = useSelector(selectSecondsTimeRange)

  const getServiceNameOptions = async () => {
    try {
      const params = { startTime, endTime };
      const data = await getServiceListApi(params);
      setServiceNameOptions(data.map(value => ({ value, label: value })));
    } catch (error) {
      console.error('获取服务列表失败:', error);
    }
  };
  
  const getNamespaceOptions = async () => {
    try {
      const data = await getNamespacesApi();
      const mapData = (data?.items || []).map(item => ({
        value: item.metadata.name,
        label: item.metadata.name,
      }));
      setNamespaceOptions(mapData);
    } catch (error) {
      console.log('获取命名空间失败:', error);
    }
  };
  
  const getEndpointNameOptions = async (serviceNameList) => {
    setEndpointNameOptions([]);
    
    try {
      const newEndpointNameOptions = await Promise.all(serviceNameList.map(async (element) => {
        const params = {
          startTime,
          endTime,
          step: getStep(startTime, endTime),
          serviceName: element,
          sortRule: 1,
        };
        const data = await getServiceEndpointNameApi(params);
        return {
          label: <span>{element}</span>,
          title: element,
          options: data.map(item => ({
            label: <span>{item?.endpoint}</span>,
            value: item?.endpoint,
          })),
        };
      }));
      
      setEndpointNameOptions(newEndpointNameOptions);
    } catch (error) {
      console.error('获取 endpoint 失败:', error);
    }
  };
  
  useEffect(() => {
    const fetchData = async () => {
      await Promise.all([getServiceNameOptions(), getNamespaceOptions()]);
    };
    
    fetchData();
  }, []);
  
  useEffect(() => {
    if (serachServiceName?.length > 0) {
      getEndpointNameOptions(serachServiceName);
    }
  }, [serachServiceName]);
  

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
      ],
    },
  ]

  const getTableData = () => {
    if (startTime && endTime) {
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
        endpointName: serachEndpointName,
        namespace: serachNamespace,
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
    }
  }

  useEffect(() => {
    getTableData()
  }, [serachServiceName, serachEndpointName, serachNamespace])

  //防抖避免跳转使用旧时间
  useDebounce(
    () => {
      getTableData()
    },
    300, // 延迟时间 300ms
    [startTime, endTime],
  )

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
      // data: [],

      loading: loading,
      onChange: handleTableChange,
      pagination: {
        pageSize: pageSize,
        pageIndex: pageIndex,
        pageCount: Math.ceil(data.length / pageSize),
      },
      emptyContent: (
        <div className="text-center">
          暂无数据
          <div className="text-left p-2">
            <div className="py-2">
              1. 如尚未监控应用，请参阅
              <a
                className="underline text-sky-500"
                target="_blank"
                href="https://originx.kindlingx.com/docs/APO%20向导式可观测性中心/安装手册/监控%20Kubernetes%20集群中的服务器和应用使用OneAgent默认OTEL探针版本/"
              >
                监控手册
              </a>
            </div>
            <div>
              2. 如已监控应用但仍无数据，请参阅
              <a
                className="underline text-sky-500"
                target="_blank"
                href="https://originx.kindlingx.com/docs/APO%20向导式可观测性中心/安装手册/运维与故障排除/APO%20服务概览无数据排查文档/#常见基础问题排查"
              >
                故障排除手册
              </a>
            </div>
          </div>
        </div>
      ),
      showLoading: false,
    }
  }, [data, pageIndex, pageSize, loading])
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
      <div className="p-2 my-2 flex flex-row w-full">
        <div className="flex flex-row items-center mr-5 text-sm min-w-[280px]">
          <span className="text-nowrap">命名空间：</span>
          <Select
            mode='multiple'
            id="namespace"
            placeholder="请选择"
            className='w-full'
            value={serachNamespace}
            onChange={(e) => setSerachNamespace(e)}
            options={namespaceOptions}
            maxTagCount={2}
            maxTagPlaceholder={(omittedValues) => `+${omittedValues.length}...`}
            allowClear
          />
        </div>
        <div className="flex flex-row items-center mr-5 text-sm min-w-[280px]">
          <span className="text-nowrap">服务名：</span>
          <Select
            mode='multiple'
            allowClear
            className='w-full'
            id="serviceName"
            placeholder="请选择"
            value={serachServiceName}
            onChange={(e) => setSerachServiceName(e)}
            options={serviceNameOptions}
            popupMatchSelectWidth={false}
            maxTagCount={2}
            maxTagPlaceholder={(omittedValues) => `+${omittedValues.length}...`}
          />
        </div>
        <div className="flex flex-row items-center mr-5 text-sm min-w-[280px]">
          <span className="text-nowrap">服务端点：</span>
          <Select
            mode='multiple'
            id="endpointName"
            placeholder="请选择"
            className='w-full'
            value={serachEndpointName}
            popupMatchSelectWidth={false}
            onChange={(e) => setSerachEndpointName(e)}
            options={endpointNameOptions}
            maxTagCount={2}
            maxTagPlaceholder={(omittedValues) => `+${omittedValues.length}...`}
            allowClear
          />
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
