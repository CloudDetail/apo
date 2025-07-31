import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Collapse } from 'antd'
import MarkdownRender from 'src/core/components/Markdown/MarkdownRender'
import IntentRecognition from './ToolRender/IntentRecognition'
import DataScopeForm from './ToolRender/DataScopeForm'
import ProcessDefinition from './ToolRender/ProcessDefinition'
import AbnormalityDetection from './ToolRender/AbnormalityDetection'
import AnomalyCausalChainTable from './ToolRender/AnomalyCausalChainTable'
import Result from './ToolRender/Result'
import Logs from './ToolRender/Logs'
import Metrics from './ToolRender/Metrics'
import styles from '../AGUIChat.module.scss'
import ToolCollapseIcon from './ToolCollapseIcon'

interface ToolRenderProps {
  toolName: string
  title: string
  content?: string
}

const AgentToolDisplay: React.FC<ToolRenderProps> = ({ toolName, title, content }) => {
  const { t } = useTranslation('pro/agent')

  const [activeKey, setActiveKey] = useState([toolName])
  const [hovered, setHovered] = useState(false)
  // 监听完成状态，3 秒后自动关闭
  useEffect(() => {
    if (content && toolName !== 'result') {
      const timer = setTimeout(() => {
        setActiveKey(null)
      }, 5000)
      return () => clearTimeout(timer)
    }
  }, [content])

  return (
    <div>
      <Collapse
        ghost
        className={styles.toolCallCollapse}
        activeKey={activeKey}
        onChange={(key) => {
          setActiveKey(key)
        }}
        expandIcon={() => <ToolCollapseIcon hovered={hovered} />}
        items={[
          {
            key: toolName,
            label: (
              <div
                className="flex items-center "
                onMouseEnter={() => setHovered(true)}
                onMouseLeave={() => setHovered(false)}
              >
                <span className="mr-2">📈</span>
                <span className="text-[var(--ant-color-text-tertiary)] text-">{title}</span>
              </div>
            ),
            children: (
              <div className="m-[6px] px-3 relative border-l-2 border-[var(--ant-color-border)] ">
                {/* <span className="absolute top-0 left-[6px]  text-[var(--ant-color-text-tertiary)] text-sm border-l border-[var(--ant-color-border)] h-full"></span> */}
                <div className="bg-[var(--ant-color-bg-elevated)] rounded-2xl p-3">
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
                      case 'logs':
                        return <Logs args={content} />
                      case 'metrics':
                        return <Metrics args={content} />
                      default:
                        return <MarkdownRender content={content} />
                    }
                  })()}
                </div>
              </div>
            ),
          },
        ]}
      />
    </div>
  )
}

export default AgentToolDisplay
