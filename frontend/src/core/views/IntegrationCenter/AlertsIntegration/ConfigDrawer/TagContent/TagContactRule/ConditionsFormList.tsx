/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Button, Flex, Form, Input, Select, Tag } from 'antd'
import { AiOutlineLine } from 'react-icons/ai'
interface ConditionsFormListProps {
  fieldName: string | number
}
const operationOptions = [
  {
    value: 'match',
    label: <span>匹配</span>,
  },
  {
    value: 'notMatch',
    label: <span>不匹配</span>,
  },
]

const ConditionsFormList = ({ fieldName }: ConditionsFormListProps) => {
  return (
    <Form.List name={[fieldName, 'conditions']}>
      {(subFields, subOpt) => (
        <div style={{ display: 'flex', flexDirection: 'column', rowGap: 16 }}>
          {subFields.map((subField, index) => (
            <Form.Item key={subField.key} rules={[{ required: true }]} style={{ marginBottom: 0 }}>
              <Flex justify="center" align="flex-start" gap={5}>
                {index > 0 && (
                  <Form.Item style={{ marginBottom: 0 }}>
                    <Tag color="processing">且</Tag>
                  </Form.Item>
                )}

                <Form.Item
                  name={[subField.name, 'fromField']}
                  rules={[{ required: true, message: '请输入比较来源字段' }]}
                  style={{ marginBottom: 0 }}
                >
                  <Input placeholder="比较来源字段" />
                </Form.Item>
                <Form.Item
                  name={[subField.name, 'operation']}
                  style={{ marginBottom: 0 }}
                  initialValue="match"
                >
                  <Select options={operationOptions} style={{ width: 120 }}></Select>
                </Form.Item>
                <Form.Item
                  name={[subField.name, 'expr']}
                  className="flex-1"
                  rules={[{ required: true, message: '请输入正则表达式' }]}
                  style={{ marginBottom: 0 }}
                >
                  <Input placeholder="正则表达式" />
                </Form.Item>
                <Form.Item style={{ marginBottom: 0 }}>
                  <Button
                    size="small"
                    danger
                    icon={<AiOutlineLine />}
                    className="flex-grow-0 flex-shrink-0"
                    onClick={() => subOpt.remove(subField.name)}
                  >
                    {/* <span className="text-xs"></span> */}
                  </Button>
                </Form.Item>
              </Flex>
            </Form.Item>
          ))}
          <Button color="primary" variant="filled" onClick={() => subOpt.add()} block>
            + 新增过滤条件
          </Button>
        </div>
      )}
    </Form.List>
  )
}
export default ConditionsFormList
