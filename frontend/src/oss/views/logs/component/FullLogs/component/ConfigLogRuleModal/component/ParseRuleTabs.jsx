/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Form, Tabs } from 'antd'
import { useEffect, useState } from 'react'
import TextArea from 'antd/es/input/TextArea'
import { AiOutlineInfoCircle } from 'react-icons/ai'
import LogStructRuleFormList from './LogStructRuleFormList'
import { useTranslation } from 'react-i18next' // 引入i18n

const ParseRuleTabs = () => {
  const { t, i18n } = useTranslation('oss/fullLogs')
  const form = Form.useFormInstance()
  const [activeKey, setActiveKey] = useState('unStructured')
  const items = [
    {
      key: 'unStructured',
      label: t('configLogRuleModal.parseRuleTabs.unStructuredLabel'),
      children: (
        <div>
          <div className="flex items-center mb-2">
            <AiOutlineInfoCircle size={16} className="mx-1" />
            <span className="text-xs text-gray-400">
              {t('configLogRuleModal.parseRuleTabs.unStructuredDescription')}
              <a
                href=
                {
                  i18n.language === 'zh'
                  ? "https://kindlingx.com/docs/APO%20向导式可观测性中心/配置指南/日志解析规则配置/"
                  : "https://docs.autopilotobservability.com/Logs%20Monitoring/Configure%20Log%20Parsing%20Rules"
                }
                className="underline"
                target="_blank"
              >
                {t('configLogRuleModal.parseRuleTabs.unStructuredHelpLink')}
              </a>
            </span>
          </div>
          <Form.Item
            name="parseRule"
            rules={[
              {
                validator: async (_, value) => {
                  if (form.getFieldValue('isStructured')) {
                    return Promise.resolve()
                  }
                  const unStructuredList = form.getFieldValue('unStructured') || []
                  if (!value && !unStructuredList[0]?.name) {
                    return Promise.reject(t('configLogRuleModal.parseRuleTabs.parseRuleError'))
                  }
                },
              },
            ]}
          >
            <TextArea
              placeholder={t('configLogRuleModal.parseRuleTabs.parseRulePlaceholder')}
              rows={3}
            />
          </Form.Item>
          <LogStructRuleFormList fieldName={'unStructured'} />
        </div>
      ),
    },
    {
      key: 'structured',
      label: t('configLogRuleModal.parseRuleTabs.structuredLabel'),
      children: (
        <>
          <div className="flex mb-2">
            <AiOutlineInfoCircle size={16} className="mx-1" />
            <span className="text-xs text-gray-400">
              {t('configLogRuleModal.parseRuleTabs.structuredDescription')}
            </span>
          </div>
          <Form.Item
            name="structuredRule"
            rules={[
              {
                validator: async (_, value) => {
                  if (form.getFieldError('isStructured') && value && !checkyIsJson(value)) {
                    return Promise.reject(t('configLogRuleModal.parseRuleTabs.structuredRuleError'))
                  }
                },
              },
            ]}
          >
            <TextArea
              placeholder={t('ConfigLogRuleModal.parseRuleTabs.structuredRulePlaceholder')}
              rows={3}
              onChange={(e) => {
                changeStructuredRule(e.target.value)
              }}
            />
          </Form.Item>
          <LogStructRuleFormList fieldName={'structured'} />
        </>
      ),
    },
  ]
  const checkyIsJson = (value) => {
    try {
      // 尝试解析输入字符串为 JSON 对象
      const parsed = JSON.parse(value)
      // 确保解析结果是对象且不是数组
      if (
        typeof parsed === 'object' &&
        parsed !== null &&
        !Array.isArray(parsed) &&
        Object.keys(parsed).length > 0
      ) {
        return true
      }
      return false
    } catch (error) {
      return false
    }
  }
  const changeStructuredRule = (value) => {
    if (checkyIsJson(value)) {
      // 遍历对象的第一层 key 和 value
      const parsed = JSON.parse(value)
      const result = Object.entries(parsed).map(([key, value]) => {
        let type
        if (typeof value === 'string') {
          type = 'String'
        } else if (typeof value === 'number') {
          type = Number.isInteger(value) ? 'Int64' : 'Float64'
        } else if (typeof value === 'boolean') {
          type = 'Bool'
        } else {
          type = 'String'
        }
        return {
          name: key,
          type: {
            key: type,
            label: type,
            value: type,
          },
        }
      })
      form.setFieldValue('structured', result)
    }
  }

  const changeTabs = (key) => {
    form.setFieldValue('isStructured', key === 'structured')
  }
  useEffect(() => {
    setActiveKey(form.getFieldValue('isStructured') ? 'structured' : 'unStructured')
  }, [form.getFieldValue('isStructured')])
  return (
    <Tabs
      defaultActiveKey="unStructured"
      items={items}
      activeKey={activeKey}
      onChange={changeTabs}
      // destroyInactiveTabPane
    />
  )
}
export default ParseRuleTabs
