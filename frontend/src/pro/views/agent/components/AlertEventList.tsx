import React, { useState, useEffect } from 'react'
import { Card, List, Tag, Spin, Empty, Button } from 'antd'
import { getAlertEventsApi } from 'src/core/api/alerts'
import { useSelector } from 'react-redux'
import { convertUTCToLocal } from 'src/core/utils/time'
import { useTranslation } from 'react-i18next'
import { AlertLevel, AlertStatus } from 'src/oss/views/alertEvents/components/AlertInfoCom'
import { useNavigate } from 'react-router-dom'

const AlertEventCard: React.FC<{ event: any }> = ({ event }) => {
  const navigate = useNavigate()

  const handleCardClick = () => {
    // 跳转到告警事件详情页面
    navigate(
      `/alerts/events/detail/${encodeURIComponent(event.alertId || event.id)}/${encodeURIComponent(event.id)}`,
    )
  }

  return (
    <Card
      size="small"
      className="mb-2 cursor-pointer hover:shadow-md transition-shadow alert-event-card w-full"
      onClick={handleCardClick}
    >
      <div className="flex items-start justify-between mb-2 min-w-0">
        <div className="flex items-center gap-2 min-w-0 flex-1">
          <AlertLevel level={event.severity} />
          <span className="font-medium text-sm truncate" title={event.name}>
            {event.name}
          </span>
        </div>
        <AlertStatus status={event.status} />
      </div>

      <div className="text-xs text-gray-500 mb-2">{convertUTCToLocal(event.createTime)}</div>

      {event.tagsDisplay && event.tagsDisplay.length > 0 && (
        <div className="flex flex-wrap gap-1 mb-2">
          {event.tagsDisplay.slice(0, 3).map((tag, index) => (
            <Tag key={index} className="text-xs">
              {tag.name} = {tag.value}
            </Tag>
          ))}
          {event.tagsDisplay.length > 3 && (
            <Tag className="text-xs">+{event.tagsDisplay.length - 3}</Tag>
          )}
        </div>
      )}

      {event.detail && (
        <div className="text-xs text-gray-600 truncate" title={event.detail}>
          {event.detail}
        </div>
      )}
    </Card>
  )
}

const AlertEventList: React.FC = () => {
  const { t } = useTranslation('oss/alertEvents')
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
          pageSize: 20,
        },
        filters: [
          {
            key: 'status',
            selected: ['firing'],
            name: t('status'),
          },
          {
            key: 'validity',
            selected: ['valid', 'other'],
            name: t('alertValidity'),
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
  }, [startTime, endTime, dataGroupId])
  const navigate = useNavigate()

  return (
    <Card
      className="h-full overflow-hidden"
      classNames={{ body: 'p-0 overflow-y-auto flex flex-col' }}
      styles={{ body: { height: 'calc(100% - 60px)' } }}
      title={
        <>
          告警事件{' '}
          <div className="text-xs text-gray-500">
            {t('showing')} {alertEvents.length} {t('activeAlerts')}
          </div>
        </>
      }
      extra={
        <Button type="link" onClick={() => navigate('/alerts/events')}>
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
            <List
              dataSource={alertEvents}
              renderItem={(event) => (
                <List.Item className="p-0 border-none">
                  <AlertEventCard event={event} />
                </List.Item>
              )}
            />
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
