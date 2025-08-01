import React from 'react'
import { Typography } from 'antd'
import { RobotOutlined } from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import styles from '../AGUIChat.module.scss'
import { AiOutlineBarChart } from 'react-icons/ai'
import { IoMdHelp } from 'react-icons/io'
import { IoAlertCircleOutline } from 'react-icons/io5'
import ShinyText from '../../Animation/ShinyText'
import AnimatedContent from '../../Animation/AnimatedContent'

const { Paragraph } = Typography

interface QuickQuestionsProps {
  onQuestionClick: (question: string) => void
}
const CustomCard = ({ children }: { children: React.ReactNode }) => {
  return (
    <div className="relative drop-shadow-xl w-[500px] h-[300px] overflow-hidden rounded-2xl bg-transparent">
      <div
        className="absolute flex items-center justify-center text-white z-[1]  rounded-2xl inset-0.5 
      bg-[#131313]  border  border-[var(--ant-color-border)]"
      >
        {children}
      </div>
      <div className="absolute w-56 h-48 bg-white blur-[50px] -left-1/3 -top-1/2"></div>
    </div>
  )
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
  const capabilities = [
    {
      icon: '🔍',
      title: t('quickQuestions.capabilities.latencyAnalysis'),
      description: t('quickQuestions.capabilities.latencyAnalysisDescription'),
    },
    {
      icon: '📉',
      title: t('quickQuestions.capabilities.errorAnalysis'),
    },
    {
      icon: '🔄',
      title: t('quickQuestions.capabilities.resourceAnalysis'),
      description: t('quickQuestions.capabilities.resourceAnalysisDescription'),
    },
    {
      icon: '🧠',
      title: t('quickQuestions.capabilities.conversationalAnalysis'),
      description: t('quickQuestions.capabilities.conversationalAnalysisDescription'),
    },
  ]
  return (
    <div className="h-full  overflow-hidden justify-around flex flex-col relative">
      <div className={styles.quickQuestionsHeader} style={{ zIndex: 1, marginTop: '20px' }}>
        <div className=" flex items-center justify-center w-full text-4xl font-semibold  my-4  gap-2">
          <RobotOutlined className="text-5xl text-[var(--ant-color-primary)]" />
          <ShinyText text={t('quickQuestions.title')} speed={3} />
        </div>
        <Paragraph className="mb-2 text-base text-[var(--ant-color-text-heading)]">
          {t('quickQuestions.subtitle')}
        </Paragraph>
      </div>
      <div className="flex flex-col items-center justify-center z-10">
        <div className="items-center text-center text-lg font-bold my-2">
          {t('quickQuestions.capabilitiesTitle')}
        </div>
        <div className="flex flex-col justify-center w-3/4 text-base text-[var(--ant-color-text-heading)]">
          <AnimatedContent>
            {capabilities.map((capability, index) => (
              <div key={index} className="my-1 flex items-center">
                <span className="mx-2">{capability.icon}</span>
                <span>{capability.title}</span>
              </div>
            ))}
          </AnimatedContent>
        </div>
      </div>

      {/* <div className="flex items-center justify-center w-full gap-11 z-20">
        <CustomCard>
          <div className="flex flex-col items-center w-full h-full">
            <div className="items-center text-center text-lg font-bold mt-4 my-0">
              <SplitText text={t('quickQuestions.capabilitiesTitle')} splitType="lines" />
            </div>
            <div className="flex flex-col w-full justify-around flex-1">
              {capabilities.map((capability, index) => (
                <div
                  className="flex items-center m-2 p-2  rounded-xl bg-[#1b1b1b] text-[var(--ant-color-text)]"
                  key={index}
                >
                  <span className="mx-2">{capability.icon}</span>
                  <span>{capability.title}</span>
                </div>
              ))}
            </div>
          </div>
        </CustomCard>
        <CustomCard>
          <div className="flex flex-col items-center w-full h-full">
            <div className="items-center text-center text-lg font-bold mt-4 my-0">
              <SplitText text={t('quickQuestions.quickQuestionsTitle')} splitType="lines" />
            </div>
            <div className="flex flex-col justify-around flex-1">
              {quickQuestions.map((capability, index) => (
                <div
                  className="z-40 flex items-center m-2 p-2 group justify-between  rounded-2xl bg-[#1b1b1b] cursor-pointer text-[var(--ant-color-text)]"
                  key={index}
                  onClick={() => onQuestionClick(capability.question)}
                >
                  <div className="flex items-center">
                    <span className="mx-2">{capability.icon}</span>
                    <span>{capability.question}</span>
                  </div>
                  <span className="text-lg mr-2 font-bold opacity-60 text-[var(--ant-color-primary)] group-hover:translate-x-1 transition-transform duration-200 cursor-pointer">
                    →
                  </span>
                </div>
              ))}
            </div>
          </div>
        </CustomCard>
      </div> */}

      <div className="flex flex-col items-center justify-center z-10  ">
        <div className="items-center text-center text-lg font-bold my-2">
          {t('quickQuestions.quickQuestionsTitle')}
        </div>
        <div className="w-3/4">
          <AnimatedContent>
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
          </AnimatedContent>
        </div>
      </div>
      {/* <div
        style={{ width: '100%', height: '100%', position: 'absolute', top: 0, left: 0, zIndex: 0 }}
      >
        <LightRays
          raysOrigin="top-center"
          raysColor="#ffffff"
          raysSpeed={1.5}
          lightSpread={0.8}
          rayLength={1.2}
          followMouse={true}
          mouseInfluence={0.1}
          noiseAmount={0.1}
          distortion={0.05}
          className="absolute top-0 left-0 w-full h-full"
        />
      </div> */}
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

export default React.memo(QuickQuestions)
