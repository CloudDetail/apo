import React, { useMemo, useState, useEffect } from 'react'
import StatusInfo from 'src/components/StatusInfo'
import BasicTable from 'src/components/Table/basicTable'
import {
  CButton,
  CCardBody,
  CCardHeader,
  CModal,
  CModalBody,
  CModalHeader,
  CModalTitle,
} from '@coreui/react'
import { FaChartLine } from 'react-icons/fa6'
import { useNavigate } from 'react-router-dom'
import { serviceMock } from 'src/components/ReactFlow/mock'
import { usePropsContext } from 'src/contexts/PropsContext'
import DelayLineChart from '../infoUni/DelayLineChart'
import { convertTime } from 'src/utils/time'
import { getServiceDsecendantRelevanceApi } from 'src/api/serviceInfo'
import { useSelector } from 'react-redux'
import { selectSecondsTimeRange } from 'src/store/reducers/timeRangeReducer'
import { getStep } from 'src/utils/step'

function DependentTable(props) {
  const { serviceList } = props
  const { serviceName, endpoint } = usePropsContext()
  const [visible, setVisible] = useState(false)
  const navigate = useNavigate()
  const [data, setData] = useState([])
  const { startTime, endTime } = useSelector(selectSecondsTimeRange)
  const [loading, setLoading] = useState(false)
  const getTableData = () => {
    setLoading(true)

    getServiceDsecendantRelevanceApi({
      startTime: startTime,
      endTime: endTime,
      service: serviceName,
      endpoint: endpoint,
      step: getStep(startTime, endTime),
    })
      .then((res) => {
        setData(res ?? [])
        setLoading(false)
      })
      .catch((error) => {
        setData([])
        setLoading(false)
      })
  }
  useEffect(() => {
    getTableData()
  }, [startTime, endTime, serviceName, endpoint])
  const column = [
    {
      title: '应用名称',
      accessor: 'serviceName',
      customWidth: 150,
    },

    {
      title: '服务端点',
      accessor: 'endpoint',
      // Cell: (props) => {
      //   console.log(props)
      //   // const navigate = useNavigate()
      //   // toServiceInfo()
      //   // TODO encode

      //   return <a onClick={navigate(
      //     `/service/info?service-name=${encodeURIComponent(serviceName)}&endpoint=${encodeURIComponent(endpoint)}&breadcrumb-name=${encodeURIComponent(serviceName)}`,
      //   )} >{props.value}</a>
      //   // window.location.reload();
      //   // window.location.href = `/service/info?service-name=${serviceName}&url=${url}&&breadcrumb-name=${serviceName}`
      // },
    },
    {
      title: (
        <div className="text-center">
          延时主要来源<div className="block">(自身/依赖)</div>
        </div>
      ),
      accessor: 'isLatencySelf',
      canExpand: false,
      customWidth: 100,
      Cell: (props) => {
        return props.value ? '自身' : '依赖'
      },
    },
    {
      title: 'RED告警',
      accessor: `REDStatus`,

      Cell: (props) => {
        // eslint-disable-next-line react/prop-types
        const { value } = props
        return <StatusInfo status={value} />
      },
    },
    {
      title: '日志错误数量',
      accessor: `logsStatus`,
      Cell: (props) => {
        // eslint-disable-next-line react/prop-types
        const { value } = props
        return <StatusInfo status={value} />
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
      Cell: (props) => {
        const { value } = props
        return <>{value !== null ? convertTime(value, 'yyyy-mm-dd hh:mm:ss') : 'N/A'} </>
      },
    },
  ]
  const toServiceInfoPage = (props) => {
    console.log(props)
    navigate(
      `/service/info?service-name=${encodeURIComponent(props.serviceName)}&endpoint=${encodeURIComponent(props.endpoint)}&breadcrumb-name=${encodeURIComponent(props.serviceName)}`,
    )
  }
  const tableProps = useMemo(() => {
    return {
      columns: column,
      data: data ?? [],
      showBorder: true,
      loading: loading,
      clickRow: toServiceInfoPage,
    }
  }, [data, startTime, endTime, loading])
  return (
    <>
      <CModal
        alignment="center"
        visible={visible}
        onClose={() => setVisible(false)}
        aria-labelledby="VerticallyCenteredExample"
      >
        <CModalHeader>
          <CModalTitle id="VerticallyCenteredExample">延时曲线全览对比图</CModalTitle>
        </CModalHeader>
        <CModalBody>
          <DelayLineChart color="rgba(154, 102, 255, 1)" multiple />
        </CModalBody>
      </CModal>
      {data && <BasicTable {...tableProps} />}
    </>
  )
}

export default DependentTable
