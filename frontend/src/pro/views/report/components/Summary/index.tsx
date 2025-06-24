import { Card, Tag } from 'antd'
import { ReactSVG } from 'react-svg'
import Error from 'src/pro/assets/icons/errorType.svg'
import Slow from 'src/pro/assets/icons/slowType.svg'
import Infra from 'src/pro/assets/icons/infraType.svg'
import { useTranslation } from 'react-i18next'
import { ReportType, TagsParams } from '../../type'
import ApplicationInfo from './ApplicationInfo'
import { marked } from 'marked'
import './github-markdown.css'
const result = [
  {
    icon: '🕒',
    label: '异常时间',
    value: '2025-05-15 12:03~12:07',
  },
  {
    icon: '📊',
    label: '异常概况',
    value: (
      <ul>
        <li>- 网络方向异常线程占比从 15% 快速增长至 70%，已持续 5 分钟</li>
        <li>- 总线程数 128，异常线程数峰值达 90</li>
      </ul>
    ),
  },
  {
    icon: '🧭',
    label: '初步结论',
    value: '系统性网络阻塞，可能访问下游超时',
  },
  {
    icon: '📎',
    label: '日志支持',
    value: '发现相关异常日志 12 条',
  },
  {
    icon: '🌐',
    label: '网络探测',
    value: 'auth-service RTT 从 25ms 升高到 850ms',
  },
]
function getReportIcon(type: ReportType) {
  return type === 'error' ? Error : type === 'slow' ? Slow : Infra
}

interface OverviewProps {
  timestamp: number
  detail: string
  reason: string
  tags: TagsParams
}
const Summary = ({
  type,
  overview,
  cause,
}: {
  type: ReportType
  overview: OverviewProps
  cause: string
}) => {
  const { t } = useTranslation('pro/report')
  const renderMarkdown = () => {
    return {
      __html: marked(overview.reason),
    }
  }
  return (
    <Card>
      <div className="flex items-center">
        <div className="rounded-md bg-[var(--ant-color-bg-layout)] p-2">
          <ReactSVG
            src={getReportIcon(type)}
            beforeInjection={(svg) => {
              svg.setAttribute('width', '32')
              svg.setAttribute('height', '32')
            }}
          />
        </div>
        <div>
          <div className="text-lg font-bold pl-2">{t(type)}</div>
          <Tag color="error" bordered={false} className="text-sm ml-1">
            {cause}
          </Tag>
        </div>
      </div>
      <ApplicationInfo tags={overview.tags} />
      {/* <div>{overview.detail}</div> */}
      <div className="p-4 pb-0">
        <div className="items-start p-3 bg-[var(--ant-color-bg-layout)] rounded-lg transition-colors">
          {/* {result.map((metric) => (
            <div key={metric.label} className="flex mb-2">
              <span className="text-xl mr-3">{metric.icon}</span>
              <div className="flex">
                <div className=" font-medium text-gray-500 dark:text-gray-400">{metric.label}</div>
                ：
                <div className="text-sm font-semibold text-gray-900 dark:text-gray-100">
                  {metric.value}
                </div>
              </div>
            </div>
          ))} */}
          <div
            dangerouslySetInnerHTML={renderMarkdown()}
            className={`markdown-body`}
            style={{ fontSize: 14, background: 'bg-[var(--ant-color-bg-layout)]', minHeight: 0 }}
          />
        </div>
      </div>
    </Card>
  )
}
export default Summary
