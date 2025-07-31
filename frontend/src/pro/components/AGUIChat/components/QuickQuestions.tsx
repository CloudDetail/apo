import React from 'react'
import { Typography } from 'antd'
import { RobotOutlined } from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import styles from '../AGUIChat.module.scss'
import { AiOutlineBarChart } from 'react-icons/ai'
import { IoMdHelp } from 'react-icons/io'
import { IoAlertCircleOutline } from 'react-icons/io5'

const { Paragraph } = Typography

interface QuickQuestionsProps {
  onQuestionClick: (question: string) => void
}

const QuickQuestions: React.FC<QuickQuestionsProps> = ({ onQuestionClick }) => {
  const { t } = useTranslation('pro/agent')

  const quickQuestions = [
    {
      icon: <AiOutlineBarChart />,
      question: t('quickQuestions.questions.latencyCheck'),
    },
    {
      icon: <IoAlertCircleOutline />,
      question: t('quickQuestions.questions.errorCheck'),
    },
    {
      icon: <IoMdHelp />,
      question: t('quickQuestions.questions.cpuUsage'),
    },
  ]

  return (
    <div className="h-full  overflow-hidden justify-around flex flex-col">
      <div className={styles.quickQuestionsHeader}>
        <div className=" flex items-center justify-center w-full text-4xl font-semibold m-0 mb-2  gap-2">
          <RobotOutlined className="text-5xl text-[var(--ant-color-primary)]" />
          <span>{t('quickQuestions.title')}</span>
        </div>
        <Paragraph className="mb-2 text-base text-[var(--ant-color-text-heading)]">
          {t('quickQuestions.subtitle')}
        </Paragraph>
      </div>
      <div className="flex flex-col items-center justify-center">
        <div className="items-center text-center text-lg font-bold my-2">
          {t('quickQuestions.capabilitiesTitle')}
        </div>
        <div className="flex flex-col justify-center w-3/4 text-base text-[var(--ant-color-text-heading)]">
          <div className="my-1">
            <span className="mx-2">🔍</span> {t('quickQuestions.capabilities.latencyAnalysis')}
          </div>
          <div className="my-1">
            <span className="mx-2">📉</span> {t('quickQuestions.capabilities.errorAnalysis')}
          </div>
          <div className="my-1">
            <span className="mx-2">🔄</span> {t('quickQuestions.capabilities.resourceAnalysis')}
          </div>
          <div className="my-1">
            <span className="mx-2">🧠</span>{' '}
            {t('quickQuestions.capabilities.conversationalAnalysis')}
          </div>
        </div>
      </div>
      <div className="flex flex-col items-center justify-center ">
        <div className="items-center text-center text-lg font-bold my-2">
          {t('quickQuestions.quickQuestionsTitle')}
        </div>
        <div className="w-3/4">
          {quickQuestions.map((question, index) => (
            <div
              key={index}
              className="w-full bg-[var(--ant-color-fill-secondary)] group my-2 pl-2 pr-4 py-2 rounded-md flex items-center justify-between hover:bg-[var(--ant-color-fill)] cursor-pointer"
              onClick={() => onQuestionClick(question.question)}
            >
              <div className="flex items-center text-sm font-medium group-hover:text-[var(--ant-color-info-text-hover)] transition-colors duration-200">
                <div className="text-lg font-bold opacity-60 text-blue-400 mr-2 ">
                  {question.icon}
                </div>
                {question.question}
              </div>
              <span className="text-lg font-bold opacity-60 text-[var(--ant-color-primary)] group-hover:translate-x-1 transition-transform duration-200">
                →
              </span>
            </div>
          ))}
        </div>
      </div>
      {/* <div className={styles.quickQuestionsGrid}>
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
      </div> */}

      {/* <div className={styles.quickQuestionsFooter}>
        <Paragraph className={styles.footerText}>或者直接在下方输入框中提出您的问题</Paragraph>
      </div> */}
    </div>
  )
}

export default QuickQuestions
