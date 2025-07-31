import React from 'react'
import { Typography } from 'antd'
import { RobotOutlined } from '@ant-design/icons'
import styles from '../AGUIChat.module.scss'
import { AiOutlineBarChart } from 'react-icons/ai'
import { IoMdHelp } from 'react-icons/io'
import { IoAlertCircleOutline } from 'react-icons/io5'

const { Paragraph } = Typography

interface QuickQuestionsProps {
  onQuestionClick: (question: string) => void
}

const QuickQuestions: React.FC<QuickQuestionsProps> = ({ onQuestionClick }) => {
  const quickQuestions = [
    // '当前环境中存在慢延时问题吗？开始时间1753232482682953 结束时间1753233382682953',
    // '应用慢延时巡检：请检查最近30分钟内系统中是否存在应用响应变慢或延时异常的情况，并分析其可能原因。',
    // '应用错误巡检：请检查最近30分钟内系统中是否存在应用异常或报错类问题，并分析其可能原因。',
    {
      icon: <AiOutlineBarChart />,
      question: '检查最近15分钟系统中是否存在应用响应变慢或延时异常的情况，并分析原因。',
    },
    {
      icon: <IoAlertCircleOutline />,
      question: '分析最近15分钟系统中是否存在应用异常或错误问题，并分析可能原因。',
    },
    { icon: <IoMdHelp />, question: 'node-16主机的CPU使用率超过80%，这是怎么回事？' },
  ]

  return (
    <div className="h-full  overflow-hidden justify-around flex flex-col">
      <div className={styles.quickQuestionsHeader}>
        <div className=" flex items-center justify-center w-full text-4xl font-semibold m-0 mb-2  gap-2">
          <RobotOutlined className="text-5xl text-[var(--ant-color-primary)]" />
          <span>Agent SRE</span>
        </div>
        <Paragraph className="mb-2 text-base text-[var(--ant-color-text-heading)]">
          🚀 让 AI 极速解决系统问题
        </Paragraph>
      </div>
      <div className="flex flex-col items-center justify-center">
        <div className="items-center text-center text-lg font-bold my-2">🎯 我能做什么？</div>
        <div className="flex flex-col justify-center w-3/4 text-base text-[var(--ant-color-text-heading)]">
          <div className="my-1">
            <span className="mx-2">🔍</span> 找出最近延时升高的服务和原因
          </div>
          <div className="my-1">
            <span className="mx-2">📉</span> 分析错误率突增的根因
          </div>
          <div className="my-1">
            <span className="mx-2">🔄</span> 分析系统资源告警的深层原因
          </div>
          <div className="my-1">
            <span className="mx-2">🧠</span> 对话式查看日志、指标、调用链
          </div>
        </div>
      </div>
      <div className="flex flex-col items-center justify-center ">
        <div className="items-center text-center text-lg font-bold my-2">✨ 快速提问</div>
        <div className="w-3/4">
          {quickQuestions.map((question, index) => (
            <div
              key={index}
              className="w-full bg-[var(--ant-color-fill-secondary)] group my-2 pl-2 pr-4 py-2 rounded-md flex items-center justify-between hover:bg-[var(--ant-color-fill)] cursor-pointer"
              onClick={() => onQuestionClick(question.question)}
            >
              <div className="flex text-sm font-medium group-hover:text-[var(--ant-color-info-text-hover)] transition-colors duration-200">
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
