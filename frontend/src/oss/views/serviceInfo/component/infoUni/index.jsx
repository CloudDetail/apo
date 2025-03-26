/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react'
import { CImage } from '@coreui/react'
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
import AlertInfoTabs from './AlertInfo/AlertInfoTabs'
import SqlMetrics from './SqlMetrics'
import EntryImpact from './EntryImpact'
import { useTranslation } from 'react-i18next'
import { Collapse, Space, Tooltip } from 'antd'
import styles from './index.module.scss'
import { useServiceInfoContext } from 'src/oss/contexts/ServiceInfoContext'
import { FiDatabase } from 'react-icons/fi'
import { IoMdInformationCircleOutline } from 'react-icons/io'
import { SlArrowDown } from 'react-icons/sl'
export default function InfoUni() {
  const { t } = useTranslation('oss/serviceInfo')
  const statusPanels = useServiceInfoContext((ctx) => ctx.statusPanels)

  const { serviceName } = usePropsContext()

  const activeTabKey = useServiceInfoContext((ctx) => ctx.activeTabKey)
  const setActiveTabKey = useServiceInfoContext((ctx) => ctx.setActiveTabKey)

  const panelStyle = {
    marginBottom: 12,
    border: 'none',
  }
  function panelLabel(key, icon, extra) {
    return (
      <Space className="h-[32px]">
        {statusPanels[key] && (
          <div className="w-[32px] h-[32px]">
            <StatusInfo status={statusPanels[key]} />
          </div>
        )}
        {icon && <div className="w-[32px] h-[32px] flex items-center justify-center">{icon}</div>}
        {t(`contentIndex.${key}`, { serviceName })}
        {extra}
      </Space>
    )
  }
  const getItems = (panelStyle) => [
    {
      key: 'impact',
      forceRender: true,
      label: panelLabel(
        'impact',
        null,
        <Tooltip title={t('entryImpact.toastMessage')}>
          <IoMdInformationCircleOutline size={20} color="#f7c01a" className="mr-1" />
        </Tooltip>,
      ),
      children: <EntryImpact />,
      style: panelStyle,
    },
    {
      key: 'alert',
      forceRender: true,
      label: panelLabel('alert'),
      children: <AlertInfoTabs />,
      style: panelStyle,
    },

    {
      key: 'instance',
      forceRender: true,
      label: panelLabel('instance'),
      children: <InstanceInfo />,
      style: panelStyle,
    },
    {
      key: 'k8s',
      forceRender: true,
      label: panelLabel('k8s'),
      children: <K8sInfo />,
      style: panelStyle,
    },
    {
      key: 'error',
      forceRender: true,
      label: panelLabel('error'),
      children: <ErrorInstanceInfo />,
      style: panelStyle,
    },
    {
      key: 'sql',
      forceRender: true,
      label: panelLabel('sql', <FiDatabase />),
      style: panelStyle,
      children: <SqlMetrics />,
    },
    {
      key: 'metrics',
      showHeader: true,
      label: panelLabel('metrics', <LuLayoutDashboard />),
      style: panelStyle,
      children: <MetricsDashboard />,
    },
    {
      key: 'polarisMetrics',
      forceRender: true,
      label: panelLabel(
        'polarisMetrics',
        <TbNorthStar />,
        <CImage src={ThumbsUp} width={32} className="ml-2" />,
      ),
      children: <PolarisMetricsInfo />,
      style: panelStyle,
    },
    {
      key: 'logs',
      showHeader: false,
      label: panelLabel(
        'logs',
        <CIcon icon={cilDescription} />,
        <Tooltip title={t('logsInfo.toastMessage')}>
          <IoMdInformationCircleOutline size={20} color="#f7c01a" className="mr-1" />
        </Tooltip>,
      ),
      children: <LogsInfo />,
      style: panelStyle,
    },
    {
      key: 'trace',
      showHeader: false,
      label: panelLabel(
        'trace',
        <PiPath />,
        <Tooltip title={t('traceInfo.toastMessage')}>
          <IoMdInformationCircleOutline size={20} color="#f7c01a" className="mr-1" />
        </Tooltip>,
      ),
      children: <TraceInfo />,
      style: panelStyle,
    },
  ]
  return (
    <>
      <Collapse
        className={styles.collapse}
        bordered={false}
        defaultActiveKey={['polarisMetrics']}
        expandIconPosition={'end'}
        items={getItems(panelStyle)}
        activeKey={activeTabKey}
        onChange={setActiveTabKey}
        expandIcon={({ isActive }) => {
          return isActive ? (
            <SlArrowDown className={'rotate-180'} size={18} />
          ) : (
            <SlArrowDown className={'rotate-0'} size={18} />
          )
        }}
      />
    </>
  )
}
