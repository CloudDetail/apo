/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Col, Form, Input, Row, Select } from 'antd'
import React from 'react'
import { AiOutlineInfoCircle } from 'react-icons/ai'
import { IoIosRemoveCircleOutline, IoMdAddCircleOutline } from 'react-icons/io'
import LogStructRuleFormList from './LogStructRuleFormList'
import { useTranslation } from 'react-i18next' // 引入i18n

const routeKeyList = [
  { label: '_container_id_', value: '_container_id_' },
  { label: '_source_', value: '_source_' },
  { label: 'container.image.name', value: 'container.image.name' },
  { label: 'container.ip', value: 'container.ip' },
  { label: 'container.name', value: 'container.name' },
  { label: 'content', value: 'content' },
  { label: 'host.ip', value: 'host.ip' },
  { label: 'host.name', value: 'host.name' },
  { label: 'k8s.namespace.name', value: 'k8s.namespace.name' },
  { label: 'k8s.node.ip', value: 'k8s.node.ip' },
  { label: 'k8s.node.name', value: 'k8s.node.name' },
  { label: 'k8s.pod.name', value: 'k8s.pod.name' },
  { label: 'k8s.pod.uid', value: 'k8s.pod.uid' },
]

function isValidKey(str) {
  // 定义正则表达式，确保开头是字母或下划线
  const regex = /^[a-zA-Z_].*$/

  return regex.test(str)
}
export default function LogRouteRuleFormList() {
  const { t } = useTranslation('oss/fullLogs') // 使用i18n
  const form = Form.useFormInstance()
  return (
    <Form.List name={'routeRule'}>
      {(fields, { add, remove }, { errors }) => (
        <>
          <Form.Item
            required
            label={
              <>
                {/* <div className="flex flex-row"> */}
                {t('configLogRuleModal.logRouteRuleFormList.matchRuleLabel')}{' '}
                <IoMdAddCircleOutline
                  onClick={() => add()}
                  size={20}
                  className="mx-2 cursor-pointer"
                />
                {/* </div>
                <div className="flex flex-row"> */}
                <AiOutlineInfoCircle size={16} className="ml-1 mr-1" />
                <span className="text-xs text-gray-400">
                  {t('configLogRuleModal.logRouteRuleFormList.matchRuleDescribeText')}
                </span>
                {/* </div> */}
              </>
            }
          >
            {fields.map((field, index) => (
              <div key={field.name} className=" px-3 pt-3 pb-0 rounded relative">
                <Row gutter={12}>
                  <Col span={11} key={index}>
                    <Form.Item
                      name={[field.name, 'key']}
                      required
                      rules={[
                        {
                          validator: async (_, value) => {
                            // 获取当前表单中所有的routeRule项
                            const routeRule = form.getFieldValue('routeRule') || []
                            // 检查是否有重复的key
                            if (!value) {
                              return Promise.reject(
                                t('configLogRuleModal.logRouteRuleFormList.errorInfo1'),
                              )
                            }
                            const duplicate = routeRule.filter(
                              (item, i) => item?.key?.key === value.key && i !== index,
                            )
                            if (duplicate.length) {
                              return Promise.reject(
                                t('configLogRuleModal.logRouteRuleFormList.errorInfo2'),
                              )
                            }
                          },
                        },
                      ]}
                    >
                      <Select
                        options={routeKeyList}
                        labelInValue
                        placeholder={t('configLogRuleModal.logRouteRuleFormList.selectPlaceholder')}
                      />
                    </Form.Item>
                  </Col>
                  <Col span={11} key={index}>
                    <Form.Item
                      name={[field.name, 'value']}
                      required
                      rules={[
                        {
                          validator: async (_, value) => {
                            // 获取当前表单中所有的routeRule项
                            const routeRule = form.getFieldValue('routeRule') || []
                            // 检查是否有重复的key
                            if (!value) {
                              return Promise.reject(
                                t('configLogRuleModal.logRouteRuleFormList.errorInfo3'),
                              )
                            }
                          },
                        },
                      ]}
                    >
                      <Input
                        placeholder={t('configLogRuleModal.logRouteRuleFormList.inputPlaceholder')}
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
