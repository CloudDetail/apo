import { marked } from 'marked'
import { RiRobot2Line } from 'react-icons/ri'
import 'github-markdown-css/github-markdown-dark.css'
import { CSpinner } from '@coreui/react'
import { Button } from 'antd'
import React from 'react'
import DataViewComponents from './DataViewComponents'
import styles from './index.module.scss'

// import remarkMath from 'remark-math'
interface AIMessageProps {
  message: string
  loading: boolean
  reportUrlList?: string[]
  components: any[]
}

const AIMessage: React.FC<AIMessageProps> = ({
  message,
  loading,
  reportUrlList = [],
  components = [],
}) => {
  const renderMarkdown = () => {
    return {
      __html: marked(message),
    }
  }

  return (
    <div>
      <div
        className="w-4/5 overflow-hidden flex mb-4 flex-shrink-0"
        style={{
          // whiteSpace: 'pre-line',
          wordBreak: 'break-all',
          overflow: 'hidden',
        }}
      >
        <div className="m-2 flex-shrink-0 flex-grow-0">
          <RiRobot2Line size={24} />
        </div>
        <div className="overflow-hidden">
          {message?.length > 0 && (
            <div
              dangerouslySetInnerHTML={renderMarkdown()}
              className={`p-3 markdown-body rounded-2xl ${loading ? styles.loading : ''}`}
              style={{ fontSize: 14, background: '#3e3f4c', minHeight: 0 }}
            />
          )}
          {reportUrlList?.map((reportUrl, index) => (
            <div>
              <Button type="link" target="_blank" href={reportUrl}>
                延时报告_{index + 1}
              </Button>
            </div>
          ))}
          {loading && message?.length === 0 && <CSpinner variant="grow" />}
        </div>
      </div>
      <div className="mx-8">
        {components?.length > 0 && <DataViewComponents components={components} />}
      </div>
    </div>
  )
}

export default React.memo(AIMessage)
