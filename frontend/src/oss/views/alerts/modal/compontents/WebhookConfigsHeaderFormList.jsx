/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Col, Form, Input, Row } from 'antd'
import React from 'react'
import { IoIosRemoveCircleOutline, IoMdAddCircleOutline } from 'react-icons/io'
const reservedHeaders = [
  'Authorization',
  'Host',
  'Content-Encoding',
  'Content-Length',
  'Content-Type',
  'User-Agent',
  'Connection',
  'Keep-Alive',
  'Proxy-Authenticate',
  'Proxy-Authorization',
  'Www-Authenticate',
  'Accept-Encoding',
  'X-Prometheus-Remote-Write-Version',
  'X-Prometheus-Remote-Read-Version',
  'X-Prometheus-Scrape-Timeout-Seconds',
  'X-Amz-Date',
  'X-Amz-Security-Token',
  'X-Amz-Content-Sha256',
]
function isValidKey(str) {
  // 定义正则表达式，确保开头是字母或下划线
  const regex = /^[a-zA-Z_].*$/

  return regex.test(str)
}
export default function WebhookConfigsHeaderFormList({ formListName }) {
  return (
    <Form.List name={formListName}>
      {(fields, { add, remove }, { errors }) => (
        <>
          <Form.Item
            label={
              <>
                Header <IoMdAddCircleOutline onClick={() => add()} size={20} className="mx-2" />
              </>
            }
          >
            {fields.map((field, index) => (
              <div key={field.key} className="bg-[#323545] px-3 pt-3 pb-0 rounded relative">
                <Row gutter={12}>
                  <Col span={11} key={index}>
                    <Form.Item
                      noStyle
                      name={[field.name, 'key']}
                      rules={[
                        {
                          validator: async (_, key) => {
                            if (!isValidKey(key)) {
                              return Promise.reject(new Error('请以字母或下划线开头'))
                            }
                            if (reservedHeaders.includes(key)) {
                              return Promise.reject(new Error('特殊意义的Header不可设置'))
                            }
                          },
                        },
                      ]}
                    >
                      <Input placeholder="key" />
                    </Form.Item>
                  </Col>
                  <Col span={11} key={index}>
                    <Form.Item noStyle name={[field.name, 'value']}>
                      <Input placeholder="value" />
                    </Form.Item>
                  </Col>
                  <Col span={1}>
                    <IoIosRemoveCircleOutline
                      size={20}
                      className="mt-1"
                      onClick={() => remove(field.name)}
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
