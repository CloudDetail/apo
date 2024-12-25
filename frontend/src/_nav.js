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
import { GrSystem } from 'react-icons/gr'
import TranslationCom from './oss/components/TranslationCom'

const translationUrl = 'oss/routes'
const commercialNav = []
const _nav = [
  {
    key: 'service',
    icon: <IoMdCloudOutline />,
    label: <TranslationCom text="servicesName" url={translationUrl} />,
    abbreviation: <TranslationCom text="servicesAbbreviationName" url={translationUrl} />,
    to: '/service',
  },
  {
    key: 'logs',
    label: <TranslationCom text="logsName" url={translationUrl} />,
    to: '/logs',
    icon: <LuFileText />,
    children: [
      {
        key: 'faultSite',
        label: <TranslationCom text="faultLogsName" url={translationUrl} />,
        to: '/logs/fault-site',
      },
      {
        key: 'full',
        label: <TranslationCom text="allLogsName" url={translationUrl} />,
        to: '/logs/full',
      },
    ],
  },
  {
    key: 'trace',
    icon: <PiPath />,
    label: <TranslationCom text="tracesName" url={translationUrl} />,
    to: '/trace',
  },
  {
    key: 'system',
    icon: <AiOutlineDashboard />,
    label: <TranslationCom text="overviewDashboardName" url={translationUrl} />,
    abbreviation: <TranslationCom text="overviewDashboardAbbreviationName" url={translationUrl} />,
    to: '/system-dashboard',
  },
  {
    key: 'basic',
    icon: <AiOutlineDashboard />,
    label: <TranslationCom text="infrastructureDashboardName" url={translationUrl} />,
    abbreviation: (
      <TranslationCom text="infrastructureDashboardAbbreviationName" url={translationUrl} />
    ),
    to: '/basic-dashboard',
  },
  {
    key: 'application',
    icon: <AiOutlineDashboard />,
    label: <TranslationCom text="applicationDashboardName" url={translationUrl} />,
    abbreviation: (
      <TranslationCom text="applicationDashboardAbbreviationName" url={translationUrl} />
    ),
    to: '/application-dashboard',
  },
  {
    key: 'middleware',
    icon: <AiOutlineDashboard />,
    label: <TranslationCom text="middlewareDashboardName" url={translationUrl} />,
    abbreviation: (
      <TranslationCom text="middlewareDashboardAbbreviationName" url={translationUrl} />
    ),
    to: '/middleware-dashboard',
  },
  {
    key: 'alerts',
    icon: <FaRegBell />,
    label: <TranslationCom text="alertsName" url={translationUrl} />,
    to: '/alerts',
  },
  {
    key: 'config',
    icon: <MdOutlineSettings />,
    label: <TranslationCom text="configurationsName" url={translationUrl} />,
    abbreviation: <TranslationCom text="configurationsAbbreviationName" url={translationUrl} />,
    to: '/config',
  },
  {
    key: 'manage',
    icon: <GrSystem />,
    label: <TranslationCom text="systemSettingsName" url={translationUrl} />,
    abbreviation: <TranslationCom text="systemSettingsAbbreviationName" url={translationUrl} />,
    children: [
      {
        key: 'userManage',
        label: <TranslationCom text="userManageName" url={translationUrl} />,
        to: '/system/user-manage',
      },
    ],
  },
]

export { _nav, commercialNav }
