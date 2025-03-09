import { CAccordionBody, CImage } from '@coreui/react'
import React, { useEffect, useState } from 'react'
import { useSelector } from 'react-redux'
import { getPolarisInferApi } from 'src/core/api/serviceInfo'
import { selectSecondsTimeRange } from 'src/core/store/reducers/timeRangeReducer'
import 'react-image-lightbox/style.css'
import Lightbox from 'react-image-lightbox'
import { getStep } from 'src/core/utils/step'
import { useDebounce } from 'react-use'
import { Skeleton, Table, Button, Select, Flex, Typography, Divider } from 'antd'
import FaultReportModal from './FaultReportModal'
import { InfoCircleFilled } from '@ant-design/icons'
import { useAlertAnalyzeContext } from 'src/pro/contexts/AlertAnalyzeContext'
import { SlowErrorType } from 'src/constants'
import { useTranslation } from 'react-i18next'
import i18next from 'i18next'
const i18nKey = 'node.polaris'

const options = [
  { value: 'network_time', label: 'network_time' },
  { value: 'cpu_time', label: 'cpu_time' },
  { value: 'lock_gc_time', label: 'lock_gc_time' },
  { value: 'disk_io_time', label: 'disk_io_time' },
  { value: 'scheduling_time', label: 'scheduling_time' },
]

function PolarisMetricsInfo({ service, endpoint }) {
  const { t } = useTranslation('pro/serviceAnalyze')
  const { selectedNodeInfo } = useAlertAnalyzeContext()
  const [image, setImage] = useState()
  const [inferCause, setInferCause] = useState()
  const { startTime, endTime } = useSelector(selectSecondsTimeRange)
  const [loading, setLoading] = useState(false)
  const [isOpen, setIsOpen] = useState(false)
  const [modalVisibility, setModalVisibility] = useState(false)
  const [selectedErrorType, setSelectedErrorType] = useState(null)
  const [defaultSelected, setDefaultSelected] = useState('')
  const columns = [
    {
      title: t(`${i18nKey}.type`),
      dataIndex: 'type',
      key: 'type',
    },
    {
      title: t(`${i18nKey}.describe`),
      dataIndex: 'describe',
      key: 'describe',
    },
  ]
  const dataSource = Object.entries(SlowErrorType).map(([key, value]) => ({
    type: key,
    key: key,
    describe: value,
  }))

  window.global = window
  const global = window
  const getData = () => {
    const language = i18next.language
    const timeZone = Intl.DateTimeFormat().resolvedOptions().timeZone
    if (startTime && endTime) {
      setLoading(true)
      getPolarisInferApi({
        startTime: startTime,
        endTime: endTime,
        service: service,
        endpoint: endpoint,
        step: getStep(startTime, endTime),
        language: language,
        timezone: timeZone,
      })
        .then((res) => {
          setImage(res?.inferMetricsPng)
          setInferCause(res?.inferCause)
          ifExistInferResult(res?.inferCause)
          setLoading(false)
        })
        .catch((error) => {
          setLoading(false)
        })
    }
  }

  function ifExistInferResult(result) {
    options.forEach((item) => {
      if (result.includes(item['value'])) {
        setDefaultSelected(item['value'])
        setSelectedErrorType(item['value'])
        return
      }
    })
  }

  const handleImageClick = () => {
    setIsOpen(true)
  }

  function handleChange(value) {
    setSelectedErrorType(value)
  }

  function handlePolarisDetails() {
    setModalVisibility(true)
  }

  // useEffect(() => {
  //   getData()
  // }, [startTime, endTime, service, endpoint])
  //防抖避免跳转使用旧时间
  useDebounce(
    () => {
      if (selectedNodeInfo.endpoint) {
        setLoading(true)
        getData()
      } else {
        setLoading(false)
      }
    },
    300, // 延迟时间 300ms
    [startTime, endTime, service, endpoint],
  )

  return (
    <div>
      {selectedNodeInfo.endpoint ? (
        <Skeleton loading={loading}>
          {/* {inferCause && (
          <div className="w-full flex flex-col">
            <div className="flex-1">{inferCause}</div>
          </div>
        )} */}
          {image && (
            <div className="w-full">
              <img
                src={'data:image/png;base64,' + image}
                alt="Example"
                onClick={handleImageClick}
                style={{ cursor: 'pointer', width: '100%', height: '400px' }}
              />
              {isOpen && (
                <Lightbox
                  imagePadding={50}
                  mainSrc={'data:image/png;base64,' + image}
                  onCloseRequest={() => setIsOpen(false)}
                />
              )}
            </div>
          )}
          <Divider />
          <p className="mt-2">{t(`${i18nKey}.toast`)}</p>
          <p className="mt-4 text-xs">
            <span></span>
            <InfoCircleFilled /> {t(`${i18nKey}.info`)}
          </p>
          <Table dataSource={dataSource} columns={columns} pagination={false} className="mt-1" />
          <div className="flex items-center mt-4">
            <div className="flex items-center w-9/12">
              <p> {t(`${i18nKey}.costTime`)}</p>
              <Select
                defaultValue={defaultSelected}
                style={{ width: '30%' }}
                onChange={handleChange}
                options={options}
              />
            </div>
            <div className="flex items-center">
              <Button onClick={handlePolarisDetails} disabled={!selectedErrorType}>
                {t(`${i18nKey}.clickToView`)}
              </Button>
            </div>
          </div>
        </Skeleton>
      ) : (
        <Flex>
          <p
            style={{
              color: '#9ca3af',
              fontSize: '12px',
            }}
          >
            {t(`${i18nKey}.selectNode`)}
          </p>
        </Flex>
      )}
      <FaultReportModal
        modalVisibility={modalVisibility}
        hidenModal={() => setModalVisibility(false)}
        service={service}
        endpoint={endpoint}
        reason={selectedErrorType}
      />
    </div>
  )
}
export default PolarisMetricsInfo
