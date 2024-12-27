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

const namespace = 'oss/routes'
const commercialNav = []
const _nav = [
  {
    key: 'service',
    icon: <IoMdCloudOutline />,
    label: <TranslationCom text="servicesName" space={namespace} />,
    abbreviation: <TranslationCom text="servicesAbbreviationName" space={namespace} />,
    to: '/service',
  },
  {
    key: 'logs',
    label: <TranslationCom text="logsName" space={namespace} />,
    to: '/logs',
    icon: <LuFileText />,
    children: [
      {
        key: 'faultSite',
        label: <TranslationCom text="faultLogsName" space={namespace} />,
        to: '/logs/fault-site',
      },
      {
        key: 'full',
        label: <TranslationCom text="allLogsName" space={namespace} />,
        to: '/logs/full',
      },
    ],
  },
  {
    key: 'trace',
    icon: <PiPath />,
    label: <TranslationCom text="tracesName" space={namespace} />,
    to: '/trace',
  },
  {
    key: 'system',
    icon: <AiOutlineDashboard />,
    label: <TranslationCom text="overviewDashboardName" space={namespace} />,
    abbreviation: <TranslationCom text="overviewDashboardAbbreviationName" space={namespace} />,
    to: '/system-dashboard',
  },
  {
    key: 'basic',
    icon: <AiOutlineDashboard />,
    label: <TranslationCom text="infrastructureDashboardName" space={namespace} />,
    abbreviation: (
      <TranslationCom text="infrastructureDashboardAbbreviationName" space={namespace} />
    ),
    to: '/basic-dashboard',
  },
  {
    key: 'application',
    icon: <AiOutlineDashboard />,
    label: <TranslationCom text="applicationDashboardName" space={namespace} />,
    abbreviation: <TranslationCom text="applicationDashboardAbbreviationName" space={namespace} />,
    to: '/application-dashboard',
  },
  {
    key: 'middleware',
    icon: <AiOutlineDashboard />,
    label: <TranslationCom text="middlewareDashboardName" space={namespace} />,
    abbreviation: <TranslationCom text="middlewareDashboardAbbreviationName" space={namespace} />,
    to: '/middleware-dashboard',
  },
  {
    key: 'alerts',
    icon: <FaRegBell />,
    label: <TranslationCom text="alertsName" space={namespace} />,
    to: '/alerts',
  },
  {
    key: 'config',
    icon: <MdOutlineSettings />,
    label: <TranslationCom text="configurationsName" space={namespace} />,
    abbreviation: <TranslationCom text="configurationsAbbreviationName" space={namespace} />,
    to: '/config',
  },
  {
    key: 'manage',
    icon: <GrSystem />,
    label: <TranslationCom text="systemSettingsName" space={namespace} />,
    abbreviation: <TranslationCom text="systemSettingsAbbreviationName" space={namespace} />,
    children: [
      {
        key: 'userManage',
        label: <TranslationCom text="userManageName" space={namespace} />,
        to: '/system/user-manage',
      },
    ],
  },
]

export { _nav, commercialNav }
