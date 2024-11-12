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

const options = [
  { value: 'network_time', label: 'network_time' },
  { value: 'cpu_time', label: 'cpu_time' },
  { value: 'lock_gc_time', label: 'lock_gc_time' },
  { value: 'disk_io_time', label: 'disk_io_time' },
  { value: 'scheduling_time', label: 'scheduling_time' },
]

function PolarisMetricsInfo({ service, endpoint }) {
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
      title: '类型',
      dataIndex: 'type',
      key: 'type',
    },
    {
      title: '含义',
      dataIndex: 'describe',
      key: 'describe',
    },
  ]

  const dataSource = [
    {
      key: '1',
      type: 'network_time',
      describe: '耗时为对外的服务调用和任何数据库和中间件，DNS等网络操作耗时',
    },
    {
      key: '2',
      type: 'CPU_time',
      describe:
        '程序代码在cpu上的执行时间.如果cpu_time过长，是由于嵌入层次过多或者循环过多导致在cpu上执行时间',
    },
    {
      key: '3',
      type: 'lock_gc_time',
      describe:
        '程序由于GC或者锁等待产生时延，如果是GC频繁,延时应该分布在不同时间点，如果是锁，一般是连续时间',
    },
    {
      key: '4',
      type: 'disk_io_time',
      describe: '程序写文件时间',
    },
    {
      key: '5',
      type: 'schedule_time',
      describe: '由于CPU争抢而产生的等待时间',
    },
  ]
  window.global = window
  const global = window
  const getData = () => {
    if (startTime && endTime) {
      setLoading(true)
      getPolarisInferApi({
        startTime: startTime,
        endTime: endTime,
        service: service,
        endpoint: endpoint,
        step: getStep(startTime, endTime),
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
          <p className="mt-2">
            请根据延时的波动曲线，从以下构成曲线中选择和延时曲线最相似的曲线作为接下来的排查方向。
          </p>
          <p className="mt-4 text-xs">
            <span></span>
            <InfoCircleFilled />{' '}
            通过eBPF技术我们已经将延时拆分成了以下几项，所以延时的变化一定是由下列某一项或多项构成
          </p>
          <Table dataSource={dataSource} columns={columns} pagination={false} className="mt-1" />
          <div className="flex items-center mt-4">
            <div className="flex items-center w-9/12">
              <p>请选择故障类型查看报告：</p>
              <Select
                defaultValue={defaultSelected}
                style={{ width: '30%' }}
                onChange={handleChange}
                options={options}
              />
            </div>
            <div className="flex items-center">
              <Button onClick={handlePolarisDetails} disabled={!selectedErrorType}>
                点击查看报告
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
            请选择疑似节点查看北极星指标和延时报告
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
