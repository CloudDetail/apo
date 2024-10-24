import { Card, Col, Form, Input, InputNumber, Row, Segmented, Switch } from 'antd'
import TextArea from 'antd/es/input/TextArea'
import React, { useState } from 'react'
import { IoIosRemoveCircleOutline, IoMdAddCircleOutline } from 'react-icons/io'
import { defaultHtml } from './defaultHTMLcontext'

export default function DingTalkConfigsFormList() {
  const tlsConfigItemsList = [
    {
      label: 'Webhook地址',
      name: 'url',
      placeholder: 'webhook URL',
      required: true,
    },
    {
      label: '加密密钥',
      name: 'secret',
      placeholder: '钉钉加密密钥',
      required: true,
    },
  ]

  return (
    <Form.List name="dingTalkConfigs" initialValue={[{ html: defaultHtml, requireTls: false }]}>
      {(fields, { add, remove }, { errors }) => (
        <>
          <Card
            title={
              <span className="flex items-center">
                钉钉通知
                {/* <IoMdAddCircleOutline onClick={() => add()} size={20} className="mx-2" /> */}
              </span>
            }
          >
            {fields.map((field, index) => (
              <div className="bg-[#323545] px-3 pt-3 pb-0 rounded relative  mb-2 mt-1">
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
                            message: item.label + '不可为空',
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
