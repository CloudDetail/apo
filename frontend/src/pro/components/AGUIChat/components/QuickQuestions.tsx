import React from 'react'
import { Typography } from 'antd'
import { RobotOutlined } from '@ant-design/icons'
import styles from '../AGUIChat.module.scss'

const { Title, Paragraph } = Typography

interface QuickQuestionsProps {
  onQuestionClick: (question: string) => void
}

const QuickQuestions: React.FC<QuickQuestionsProps> = ({ onQuestionClick }) => {
  const quickQuestions = [
    '当前环境中存在慢延时问题吗？开始时间1753232482682953 结束时间1753233382682953',
    '应用慢延时巡检：请检查最近30分钟内系统中是否存在应用响应变慢或延时异常的情况，并分析其可能原因。',
    '应用错误巡检：请检查最近30分钟内系统中是否存在应用异常或报错类问题，并分析其可能原因。',
  ]

  return (
    <div className={styles.quickQuestionsContainer}>
      <div className={styles.quickQuestionsHeader}>
        <div className=" flex items-center justify-center w-full text-4xl font-semibold m-0 mb-4  gap-2">
          <RobotOutlined className="text-5xl text-[var(--ant-color-primary)]" />
          <span>Agent SRE</span>
        </div>
        <Paragraph className={styles.quickQuestionsSubtitle}>
          选择您关心的问题，我将为您提供专业的SRE支持
        </Paragraph>
      </div>

      <div className={styles.quickQuestionsGrid}>
        {quickQuestions.map((question, index) => (
          <div
            key={index}
            className={styles.questionItem}
            onClick={() => onQuestionClick(question)}
          >
            <span className={styles.questionText}>{question}</span>
            <span className={styles.questionArrow}>→</span>
          </div>
        ))}
      </div>

      <div className={styles.quickQuestionsFooter}>
        <Paragraph className={styles.footerText}>或者直接在下方输入框中提出您的问题</Paragraph>
      </div>
    </div>
  )
}

export default QuickQuestions
