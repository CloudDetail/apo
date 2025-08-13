import React from 'react'
import { Button, Tooltip } from 'antd'
import { ReloadOutlined } from '@ant-design/icons'
import styles from '../AGUIChat.module.scss'
import { useTranslation } from 'react-i18next'

interface ChatHeaderProps {
  siderCollapsed: boolean
  onToggleSider: () => void
  onOpenSettings: () => void
  onRefreshSessions: () => void
}

const ChatHeader: React.FC<ChatHeaderProps> = ({ onRefreshSessions }) => {
  const { t } = useTranslation('pro/agent')

  return (
    <div className={styles.header}>
      <div className={styles.headerLeft}>
        {/* <Tooltip title={siderCollapsed ? '展开侧边栏' : '收起侧边栏'}>
          <Button
            type="text"
            icon={siderCollapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
            onClick={onToggleSider}
            className={styles.siderTrigger}
          />
        </Tooltip> */}
        <h2 className={styles.headerTitle}>{t('header.title')}</h2>
      </div>
      <div className={styles.headerActions}>
        {/* <Tooltip title="设置">
          <Button
            type="text"
            icon={<SettingOutlined />}
            onClick={onOpenSettings}
            className={styles.iconBtn}
          />
        </Tooltip> */}

        {/* 会话刷新按钮 - 仅在启用会话管理时显示 */}

        <Tooltip title={t('header.refresh')}>
          <Button
            type="text"
            icon={<ReloadOutlined />}
            onClick={onRefreshSessions}
            className={styles.iconBtn}
          />
        </Tooltip>
      </div>
    </div>
  )
}

export default ChatHeader
