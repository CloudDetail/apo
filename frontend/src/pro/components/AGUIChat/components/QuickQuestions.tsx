import React from 'react'
import { Alert, Button, Typography } from 'antd'
import { useTranslation } from 'react-i18next'
import styles from '../AGUIChat.module.scss'
import { AiOutlineBarChart } from 'react-icons/ai'
import { IoMdHelp } from 'react-icons/io'
import { IoAlertCircleOutline, IoWarningOutline } from 'react-icons/io5'
import ShinyText from '../../Animation/ShinyText'
import AnimatedContent from '../../Animation/AnimatedContent'
import { ReactSVG } from 'react-svg'
import agentIcon from 'src/pro/assets/agent.svg'
import { useAGUIPermission } from 'src/pro/contexts/AGUIContext'
import { SettingOutlined } from '@ant-design/icons'
import AgentModel from './AgentModel'
import { useDeployMode } from 'src/core/contexts/DeployModeContext'

const { Paragraph } = Typography

interface QuickQuestionsProps {
  onQuestionClick: (question: string) => void
}

const QuickQuestions: React.FC<QuickQuestionsProps> = ({ onQuestionClick }) => {
  const { hasPermission, permissionLoading, dataPlaneExist, defaultModel } = useAGUIPermission()
  const { t } = useTranslation('pro/agent')
  const { deployMode } = useDeployMode()
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
    },
    {
      icon: '📉',
      title: t('quickQuestions.capabilities.errorAnalysis'),
    },
    {
      icon: '🔄',
      title: t('quickQuestions.capabilities.resourceAnalysis'),
    },
    {
      icon: '🧠',
      title: t('quickQuestions.capabilities.conversationalAnalysis'),
    },
  ]
  return (
    <div className="h-full  overflow-y-scroll justify-around flex flex-col relative">
      <div className={styles.quickQuestionsHeader} style={{ zIndex: 1, marginTop: '20px' }}>
        <div className=" flex items-center justify-center w-full text-4xl font-semibold  my-4  gap-2">
          {/* <RobotOutlined className="text-5xl text-[var(--ant-color-primary)]" /> */}
          <ReactSVG
            src={agentIcon}
            beforeInjection={(svg) => {
              svg.setAttribute('width', '64')
              svg.setAttribute('height', '64')
            }}
          />
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
      {!permissionLoading && hasPermission && dataPlaneExist && defaultModel && (
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
      )}
      {!permissionLoading && !hasPermission && (
        <>
          {' '}
          <div className="ai-permission-alert flex items-center justify-center">
            <Alert
              message={t('quickQuestions.alerts.agentNotInstalledTitle')}
              description={
                <div className="alert-description">
                  {t('quickQuestions.alerts.agentNotInstalledDesc')}
                </div>
              }
              type="error"
              className="items-center w-3/4"
              showIcon
              icon={<IoWarningOutline size={30} />}
              action={
                <Button
                  type="primary"
                  size="large"
                  icon={<SettingOutlined />}
                  onClick={() => {
                    const url =
                      deployMode === 'syncause'
                        ? '/#/probe-management/settings'
                        : '/#/integration/data/settings'
                    window.open(url, '_blank')
                  }}
                >
                  {t('installNow')}
                </Button>
              }
            />
          </div>
        </>
      )}

      {!permissionLoading && !dataPlaneExist && (
        <div className="ai-permission-alert  flex items-center justify-center ">
          <Alert
            message={t('quickQuestions.alerts.dataSourceNotConnectedTitle')}
            description={
              <div className="alert-description">
                {t('quickQuestions.alerts.dataSourceNotConnectedDesc')}
              </div>
            }
            type="warning"
            className="items-center w-3/4"
            showIcon
            icon={<IoWarningOutline size={30} />}
            action={
              <Button
                type="primary"
                size="large"
                icon={<SettingOutlined />}
                onClick={() => {
                  window.open('/#/dataplane?showSettings=true', '_blank')
                }}
              >
                {t('connectNow')}
              </Button>
            }
          />
        </div>
      )}
      {!permissionLoading && !defaultModel && (
        <div className="ai-permission-alert  flex items-center justify-center ">
          <Alert
            message={t('quickQuestions.alerts.defaultModelNotConfiguredTitle')}
            type="warning"
            className="items-center w-3/4"
            showIcon
            icon={<IoWarningOutline size={30} />}
            action={
              <Button>
                <AgentModel />
              </Button>
            }
          />
        </div>
      )}
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
