import React from 'react'
import { AGUIProvider } from '../../contexts/AGUIContext'
import AGUIChat from '../../components/AGUIChat'
import { Splitter } from 'antd'
import AlertEventList from './components/AlertEventList'
import './index.css'
import { contentHeight } from 'src/constants'

const AgentPage: React.FC = () => {
  return (
    <AGUIProvider
      initialConfig={{
        baseUrl: (import.meta as any).env?.VITE_AGUI_BASE_URL || '/ai/api/v1/srechat',
        // 'http://localhost:8081/mock',
        apiKey: (import.meta as any).env?.VITE_AGUI_API_KEY || '',
      }}
    >
      <div className="overflow-hidden" style={{ height: contentHeight }}>
        <Splitter
          style={{ height: 'calc(100vh - 120px)', overflow: 'hidden' }}
          onResize={(sizes) => {
            console.log('Splitter resized:', sizes)
          }}
        >
          <Splitter.Panel defaultSize="30%" collapsible>
            <AlertEventList />
          </Splitter.Panel>
          <Splitter.Panel collapsible>
            <div className="agent-container h-full overflow-hidden">
              <AGUIChat />
            </div>
          </Splitter.Panel>
        </Splitter>
      </div>
    </AGUIProvider>
  )
}

export default AgentPage

// 如果需要在其他组件中使用Context，可以这样做：
/*
import { useAGUIChat, useAGUISession, useAGUIState } from '../../contexts/AGUIContext'

const SomeComponent: React.FC = () => {
  // 只使用聊天相关的状态
  const { messages, loading, sendMessage } = useAGUIChat()
  
  // 只使用会话相关的状态
  const { sessions, currentSession, createSession } = useAGUISession()
  
  // 只使用应用状态
  const { error, isConnected } = useAGUIState()
  
  return (
    <div>
      {error && <div>错误: {error}</div>}
      {!isConnected && <div>连接断开</div>}
      <div>当前会话: {currentSession?.title}</div>
      <div>消息数量: {messages.length}</div>
    </div>
  )
}
*/
