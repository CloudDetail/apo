import { Button, Card, Col, Form, Input, Row, Segmented, Select, Space, Switch } from 'antd'
import TextArea from 'antd/es/input/TextArea'
import React, { useEffect, useState } from 'react'
import { IoIosRemoveCircleOutline, IoMdAddCircleOutline } from 'react-icons/io'
import WebhookConfigsHeaderFormList from './WebhookConfigsHeaderFormList'

export default function WebhookConfigsFormList() {
  const [authType, setAuthType] = useState()
  const form = Form.useFormInstance()
  const formAuthType = Form.useWatch(['webhookConfigs'], form)
  // console.log(formAuthType)
  // const tlsConfigItemsList = [
  //   {
  //     label: 'URL',
  //     name: 'url',
  //     placeholder: 'Webhook地址',
  //     required: true,
  //     span: 24,
  //   },
  //   {
  //     label: '身份认证方式',
  //     name: 'authType',
  //     span: 24,
  //     children: (filed) => {
  //       return (
  //         <Select
  //           placeholder="身份认证方式"
  //           options={[
  //             { label: '用户密码', value: 'user' },
  //             { label: 'Token验证', value: 'token' },
  //           ]}
  //         />
  //       )
  //     },
  //   },
  //   // {
  //   //   name: 'auth',
  //   //   span: 24,
  //   //   noStyle: true,
  //   //   children: (e) => {
  //   //     console.log(e)
  //   //     return (
  //   //       <Row gutter={20}>
  //   //         {/* {({ getFieldValue }) => {
  //   //           const notifyType = getFieldValue('notifyType')
  //   //           return (
  //   //             <>
  //   //               {notifyType === 'email' && <EmailConfigsFormList />}
  //   //               {notifyType === 'webhook' && <WebhookConfigsFormList />}
  //   //             </>
  //   //           )
  //   //         }} */}
  //   //         {authType ? 1 : 2}
  //   //         {/* <Col span={12}>
  //   //           <Form.Item label="身份认证用户名" name="basicAuthUsername">
  //   //             <Input placeholder="身份认证用户名" />
  //   //           </Form.Item>
  //   //         </Col>
  //   //         <Col span={12}>
  //   //           <Form.Item label="身份认证密码" name="smarthostPort">
  //   //             <Input placeholder="身份认证密码" />
  //   //           </Form.Item>
  //   //         </Col> */}
  //   //       </Row>
  //   //     )
  //   //   },
  //   // },
  //   {
  //     label: '身份认证用户名',
  //     name: 'basicAuthUsername',
  //     placeholder: '身份认证用户名',
  //     hide: (authType) => {
  //       return authType !== 'user'
  //     },
  //   },
  //   {
  //     label: '身份认证密码',
  //     name: 'basicAuthPassword',
  //     placeholder: '身份认证密码',
  //     hide: () => {
  //       return authType !== 'user'
  //     },
  //   },
  //   {
  //     label: 'Token验证',
  //     name: 'bearerToken',
  //     placeholder: 'Token验证',
  //     span: 24,
  //     hide: () => {
  //       return authType !== 'token'
  //     },
  //   },
  //   {
  //     span: 24,
  //     noStyle: true,
  //     children: (filed) => {
  //       console.log()
  //       return <WebhookConfigsHeaderFormList formListName={[filed.name, 'webhookConfigsHeader']} />
  //     },
  //   },
  // ]
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
              return Promise.reject(new Error('至少设置1个邮件通知'))
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
                Webhook配置
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
                        label="URL"
                        rules={[
                          {
                            required: true,
                            message: 'URL不可为空',
                          },
                        ]}
                      >
                        <Input placeholder="Webhook URL" />
                      </Form.Item>
                    </Col>
                    <Col span={24}>
                      <Form.Item {...field} name={[field.name, 'authType']} label="身份认证方式">
                        <Select
                          placeholder="身份认证方式"
                          options={[
                            { label: '用户密码', value: 'user' },
                            { label: 'Token验证', value: 'token' },
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
                            label="身份认证用户名"
                          >
                            <Input placeholder="身份认证用户名" />
                          </Form.Item>
                        </Col>
                        <Col span={12}>
                          <Form.Item
                            {...field}
                            name={[field.name, 'basicAuthPassword']}
                            label="身份认证密码"
                          >
                            <Input placeholder="身份认证密码" />
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
                            label="Token验证"
                          >
                            <Input placeholder="Token验证" />
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
