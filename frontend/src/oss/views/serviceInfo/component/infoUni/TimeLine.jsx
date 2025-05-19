/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState, useEffect, useRef } from 'react'
import { usePropsContext } from 'src/core/contexts/PropsContext'
import { convertTime, TimestampToISO } from 'src/core/utils/time'
import './timeline.css'
import { TimeLineTypeApiMap } from 'src/constants'
import { useTranslation } from 'react-i18next'
import { Button, ConfigProvider, List, Slider, Tooltip } from 'antd'

const Timeline = (props) => {
  const { t } = useTranslation('oss/serviceInfo')
  const { type, startTime, endTime, instance, pid, nodeName, containerId } = props
  const { serviceName, endpoint } = usePropsContext()
  const [chronoList, setChronoList] = useState([])
  const [loading, setLoading] = useState(false)
  const [sliderValue, setSliderValue] = useState([startTime, endTime])
  const timelineRef = useRef(null)
  const [marks, setMarks] = useState({})

  useEffect(() => {
    const result = {}
    let segments = 6
    const step = (endTime - startTime) / segments

    for (let i = 0; i <= segments; i++) {
      const timestampMicro = Math.round(startTime + step * i)
      result[timestampMicro] = convertTime(timestampMicro, 'HH:MM')
    }

    setMarks(result)
  }, [startTime, endTime])

  useEffect(() => {
    if (startTime && endTime) {
      setLoading(true)
      const params = nodeName
        ? {
            pid,
            nodeName,
            containerId,
          }
        : { instance }
      TimeLineTypeApiMap[type]({
        startTime: sliderValue[0],
        endTime: sliderValue[1],
        service: serviceName,
        endpoint: endpoint,
        ...params,
      })
        .then((res) => {
          // 函数式更新
          setChronoList(res)
          setLoading(false)
        })
        .catch((error) => {
          setLoading(false)
          // setChronoList([])
        })
    }
  }, [sliderValue, type, serviceName, endpoint, instance])

  const scrollTimeline = (direction) => {
    const scrollAmount = 100
    if (direction === 'left') {
      timelineRef.current.scrollBy({ left: -scrollAmount, behavior: 'smooth' })
    } else {
      timelineRef.current.scrollBy({ left: scrollAmount, behavior: 'smooth' })
    }
  }

  const toPage = (item) => {
    const isValidItem = item && typeof item === 'object'
    const from = TimestampToISO(sliderValue[0])
    const to = TimestampToISO(sliderValue[1])

    const query = new URLSearchParams()

    if (isValidItem) {
      query.set('service', item.serviceName)
      query.set('endpoint', item.endpoint)
      query.set('instance', item.instanceId)
      query.set('traceId', item.traceId)
    } else {
      query.set('instance', instance)
      query.set('service', serviceName)
    }

    switch (type) {
      case 'errorLogs':
      case 'logsInfo':
        query.set('logs-from', from)
        query.set('logs-to', to)
        break

      case 'traceLogs':
        query.set('trace-from', from)
        query.set('trace-to', to)
        if (isValidItem) {
          sessionStorage.setItem('openJaegerModalAfterLoad', 'true')
        }
        break

      default:
        break
    }

    const basePath = type === 'traceLogs' ? '#/trace/fault-site' : '#/logs/fault-site'
    const fullUrl = `${window.location.origin}/${basePath}?${query.toString()}`
    window.open(fullUrl, '_blank')
  }

  return (
    <div className="w-full flex flex-col text-xs h-[260px]">
      <ConfigProvider
        theme={{
          components: {
            Slider: {
              trackBg: '#69b1ff',
              trackHoverBg: '#69b1ff',
              dotActiveBorderColor: '#91caff',
              handleColor: '#91caff',
              handleActiveColor: '#1677ff',
            },
          },
        }}
      >
        <Slider
          className="mx-4 text-xs"
          tooltip={{
            formatter: (value) => convertTime(value, 'yyyy-mm-dd hh:mm:ss'),
          }}
          range
          defaultValue={[startTime, endTime]}
          min={startTime}
          max={endTime}
          marks={marks}
          onChangeComplete={setSliderValue}
        />
      </ConfigProvider>
      <div className="min-h-7 flex-grow relative text-center">
        <List
          loading={loading}
          itemLayout="horizontal"
          dataSource={chronoList}
          renderItem={(item, index) => (
            <List.Item
              className="cursor-pointer text-center p-1 text-xs "
              onClick={() => toPage(item)}
            >
              {item.startTime && (
                <List.Item.Meta
                  title={
                    <span className="text-[var(--ant-color-text-secondary)] text-xs">
                      {convertTime(item.startTime, 'yyyy-mm-dd hh:mm:ss') +
                        ' ' +
                        t(`timeLine.${type}`)}
                    </span>
                  }
                />
              )}
            </List.Item>
          )}
        />
        {chronoList?.length > 0 && (
          <Tooltip
            title={t('timeLine.viewAllTooltip', {
              type: t(`timeLine.${type}`),
              startTime: convertTime(sliderValue[0], 'yyyy-mm-dd hh:mm:ss'),
              endTime: convertTime(sliderValue[1], 'yyyy-mm-dd hh:mm:ss'),
            })}
          >
            <Button color="primary" variant="outlined" onClick={() => toPage()}>
              {t('timeLine.viewAll', { type: t(`timeLine.${type}`) })}
            </Button>
          </Tooltip>
        )}
      </div>
    </div>
  )
}

export default Timeline
