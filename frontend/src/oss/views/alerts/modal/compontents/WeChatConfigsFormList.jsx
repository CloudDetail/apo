import { Card, Form, Input } from 'antd'
import React from 'react'
import { defaultHtml } from './defaultHTMLcontext'
import { useTranslation } from 'react-i18next' // 添加i18n

export default function WeChatConfigsFormList({ tip, setTip }) {
  const { t } = useTranslation('oss/alert') // 使用i18n
  const tlsConfigItemsList = [
    {
      label: t('wechatConfigsFormList.api_url'),
      name: 'api_url',
      placeholder: t('wechatConfigsFormList.api_url'),
      required: true,
      defaultUrl: 'https://qyapi.weixin.qq.com/cgi-bin/',
    },
    {
      label: t('wechatConfigsFormList.api_secret'),
      name: 'api_secret',
      placeholder: t('wechatConfigsFormList.api_secret'),
      required: true,
    },
    {
      label: t('wechatConfigsFormList.corp_id'),
      name: 'corp_id',
      placeholder: t('wechatConfigsFormList.corp_id'),
      required: true,
    },
    {
      label: t('wechatConfigsFormList.agent_id'),
      name: 'agent_id',
      placeholder: t('wechatConfigsFormList.agent_id'),
      required: true,
    },
    {
      list: [
        {
          label: t('wechatConfigsFormList.to_user'),
          name: 'to_user',
          placeholder: t('wechatConfigsFormList.to_user'),
        },
        {
          label: t('wechatConfigsFormList.to_party'),
          name: 'to_party',
          placeholder: t('wechatConfigsFormList.to_party'),
        },
        {
          label: t('wechatConfigsFormList.to_tag'),
          name: 'to_tag',
          placeholder: t('wechatConfigsFormList.to_tag'),
        },
      ],
      group: 'notification',
      label: t('wechatConfigsFormList.notificationType'),
      name: 'notificationType',
    },
  ]
  return (
    <Form.List name="wechatConfigs" initialValue={[{ html: defaultHtml, requireTls: false }]}>
      {(fields) => (
        <>
          <Card title={t('wechatConfigsFormList.title')}>
            {fields.map((field) => (
              <div key={field.key}>
                {tlsConfigItemsList.map((item) => {
                  if (!item.group) {
                    return (
                      <Form.Item
                        key={item.name}
                        label={item.label}
                        name={[field.name, item.name]}
                        rules={[
                          {
                            required: item.required,
                            message: `${item.label} ${t('wechatConfigsFormList.empty')}`,
                          },
                        ]}
                        initialValue={item.defaultUrl}
                      >
                        <Input placeholder={item.placeholder} defaultValue={item.defaultUrl} />
                      </Form.Item>
                    )
                  } else {
                    return (
                      <Form.Item
                        name="notificationType"
                        key={item.group}
                        label={item.label}
                        required
                        validateTrigger={['onBlur']}
                      >
                        <Card>
                          {item.list.map((subItem) => (
                            <Form.Item
                              key={subItem.name}
                              label={subItem.label}
                              name={[field.name, subItem.name]}
                            >
                              <Input placeholder={subItem.placeholder} />
                            </Form.Item>
                          ))}
                        </Card>
                        {<p style={{ color: '#CA4547' }}>{tip}</p>}
                      </Form.Item>
                    )
                  }
                })}
              </div>
            ))}
          </Card>
        </>
      )}
    </Form.List>
  )
}
