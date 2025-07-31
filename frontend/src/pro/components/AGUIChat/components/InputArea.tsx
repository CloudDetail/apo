import React, { useRef } from 'react'
import { Button, Input } from 'antd'
import { SendOutlined, StopOutlined } from '@ant-design/icons'
import { useTranslation } from 'react-i18next'

const { TextArea } = Input

interface InputAreaProps {
  inputText: string
  loading: boolean
  isRunning: boolean
  runId: string | null
  onInputChange: (text: string) => void
  onSendMessage: (props?: { question?: string }) => void
  onCancelRun: () => void
}

const InputArea: React.FC<InputAreaProps> = ({
  inputText,
  loading,
  isRunning,
  runId,
  onInputChange,
  onSendMessage,
  onCancelRun,
}) => {
  const inputRef = useRef<any>(null)
  const { t } = useTranslation('pro/agent')

  // 处理输入框回车
  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      onSendMessage()
    }
  }

  // 处理发送按钮点击
  const handleSendClick = () => {
    onSendMessage()
  }

  return (
    <div className="flex items-center gap-2 px-4 py-2  border-[var(--ant-color-border)]">
      <TextArea
        ref={inputRef}
        className="flex-1"
        placeholder={t('inputArea.placeholder')}
        value={inputText}
        onChange={(e) => onInputChange(e.target.value)}
        onKeyPress={handleKeyPress}
        disabled={loading}
        autoSize={{ minRows: 1, maxRows: 4 }}
      />
      {isRunning && runId ? (
        <Button
          type="text"
          icon={<StopOutlined />}
          className="flex-shrink-0 text-red-500 hover:text-red-600 hover:bg-red-50"
          onClick={onCancelRun}
        />
      ) : (
        <Button
          type="text"
          icon={<SendOutlined />}
          className="flex-shrink-0 text-blue-500 hover:text-blue-600 hover:bg-blue-50 disabled:text-gray-400 disabled:hover:bg-transparent"
          onClick={handleSendClick}
          disabled={!inputText.trim() || loading}
        />
      )}
    </div>
  )
}

export default InputArea
