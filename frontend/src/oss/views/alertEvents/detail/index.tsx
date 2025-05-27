/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */
import { Button, Splitter, theme } from 'antd'
import CurrentEventDetail from './CurrentEventDetail'
import Events from './Events'
import { useDebounce } from 'react-use'
import { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import { getAlertEventDetailApi, resolveAlertApi } from 'src/core/api/alerts'
import { FaCheck, FaLocationDot } from 'react-icons/fa6'
import { IoIosNotificationsOff } from 'react-icons/io'
import LoadingSpinner from 'src/core/components/Spinner'
import { useTranslation } from 'react-i18next'
import SilentAlert from './SilentAlert'
const AlertEventDetailPage = () => {
  const { t } = useTranslation('oss/alertEvents')
  const { alertId, eventId } = useParams()
  const [loading, setLoading] = useState(true)
  const [locateEvent, setLocateEvent] = useState(true)
  const [pagination, setPagination] = useState({
    pageIndex: 1,
    pageSize: 10,
    total: 0,
  })
  const [alertCheckId, setAlertCheckId] = useState(null)
  const [alertEvents, setAlertEvents] = useState([])
  const [detail, setDetail] = useState(null)
  const { useToken } = theme
  const { token } = useToken()

  const getAlertEvents = (locateEvent = false, pageIndex = null) => {
    setLoading(true)
    getAlertEventDetailApi({
      alertId,
      eventId,
      pagination: {
        currentPage: pageIndex || pagination.pageIndex,
        pageSize: pagination.pageSize,
      },
      locateEvent,
    })
      .then((res) => {
        setAlertCheckId(res?.alertCheckId)
        setDetail(res?.currentEvent)
        setAlertEvents(res?.events || [])
        setPagination({
          pageIndex: res?.pagination.currentPage,
          total: res?.pagination.total || 0,
          pageSize: res?.pagination.pageSize,
        })
      })
      .finally(() => {
        setLoading(false)
      })
  }
  useDebounce(
    () => {
      if (alertId && eventId) {
        getAlertEvents(true)
      }
    },
    300,
    [alertId, eventId],
  )
  const goToAlertLocation = () => {
    getAlertEvents(true)
    setLocateEvent(true)
    document.getElementById(eventId)?.scrollIntoView({ behavior: 'smooth', block: 'center' })
  }
  useEffect(() => {
    if (locateEvent)
      document.getElementById(eventId)?.scrollIntoView({ behavior: 'smooth', block: 'center' })
  }, [alertEvents])
  const onResolvedAlert = () => {
    setLoading(true)
    resolveAlertApi({ alertId })
      .then((res) => {
        getAlertEvents()
      })
      .catch(() => {
        setLoading(false)
      })
  }
  return (
    <Splitter
      layout="vertical"
      style={{
        height: `calc(100vh - ${import.meta.env.VITE_APP_CODE_VERSION === 'CE' ? 60 : 100}px)`,
        boxShadow: '0 0 10px rgba(0, 0, 0, 0.1)',
        // border: '1px solid rgb(229, 231, 235)',
      }}
    >
      <Splitter.Panel collapsible className="overflow-hidden" defaultSize={'50%'}>
        <CurrentEventDetail detail={detail} alertCheckId={alertCheckId} />
      </Splitter.Panel>
      <Splitter.Panel collapsible>
        <div
          className="flex flex-col h-full rounded-xl p-2 relative text-xs"
          style={{ backgroundColor: token.colorBgContainer }}
        >
          <LoadingSpinner loading={loading} />
          <div className="font-bold flex justify-between items-center">
            {t('historyTitle')}
            <div className="flex items-center">
              {detail?.lastStatus === 'resolved' ? (
                <Button
                  // color="green"
                  variant="outlined"
                  className="ml-2"
                  classNames={{ icon: 'flex items-center' }}
                  icon={<FaCheck size={20} />}
                  style={{
                    color: token.colorSuccessText,
                    backgroundColor: token.colorSuccessBg,
                    borderColor: token.colorSuccessBorder
                  }}
                  onMouseOver={(e) => {e.currentTarget.style.backgroundColor = token.colorSuccessBgHover; e.currentTarget.style.color = token.colorSuccessTextActive}}
                  onMouseLeave={(e) => {e.currentTarget.style.backgroundColor = token.colorSuccessBg}}
                >
                  {t('alertResolved')}
                </Button>
              ) : (
                <></>
                // <Button
                //   // color="green"
                //   variant="outlined"
                //   className="ml-2"
                //   classNames={{ icon: 'flex items-center' }}
                //   icon={<IoIosNotificationsOff size={20} />}
                //   onClick={onResolvedAlert}
                //   style={{
                //     color: token.colorSuccessText,
                //     backgroundColor: token.colorSuccessBg,
                //     borderColor: token.colorSuccessBorder
                //   }}
                //   onMouseOver={(e) => {e.currentTarget.style.backgroundColor = token.colorSuccessBgHover; e.currentTarget.style.color = token.colorSuccessTextActive}}
                //   onMouseLeave={(e) => {e.currentTarget.style.backgroundColor = token.colorSuccessBg}}
                // >
                //   {t('onResolved')}
                // </Button>
              )}

              <SilentAlert alertId={alertId} />
              <Button
                color="primary"
                variant="outlined"
                className="ml-2"
                icon={<FaLocationDot />}
                onClick={() => goToAlertLocation()}
                // style={{ backgroundColor: token.colorFillTertiary }}
              >
                {t('location')}
              </Button>
            </div>
          </div>
          <Events
            eventId={eventId}
            alertCheckId={alertCheckId}
            alertEvents={alertEvents}
            pageSize={pagination.pageSize}
            pageIndex={pagination.pageIndex}
            total={pagination.total}
            changePagination={(pageIndex, pageSize) => {
              setPagination({ ...pagination, pageIndex, pageSize })
              getAlertEvents(false, pageIndex)
              setLocateEvent(false)
            }}
          />
        </div>
      </Splitter.Panel>
    </Splitter>
  )
}
export default AlertEventDetailPage
