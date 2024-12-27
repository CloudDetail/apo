/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Card, Form, Input } from 'antd'
import React from 'react'
import { defaultHtml } from './defaultHTMLcontext'
import { useTranslation } from 'react-i18next' // 添加i18n

export default function DingTalkConfigsFormList() {
  const { t } = useTranslation('oss/alert') // 使用i18n
  const tlsConfigItemsList = [
    {
      label: t('dingTalkConfigsFormList.url'),
      name: 'url',
      placeholder: t('dingTalkConfigsFormList.url'),
      required: true,
    },
    {
      label: t('dingTalkConfigsFormList.secret'),
      name: 'secret',
      placeholder: t('dingTalkConfigsFormList.secret'),
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
                {t('dingTalkConfigsFormList.title')}
                {/* <IoMdAddCircleOutline onClick={() => add()} size={20} className="mx-2" /> */}
              </span>
            }
          >
            {fields.map((field, index) => (
              <div key={field.key}>
                {tlsConfigItemsList.map((item) => (
                  <Form.Item
                    key={item.name}
                    label={item.label}
                    name={[field.name, item.name]}
                    rules={[
                      {
                        required: item.required,
                        message: `${item.label} ${t('dingTalkConfigsFormList.empty')}`,
                      },
                    ]}
                  >
                    <Input placeholder={item.placeholder} />
                  </Form.Item>
                ))}
              </div>
            ))}
          </Card>
        </>
      )}
    </Form.List>
  )
}
