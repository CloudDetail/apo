/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Button, Card, ConfigProvider, Form, FormListFieldData, Input, Segmented } from 'antd'
import { IoMdClose } from 'react-icons/io'
import { MdOutlineEdit, MdPreview } from 'react-icons/md'
import TagRulePreview from './TagRulePreview'
import ConditionsFormList from './ConditionsFormList'
import TargetTagSelector from './TargetTagSelector'
import SchemaSource from './SchemaSource'
import styles from './segmented.module.scss'
import { useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'

interface TagRuleFormListComProps {
  field: FormListFieldData
  remove: any
  readOnly: boolean
  onEdit: any
}
const TagRuleFormListCom = ({ field, remove, readOnly, onEdit }: TagRuleFormListComProps) => {
  const { t } = useTranslation('core/alertsIntegration')
  const [showPreview, setShowPreview] = useState(false)
  const form = Form.useFormInstance()
  const formListRef = useRef(null)
  const rType = Form.useWatch(['enrichRuleConfigs', field.name, 'rType'])
  const rTypeOptions = [
    {
      value: 'tagMapping',
      label: t('tagMapping'),
    },
    {
      value: 'staticEnrich',
      label: t('staticEnrich'),
    },
  ]
  return (
    <Card
      size="small"
      title={`${t('rule')} ${field.name + 1} - ${rTypeOptions.find((item) => item.value === rType)?.label}`}
      key={field.key}
      ref={formListRef}
      extra={
        <>
          <Button
            size="small"
            color="primary"
            variant="text"
            icon={showPreview || readOnly ? <MdOutlineEdit /> : <MdPreview />}
            onClick={() => {
              if (readOnly) {
                onEdit(formListRef)
                setShowPreview(false)
              } else {
                form
                  .validateFields([['enrichRuleConfigs', field.name]], { recursive: true })
                  .then(() => {
                    setShowPreview(!showPreview)
                  })
              }
            }}
          >
            {showPreview || readOnly ? t('edit') : t('preview')}
          </Button>
          {readOnly || (
            <Button
              size="small"
              danger
              icon={<IoMdClose />}
              className="flex-grow-0 flex-shrink-0"
              onClick={() => remove(field.name)}
            ></Button>
          )}
        </>
      }
      type="inner"
    >
      {showPreview || readOnly ? (
        <TagRulePreview index={field.name} />
      ) : (
        <>
          <ConfigProvider
            theme={{
              components: {
                Segmented: {
                  // itemActiveBg: '#1c2b4a',
                  // itemSelectedBg: '#1c2b4a',
                  // trackBg: '#1e2635',
                  // itemSelectedColor: '#4d82ff',
                  // itemColor: 'rgba(255,255,255, 0.4)',
                },
              },
            }}
          >
            <Form.Item name="enrichRuleId" hidden>
              <Input></Input>
            </Form.Item>
            <Form.Item
              label={t('ruleType')}
              name={[field.name, 'rType']}
              initialValue={'tagMapping'}
            >
              <Segmented options={rTypeOptions} className={styles.segmented} />
            </Form.Item>
          </ConfigProvider>

          <Form.Item label={t('conditions')}>
            <ConditionsFormList fieldName={field.name} />
          </Form.Item>
          <Form.Item
            name={[field.name, 'fromField']}
            label={t('extractedField')}
            labelCol={{ span: 5, offset: 1 }}
            required
            rules={[{ required: true }]}
          >
            <Input placeholder={t('extractedFieldRequired')}></Input>
          </Form.Item>
          <Form.Item
            noStyle
            shouldUpdate={(prevValues, curValues) => {
              return (
                prevValues.enrichRuleConfigs[field.name]?.rType !==
                curValues.enrichRuleConfigs[field.name]?.rType
              )
            }}
            // dependencies={['items', field.name, 'rType']}
          >
            {() =>
              form.getFieldValue(['enrichRuleConfigs', field.name, 'rType']) === 'tagMapping' ? (
                <>
                  <Form.Item
                    name={[field.name, 'fromRegex']}
                    label={t('fromRegex')}
                    labelCol={{ span: 5, offset: 1 }}
                    // required
                    // rules={[{ required: true }]}
                  >
                    <Input placeholder={t('fromRegexRequired')}></Input>
                  </Form.Item>
                  {/* 目标字段 */}
                  <TargetTagSelector fieldName={field.name} />
                </>
              ) : (
                <>
                  {/* 来源数据表和列 */}
                  <SchemaSource fieldName={field.name} />
                </>
              )
            }
          </Form.Item>
        </>
      )}
    </Card>
  )
}
export default TagRuleFormListCom
