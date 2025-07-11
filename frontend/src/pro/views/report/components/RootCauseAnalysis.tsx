import { Table } from 'antd'
import { LuCircleAlert } from 'react-icons/lu'
import { useSelector } from 'react-redux'
import { RootCauseTyoe } from '../type'
import { useTranslation } from 'react-i18next'
const getSeverityColor = (level: string, theme: 'light' | 'dark') => {
  const isDark = theme === 'dark'

  const colorMap: Record<string, { bg: string; text: string; border: string }> = {
    major: {
      bg: isDark ? 'bg-red-900/30' : 'bg-red-100',
      text: isDark ? 'text-red-200' : 'text-red-800',
      border: isDark ? 'border-red-800' : 'border-red-200',
    },
    minor: {
      bg: isDark ? 'bg-amber-900/30' : 'bg-amber-100',
      text: isDark ? 'text-amber-200' : 'text-amber-800',
      border: isDark ? 'border-amber-800' : 'border-amber-200',
    },
    normal: {
      bg: isDark ? 'bg-green-900/30' : 'bg-green-100',
      text: isDark ? 'text-green-200' : 'text-green-800',
      border: isDark ? 'border-green-800' : 'border-green-200',
    },
    default: {
      bg: isDark ? 'bg-gray-700' : 'bg-gray-100',
      text: isDark ? 'text-gray-200' : 'text-gray-800',
      border: isDark ? 'border-gray-600' : 'border-gray-200',
    },
  }

  const { bg, text, border } = colorMap[level] || colorMap.default
  return `${bg} ${text} ${border}`
}
type Level = 'major' | 'minor' | 'normal'
interface RootCauseAnalysisItem {
  level: Level
  reason: string
  type: RootCauseTyoe
}
interface RootCauseAnalysisData {
  cause: string
  other: RootCauseAnalysisItem[]
}
const RootCauseAnalysis = ({ cause, other }: RootCauseAnalysisData) => {
  const { t } = useTranslation('pro/report')
  const { theme } = useSelector((state) => state.settingReducer)
  const columns = [
    {
      title: t('errorLevel'),
      dataIndex: 'level',
      key: 'level',
      width: 120,
      render: (value) => (
        <div
          className={`inline-block w-[80px] text-xs text-center px-2 py-1 rounded-full ${getSeverityColor(value, theme)}`}
        >
          {t(`level.${value}`)}
        </div>
      ),
    },
    {
      title: t('type'),
      dataIndex: 'type',
      key: 'type',
      width: 200,
      render: (_, record) => {
        return (
          <div className="flex items-center">
            <div
              className={`inline-block w-2 h-2 rounded-full mr-2 ${record.level === 'major' ? 'bg-red-500' : record.level === 'minor' ? 'bg-amber-500' : 'bg-green-500'}`}
            ></div>
            <span className={`px-2 py-1 text-xs rounded-full `}>
              {t(`rootCauseType.${record.type}`)}
            </span>
          </div>
        )
      },
    },
    {
      title: t('reason'),
      dataIndex: 'reason',
      key: 'reason',
      render: (value) => <>{value}</>,
    },
  ]

  return (
    <>
      <div className="text-xs">
        <div className="border-l-4 border-red-500 pl-4 py-2 bg-[var(--custom-alert-error-bg)] rounded-r-md">
          <div className="flex items-start space-x-2 mb-2">
            <LuCircleAlert className="h-5 w-5 text-red-500 mt-0.5" />
            <h3 className="text-base text-[var(--ant-text-color)]">{t('main')}</h3>
          </div>
          <ul className="list-disc pl-6 space-y-1">
            <li className="text-[var(--ant-color-text)] text-sm">{cause}</li>
          </ul>
        </div>
        <Table columns={columns} dataSource={other} pagination={false} />
      </div>
    </>
  )
}
export default RootCauseAnalysis
