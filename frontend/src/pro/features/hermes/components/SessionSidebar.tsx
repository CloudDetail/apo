import { Button, Dropdown, Empty, Input, Modal, Skeleton, Tooltip } from 'antd'
import { AppstoreOutlined, MenuFoldOutlined, MoreOutlined, PlusOutlined } from '@ant-design/icons'
import dayjs from 'dayjs'
import type { HermesSessionBookmark } from '../client/types.ts'
import styles from '../HermesAgentWorkspace.module.scss'

interface SessionSidebarProps {
  sessions: HermesSessionBookmark[]
  activeSessionId?: string
  loading: boolean
  onSelect: (storedSessionId: string) => void
  onNew: () => void
  onRename: (storedSessionId: string, title: string) => Promise<void>
  onDelete: (storedSessionId: string) => Promise<void>
  onOpenCapabilities?: () => void
  onClose?: () => void
}

export default function SessionSidebar({
  sessions,
  activeSessionId,
  loading,
  onSelect,
  onNew,
  onRename,
  onDelete,
  onOpenCapabilities,
  onClose,
}: SessionSidebarProps) {
  const rename = (session: HermesSessionBookmark) => {
    let title = session.title
    Modal.confirm({
      title: '重命名会话',
      content: (
        <Input
          defaultValue={session.title}
          maxLength={80}
          onChange={(event) => {
            title = event.target.value
          }}
        />
      ),
      okText: '保存',
      cancelText: '取消',
      onOk: () => onRename(session.storedSessionId, title),
    })
  }

  const remove = (session: HermesSessionBookmark) => {
    Modal.confirm({
      title: '删除会话',
      content: `确认删除“${session.title}”吗？删除后无法从当前会话列表恢复。`,
      okText: '删除',
      okButtonProps: { danger: true },
      cancelText: '取消',
      onOk: () => onDelete(session.storedSessionId),
    })
  }

  return (
    <aside className={styles.sessionSidebar}>
      <div className={styles.sidebarHeader}>
        <div>
          <span className={styles.eyebrow}>Hermes 运维助手</span>
          <strong>运维会话</strong>
        </div>
        {onClose && (
          <Tooltip title="收起会话列表">
            <Button type="text" icon={<MenuFoldOutlined />} onClick={onClose} />
          </Tooltip>
        )}
      </div>

      <Button
        type="primary"
        icon={<PlusOutlined />}
        block
        className={styles.newSessionButton}
        onClick={onNew}
      >
        新建会话
      </Button>

      <div className={styles.sessionList}>
        {loading ? (
          <Skeleton active paragraph={{ rows: 6 }} title={false} />
        ) : sessions.length === 0 ? (
          <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无历史会话" />
        ) : (
          sessions.map((session) => (
            <button
              type="button"
              key={session.storedSessionId}
              className={`${styles.sessionItem} ${
                activeSessionId === session.storedSessionId ? styles.activeSessionItem : ''
              }`}
              onClick={() => onSelect(session.storedSessionId)}
            >
              <span className={`${styles.sessionStatus} ${styles[session.status] || ''}`} />
              <span className={styles.sessionContent}>
                <strong title={session.title}>{session.title}</strong>
                <small title={dayjs(session.updatedAt).format('YYYY-MM-DD HH:mm:ss')}>
                  {dayjs(session.updatedAt).format('MM-DD HH:mm')}
                </small>
              </span>
              <Dropdown
                trigger={['click']}
                menu={{
                  items: [
                    { key: 'rename', label: '重命名', onClick: () => rename(session) },
                    { key: 'delete', label: '删除', danger: true, onClick: () => remove(session) },
                  ],
                }}
              >
                <span
                  role="button"
                  tabIndex={0}
                  className={styles.sessionMore}
                  onClick={(event) => event.stopPropagation()}
                  onKeyDown={(event) => event.stopPropagation()}
                >
                  <MoreOutlined />
                </span>
              </Dropdown>
            </button>
          ))
        )}
      </div>
      {onOpenCapabilities && (
        <div className={styles.capabilityEntry}>
          <Button
            type="text"
            block
            icon={<AppstoreOutlined />}
            className={styles.capabilityEntryButton}
            onClick={onOpenCapabilities}
          >
            运维任务库
          </Button>
          <small>复用已验证的运维任务方案</small>
        </div>
      )}
    </aside>
  )
}
