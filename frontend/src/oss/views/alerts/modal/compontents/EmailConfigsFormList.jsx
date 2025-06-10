/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Card, Col, Form, Input, InputNumber, Row, Segmented, Switch, theme } from 'antd'
import TextArea from 'antd/es/input/TextArea'
import React, { useState } from 'react'
import { IoIosRemoveCircleOutline, IoMdAddCircleOutline } from 'react-icons/io'
import { defaultHtml } from './defaultHTMLcontext'
import { useTranslation } from 'react-i18next'

export default function EmailConfigsFormList() {
  const { t } = useTranslation('oss/alert')
  const { useToken } = theme
  const { token } = useToken()
  const labelCol = { span: 8 }
  const tlsConfigItemsList = [
    {
      label: t('emailConfigsFormList.to'),
      name: 'to',
      placeholder: t('emailConfigsFormList.to'),
      required: true,
    },
    {
      label: t('emailConfigsFormList.from'),
      name: 'from',
      placeholder: t('emailConfigsFormList.from'),
      required: true,
    },
    {
      label: t('emailConfigsFormList.smarthost'),
      name: 'smarthost',
      placeholder: t('emailConfigsFormList.smarthost'),
      required: true,
    },
    {
      label: t('emailConfigsFormList.smarthostPort'),
      name: 'smarthostPort',
      placeholder: t('emailConfigsFormList.smarthostPort'),
      required: true,
      type: 'number',
    },
    {
      label: t('emailConfigsFormList.authUsername'),
      name: 'authUsername',
      placeholder: t('emailConfigsFormList.authUsername'),
    },
    {
      label: t('emailConfigsFormList.authPassword'),
      name: 'authPassword',
      placeholder: t('emailConfigsFormList.authPassword'),
    },
    {
      label: t('emailConfigsFormList.requireTls'),
      name: 'requireTls',
      type: 'boolean',
      layout: 'horizontal',
      valuePropName: 'checked',
      span: 24,
    },
    {
      label: t('emailConfigsFormList.html'),
      name: 'html',
      type: 'textarea',
      placeholder: t('emailConfigsFormList.html'),
    },
    {
      label: t('emailConfigsFormList.text'),
      name: 'text',
      type: 'textarea',
      placeholder: t('emailConfigsFormList.text'),
    },
  ]

  return (
    <Form.List name="emailConfigs" initialValue={[{ html: defaultHtml, requireTls: false }]}>
      {(fields, { add, remove }, { errors }) => (
        <>
          <Card
            title={
              <span className="flex items-center">
                {t('emailConfigsFormList.title')}
                {/* <IoMdAddCircleOutline onClick={() => add()} size={20} className="mx-2" /> */}
              </span>
            }
          >
            {fields.map((field, index) => (
              <div className="px-3 pt-3 pb-0 rounded relative  mb-2 mt-1" style={{ backgroundColor: token.colorFillQuaternary }}>
                {index > 0 && (
                  <IoIosRemoveCircleOutline
                    size={20}
                    className="mt-1 absolute -right-2 -top-2"
                    onClick={() => remove(field.name)}
                  />
                )}
                <Row gutter={12}>
                  {tlsConfigItemsList.map((item, index) => (
                    <Col span={item.span ?? 12} key={index}>
                      <Form.Item
                        {...field}
                        label={item.label}
                        name={[field.name, item.name]}
                        className={item.className}
                        rules={[
                          {
                            required: item.required,
                            message: item.label + t('emailConfigsFormList.empty'),
                          },
                          ...(item.rules ?? []),
                        ]}
                        layout={item.layout}
                        valuePropName={item.valuePropName}
                      >
                        {item.children ? (
                          item.children
                        ) : item.type === 'boolean' ? (
                          <Switch disabled={item.disabled} />
                        ) : item.type === 'textarea' ? (
                          <TextArea
                            placeholder={item.placeholder}
                            defaultValue={item.defaultValue}
                          />
                        ) : item.type === 'number' ? (
                          <InputNumber placeholder={item.placeholder} className="w-full" />
                        ) : (
                          <Input placeholder={item.placeholder} />
                        )}
                      </Form.Item>
                    </Col>
                  ))}
                </Row>
              </div>
            ))}
          </Card>
        </>
      )}
    </Form.List>
  )
}
