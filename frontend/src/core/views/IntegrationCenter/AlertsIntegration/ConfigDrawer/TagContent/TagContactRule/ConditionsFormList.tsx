/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Button, Flex, Form, Input, Select, Tag } from 'antd'
import { useTranslation } from 'react-i18next'
import { AiOutlineLine } from 'react-icons/ai'
interface ConditionsFormListProps {
  fieldName: string | number
}

const ConditionsFormList = ({ fieldName }: ConditionsFormListProps) => {
  const { t } = useTranslation('core/alertsIntegration')
  const operationOptions = [
    {
      value: 'match',
      label: <span>{t('match')}</span>,
    },
    {
      value: 'notMatch',
      label: <span>{t('misMatch')}</span>,
    },
  ]
  return (
    <Form.List name={[fieldName, 'conditions']}>
      {(subFields, subOpt) => (
        <div style={{ display: 'flex', flexDirection: 'column', rowGap: 16 }}>
          {subFields.map((subField, index) => (
            <Form.Item key={subField.key} rules={[{ required: true }]} style={{ marginBottom: 0 }}>
              <Flex justify="center" align="flex-start" gap={5}>
                {index > 0 && (
                  <Form.Item style={{ marginBottom: 0 }}>
                    <Tag color="processing">{t('and')}</Tag>
                  </Form.Item>
                )}

                <Form.Item
                  name={[subField.name, 'fromField']}
                  rules={[{ required: true, message: t('fromFieldRequired') }]}
                  style={{ marginBottom: 0 }}
                >
                  <Input placeholder={t('fromField')} />
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
                  rules={[{ required: true, message: t('exprRequired') }]}
                  style={{ marginBottom: 0 }}
                >
                  <Input placeholder={t('expr')} />
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
            + {t('addConditions')}
          </Button>
        </div>
      )}
    </Form.List>
  )
}
export default ConditionsFormList
