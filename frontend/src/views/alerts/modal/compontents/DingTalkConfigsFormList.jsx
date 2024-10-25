import { Card, Form, Input } from 'antd';
import React from 'react';
import { defaultHtml } from './defaultHTMLcontext';

export default function DingTalkConfigsFormList() {
  const tlsConfigItemsList = [
    {
      label: 'Webhook地址',
      name: 'url',
      placeholder: 'webhook URL',
      required: true,
    },
    {
      label: '加签密钥',
      name: 'secret',
      placeholder: '钉钉加签密钥',
      required: true,
    },
  ];

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
              <div key={field.key}>
                {tlsConfigItemsList.map((item) => (
                  <Form.Item
                    key={item.name}
                    label={item.label}
                    name={[field.name, item.name]} 
                    rules={[
                      {
                        required: item.required,
                        message: `${item.label} 是必填项`,
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
  );
}
