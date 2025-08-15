import React, { useState, useEffect } from 'react'
import { Card, List, Spin, Empty, Button, Tag } from 'antd'
import { getAlertEventsApi } from 'src/core/api/alerts'
import { useSelector } from 'react-redux'
import { convertUTCToLocal } from 'src/core/utils/time'
import { useTranslation } from 'react-i18next'
import styles from './AlertEventList.module.scss'
import { FaAngleRight } from 'react-icons/fa'
import { AlertLevel } from 'src/oss/views/alertEvents/components/AlertInfoCom'
import { LuClock2, LuTag } from 'react-icons/lu'
import { FaMagnifyingGlass } from 'react-icons/fa6'
import DateTimeRangePicker from 'src/core/components/DateTime/DateTimeRangePicker'
import RefreshDateTime from 'src/core/components/DateTime/RefreshDateTime'

const AlertEventCard: React.FC<{ event: any }> = ({ event }) => {
  const { t } = useTranslation('pro/agent')

  const { startTime, endTime } = useSelector((state: any) => state.timeRange)
  const handleButtonClick = (e: React.MouseEvent<HTMLButtonElement>) => {
    e.stopPropagation()
    // 跳转到告警事件详情页面
    window.open(
      `/#/alerts/events/detail/${encodeURIComponent(event.alertId || event.id)}/${encodeURIComponent(event.id)}`,
      '_blank',
    )
  }
  const handleCardClick = () => {
    window.open(
      `/#/report?alertEventId=${encodeURIComponent(event.id)}&startTime=${startTime}&endTime=${endTime}`,
      '_blank',
    )
  }

  return (
    <Card
      size="small"
      className={`group mb-2 cursor-pointer hover:shadow-md transition-shadow w-full ${styles.alertEventCard}
`}
      onClick={handleCardClick}
    >
      <div className="flex items-start justify-between mb-2 min-w-0">
        <div className="flex items-center gap-2 min-w-0 flex-1">
          <span
            className="font-medium text-sm truncate group-hover:text-[var(--ant-color-info-text-hover)] transition-colors duration-200"
            title={event.name}
          >
            {event.name}
          </span>
        </div>

        <AlertLevel level={event.severity} />
      </div>

      <div className="text-xs text-[var(--ant-color-text-secondary)] mb-2 flex items-center gap-1">
        <FaMagnifyingGlass className="mr-3" />
        <span>{t('reasonDirection')}：</span>
        <span className="text-[var(--ant-color-text-heading)]">{event.alertDirection}</span>
      </div>
      <div className="text-xs text-[var(--ant-color-text-secondary)] mb-2 flex items-center gap-3">
        <LuClock2 className="text-sm" /> {convertUTCToLocal(event.createTime)}
      </div>

      {event.tagsDisplay && event.tagsDisplay.length > 0 && (
        <div className="flex gap-2 items-start">
          <LuTag className="shrink-0 mt-1 text-[var(--ant-color-text-secondary)]" />
          <div className="flex flex-wrap gap-1 mb-2 w-0 flex-1 overflow-hidden">
            {event.tagsDisplay.slice(0, 3).map((tag, index) => (
              <Tag
                key={index}
                className="text-xs text-[var(--ant-color-text-secondary)] truncate"
                bordered={false}
              >
                {tag.value}
              </Tag>
            ))}
            {event.tagsDisplay.length > 3 && (
              <Tag className="text-xs text-[var(--ant-color-text-secondary)]" bordered={false}>
                +{event.tagsDisplay.length - 3}
              </Tag>
            )}
          </div>
        </div>
      )}

      {/* {event.detail && (
        <div className="text-xs text-gray-600 truncate" title={event.detail}>
          {event.detail}
        </div>
      )} */}
      <div className="flex justify-between">
        <Button size="small" type="link" onClick={handleButtonClick} className="text-xs">
          {t('viewAlertDetails')}
        </Button>
        <div className="flex items-center gap-1.5 text-gray-400 group-hover:text-gray-300 transition-colors duration-200">
          <span className="text-xs">{t('viewReport')}</span>
          <FaAngleRight className="w-3.5 h-3.5 transform group-hover:translate-x-1 transition-transform duration-200" />
        </div>
      </div>
    </Card>
  )
}

const AlertEventList: React.FC = () => {
  const { t } = useTranslation('pro/agent')
  const [alertEvents, setAlertEvents] = useState([])
  const [loading, setLoading] = useState(false)
  const { startTime, endTime } = useSelector((state: any) => state.timeRange)
  const { dataGroupId } = useSelector((state: any) => state.dataGroupReducer)
  const fetchAlertEvents = async () => {
    if (!startTime || !endTime || dataGroupId === null) return

    setLoading(true)
    try {
      const response = await getAlertEventsApi({
        startTime,
        endTime,
        pagination: {
          currentPage: 1,
          pageSize: 51,
        },
        filters: [
          {
            key: 'status',
            selected: ['firing'],
          },
          {
            key: 'validity',
            selected: ['valid', 'other'],
          },
          {
            key: 'workflow.alert_direction',
            matchExpr: '*',
          },
        ],
        groupId: dataGroupId,
      })

      setAlertEvents((response as any)?.events || [])
    } catch (error) {
      console.error('Failed to fetch alert events:', error)
      setAlertEvents([])
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchAlertEvents()
    localStorage.setItem('refreshKey', '1m')
  }, [startTime, endTime, dataGroupId])
  const { rangeTypeKey } = useSelector((state) => state.timeRange)
  return (
    <Card
      className="h-full overflow-hidden"
      classNames={{ body: 'p-0 overflow-y-auto flex flex-col' }}
      styles={{ body: { height: 'calc(100% - 60px)' } }}
      title={
        <div className="relative">
          {t('alertIntelligentAnalysisReport')}
          <div className="flex items-center gap-2">
            <DateTimeRangePicker size="small" />
            {rangeTypeKey && <RefreshDateTime size="small" />}
          </div>
        </div>
      }
      extra={
        <Button type="link" onClick={() => window.open('/#/alerts/events', '_blank')}>
          {t('viewAll')}
        </Button>
      }
    >
      <div className="h-full flex flex-col">
        <div className="flex-1 overflow-auto p-3 alert-event-list">
          {loading ? (
            <div className="flex items-center justify-center h-full">
              <Spin size="large" />
            </div>
          ) : alertEvents.length > 0 ? (
            <>
              <List
                dataSource={alertEvents}
                renderItem={(event) => (
                  <List.Item className="p-0 border-none">
                    <AlertEventCard event={event} />
                  </List.Item>
                )}
              />
              {alertEvents.length > 50 && (
                <Button type="link" onClick={() => window.open('/alerts/events', '_blank')}>
                  {t('showingFirstFifty')}
                </Button>
              )}
            </>
          ) : (
            <Empty
              description={t('noAlertEvents')}
              className="mt-8"
              image={Empty.PRESENTED_IMAGE_SIMPLE}
            />
          )}
        </div>
      </div>
    </Card>
  )
}

export default AlertEventList
