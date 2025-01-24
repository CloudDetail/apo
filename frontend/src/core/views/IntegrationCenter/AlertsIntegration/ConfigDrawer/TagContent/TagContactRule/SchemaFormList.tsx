/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Flex, Form, Input, Select } from 'antd'
import { useEffect } from 'react'
import TargetTagSelector from './TargetTagSelector'
import { useAlertIntegrationContext } from 'src/core/contexts/AlertIntegrationContext'

interface SchemaFormListProps {
  fieldName: string | number
}
const SchemaFormList = ({ fieldName }: SchemaFormListProps) => {
  const form = Form.useFormInstance()
  const schemaObject = Form.useWatch(['enrichRuleConfigs', fieldName, 'schemaObject'], form)
  const enrichRuleConfigs = Form.useWatch(['enrichRuleConfigs', fieldName], form)
  const schemas = useAlertIntegrationContext((ctx) => ctx.schemas)
  const getOptions = () => {
    if (!schemas || schemaObject?.length !== 2) return []

    // 找到匹配的 schema 对象
    const schemaCols = schemas[schemaObject[0]]
    console.log(schemaCols)
    if (!schemaCols) return []

    // 返回子节点并根据条件设置 disabled
    return schemaCols.map((child) => ({
      label: child,
      value: child,
      disabled: child === schemaObject[1] ? true : false,
    }))
  }
  useEffect(() => {
    if (enrichRuleConfigs) {
      const schemaFields = enrichRuleConfigs?.schemaFields || []
      const schemaTargets = enrichRuleConfigs?.schemaTargets || []

      let updatedSchemaTargets =
        schemaTargets?.filter((item) => schemaFields.includes(item.schemaField)) || []
      const keysInSchema = new Set(schemaTargets?.map((item) => item.schemaField))
      const missingKeys = schemaFields?.filter((key) => !keysInSchema.has(key))

      missingKeys?.forEach((key) => updatedSchemaTargets.push({ schemaField: key, targetTag: {} }))

      if (schemaObject?.length === 2) {
        const keyToRemove = schemaObject[1]
        updatedSchemaTargets = updatedSchemaTargets.filter(
          (item) => item.schemaField !== keyToRemove,
        )
        form.setFieldValue(
          ['enrichRuleConfigs', fieldName, 'schemaFields'],
          schemaFields.filter((item) => item !== keyToRemove),
        )
      }

      form.setFieldValue(['enrichRuleConfigs', fieldName, 'schemaTargets'], updatedSchemaTargets)
    }
  }, [enrichRuleConfigs])

  return (
    <>
      <Form.Item
        name={[fieldName, 'schemaFields']}
        label="源字段与目标字段"
        rules={[{ required: true, message: '请选择或输入目标字段' }]}
      >
        <Select mode="multiple" allowClear options={getOptions()}></Select>
      </Form.Item>
      <Form.Item label=" ">
        <Form.List name={[fieldName, 'schemaTargets']}>
          {(subFields, subOpt) => (
            <div style={{ display: 'flex', flexDirection: 'column', rowGap: 16 }}>
              {subFields.map((subField, index) => (
                <Form.Item
                  key={subField.key}
                  rules={[{ required: true }]}
                  style={{ marginBottom: 0 }}
                >
                  <Flex justify="center" align="flex-start" gap={5}>
                    <Form.Item
                      name={[subField.name, 'schemaField']}
                      style={{ marginBottom: 0 }}
                      className="w-[200px] flex-shrink-0"
                    >
                      <Input className="w-[200px]" placeholder="源字段" readOnly />
                    </Form.Item>
                    <div className="flex-shrink-0 mt-1">提取后字段</div>
                    <TargetTagSelector fieldName={subField.name} noLabel />
                  </Flex>
                </Form.Item>
              ))}
            </div>
          )}
        </Form.List>
      </Form.Item>
    </>
  )
}
export default SchemaFormList
