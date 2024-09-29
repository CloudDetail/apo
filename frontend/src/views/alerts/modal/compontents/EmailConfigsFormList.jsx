import { Card, Col, Form, Input, Row, Segmented, Switch } from 'antd'
import TextArea from 'antd/es/input/TextArea'
import React, { useState } from 'react'
import { IoIosRemoveCircleOutline, IoMdAddCircleOutline } from 'react-icons/io'

export default function EmailConfigsFormList() {
  const labelCol = { span: 8 }
  const tlsConfigItemsList = [
    {
      label: '接收通知邮箱地址',
      name: 'to',
      placeholder: '接收告警通知邮件的邮箱地址',
      required: true,
    },
    {
      label: '发送通知邮箱地址',
      name: 'from',
      placeholder: '发送告警通知邮件的发件人地址',
      required: true,
    },
    {
      label: 'SMTP服务器Host',
      name: 'smarthost',
      placeholder: 'SMTP服务器Host',
      required: true,
    },
    {
      label: 'SMTP服务器Port',
      name: 'smarthostPort',
      placeholder: 'SMTP服务器Port',
      required: true,
    },
    {
      label: 'SMTP用户名',
      name: 'authUsername',
      placeholder: 'SMTP用户名',
    },
    {
      label: 'SMTP密码',
      name: 'authPassword',
      placeholder: 'SMTP密码',
    },
    {
      label: '强制启用 TLS 安全传输',
      name: 'requireTls',
      type: 'boolean',
      disabled: true,
      layout: 'horizontal',
      valuePropName: 'checked',
      span: 24,
    },
    {
      label: '告警邮件HTML正文',
      name: 'html',
      type: 'textarea',
      placeholder: '告警邮件的 HTML 格式的正文内容',
    },
    {
      label: '告警邮件文本正文',
      name: 'text',
      type: 'textarea',
      placeholder: '告警邮件的文本格式的正文内容',
    },
  ]

  return (
    <Form.List name="emailConfigs" initialValue={[{ to: '' }]}>
      {(fields, { add, remove }, { errors }) => (
        <>
          <Card
            title={
              <span className="flex items-center">
                邮件通知
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
                        ]}
                        layout={item.layout}
                        valuePropName={item.valuePropName}
                      >
                        {item.children ? (
                          item.children
                        ) : item.type === 'boolean' ? (
                          <Switch defaultChecked disabled={item.disabled} />
                        ) : item.type === 'textarea' ? (
                          <TextArea placeholder={item.placeholder} />
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
