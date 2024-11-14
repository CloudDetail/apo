import React from 'react'
import CIcon from '@coreui/icons-react'
import { cilSpeedometer } from '@coreui/icons'
import { CBadge } from '@coreui/react'
import { PiPath } from 'react-icons/pi'
import { LuFileText } from 'react-icons/lu'
import { AiOutlineDashboard } from 'react-icons/ai'
import { FaRegBell } from 'react-icons/fa'
import { MdOutlineSettings } from 'react-icons/md'
import { IoIosTrendingUp, IoMdCloudOutline } from 'react-icons/io'
import { TbWaveSawTool } from 'react-icons/tb'
const commercialNav = [
  {
    key: 'alertAnalyze',
    icon: <IoIosTrendingUp />,
    label: (
      <>
        告警分析
        <CBadge color="info" className="ml-2">
          免费公测
        </CBadge>
      </>
    ),
    to: '/alert-analyze',
  },
]

const VITE_SHOW_MIDWARE_HEALTH = import.meta.env.VITE_SHOW_MIDWARE_HEALTH === "special"

let _nav = [
  { key: 'service', icon: <IoMdCloudOutline />, label: '服务概览', to: '/service' },
  {
    key: 'logs',
    label: '日志检索',
    to: '/logs',
    icon: <LuFileText />,
    children: [
      { key: 'faultSite', label: '故障现场日志', to: '/logs/fault-site' },
      { key: 'full', label: '全量日志', to: '/logs/full' },
    ],
  },
  { key: 'trace', icon: <PiPath />, label: '链路追踪', to: '/trace' },
  {
    key: 'system',
    icon: <AiOutlineDashboard />,
    label: '全局资源大盘',
    abbreviation: '全局资源',
    to: '/system-dashboard',
  },
  {
    key: 'basic',
    icon: <AiOutlineDashboard />,
    label: '应用基础设施大盘',
    abbreviation: '基础设施',
    to: '/basic-dashboard',
  },
  {
    key: 'application',
    icon: <AiOutlineDashboard />,
    label: '应用指标大盘',
    abbreviation: '应用指标',
    to: '/application-dashboard',
  },
  {
    key: 'middleware',
    icon: <AiOutlineDashboard />,
    label: '中间件大盘',
    abbreviation: '中间件',
    to: '/middleware-dashboard',
  },
  {
    key: 'mysql',
    icon: <CIcon icon={cilSpeedometer} />,
    label: 'MySQL 大盘',
    abbreviation: 'MySQL',
    to: '/mysql-dashboard',
  },
  {
    key: 'alerts',
    icon: <FaRegBell />,
    label: '告警规则',
    to: '/alerts',
  },
  {
    key: 'config',
    icon: <MdOutlineSettings />,
    label: '配置中心',
    to: '/config',
  },
  {
    key: 'healthy',
    label: '服务健康状态',
    abbreviation: '服务健康',
    to: '/service-healthy',
    icon: <TbWaveSawTool />,
  },
]

if (!VITE_SHOW_MIDWARE_HEALTH) {
  _nav = _nav.filter((item) => item.key !== 'middleware' && item.key !== 'healthy')
}


export { _nav, commercialNav }
