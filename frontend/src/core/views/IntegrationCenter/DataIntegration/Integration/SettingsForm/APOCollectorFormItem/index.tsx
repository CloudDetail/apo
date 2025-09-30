/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Button, Card, Divider, Form, Input, Space, Typography } from 'antd'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { portsDefault } from 'src/core/views/IntegrationCenter/constant'

const APOCollectorFormItem = ({
  defaultGatewayAddr,
  isMinimal = false,
}: {
  defaultGatewayAddr?: string
  isMinimal?: boolean
}) => {
  const { t } = useTranslation('core/dataIntegration')
  const [showMore, setShowMore] = useState(false)
  return (
    <>
      <Divider></Divider>
      <Typography.Title level={5}>{t('apoCollectorSetting')}</Typography.Title>
      <div className="px-3">
        {!isMinimal && (
          <Form.Item
            label={
              <Space>
                {t('collectorGatewayAddr')}
                <span className=" text-[var(--ant-color-text-secondary)] text-xs">
                  {t('collectorGatewayAddrHint')}
                </span>
              </Space>
            }
            name={['apoCollector', 'collectorGatewayAddr']}
            rules={[{ required: true }]}
          >
            <Input />
          </Form.Item>
        )}

        {showMore ? (
          <Card type="inner" title={t('advanced')} size="small">
            {isMinimal && (
              <Form.Item
                label={
                  <Space>
                    {t('collectorGatewayAddr')}
                    <span className=" text-[var(--ant-color-text-secondary)] text-xs">
                      {t('collectorGatewayAddrHint')}
                    </span>
                  </Space>
                }
                name={['apoCollector', 'collectorGatewayAddr']}
                initialValue={defaultGatewayAddr}
              >
                <Input />
              </Form.Item>
            )}
            <Form.Item
              label={t('collectorAddr')}
              name={['apoCollector', 'collectorAddr']}
              initialValue={'apo-nginx-proxy-svc'}
            >
              <Input />
            </Form.Item>
            {portsDefault.map((item) => (
              <Form.Item
                name={['apoCollector', 'ports', item.key]}
                label={
                  <Space>
                    {item.key}
                    <span className="text-xs text-[var(--ant-color-text-secondary)]">
                      {item.descriptions}
                    </span>
                  </Space>
                }
                rules={[{ required: true }]}
                initialValue={item.value}
              >
                <Input />
              </Form.Item>
            ))}

            {/* <Form.List name={['apoCollector', 'ports']} initialValue={portsDefault}>
              {(fields, { add, remove }) => (
                <>
                  {fields.map(({ key, name, ...restField }) => (
                    <div className="flex">
                      <Form.Item {...restField} name={[name, 'key']} hidden>
                        <Input />
                      </Form.Item>
                      <Form.Item
                        {...restField}
                        name={[name, 'title']}
                        label={' '}
                        className="w-[200px]"
                      >
                        <Input readOnly variant="borderless" />
                      </Form.Item>
                      <Form.Item
                        {...restField}
                        name={[name, 'value']}
                        label={t('port')}
                        rules={[{ required: true }]}
                        className="w-[80px]"
                      >
                        <Input />
                      </Form.Item>
                      <Form.Item
                        {...restField}
                        name={[name, 'descriptions']}
                        label=" "
                        className="flex-grow"
                      >
                        <Input readOnly variant="borderless" />
                      </Form.Item>
                    </div>
                  ))}
                </>
              )}
            </Form.List> */}
          </Card>
        ) : (
          <Button color="primary" variant="outlined" onClick={() => setShowMore(true)}>
            {t('advanced')}
          </Button>
        )}
      </div>
      <Divider></Divider>
    </>
  )
}
export default APOCollectorFormItem
