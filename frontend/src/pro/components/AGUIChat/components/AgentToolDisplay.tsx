import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Collapse } from 'antd'
import { CheckCircleOutlined } from '@ant-design/icons'
import MarkdownRender from 'src/core/components/Markdown/MarkdownRender'
import IntentRecognition from './ToolRender/IntentRecognition'
import DataScopeForm from './ToolRender/DataScopeForm'
import ProcessDefinition from './ToolRender/ProcessDefinition'
import AbnormalityDetection from './ToolRender/AbnormalityDetection'
import AnomalyCausalChainTable from './ToolRender/AnomalyCausalChainTable'
import Result from './ToolRender/Result'

interface ToolRenderProps {
  toolName: string
  title: string
  content?: string
}

const AgentToolDisplay: React.FC<ToolRenderProps> = ({ toolName, title, content }) => {
  const { t } = useTranslation('pro/agent')

  const [activeKey, setActiveKey] = useState([toolName])

  // 监听完成状态，3 秒后自动关闭
  useEffect(() => {
    if (content) {
      const timer = setTimeout(() => {
        setActiveKey(null)
      }, 5000)
      return () => clearTimeout(timer)
    }
  }, [content])

  const getStatusIcon = () => {
    // if (displayToolCall.status === 'error') {
    //   return <CloseCircleOutlined className="text-red-500 ml-2" />
    // }
    // if (displayToolCall.status === 'completed' || displayToolCall.result) {
    return <CheckCircleOutlined className="text-green-500 ml-2" />
    // }
    // return <LoadingOutlined className="text-blue-500 ml-2" />
  }

  return (
    <div>
      <Collapse
        ghost
        activeKey={activeKey}
        onChange={(key) => {
          setActiveKey(key)
        }}
        items={[
          {
            key: toolName,
            label: (
              <div className="flex items-center ">
                <span className="mr-2">📈</span>
                <span className="text-[var(--ant-color-text-secondary)] text-">{title}</span>
                {getStatusIcon()}
              </div>
            ),
            children: (
              <div className="pl-3 border-l border-[var(--ant-color-border)]">
                {(() => {
                  switch (toolName) {
                    case 'intentRecognition':
                      return <IntentRecognition args={content} />
                    case 'dataScope':
                      return <DataScopeForm args={content} />
                    case 'processDefinition':
                      return <ProcessDefinition args={content} />
                    case 'abnormalityDetection':
                      return <AbnormalityDetection args={content} />
                    case 'causalChain':
                    case 'filterCausal':
                    case 'ebpfEvidence':
                      return <AnomalyCausalChainTable args={content} />
                    case 'result':
                      return <Result args={content} />
                    default:
                      return <MarkdownRender content={content} />
                  }
                })()}
              </div>
            ),
          },
        ]}
      />
    </div>
  )
}

export default AgentToolDisplay
