import { Card, Form, Input } from 'antd';
import React from 'react';
import { defaultHtml } from './defaultHTMLcontext';

export default function WeChatConfigsFormList() {
    const tlsConfigItemsList = [
        {
            label: 'API地址',
            name: 'api_url',
            placeholder: 'wechat URL',
            required: true,
            defaultUrl: "https://qyapi.weixin.qq.com/cgi-bin/"
        },
        {
            label: 'API 密钥',
            name: 'api_secret',
            placeholder: '微信api密钥',
            required: true,
        },
        {
            label: '企业 ID',
            name: 'corp_id',
            placeholder: '企业 ID',
            required: true
        },
        {
            label: '应用 ID',
            name: 'agent_id',
            placeholder: '应用 ID',
            required: true
        },
        {
            list: [
                {
                    label: '通知到用户',
                    name: 'to_user',
                    placeholder: '支持输入用户名或@all，多个接收者用‘|’分隔',
                },
                {
                    label: '通知到部门',
                    name: 'to_party',
                    placeholder: '指定接收消息的部门，部门ID列表，多个接收者用‘|’分隔',
                },
                {
                    label: '通知到标签',
                    name: 'to_tag',
                    placeholder: '指定接收消息的标签，标签ID列表，多个接收者用‘|’分隔',
                }
            ],
            group: 'notification',
            label: '通知方式',
            name: 'notificationType'
        }
    ];

    return (
        <Form.List name="wechatConfigs" initialValue={[{ html: defaultHtml, requireTls: false }]}>
            {(fields) => (
                <>
                    <Card title="微信通知">
                        {fields.map((field) => (
                            <div key={field.key}>
                                {tlsConfigItemsList.map((item) => {
                                    if (!item.group) {
                                        return (
                                            <Form.Item
                                                key={item.name}
                                                label={item.label}
                                                name={[field.name, item.name]}
                                                rules={[{ required: item.required, message: `${item.label} 是必填项` }]}
                                                initialValue={item.defaultUrl}
                                            >
                                                <Input placeholder={item.placeholder} defaultValue={item.defaultUrl} />
                                            </Form.Item>
                                        );
                                    } else {
                                        return (
                                            <Form.Item
                                                key={item.group}
                                                label={item.label}
                                                required
                                                validateTrigger={['onBlur', 'onChange']}
                                                rules={[
                                                    {
                                                        validator: async (_, __) => {
                                                            const formInstance = Form.useFormInstance();
                                                            const toUser = formInstance.getFieldValue([field.name, 'to_user']);
                                                            const toParty = formInstance.getFieldValue([field.name, 'to_party']);
                                                            const toTag = formInstance.getFieldValue([field.name, 'to_tag']);
                                                            if (!toUser && !toParty && !toTag) {
                                                                throw new Error('至少填写一个通知方式');
                                                            }
                                                        }
                                                    }
                                                ]}
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
                                            </Form.Item>
                                        );
                                    }
                                })}
                            </div>
                        ))}
                    </Card>
                </>
            )}
        </Form.List>
    );
}
