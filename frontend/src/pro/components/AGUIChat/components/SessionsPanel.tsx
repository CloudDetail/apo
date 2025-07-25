import React from 'react'
import { Layout } from 'antd'

const { Sider } = Layout

interface Session {
  id: string
  title: string
  updatedAt: Date | string
}

interface SessionsPanelProps {
  sessions: Session[]
  currentSession: Session | null
  siderCollapsed: boolean
  onCreateSession: () => void
  onSelectSession: (sessionId: string) => void
  onDeleteSession: (sessionId: string) => void
}

const SessionsPanel: React.FC<SessionsPanelProps> = ({
  sessions,
  currentSession,
  siderCollapsed,
  onCreateSession,
  onSelectSession,
  onDeleteSession,
}) => {
  // 安全地获取日期字符串
  const getDateString = (timestamp: Date | string) => {
    try {
      if (timestamp instanceof Date) {
        return timestamp.toLocaleDateString()
      } else if (timestamp) {
        return new Date(timestamp).toLocaleDateString()
      }
      return '无效日期'
    } catch {
      return '无效日期'
    }
  }

  return (
    <Sider
      width={300}
      collapsed={siderCollapsed}
      collapsedWidth={0}
      trigger={null}
      collapsible
      className="sessions-sider"
    >
      <div className="sessions-panel">
        <div className="sessions-header">
          <button onClick={onCreateSession} className="new-session-btn">
            ➕ 新建对话
          </button>
        </div>

        <div className="sessions-list">
          {sessions.map((session) => (
            <div
              key={session.id}
              className={`session-item ${currentSession?.id === session.id ? 'active' : ''}`}
              onClick={() => onSelectSession(session.id)}
            >
              <div className="session-info">
                <div className="session-title">{session.title}</div>
                <div className="session-date">{getDateString(session.updatedAt)}</div>
              </div>
              <button
                className="delete-btn"
                onClick={(e) => {
                  e.stopPropagation()
                  onDeleteSession(session.id)
                }}
              >
                🗑️
              </button>
            </div>
          ))}
        </div>
      </div>
    </Sider>
  )
}

export default SessionsPanel
