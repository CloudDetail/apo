/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useEffect, useRef, useState, useMemo } from 'react'
import { CAccordion, CAccordionHeader, CAccordionItem, CImage } from '@coreui/react'
import StatusInfo from 'src/core/components/StatusInfo'
import InstanceInfo from './InstanceInfo'
import ErrorInstanceInfo from './ErrorInstance/ErrorInstanceInfo'
import MetricsDashboard from './MetricsDashboard'
import LogsInfo from './LogsInfo'
import TraceInfo from './TraceInfo'
import K8sInfo from './K8sInfo'
import { LuLayoutDashboard } from 'react-icons/lu'
import { TbNorthStar } from 'react-icons/tb'
import { cilDescription } from '@coreui/icons'
import CIcon from '@coreui/icons-react'
import { PiPath } from 'react-icons/pi'
import ThumbsUp from 'src/core/assets/icons/thumbsUp.svg'
import { usePropsContext } from 'src/core/contexts/PropsContext'
import PolarisMetricsInfo from './PolarisMetricsInfo/PolarisMetricsInfo'
import { selectProcessedTimeRange } from 'src/core/store/reducers/timeRangeReducer'
import { useSelector } from 'react-redux'
import AlertInfoTabs from './AlertInfo/AlertInfoTabs'
import SqlMetrics from './SqlMetrics'
import EntryImpact from './EntryImpact'
import { useTranslation } from 'react-i18next'

export default function InfoUni() {
  const { t, i18n } = useTranslation('oss/serviceInfo')
  const [openedPanels, setOpenedPanels] = useState({
    polarisMetrics: true,
  })
  const [statusPanels, setStatusPanels] = useState({
    instance: 'unknown',
    k8s: 'unknown',
    error: 'unknown',
    alert: 'unknown',
    impact: 'unknown',
  })
  const [dashboardVariable, setDashboardVariable] = useState(null)
  const { serviceName } = usePropsContext()
  const storeTimeRange = useSelector((state) => state.timeRange)

  const refs = useRef({})
  const { startTime, endTime } = useSelector(selectProcessedTimeRange)
  const handleToggle = (key) => {
    console.log(key, openedPanels)
    if (!openedPanels[key]) {
      const buttonHtml = refs.current[key].querySelector('button')
      if (buttonHtml) {
        if (buttonHtml.getAttribute('aria-expanded') === 'true') {
          buttonHtml.click()
        }
      }
    }
    setOpenedPanels((prev) => ({
      ...prev,
      [key]: true,
    }))
  }
  useEffect(() => {
    console.log(openedPanels)
  }, [openedPanels])
  const handlePanelStatus = (key, status) => {
    if (status !== 'normal') {
      handleToggle(key)
    }
    setStatusPanels((prev) => ({
      ...prev,
      [key]: status,
    }))
  }

  const mockAccordionList = [
    {
      key: 'impact',
      loadBeforeOpen: true,
      name: t('contentIndex.impactAnalysis', { serviceName }),
      component: EntryImpact,

      componentProps: {
        handlePanelStatus: (status) => handlePanelStatus('impact', status),
      },
    },
    {
      key: 'alert',
      loadBeforeOpen: true,
      name: t('contentIndex.alertEvents', { serviceName }),
      component: AlertInfoTabs,
      componentProps: {
        handlePanelStatus: (status) => handlePanelStatus('alert', status),
      },
    },

    {
      key: 'instance',
      loadBeforeOpen: true,
      name: t('contentIndex.serviceEndpointInstance', { serviceName }),
      component: InstanceInfo,
      componentProps: {
        handleToggle,
        handlePanelStatus: (status) => handlePanelStatus('instance', status),
        prepareVariable: (props) => setDashboardVariable(props),
      },
    },
    {
      key: 'k8s',
      loadBeforeOpen: true,
      name: t('contentIndex.k8sEvents', { serviceName }),
      component: K8sInfo,
      componentProps: {
        handleToggle,
        handlePanelStatus: (status) => handlePanelStatus('k8s', status),
      },
    },
    {
      key: 'error',
      loadBeforeOpen: true,
      name: t('contentIndex.errorInstances', { serviceName }),
      component: ErrorInstanceInfo,
      componentProps: {
        handleToggle,
        handlePanelStatus: (status) => handlePanelStatus('error', status),
      },
    },
    {
      key: 'sql',
      loadBeforeOpen: true,
      name: t('contentIndex.databaseCalls', { serviceName }),
      component: SqlMetrics,
    },
    {
      key: 'metrics',
      showHeader: true,
      icon: <LuLayoutDashboard />,
      name: t('contentIndex.customMetricsDashboard'),
      component: MetricsDashboard,
      componentProps: {
        variable: dashboardVariable,
        startTime,
        endTime,
        storeTimeRange,
      },
    },
    {
      key: 'polarisMetrics',
      showHeader: true,
      defaultActibe: true,
      icon: <TbNorthStar />,
      name: (
        <div className="inline-flex items-center">
          {t('contentIndex.polarisMetricsRecommendation')}
          <CImage src={ThumbsUp} width={32} className="ml-2" />
        </div>
      ),
      component: PolarisMetricsInfo,
    },
    {
      key: 'logs',
      showHeader: true,
      icon: <CIcon icon={cilDescription} />,
      name: t('contentIndex.logs'),
      component: LogsInfo,
    },
    {
      key: 'trace',
      showHeader: true,
      icon: <PiPath />,
      name: t('contentIndex.trace'),
      component: TraceInfo,
    },
  ]

  const [accordionList, setAccordionList] = useState(mockAccordionList)

  useEffect(() => {
    setAccordionList([])
    setAccordionList(mockAccordionList)
  }, [serviceName])

  return (
    <>
      {accordionList.map((item, index) => {
        const MemoizedComponent = useMemo(() => item.component, [item.component])
        const componentProps = { ...item.componentProps, variable: dashboardVariable }

        return (
          <CAccordion
            key={item.key}
            activeItemKey={openedPanels[item.key] ? item.key : ''}
            className="mb-2.5"
          >
            <CAccordionItem itemKey={item.key}>
              <CAccordionHeader
                ref={(el) => (refs.current[item.key] = el)}
                style={{ backgroundColor: item.status === 'error' ? '' : '' }}
                onClick={() => handleToggle(item.key)}
              >
                {statusPanels[item.key] && (
                  <div className="w-[32px] h-[32px]">
                    <StatusInfo status={statusPanels[item.key]} />
                  </div>
                )}
                {item.icon && item.icon}
                <span className="ml-2">{item.name}</span>
              </CAccordionHeader>
              <MemoizedComponent {...componentProps} />
            </CAccordionItem>
          </CAccordion>
        )
      })}
    </>
  )
}
