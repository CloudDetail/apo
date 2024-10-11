import React, { useState, useEffect, useRef } from 'react'
import CIcon from '@coreui/icons-react'
import { cilChevronCircleLeftAlt, cilChevronCircleRightAlt } from '@coreui/icons'
import { useNavigate } from 'react-router-dom'
import { usePropsContext } from 'src/contexts/PropsContext'
import { convertTime, splitTimeRange, TimestampToISO } from 'src/utils/time'
import './timeline.css'
import { CListGroup, CListGroupItem } from '@coreui/react'
import Empty from 'src/components/Empty/Empty'
import { TimeLineTypeApiMap, TimeLineTypeTitleMap } from 'src/constants'
import LoadingSpinner from 'src/components/Spinner'

const Timeline = (props) => {
  const navigate = useNavigate()
  const { type, startTime, endTime, instance, pid, nodeName, containerId } = props
  const { serviceName, endpoint } = usePropsContext()
  const [chronoList, setChronoList] = useState([])
  const [activeKey, setActiveKey] = useState(-1)
  const [loading, setLoading] = useState(false)
  const timelineRef = useRef(null)

  useEffect(() => {
    const result = splitTimeRange(startTime, endTime)
    const newChronoList = result.map((time) => ({
      title: convertTime(time.start, 'HH:MM'),
      contentTitle: '错误日志',
      list: null,
      start: time.start,
      end: time.end,
    }))
    setChronoList(newChronoList)
    setActiveKey(0)
  }, [startTime, endTime])

  useEffect(() => {
    if (activeKey > -1 && chronoList.length > 0) {
      const item = chronoList[activeKey]
      if (item && !item.list && startTime && endTime) {
        setLoading(true)
        const params = nodeName
          ? {
              pid,
              nodeName,
              containerId,
            }
          : { instance }
        TimeLineTypeApiMap[type]({
          startTime: item.start,
          endTime: item.end,
          service: serviceName,
          endpoint: endpoint,
          ...params,
        })
          .then((res) => {
            // 函数式更新
            setChronoList((prevChronoList) => {
              const newChronoList = [...prevChronoList]
              newChronoList[activeKey].list = res ?? []
              return newChronoList
            })
            setLoading(false)
          })
          .catch((error) => {
            setLoading(false)
            // setChronoList([])
          })
      }
    }
  }, [activeKey, chronoList, type, serviceName, endpoint, instance])

  const scrollTimeline = (direction) => {
    const scrollAmount = 100
    if (direction === 'left') {
      timelineRef.current.scrollBy({ left: -scrollAmount, behavior: 'smooth' })
    } else {
      timelineRef.current.scrollBy({ left: scrollAmount, behavior: 'smooth' })
    }
  }

  const toPage = (item) => {
    let url = '/logs/fault-site?'
    switch (type) {
      case 'errorLogs':
        url = `/logs/fault-site?service=${item.serviceName}&endpoint=${item.endpoint}&instance=${item.instanceId}&traceId=${item.traceId}&logs-from=${TimestampToISO(chronoList[activeKey].start)}&logs-to=${TimestampToISO(chronoList[activeKey].end)}`
        break
      case 'logsInfo':
        url = `/logs/fault-site?service=${item.serviceName}&endpoint=${item.endpoint}&instance=${item.instanceId}&traceId=${item.traceId}&logs-from=${TimestampToISO(chronoList[activeKey].start)}&logs-to=${TimestampToISO(chronoList[activeKey].end)}`
        break
      case 'traceLogs':
        url = `/trace?service=${item.serviceName}&endpoint=${item.endpoint}&instance=${item.instanceId}&traceId=${item.traceId}&trace-from=${TimestampToISO(chronoList[activeKey].start)}&trace-to=${TimestampToISO(chronoList[activeKey].end)}`
        sessionStorage.setItem('openJaegerModalAfterLoad', 'true')
        break
      default:
        break
    }
    // url += `&from=${TimestampToISO(item.startTime)}&to=${TimestampToISO(item.endTime)}`
    navigate(url)
  }

  return (
    <div className="w-full flex flex-col h-full">
      <div className="timeline-container flex-grow-0 flex-shrink-0">
        <CIcon
          icon={cilChevronCircleLeftAlt}
          width={30}
          className="opacity-40 mx-2 cursor-pointer"
          onClick={() => scrollTimeline('left')}
        />
        <div className="timeline" ref={timelineRef}>
          {chronoList.map((item, index) => (
            <div key={index} className="timeline-segment">
              <div
                className={`timeline-item ${activeKey === index ? 'active' : ''}`}
                onClick={() => setActiveKey(index)}
              >
                <div className="timeline-item-title">{item.title}</div>
                <button className="timeline-button"></button>
              </div>
              <div className="timeline-line"></div>
            </div>
          ))}
        </div>
        <CIcon
          icon={cilChevronCircleRightAlt}
          width={30}
          className="opacity-40 mx-2 cursor-pointer"
          onClick={() => scrollTimeline('right')}
        />
      </div>

      <div className="min-h-7 flex-grow relative">
        <LoadingSpinner loading={loading}></LoadingSpinner>
        {!loading && chronoList[activeKey]?.list?.length === 0 && <Empty />}
        {chronoList[activeKey]?.list?.length > 0 && (
          <CListGroup className="h-full overflow-y-auto">
            {chronoList[activeKey].list.map((item, index) => (
              <CListGroupItem className="cursor-pointer" key={index} onClick={() => toPage(item)}>
                <div className="text-center">
                  {convertTime(item.startTime, 'yyyy-mm-dd hh:mm:ss')} {TimeLineTypeTitleMap[type]}
                </div>
              </CListGroupItem>
            ))}
          </CListGroup>
        )}
      </div>
    </div>
  )
}

export default Timeline
