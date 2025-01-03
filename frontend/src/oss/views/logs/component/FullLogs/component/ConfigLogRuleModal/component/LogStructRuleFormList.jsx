/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Col, Form, Input, Row, Select } from 'antd'
import { IoIosRemoveCircleOutline, IoMdAddCircleOutline } from 'react-icons/io'
const routeKeyList = [
  { value: 'Int8', label: 'Int8' },
  { value: 'Int16', label: 'Int16' },
  { value: 'Int32', label: 'Int32' },
  { value: 'Int64', label: 'Int64' },
  { value: 'Int128', label: 'Int128' },
  { value: 'Int256', label: 'Int256' },
  { value: 'UInt8', label: 'UInt8' },
  { value: 'UInt16', label: 'UInt16' },
  { value: 'UInt32', label: 'UInt32' },
  { value: 'UInt64', label: 'UInt64' },
  { value: 'UInt128', label: 'UInt128' },
  { value: 'UInt256', label: 'UInt256' },
  { value: 'Float32', label: 'Float32' },
  { value: 'Float64', label: 'Float64' },
  { value: 'Date', label: 'Date' },
  { value: 'Date32', label: 'Date32' },
  { value: 'DateTime', label: 'DateTime' },
  { value: 'DateTime64', label: 'DateTime64' },
  { value: 'String', label: 'String' },
  { value: 'FixedString(N)', label: 'FixedString(N)' },
  { value: 'Bool', label: 'Bool' },
]

export default function LogStructRuleFormList({ fieldName }) {
  const form = Form.useFormInstance()
  return (
    <Form.List name={fieldName}>
      {(fields, { add, remove }, { errors }) => (
        <>
          <Form.Item
            required={form.getFieldValue('isStructured')}
            label={
              <>
                {/* <div className="flex flex-row"> */}
                日志字段数据类型{' '}
                <IoMdAddCircleOutline
                  onClick={() =>
                    add({
                      name: '',
                      type: {
                        key: 'String',
                        label: 'String',
                        value: 'String',
                      },
                    })
                  }
                  size={20}
                  className="mx-2 cursor-pointer"
                />
              </>
            }
          >
            {fields.map((field, index) => (
              <div key={field.name} className=" px-3 pt-3 pb-0 rounded relative">
                <Row gutter={12}>
                  <Col span={11}>
                    <Form.Item
                      name={[field.name, 'name']}
                      required
                      rules={[
                        {
                          validator: async (_, value) => {
                            // 获取当前表单中所有的routeRule项
                            const tableFields = form.getFieldValue(fieldName) || []
                            // // 检查是否有重复的key
                            const isStructured = form.getFieldValue('isStructured')
                            if (isStructured) {
                              if (!value) {
                                return Promise.reject('字段名不可为空')
                              }
                            } else if (!form.getFieldValue('parseRule') && !value) {
                              return Promise.reject('字段名不可为空')
                            }
                            const duplicate = tableFields.filter(
                              (item, i) => item?.name === value && i !== index,
                            )
                            if (duplicate.length) {
                              return Promise.reject('已存在相同的字段名')
                            }
                          },
                        },
                      ]}
                    >
                      <Input placeholder="字段名" />
                    </Form.Item>
                  </Col>
                  <Col span={11}>
                    <Form.Item
                      // noStyle
                      name={[field.name, 'type']}
                      required
                      rules={[
                        {
                          validator: async (_, value) => {
                            // 获取当前表单中所有的routeRule项
                            // // 检查是否有重复的key
                            const isStructured = form.getFieldValue('isStructured')
                            if (isStructured) {
                              if (!value) {
                                return Promise.reject('字段类型不可为空')
                              }
                            } else if (!form.getFieldValue('parseRule') && !value) {
                              return Promise.reject('字段类型不可为空')
                            }
                          },
                        },
                      ]}
                    >
                      <Select
                        options={routeKeyList}
                        labelInValue
                        placeholder="选择匹配规则Key"
                        defaultValue={{
                          key: 'String',
                          label: 'String',
                          value: 'String',
                        }}
                      />
                    </Form.Item>
                  </Col>
                  <Col span={1}>
                    <IoIosRemoveCircleOutline
                      size={20}
                      className="mt-1 cursor-pointer"
                      onClick={() => {
                        remove(field.name)
                      }}
                    />
                  </Col>
                </Row>
              </div>
            ))}
          </Form.Item>
        </>
      )}
    </Form.List>
  )
}
