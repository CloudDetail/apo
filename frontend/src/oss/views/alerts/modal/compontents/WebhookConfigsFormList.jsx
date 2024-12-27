/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Button, Card, Col, Form, Input, Row, Segmented, Select, Space, Switch } from 'antd'
import TextArea from 'antd/es/input/TextArea'
import React, { useEffect, useState } from 'react'
import { IoIosRemoveCircleOutline, IoMdAddCircleOutline } from 'react-icons/io'
import WebhookConfigsHeaderFormList from './WebhookConfigsHeaderFormList'
import { useTranslation } from 'react-i18next' // 引入i18n

export default function WebhookConfigsFormList() {
  const [authType, setAuthType] = useState()
  const form = Form.useFormInstance()
  const formAuthType = Form.useWatch(['webhookConfigs'], form)
  const { t } = useTranslation('oss/alert') // 使用i18n

  useEffect(() => {
    if (formAuthType?.length > 0) setAuthType(formAuthType[0].authType)
  }, [formAuthType])
  return (
    <Form.List
      name="webhookConfigs"
      initialValue={[{ to: '', host: '', port: '' }]}
      rules={[
        {
          validator: async (_, names) => {
            if (!names || names.length < 1) {
              return Promise.reject(new Error(t('webhookConfigsFormList.addWebhook')))
            }
          },
        },
      ]}
    >
      {(fields, { add, remove }, { errors }) => (
        <>
          <Card
            title={
              <span className="flex items-center">
                {t('webhookConfigsFormList.title')}
                {/* <IoMdAddCircleOutline onClick={() => add()} size={20} className="mx-2" /> */}
              </span>
            }
          >
            {fields.map((field, index) => (
              <div
                key={field.key}
                className="bg-[#323545] px-3 pt-3 pb-0 rounded relative mb-2 mt-1"
              >
                {index > 0 && (
                  <IoIosRemoveCircleOutline
                    size={20}
                    className="mt-1 absolute -right-2 -top-2"
                    onClick={() => remove(field.name)}
                  />
                )}
                <Row gutter={12}>
                  <>
                    <Col span={24}>
                      <Form.Item
                        {...field}
                        name={[field.name, 'url']}
                        label={t('webhookConfigsFormList.url')}
                        rules={[
                          {
                            required: true,
                            message:
                              t('webhookConfigsFormList.url') +
                              ' ' +
                              t('modifyAlertNotifyModal.invalidNotifyType'),
                          },
                        ]}
                      >
                        <Input placeholder="Webhook URL" />
                      </Form.Item>
                    </Col>
                    <Col span={24}>
                      <Form.Item
                        {...field}
                        name={[field.name, 'authType']}
                        label={t('webhookConfigsFormList.authType')}
                      >
                        <Select
                          placeholder={t('webhookConfigsFormList.authType')}
                          options={[
                            { label: t('webhookConfigsFormList.userAuth'), value: 'user' },
                            { label: t('webhookConfigsFormList.tokenAuth'), value: 'token' },
                          ]}
                          onChange={setAuthType}
                        />
                      </Form.Item>
                    </Col>
                    {authType === 'user' && (
                      <>
                        <Col span={12}>
                          <Form.Item
                            {...field}
                            name={[field.name, 'basicAuthUsername']}
                            label={t('webhookConfigsFormList.username')}
                          >
                            <Input placeholder={t('webhookConfigsFormList.username')} />
                          </Form.Item>
                        </Col>
                        <Col span={12}>
                          <Form.Item
                            {...field}
                            name={[field.name, 'basicAuthPassword']}
                            label={t('webhookConfigsFormList.password')}
                          >
                            <Input placeholder={t('webhookConfigsFormList.password')} />
                          </Form.Item>
                        </Col>
                      </>
                    )}
                    {authType === 'token' && (
                      <>
                        <Col span={24}>
                          <Form.Item
                            {...field}
                            name={[field.name, 'bearerToken']}
                            label={t('webhookConfigsFormList.token')}
                          >
                            <Input placeholder={t('webhookConfigsFormList.token')} />
                          </Form.Item>
                        </Col>
                      </>
                    )}
                    {/* <Col span={24}>
                      <Form.Item noStyle {...field}>
                        <WebhookConfigsHeaderFormList
                          formListName={[field.name, 'webhookConfigsHeader']}
                        />
                      </Form.Item>
                    </Col> */}
                  </>
                </Row>
              </div>
            ))}
          </Card>
        </>
      )}
    </Form.List>
  )
}
