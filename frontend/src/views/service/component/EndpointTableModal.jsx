import { CModal, CModalBody, CModalHeader, CModalTitle } from '@coreui/react'
import { Tooltip } from 'antd'
import React, { useState, useEffect, useMemo } from 'react'
import { AiOutlineInfoCircle } from 'react-icons/ai'
import { useNavigate } from 'react-router-dom'
import { getEndpointTableApi } from 'src/api/service'
import BasicTable from 'src/components/Table/basicTable'
import TempCell from 'src/components/Table/TempCell'
import { DelaySourceTimeUnit } from 'src/constants'
import { getStep } from 'src/utils/step'

function EndpointTableModal(props) {
  const { visible, serviceName, closeModal, timeRange } = props
  const [loading, setLoading] = useState(false)
  const [data, setData] = useState([])
  const navigate = useNavigate()
  const [pageIndex, setPageIndex] = useState(1)
  const [pageSize, setPageSize] = useState(10)

  useEffect(() => {
    if (visible && serviceName) {
      setLoading(true)
      // 记录请求的时间范围，以便后续趋势图补0
      getEndpointTableApi({
        startTime: timeRange.startTime,
        endTime: timeRange.endTime,
        step: getStep(timeRange.startTime, timeRange.endTime),
        serviceName: serviceName,
        sortRule: 1,
      })
        .then((res) => {
          if (res && res?.length > 0) {
            setData(res)
          } else {
            setData([])
          }
          setLoading(false)
        })
        .catch(() => {
          setData([])
          setLoading(false)
        })
    }
  }, [visible, serviceName, timeRange])
  const column = [
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
          <div className="flex flex-row ">
            <AiOutlineInfoCircle size={16} />
            <div className="text-center">
              延时主要来源<div className="block text-[10px]">(自身/依赖/未知)</div>
            </div>
          </div>
        </Tooltip>
      ),
      accessor: 'delaySource',
      canExpand: false,
      customWidth: 110,
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
          <TempCell type="latency" data={value} timeRange={timeRange} />
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
        return <TempCell type="errorRate" data={value} timeRange={timeRange} />
      },
    },
    {
      title: '吞吐量',
      accessor: `tps`,
      minWidth: 140,

      Cell: (props) => {
        // eslint-disable-next-line react/prop-types
        const { value } = props
        return <TempCell type="tps" data={value} timeRange={timeRange} />
      },
    },
  ]
  const toServiceInfoPage = (props) => {
    navigate(
      `/service/info?service-name=${encodeURIComponent(serviceName)}&endpoint=${encodeURIComponent(props.endpoint)}&breadcrumb-name=${encodeURIComponent(serviceName)}`,
    )
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
      loading: loading,
      onChange: handleTableChange,
      clickRow: toServiceInfoPage,
      pagination: {
        pageSize: pageSize,
        pageIndex: pageIndex,
        pageCount: Math.ceil(data.length / pageSize),
      },
    }
  }, [data, loading, pageIndex, pageSize])
  return (
    <CModal
      visible={visible}
      alignment="center"
      size="xl"
      className="absolute-modal"
      onClose={closeModal}
    >
      <CModalHeader>
        <CModalTitle>{serviceName}下所有服务端点数据</CModalTitle>
      </CModalHeader>

      <CModalBody className="text-sm h-4/5">
        <BasicTable {...tableProps} />
      </CModalBody>
    </CModal>
  )
}

export default EndpointTableModal
