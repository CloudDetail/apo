import React, { useMemo, useEffect, useState } from 'react'
import BasicTable from 'src/components/Table/basicTable'
import { CCard } from '@coreui/react'
import { useSelector } from 'react-redux'
import { selectProcessedTimeRange } from 'src/store/reducers/timeRangeReducer'
import { Input } from 'antd'
import { getServiceHealthyApi } from 'src/api/serviceHealthy'
import { FaCircle } from 'react-icons/fa'
import LoadingSpinner from 'src/components/Spinner'
export default function ServiceHealthyPage() {
  const [data, setData] = useState([])
  const [serachServiceName, setSerachServiceName] = useState()
  const [serachEndpointName, setSerachEndpointName] = useState()
  const [serachNamespace, setSerachNamespace] = useState()
  const [loading, setLoading] = useState(false)
  const StatusColorMap = {
    yellow: '#f9bb07',
    green: '#24d160',
    red: '#ff3366',
  }
  const column = [
    {
      title: '服务名称',
      accessor: 'serviceName',
      customWidth: 250,
    },
    {
      title: '服务状态',
      accessor: 'status',
      customWidth: 120,
      Cell: ({ value }) => {
        return (
          <div className="p-2 w-full justify-center flex items-center h-full">
            <div>
              <FaCircle color={StatusColorMap[value]} />
            </div>
          </div>
        )
      },
    },
    {
      title: '服务健康总分',
      accessor: 'score',
      customWidth: 120,
    },
    {
      title: '得分详情',
      accessor: 'scoreDetail',
      hide: true,
      isNested: true,
      children: [
        {
          title: '得分原因',
          accessor: 'key',
          customWidth: 120,
        },
        {
          title: '得分',
          accessor: `score`,
          customWidth: 120,
        },
        {
          title: '得分详情',
          accessor: `detail`,
          justifyContent: 'left',
        },
      ],
    },
  ]
  const { startTime, endTime } = useSelector(selectProcessedTimeRange)
  const getTableData = () => {
    if (startTime && endTime) {
      setLoading(true)
      getServiceHealthyApi({
        startTime: startTime,
        endTime: endTime,
        serviceName: serachServiceName,
        endpointName: serachEndpointName,
        namespace: serachNamespace,
      })
        .then((res) => {
          //   if (res && res?.length > 0) {
          //     setData(res)
          //   } else {
          //     setData([])
          //   }
          setData(res?.serviceList ?? [])
          setLoading(false)
        })
        .catch(() => {
          setData([])
          setLoading(false)
        })
    }
  }
  useEffect(() => {
    getTableData()
  }, [startTime, endTime])

  const handleKeyDown = (event) => {
    getTableData()
  }
  const changeSearchValue = (event) => {
    switch (event.target.id) {
      case 'serviceName':
        setSerachServiceName(event.target.value)
        return
      case 'endpointName':
        setSerachEndpointName(event.target.value)
        return
      case 'namespace':
        setSerachNamespace(event.target.value)
        return
      default:
        break
    }
  }
  const tableProps = useMemo(() => {
    return {
      columns: column,
      data: data,
    }
  }, [data])
  return (
    <div style={{ width: '100%', overflow: 'hidden' }}>
      <LoadingSpinner loading={loading} />
      <div className="p-2 my-2 flex flex-row">
        <div className="flex flex-row items-center mr-5 text-sm">
          <span className="text-nowrap">服务名称：</span>
          <Input
            id="serviceName"
            placeholder="检索"
            value={serachServiceName}
            onChange={changeSearchValue}
            onPressEnter={handleKeyDown}
          />
        </div>
        {/* <div className="flex flex-row items-center mr-5 text-sm">
          <span className="text-nowrap">服务端点：</span>
          <Input
            id="endpointName"
            placeholder="检索"
            value={serachEndpointName}
            onChange={changeSearchValue}
            onPressEnter={handleKeyDown}
          />
        </div>
        <div className="flex flex-row items-center mr-5 text-sm">
          <span className="text-nowrap">命名空间：</span>
          <Input
            id="namespace"
            placeholder="检索"
            value={serachNamespace}
            onChange={changeSearchValue}
            onPressEnter={handleKeyDown}
          />
        </div> */}
        <div>{/* <ThresholdCofigModal /> */}</div>
      </div>

      <CCard style={{ height: 'calc(100vh - 150px)' }}>
        <div className="mb-4 h-full p-2 text-xs justify-between">
          <BasicTable {...tableProps} />
        </div>
      </CCard>
    </div>
  )
}
