import { Button, message as antdMessage } from 'antd'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { FaRegCopy, FaThumbsDown, FaThumbsUp } from 'react-icons/fa'
import { FiThumbsDown, FiThumbsUp } from 'react-icons/fi'
import { PiShareFatBold } from 'react-icons/pi'
import {
  buildFinalMessageShareUrl,
  getFinalMessageCopySuccessKey,
  getFinalMessageCopyText,
  isFinalMessageFeedbackSelected,
  type FinalMessageCopyTarget,
  type FinalMessageFeedbackTarget,
  type FinalMessageSuffixMessageLike,
} from 'src/pro/api/conversationHistoryView'

const copyTextToClipboard = async (value: string, successText: string, errorText: string) => {
  try {
    if (navigator.clipboard && window.isSecureContext) {
      await navigator.clipboard.writeText(value)
      antdMessage.success(successText)
      return
    }

    const textArea = document.createElement('textarea')
    textArea.value = value
    textArea.style.position = 'absolute'
    textArea.style.left = '-999999px'
    document.body.prepend(textArea)
    textArea.select()

    try {
      document.execCommand('copy')
      antdMessage.success(successText)
    } catch (error) {
      antdMessage.error(errorText)
      console.error(error)
    } finally {
      textArea.remove()
    }
  } catch (error) {
    antdMessage.error(errorText)
    console.error(error)
  }
}

const FinalMessageSuffix = ({
  message,
  activeThreadId,
}: {
  message: FinalMessageSuffixMessageLike
  activeThreadId?: string
}) => {
  const { t } = useTranslation('pro/agent')
  const copyText = getFinalMessageCopyText(message)
  const [selectedFeedback, setSelectedFeedback] = useState<FinalMessageFeedbackTarget | undefined>()

  const copyWithFeedback = (value: string, target: FinalMessageCopyTarget) =>
    copyTextToClipboard(value, t(getFinalMessageCopySuccessKey(target)), t('messageActions.copyFailed'))

  const positiveSelected = isFinalMessageFeedbackSelected(selectedFeedback, 'positive')
  const negativeSelected = isFinalMessageFeedbackSelected(selectedFeedback, 'negative')

  const handleCopy = () => {
    if (!copyText) return
    void copyWithFeedback(copyText, 'response')
  }

  const handleShare = () => {
    if (typeof window === 'undefined') return
    void copyWithFeedback(
      buildFinalMessageShareUrl({
        href: window.location.href,
        threadId: activeThreadId,
      }),
      'share',
    )
  }

  return (
    <div className="flex items-center justify-start ml-5 mt-3 text-[var(--ant-color-text-secondary)]">
      <Button
        type="text"
        icon={<FaRegCopy />}
        className="text-[var(--ant-color-text-secondary)]"
        aria-label="Copy response"
        disabled={!copyText}
        onClick={handleCopy}
      />
      <Button
        type="text"
        icon={<PiShareFatBold />}
        className="text-[var(--ant-color-text-secondary)]"
        aria-label="Copy conversation link"
        onClick={handleShare}
      />
      <Button
        type="text"
        icon={positiveSelected ? <FaThumbsUp /> : <FiThumbsUp />}
        className={
          positiveSelected
            ? 'text-[var(--ant-color-primary)]'
            : 'text-[var(--ant-color-text-secondary)]'
        }
        aria-label="Positive feedback"
        aria-pressed={positiveSelected}
        onClick={() => setSelectedFeedback('positive')}
      />
      <Button
        type="text"
        icon={negativeSelected ? <FaThumbsDown /> : <FiThumbsDown />}
        className={
          negativeSelected
            ? 'text-[var(--ant-color-primary)]'
            : 'text-[var(--ant-color-text-secondary)]'
        }
        aria-label="Negative feedback"
        aria-pressed={negativeSelected}
        onClick={() => setSelectedFeedback('negative')}
      />
    </div>
  )
}

export default FinalMessageSuffix
