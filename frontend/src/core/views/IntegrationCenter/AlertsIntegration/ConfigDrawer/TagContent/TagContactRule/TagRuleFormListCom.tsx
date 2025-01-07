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

const rTypeOptions = [
  {
    value: 'tagMapping',
    label: '提取标签',
  },
  {
    value: 'staticEnrich',
    label: '映射标签',
  },
]
interface TagRuleFormListComProps {
  field: FormListFieldData
  remove: any
  readOnly: boolean
  onEdit: any
}
const TagRuleFormListCom = ({ field, remove, readOnly, onEdit }: TagRuleFormListComProps) => {
  const [showPreview, setShowPreview] = useState(false)
  const form = Form.useFormInstance()
  const formListRef = useRef(null)
  const rType = Form.useWatch(['enrichRuleConfigs', field.name, 'rType'])

  return (
    <Card
      size="small"
      title={`规则 ${field.name + 1} - ${rTypeOptions.find((item) => item.value === rType)?.label}`}
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
            {showPreview || readOnly ? '编辑' : '标签规则预览'}
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
                  itemActiveBg: '#1c2b4a',
                  itemSelectedBg: '#1c2b4a',
                  trackBg: '#1e2635',
                  itemSelectedColor: '#4d82ff',
                  itemColor: 'rgba(255,255,255, 0.4)',
                },
              },
            }}
          >
            <Form.Item name="enrichRuleId" hidden>
              <Input></Input>
            </Form.Item>
            <Form.Item label="规则类型" name={[field.name, 'rType']} initialValue={'tagMapping'}>
              <Segmented options={rTypeOptions} className={styles.segmented} />
            </Form.Item>
          </ConfigProvider>

          <Form.Item label="过滤条件">
            <ConditionsFormList fieldName={field.name} />
          </Form.Item>
          <Form.Item
            name={[field.name, 'fromField']}
            label="提取字段"
            labelCol={{ span: 3, offset: 1 }}
            required
            rules={[{ required: true }]}
          >
            <Input placeholder="请输入希望提取的字段"></Input>
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
                    label="提取正则表达式"
                    labelCol={{ span: 3, offset: 1 }}
                    // required
                    // rules={[{ required: true }]}
                  >
                    <Input placeholder="请输入提取正则表达式"></Input>
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
