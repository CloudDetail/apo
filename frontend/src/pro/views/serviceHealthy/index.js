import React, { useMemo, useEffect, useState } from 'react'
import BasicTable from 'src/core/components/Table/basicTable'
import { CCard } from '@coreui/react'
import { useSelector } from 'react-redux'
import { selectProcessedTimeRange } from 'src/core/store/reducers/timeRangeReducer'
import { Input } from 'antd'
import { FaCircle } from 'react-icons/fa'
import LoadingSpinner from 'src/core/components/Spinner'
import BorderTag from 'src/core/components/Tag/BorderTag'
import { getServiceHealthyApi } from 'src/pro/api/serviceHealthy'
import { useTranslation } from 'react-i18next'
export default function ServiceHealthyPage() {
  const { t } = useTranslation('pro/healthy')
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
  const scoreMapList = [
    { name: t('severeAbnormal'), type: 'error' },
    { name: t('moderateAbnormal'), type: 'magenta' },
    { name: t('mildAbnormal'), type: 'warning' },
    { name: t('healthy'), type: 'success' },
  ]
  const column = [
    {
      title: t('serviceName'),
      accessor: 'serviceName',
      customWidth: 250,
    },
    {
      title: t('serviceStatus'),
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
      title: t('serviceHealthScore'),
      accessor: 'percentScore',
      customWidth: 120,
    },
    {
      title: t('checkItem'),
      accessor: 'key',
      customWidth: 120,
    },
    {
      title: t('healthStatus'),
      accessor: `score`,
      customWidth: 120,
      Cell: ({ value }) => {
        return <BorderTag {...scoreMapList[value]} />
      },
    },
    {
      title: t('statusDetails'),
      accessor: `detail`,
      justifyContent: 'left',
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
          <span className="text-nowrap">{t('serviceName')}：</span>
          <Input
            id="serviceName"
            placeholder={t('search')}
            value={serachServiceName}
            onChange={changeSearchValue}
            onPressEnter={handleKeyDown}
          />
        </div>
      </div>

      <CCard style={{ height: 'calc(100vh - 150px)' }}>
        <div className="mb-4 h-full p-2 text-xs justify-between">
          <BasicTable {...tableProps} />
        </div>
      </CCard>
    </div>
  )
}
