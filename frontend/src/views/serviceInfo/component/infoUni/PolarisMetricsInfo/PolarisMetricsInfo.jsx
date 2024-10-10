import { CAccordionBody, CImage } from '@coreui/react'
import React, { useEffect, useState } from 'react'
import { useSelector } from 'react-redux'
import { getPolarisInferApi } from 'src/api/serviceInfo'
import { usePropsContext } from 'src/contexts/PropsContext'
import { selectSecondsTimeRange } from 'src/store/reducers/timeRangeReducer'
import 'react-image-lightbox/style.css'
import Lightbox from 'react-image-lightbox'
import { getStep } from 'src/utils/step'
import GlossaryTable from './GlossaryTable'
import { useDebounce } from 'react-use'

function PolarisMetricsInfo() {
  const [image, setImage] = useState()
  const [inferCause, setInferCause] = useState()
  const { serviceName, endpoint } = usePropsContext()
  const { startTime, endTime } = useSelector(selectSecondsTimeRange)
  const [loading, setLoading] = useState(false)
  const [isOpen, setIsOpen] = useState(false)
  window.global = window
  const global = window
  const getData = () => {
    if (startTime && endTime) {
      setLoading(true)
      getPolarisInferApi({
        startTime: startTime,
        endTime: endTime,
        service: serviceName,
        endpoint: endpoint,
        step: getStep(startTime, endTime),
      })
        .then((res) => {
          setImage(res?.inferMetricsPng)
          setInferCause(res?.inferCause)
          setLoading(false)
        })
        .catch((error) => {
          setLoading(false)
        })
    }
  }
  const handleImageClick = () => {
    setIsOpen(true)
  }
  // useEffect(() => {
  //   getData()
  // }, [startTime, endTime, serviceName, endpoint])
  //防抖避免跳转使用旧时间
  useDebounce(
    () => {
      getData()
    },
    300, // 延迟时间 300ms
    [startTime, endTime, serviceName, endpoint],
  )
  return (
    // {

    //   !loading && image?.length === 0 && inferCause?.length ===0 &&
    // }
    <CAccordionBody className="flex flex-row ">
      {image && (
        <div className="w-1/3">
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
      {inferCause && (
        <div className="w-2/3 px-5 py-2 flex flex-col">
          <div className="flex-1">{inferCause}</div>
          <div className="flex-1 text-xs">
            <GlossaryTable />
          </div>
        </div>
      )}
    </CAccordionBody>
  )
}
export default PolarisMetricsInfo
