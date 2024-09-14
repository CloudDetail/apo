import React, { useEffect, useRef, useState, useMemo } from 'react'
import { CAccordion, CAccordionHeader, CAccordionItem, CImage } from '@coreui/react'
import StatusInfo from 'src/components/StatusInfo'
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
import ThumbsUp from 'src/assets/icons/thumbsUp.svg'
import { usePropsContext } from 'src/contexts/PropsContext'
import PolarisMetricsInfo from './PolarisMetricsInfo/PolarisMetricsInfo'
import { selectProcessedTimeRange } from 'src/store/reducers/timeRangeReducer'
import { useSelector } from 'react-redux'
import AlertInfoTabs from './AlertInfo/AlertInfoTabs'
import SqlMetrics from './SqlMetrics'
import EntryImpact from './EntryImpact'

export default function InfoUni() {
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
      key: 'alert',
      loadBeforeOpen: true,
      name: <>{serviceName}的告警事件</>,
      component: AlertInfoTabs,
      componentProps: {
        handlePanelStatus: (status) => handlePanelStatus('alert', status),
      },
    },

    {
      key: 'instance',
      loadBeforeOpen: true,
      name: `${serviceName}的服务端点实例`,
      component: InstanceInfo,
      componentProps: {
        handleToggle,
        handlePanelStatus: (status) => handlePanelStatus('instance', status),
        prepareVariable: (props) => setDashboardVariable(props),
      },
    },
    {
      key: 'impact',
      loadBeforeOpen: true,
      name: <>{serviceName}的影响面分析</>,
      component: EntryImpact,

      componentProps: {
        handlePanelStatus: (status) => handlePanelStatus('impact', status),
      },
    },
    {
      key: 'k8s',
      loadBeforeOpen: true,
      name: `${serviceName}的k8s事件`,
      component: K8sInfo,
      componentProps: {
        handleToggle,
        handlePanelStatus: (status) => handlePanelStatus('k8s', status),
      },
    },
    {
      key: 'error',
      loadBeforeOpen: true,
      name: `${serviceName}的错误实例`,
      component: ErrorInstanceInfo,
      componentProps: {
        handleToggle,
        handlePanelStatus: (status) => handlePanelStatus('error', status),
      },
    },
    {
      key: 'sql',
      loadBeforeOpen: true,
      name: `${serviceName}的数据库调用`,
      component: SqlMetrics,
    },
    {
      key: 'metrics',
      showHeader: true,
      icon: <LuLayoutDashboard />,
      name: '用户自定义指标大盘',
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
          基于北极星因果指标的主要原因推荐: 根据数据确认是应用自身的问题，还是下游依赖的问题
          <CImage src={ThumbsUp} width={32} className="ml-2" />
        </div>
      ),
      component: PolarisMetricsInfo,
    },
    {
      key: 'logs',
      showHeader: true,
      icon: <CIcon icon={cilDescription} />,
      name: '日志',
      component: LogsInfo,
    },
    {
      key: 'trace',
      showHeader: true,
      icon: <PiPath />,
      name: '链路追踪',
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
