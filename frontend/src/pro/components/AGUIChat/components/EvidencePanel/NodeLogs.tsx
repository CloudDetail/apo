import React, { useState, useMemo } from 'react'
import { theme, Drawer, Button, Input } from 'antd'
import { SnippetsOutlined, ExpandOutlined, SearchOutlined, CopyOutlined } from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import dayjs from 'dayjs'
import type { FindingsLogItem } from '../../../../api/conversation'

const { useToken } = theme

interface NodeLogsProps {
  logs: FindingsLogItem[]
}

const NodeLogs: React.FC<NodeLogsProps> = ({ logs }) => {
  const { token } = useToken()
  const { t } = useTranslation('pro/agui')
  const [drawerOpen, setDrawerOpen] = useState(false)
  const [searchText, setSearchText] = useState('')
  const [levelFilter, setLevelFilter] = useState<string>('all')

  // 识别日志级别 (仅用于 Drawer)
  const getLogLevel = (message: string) => {
    const upperMessage = message.toUpperCase()
    if (upperMessage.includes('ERROR') || upperMessage.includes('FATAL')) return 'error'
    if (upperMessage.includes('WARN')) return 'warn'
    if (upperMessage.includes('INFO')) return 'info'
    if (upperMessage.includes('DEBUG')) return 'debug'
    return 'info'
  }

  // 获取日志级别标签 (仅用于 Drawer)
  const getLevelTag = (level: string) => {
    const colors = {
      error: 'red',
      warn: 'orange',
      info: 'blue',
      debug: 'green',
    }
    return colors[level as keyof typeof colors] || 'default'
  }

  // 过滤和搜索日志 (仅用于 Drawer)
  const filteredLogs = useMemo(() => {
    return logs.filter((log) => {
      const level = getLogLevel(log.message)
      const matchesLevel = levelFilter === 'all' || level === levelFilter
      const matchesSearch =
        searchText === '' || log.message.toLowerCase().includes(searchText.toLowerCase())
      return matchesLevel && matchesSearch
    })
  }, [logs, levelFilter, searchText])

  // 复制日志内容 (仅用于 Drawer)
  const copyLogContent = (log: FindingsLogItem) => {
    const timestamp = dayjs(Number(log.timestamp) / 1000).format('YYYY-MM-DD HH:mm:ss.SSS')
    const content = `[${timestamp}] ${log.message}`
    navigator.clipboard.writeText(content)
  }

  // 复制所有日志 (仅用于 Drawer)
  const copyAllLogs = () => {
    const content = filteredLogs
      .map((log) => {
        const timestamp = dayjs(Number(log.timestamp) / 1000).format('YYYY-MM-DD HH:mm:ss.SSS')
        return `[${timestamp}] ${log.message}`
      })
      .join('\n')
    navigator.clipboard.writeText(content)
  }

  return (
    <div className="space-y-3">
      <div className="space-y-2">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <SnippetsOutlined
              style={{
                fontSize: 16,
                color: token.colorTextSecondary,
              }}
            />
            <span
              className="text-base font-medium"
              style={{
                color: token.colorTextBase,
              }}
            >
              {t('evidencePanel.logs')}
            </span>
          </div>
          <Button
            type="text"
            size="small"
            icon={<ExpandOutlined />}
            onClick={() => setDrawerOpen(true)}
            style={{ color: token.colorTextSecondary }}
          />
        </div>
        <div
          style={{
            height: 1,
            backgroundColor: token.colorBorder,
          }}
        />
      </div>
      <div
        className="max-h-[500px] overflow-y-auto border rounded"
        style={{ backgroundColor: token.colorBgContainer }}
      >
        {logs?.map((log, index) => {
          return (
            <div
              key={index}
              className="flex gap-2 p-2 border-b last:border-b-0"
              style={{ borderColor: token.colorBorderSecondary }}
            >
              <div className="flex-1 min-w-0">
                <div className="text-xs mb-1" style={{ color: token.colorTextSecondary }}>
                  {dayjs(Number(log.timestamp) / 1000).format('YYYY-MM-DD HH:mm:ss')}
                </div>
                <div
                  className="text-xs cursor-pointer"
                  style={{
                    color: token.colorText,
                    display: '-webkit-box',
                    WebkitLineClamp: 2,
                    WebkitBoxOrient: 'vertical',
                    overflow: 'hidden',
                    wordBreak: 'break-word',
                  }}
                >
                  {log.message}
                </div>
              </div>
            </div>
          )
        })}
      </div>

      {/* 全量日志展示侧边栏 */}
      <Drawer
        title={
          <div className="flex items-center justify-between">
            <span>{t('evidencePanel.logs')}</span>
            <div className="flex gap-2">
              <Input
                placeholder="搜索日志内容..."
                prefix={<SearchOutlined />}
                value={searchText}
                onChange={(e) => setSearchText(e.target.value)}
                size="small"
                allowClear
                style={{ width: 200 }}
              />
              {/* <Select
                value={levelFilter}
                onChange={setLevelFilter}
                size="small"
                style={{ width: 100 }}
              >
                <Select.Option value="all">全部</Select.Option>
                <Select.Option value="error">ERROR</Select.Option>
                <Select.Option value="warn">WARN</Select.Option>
                <Select.Option value="info">INFO</Select.Option>
                <Select.Option value="debug">DEBUG</Select.Option>
              </Select> */}
              {/* <Button
                size="small"
                icon={<CopyOutlined />}
                onClick={copyAllLogs}
                title="复制所有日志"
              /> */}
            </div>
          </div>
        }
        placement="right"
        width="80%"
        zIndex={100000}
        open={drawerOpen}
        onClose={() => setDrawerOpen(false)}
        destroyOnClose={true}
        styles={{
          body: {
            padding: 0,
            height: '100%',
          },
        }}
      >
        <div
          style={{
            height: '100%',
            overflow: 'auto',
            backgroundColor: token.colorBgContainer,
            fontFamily: 'Monaco, Menlo, "Ubuntu Mono", monospace',
          }}
        >
          {filteredLogs?.map((log, index) => {
            const timestamp = dayjs(Number(log.timestamp) / 1000).format('YYYY-MM-DD HH:mm:ss.SSS')

            return (
              <div
                key={index}
                className="group hover:var(--ant-color-row-hover-bg) border-b"
                style={{ borderColor: token.colorBorderSecondary }}
              >
                <div className="flex items-start gap-3 p-3">
                  {/* 行号 */}
                  <div
                    className="text-xs font-mono text-right select-none"
                    style={{
                      color: token.colorTextTertiary,
                      minWidth: '40px',
                      paddingTop: '1px',
                    }}
                  >
                    {index + 1}
                  </div>

                  {/* 时间戳 */}
                  <div
                    className="text-sm font-mono whitespace-nowrap select-all"
                    style={{
                      color: token.colorTextSecondary,
                      minWidth: '160px',
                    }}
                  >
                    {timestamp}
                  </div>

                  {/* 日志级别标签 */}
                  {/* <Tag
                    color={getLevelTag(level)}
                    style={{
                      margin: 0,
                      minWidth: '60px',
                      textAlign: 'center',
                      fontSize: '11px',
                      fontWeight: 'bold',
                    }}
                  >
                    {level.toUpperCase()}
                  </Tag> */}

                  {/* 日志内容 */}
                  <div className="flex-1 min-w-0">
                    <div
                      className="text-xs font-mono whitespace-pre-wrap select-all"
                      style={{
                        color: token.colorText,
                        wordBreak: 'break-word',
                        lineHeight: '1.5',
                      }}
                    >
                      {log.message}
                    </div>
                  </div>

                  {/* 复制按钮 */}
                  <Button
                    type="text"
                    size="small"
                    icon={<CopyOutlined />}
                    className="opacity-0 group-hover:opacity-100 transition-opacity"
                    onClick={() => copyLogContent(log)}
                    title="复制此条日志"
                    style={{
                      color: token.colorTextSecondary,
                      minWidth: 'auto',
                      padding: '0 6px',
                    }}
                  />
                </div>
              </div>
            )
          })}

          {filteredLogs.length === 0 && (
            <div className="text-center py-16" style={{ color: token.colorTextSecondary }}>
              {logs.length === 0 ? '暂无日志数据' : '没有匹配的日志'}
            </div>
          )}
        </div>
      </Drawer>
    </div>
  )
}

export default NodeLogs
