import { useAGUIChat } from 'src/pro/contexts/AGUIContext'
import IntentRecognition from './IntentRecognition'
import { useEffect, useState } from 'react'
import DataScopeForm from './DataScopeForm'
import ProcessDefinition from './ProcessDefinition'
import AbnormalityDetection from './AbnormalityDetection'
import AnomalyCausalChainTable from './AnomalyCausalChainTable'
import Result from './Result'
import { useTranslation } from 'react-i18next'
import { Collapse } from 'antd'
import { LoadingOutlined, CheckCircleOutlined, CloseCircleOutlined } from '@ant-design/icons'
import MarkdownRender from 'src/core/components/Markdown/MarkdownRender'

interface ToolRenderProps {
  toolCall?: {
    id: string
    name: string
    args: string
    result?: any
    status?: 'pending' | 'completed' | 'error'
  }
  isCurrent?: boolean
}

const ToolRender: React.FC<ToolRenderProps> = ({ toolCall, isCurrent = false }) => {
  const { currentToolCall } = useAGUIChat()
  const displayToolCall = toolCall || currentToolCall
  const { t } = useTranslation('pro/agent')

  const [argsObj, setArgsObj] = useState({})
  const [activeKey, setActiveKey] = useState([displayToolCall?.id])
  useEffect(() => {
    if (!displayToolCall) return

    // 参数解析
    try {
      if (displayToolCall.args && displayToolCall.args.trim()) {
        const parsed = JSON.parse(displayToolCall.args)
        setArgsObj(parsed)
      }
    } catch (e) {
      console.log('参数解析失败:', e)
    }
  }, [displayToolCall?.args])

  // 监听完成状态，3 秒后自动关闭
  useEffect(() => {
    if (displayToolCall?.status === 'completed' && displayToolCall.name !== 'result') {
      const timer = setTimeout(() => {
        setActiveKey(null)
      }, 5000)
      return () => clearTimeout(timer)
    }
  }, [displayToolCall?.status, displayToolCall?.id])

  const getStatusText = () => {
    if (displayToolCall.status === 'error') return '执行失败'
    if (displayToolCall.status === 'completed' || displayToolCall.result) return '已完成'
    return '执行中...'
  }

  const getStatusClasses = () => {
    if (displayToolCall.status === 'error')
      return 'bg-[var(--ant-color-error-bg)] text-[var(--ant-color-error)]'
    if (displayToolCall.status === 'completed' || displayToolCall.result)
      return 'bg-[var(--ant-color-success-bg)] text-[var(--ant-color-success)]'
    return 'bg-[var(--ant-color-warning-bg)] text-[var(--ant-color-warning)]'
  }

  const getStatusIcon = () => {
    if (displayToolCall.status === 'error') {
      return <CloseCircleOutlined className="text-red-500 ml-2" />
    }
    if (displayToolCall.status === 'completed' || displayToolCall.result) {
      return <CheckCircleOutlined className="text-green-500 ml-2" />
    }
    return <LoadingOutlined className="text-blue-500 ml-2" />
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
            key: displayToolCall.id,
            label: (
              <div className="flex items-center ">
                <span className="mr-2">🛠️</span>
                <span className="text-[var(--ant-color-text-secondary)] text-">
                  {t(displayToolCall.name)}
                </span>
                {getStatusIcon()}
              </div>
            ),
            children: (
              <div className="pl-3 border-l border-[var(--ant-color-border)]">
                {(() => {
                  switch (displayToolCall.name) {
                    case 'intentRecognition':
                      return <IntentRecognition args={displayToolCall.args} />
                    case 'dataScope':
                      return (
                        <DataScopeForm args={argsObj} status={displayToolCall.status || 'render'} />
                      )
                    case 'processDefinition':
                      return <ProcessDefinition args={displayToolCall.args} />
                    case 'abnormalityDetection':
                      return <AbnormalityDetection args={argsObj} />
                    case 'causalChain':
                    case 'filterCausal':
                    case 'ebpfEvidence':
                      return <AnomalyCausalChainTable args={argsObj} />
                    case 'result':
                      return <Result args={argsObj} />
                    default:
                      return <MarkdownRender content={displayToolCall.args} />
                  }
                })()}
              </div>
            ),
          },
        ]}
      />
      {/* <div className="text-sm text-[var(--ant-color-text-secondary)]">
        {t(displayToolCall.name)}
      </div> */}
      {/* <div className="p-3">
        {displayToolCall.status === 'pending' && (
          <div className="flex items-center gap-2 mb-3">
            <Spin />
            <span className="text-sm text-[var(--ant-color-text-secondary)]">AI 正在思考...</span>
          </div>
        )}

        {displayToolCall.name === 'intentRecognition' && (
          <IntentRecognition args={displayToolCall.args} />
        )}
        {displayToolCall.name === 'dataScope' && (
          <DataScopeForm args={argsObj} status={displayToolCall.status || 'render'} />
        )}
        {displayToolCall.name === 'processDefinition' && (
          <ProcessDefinition args={displayToolCall.args} />
        )}
        {displayToolCall.name === 'abnormalityDetection' && <AbnormalityDetection args={argsObj} />}
        {['causalChain', 'filterCausal', 'ebpfEvidence'].includes(displayToolCall.name) && (
          <AnomalyCausalChainTable args={argsObj} />
        )}
        {displayToolCall.name === 'result' && <Result args={argsObj} />}
      </div> */}
      {/* {displayToolCall.result && (
        <div className="p-3">
          <div className="font-medium mb-2 text-sm text-[var(--ant-color-text)]">结果:</div>
          <pre className="p-2 rounded font-mono text-xs overflow-x-auto whitespace-pre-wrap bg-[var(--ant-color-bg-layout)] border border-[var(--ant-color-border)]">
            {JSON.stringify(displayToolCall.result, null, 2)}
          </pre>
        </div>
      )} */}
      {/* {displayToolCall.status === 'error' && (
        <div className="p-3 border-t border-[var(--ant-color-error-border)] bg-[var(--ant-color-error-bg)] text-[var(--ant-color-error)]">
          <div className="font-medium mb-2 text-sm">错误:</div>
          <div className="font-medium">工具调用执行失败</div>
        </div>
      )} */}
    </div>
  )
}

export default ToolRender
