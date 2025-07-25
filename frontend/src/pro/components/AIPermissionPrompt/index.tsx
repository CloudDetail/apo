import React from 'react'
import { Card, Button, Alert, Spin, Typography } from 'antd'
import { RobotOutlined, SettingOutlined } from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import './AIPermissionPrompt.css'
import { IoWarningOutline } from 'react-icons/io5'
import { useNavigate } from 'react-router-dom'

const { Title, Paragraph, Text } = Typography

interface AIPermissionPromptProps {
  isLoading?: boolean
  error?: string | null
  onRefresh?: () => void
}

const AIPermissionPrompt: React.FC<AIPermissionPromptProps> = ({
  isLoading = false,
  error = null,
  onRefresh,
}) => {
  const { t } = useTranslation('pro/chatAI')
  const navigate = useNavigate()
  const onConfigure = () => {
    navigate('/probe-installer')
  }

  if (isLoading) {
    return (
      <div className="ai-permission-prompt">
        <div className="ai-permission-loading">
          <Spin size="large" />
          <Text className="loading-text">{t('checkingPermission', '正在检查权限...')}</Text>
        </div>
      </div>
    )
  }

  // if (error) {
  //   return (
  //     <div className="ai-permission-prompt">
  //       <Card className="ai-permission-card">
  //         <Alert
  //           message={t('permissionCheckError', '权限检查失败')}
  //           description={error}
  //           type="error"
  //           showIcon
  //           action={
  //             <Button size="small" onClick={onRefresh}>
  //               {t('retry', '重试')}
  //             </Button>
  //           }
  //         />
  //       </Card>
  //     </div>
  //   )
  // }

  return (
    <div className="ai-permission-prompt">
      <Card className="ai-permission-card">
        <div className="ai-permission-content">
          <div className="ai-permission-header">
            <RobotOutlined className="ai-icon" />
            <Title level={2} className="ai-title">
              {t('aiAssistant', 'AI 智能助手')}
            </Title>
          </div>

          <div className="ai-permission-description">
            <Paragraph className="description-text">
              {t(
                'aiDescription',
                '智能运维助手，为您提供专业的SRE支持，包括系统性能分析、告警根因分析、故障排查等功能。',
              )}
            </Paragraph>
          </div>

          <div className="ai-permission-alert">
            <Alert
              message={t('未安装minimal-agent')}
              description={
                <div className="alert-description">
                  minimal-agent为智能体运行的基础，轻松两步安装
                </div>
              }
              type="warning"
              className="items-center"
              showIcon
              icon={<IoWarningOutline size={30} />}
              action={
                <Button
                  type="primary"
                  size="large"
                  icon={<SettingOutlined />}
                  onClick={onConfigure}
                >
                  {t('installNow', '立即配置')}
                </Button>
              }
            />
          </div>

          {/* <div className="ai-permission-actions">
            <Space size="middle">
              <Button type="primary" size="large" icon={<SettingOutlined />} onClick={onConfigure}>
                {t('configureNow', '立即配置')}
              </Button>
              <Button
                size="large"
                icon={<BookOutlined />}
                onClick={() => {
                  // 打开配置手册链接
                  window.open('/docs/ai-configuration', '_blank')
                }}
              >
                {t('viewGuide', '查看手册')}
              </Button>
            </Space>
          </div> */}

          {/* <div className="ai-permission-features">
            <Title level={4}>{t('aiFeatures', 'AI 助手功能')}</Title>
            <div className="features-grid">
              <div className="feature-item">
                <div className="feature-icon">🔍</div>
                <div className="feature-content">
                  <Text strong>{t('performanceAnalysis', '性能分析')}</Text>
                  <Text type="secondary">
                    {t('performanceAnalysisDesc', '智能分析系统性能瓶颈')}
                  </Text>
                </div>
              </div>
              <div className="feature-item">
                <div className="feature-icon">🚨</div>
                <div className="feature-content">
                  <Text strong>{t('alertAnalysis', '告警分析')}</Text>
                  <Text type="secondary">{t('alertAnalysisDesc', '快速定位告警根因')}</Text>
                </div>
              </div>
              <div className="feature-item">
                <div className="feature-icon">🔧</div>
                <div className="feature-content">
                  <Text strong>{t('troubleshooting', '故障排查')}</Text>
                  <Text type="secondary">{t('troubleshootingDesc', '自动化故障诊断流程')}</Text>
                </div>
              </div>
              <div className="feature-item">
                <div className="feature-icon">📊</div>
                <div className="feature-content">
                  <Text strong>{t('capacityPlanning', '容量规划')}</Text>
                  <Text type="secondary">{t('capacityPlanningDesc', '智能容量规划建议')}</Text>
                </div>
              </div>
            </div>
          </div> */}
        </div>
      </Card>
    </div>
  )
}

export default AIPermissionPrompt
