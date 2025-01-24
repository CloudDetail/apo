/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Cascader, Form } from 'antd'
import SchemaFormList from './SchemaFormList'
import { useAlertIntegrationContext } from 'src/core/contexts/AlertIntegrationContext'
interface SchemaSourceProps {
  fieldName: string | number
}
const SchemaSource = ({ fieldName }: SchemaSourceProps) => {
  const schemas = useAlertIntegrationContext((ctx) => ctx.schemas)
  const getOptions = () => {
    return Object.entries(schemas).map(([key, value]) => ({
      label: key,
      value: key,
      children: ((value as any[]) || []).map((col) => ({
        label: col,
        value: col,
      })),
    }))
  }
  return (
    <>
      <Form.Item
        label="映射表和映射列"
        extra="提取字段映射至映射表中某个列"
        style={{ marginBottom: 0 }}
        name={[fieldName, 'schemaObject']}
        required
        rules={[{ required: true }]}
      >
        <Cascader style={{ width: '100%' }} options={getOptions()} />
      </Form.Item>
      <SchemaFormList fieldName={fieldName} />
    </>
  )
}
export default SchemaSource
