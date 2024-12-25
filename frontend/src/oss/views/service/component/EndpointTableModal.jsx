import { CModal, CModalBody, CModalHeader, CModalTitle } from '@coreui/react'
import { Tooltip } from 'antd'
import React, { useState, useEffect, useMemo } from 'react'
import { AiOutlineInfoCircle } from 'react-icons/ai'
import { useNavigate } from 'react-router-dom'
import { getEndpointTableApi } from 'core/api/service'
import BasicTable from 'src/core/components/Table/basicTable'
import TempCell from 'src/core/components/Table/TempCell'
import { DelaySourceTimeUnit } from 'src/constants'
import { getStep } from 'src/core/utils/step'
import { useTranslation } from 'react-i18next'

function EndpointTableModal(props) {
  const { i18n, t } = useTranslation('oss/service')
  const { visible, serviceName, closeModal, timeRange } = props
  const [loading, setLoading] = useState(false)
  const [data, setData] = useState([])
  const navigate = useNavigate()
  const [pageIndex, setPageIndex] = useState(1)
  const [pageSize, setPageSize] = useState(10)

  const currentLanguage = i18n.language

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
      title: t('EndpointTableModal.columns.endpoint'),
      accessor: 'endpoint',
      canExpand: false,
    },
    {
      title: (
        <Tooltip
          title={
            <div>
              <p className="text-[#D3D3D3]">
                {t('index.serviceTableColumns.serviceDetails.delaySource.title1')}
              </p>
              <p className="text-[#D3D3D3] mt-2">
                {t('index.serviceTableColumns.serviceDetails.delaySource.title2')}
              </p>
              <p className="text-[#D3D3D3] mt-2">
                {t('index.serviceTableColumns.serviceDetails.delaySource.title3')}
              </p>
            </div>
          }
        >
          <div className="flex flex-row justify-center items-center">
            <div>
              <div className="text-center flex flex-row ">
                {t('EndpointTableModal.columns.delaySource.title4')}
              </div>
              <div className="block text-[10px] text-center">
                {t('EndpointTableModal.columns.delaySource.title5')}
              </div>
            </div>
            <AiOutlineInfoCircle size={16} className="ml-2" />
          </div>
        </Tooltip>
      ),
      accessor: 'delaySource',
      canExpand: false,
      customWidth: 200,
      Cell: (props) => {
        const { value } = props
        return <>{DelaySourceTimeUnit[value]}</>
      },
    },
    {
      title: t('EndpointTableModal.columns.latency'),
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
      title: t('EndpointTableModal.columns.errorRate'),
      accessor: `errorRate`,

      minWidth: 140,
      Cell: (props) => {
        // eslint-disable-next-line react/prop-types
        const { value } = props
        return <TempCell type="errorRate" data={value} timeRange={timeRange} />
      },
    },
    {
      title: t('EndpointTableModal.columns.tps'),
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
        <CModalTitle>
          {currentLanguage === 'zh' ? (
            <span>
              {serviceName} {t('EndpointTableModal.serviceEndpointDataText')}
            </span>
          ) : (
            <span>
              {' '}
              {t('EndpointTableModal.serviceEndpointDataText')} {serviceName}
            </span>
          )}
        </CModalTitle>
      </CModalHeader>

      <CModalBody className="text-sm h-4/5">
        <BasicTable {...tableProps} />
      </CModalBody>
    </CModal>
  )
}

export default EndpointTableModal
