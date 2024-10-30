import { Card, Form, Input, Modal, Select, Tag, Tooltip } from 'antd'
import _, { reject } from 'lodash'
import React, { useEffect, useState } from 'react'
import { useSelector } from 'react-redux'
import { addAlertNotifyApi, updateAlertNotifyApi } from 'src/api/alerts'
import { showToast } from 'src/utils/toast'
import EmailConfigsFormList from './compontents/EmailConfigsFormList'
import WebhookConfigsFormList from './compontents/WebhookConfigsFormList'
import DingTalkConfigsFormList from './compontents/DingTalkConfigsFormList'
import WeChatConfigsFormList from './compontents/WeChatConfigsFormList'
export default function ModifyAlertNotifyModal({
  modalVisible,
  notifyInfo = null,
  closeModal,
  refresh,
}) {
  const [form] = Form.useForm()
  const updateAlertNotify = (amConfigReceiver, type) => {
    let api = addAlertNotifyApi;
    let params = typeof type === 'undefined'
      ? { amConfigReceiver }
      : { amConfigReceiver, type };

    if (notifyInfo) {
      api = updateAlertNotifyApi;
      params.oldName = notifyInfo.name;
    }

    api(params).then(() => {
      showToast({
        title: '保存告警通知配置成功',
        color: 'success',
      });
      closeModal();
      refresh();
    });
  };

  const saveRule = () => {
    console.log(form.getFieldsValue(true));
    form
      .validateFields()
      .then(() => {
        const formState = form.getFieldsValue(true);
        let amConfigReceiver = {
          name: formState.name,
        };

        if (formState.notifyType === 'email') {
          amConfigReceiver.emailConfigs = formState.emailConfigs?.map((item) => {
            let config = Object.keys(item).reduce((acc, key) => {
              if (item[key] !== '' && !['smarthost', 'smarthostPort'].includes(key)) {
                acc[key] = item[key];
              }
              return acc;
            }, {});
            if (item.smarthost && item.smarthostPort) {
              config.smarthost = item.smarthost + ':' + item.smarthostPort;
            }
            config.tlsConfig = {
              insecureSkipVerify: true,
            };
            config.requireTls = item.requireTls;
            return config;
          }).filter(Boolean);
        } else if (formState.notifyType === 'webhook') {
          amConfigReceiver.webhookConfigs = formState.webhookConfigs?.map((item) => {
            let config = {};
            if (item.url) config.url = item.url;
            config.httpConfig = {};
            if (item.authType === 'user' && item.basicAuthUsername && item.basicAuthPassword) {
              config.httpConfig.basicAuth = {
                username: item.basicAuthUsername,
                password: item.basicAuthPassword,
              };
            } else if (item.authType === 'token' && item.bearerToken) {
              config.httpConfig.bearerToken = item.bearerToken;
            }

            if (item.webhookConfigsHeader?.length > 0) {
              let headers = {};
              item.webhookConfigsHeader?.forEach((header) => {
                headers[header?.key] = { values: [header.value] };
              });
              config.httpConfig.httpHeaders = headers;
            }
            config.httpConfig.tlsConfig = {
              insecureSkipVerify: true,
            };
            return config;
          }).filter(Boolean);
        } else if (formState.notifyType === 'dingtalk') {
          amConfigReceiver.dingTalkConfigs = formState.dingTalkConfigs?.map((item) => {
            let config = {};
            if (item.url) config.url = item.url;
            if (item.secret) config.secret = item.secret;
            return config;
          }).filter(Boolean);
        } else if (formState.notifyType === 'wechat') {
          amConfigReceiver.wechatConfigs = formState.wechatConfigs?.map((item) => {
            let flag = false;
            let config = {};
            if (item.api_url) config.api_url = item.api_url;
            if (item.api_secret) config.api_secret = item.api_secret;
            if (item.corp_id) config.corp_id = item.corp_id;
            if (item.agent_id) config.agent_id = item.agent_id;
            if (item.to_user) {
              config.to_user = item.to_user;
              flag = true;
            }
            if (item.to_party) {
              config.to_party = item.to_party;
              flag = true;
            }
            if (item.to_tag) {
              config.to_tag = item.to_tag;
              flag = true;
            }
            if (!flag) {
              return null;
            }
            return config;
          }).filter(Boolean);
        }

        if (
          (formState.notifyType === 'email' && !amConfigReceiver.emailConfigs?.length) ||
          (formState.notifyType === 'webhook' && !amConfigReceiver.webhookConfigs?.length) ||
          (formState.notifyType === 'dingtalk' && !amConfigReceiver.dingTalkConfigs?.length) ||
          (formState.notifyType === 'wechat' && !amConfigReceiver.wechatConfigs?.length)
        ) {
          return;
        }

        updateAlertNotify(amConfigReceiver, formState.notifyType);
      })
      .catch((error) => console.log(error));
  };

  useEffect(() => {
    //console.log(notifyInfo)
    if (notifyInfo && modalVisible) {
      const emailConfigs = notifyInfo?.emailConfigs?.map((config) => {
        //编辑的时候就校验 端口号不允许冒号存在
        let smarthost, smarthostPort
        const lastColonIndex = (config.smarthost ?? '').lastIndexOf(':')
        if (lastColonIndex > -1) {
          smarthost = config.smarthost.slice(0, lastColonIndex)
          smarthostPort = config.smarthost.slice(lastColonIndex + 1)
        }
        return {
          to: config.to,
          from: config.from,
          smarthost,
          smarthostPort,
          authUsername: config.authUsername,
          authPassword: config.authPassword,
          html: config.html,
          text: config.text,
          requireTls: config.requireTls,
        }
      })
      const webhookConfigs = notifyInfo?.webhookConfigs?.map((config) => {
        let authType
        if (config?.httpConfig.basicAuth) {
          authType = 'user'
        }
        if (config?.httpConfig.bearerToken) {
          authType = 'token'
        }
        // console.log(config.httpConfig?.httpHeaders)
        // const webhookConfigsHeader = Object.entries(config.httpConfig?.httpHeaders)?.map(
        //   ([key, value]) => {
        //     console.log({
        //       {
        //       key,
        //       value,
        //     }
        //     })
        //     return Object.entries(header).map(([key, value]) => ({
        //       key,
        //       value,
        //     }))
        //   },
        // )
        // const webhookConfigsHeader = Object.entries(config.httpConfig?.httpHeaders)?.map((header) => {
        //   console.log(header)
        //   return Object.entries(header).map(([key, value]) => ({
        //     key,
        //     value,
        //   }))
        // })
        return {
          url: config.url,
          authType,
          basicAuthUsername: config?.httpConfig?.basicAuth?.username,
          basicAuthPassword: config?.httpConfig?.basicAuth?.password,
          bearerToken: config?.httpConfig.bearerToken,
          // webhookConfigsHeader: webhookConfigsHeader,
        }
      })
      const dingTalkConfigs = notifyInfo?.dingTalkConfigs?.map((config) => {
        return {
          url: config.url,
          secret: config.secret
        }
      })
      const wechatConfigs = notifyInfo?.wechatConfigs?.map((config) => {
        return {
          api_url: config.api_url,
          api_secret: config.api_secret,
          corp_id: config.corp_id,
          agent_id: config.agent_id,
          to_user: config.to_user,
          to_party: config.to_party,
          to_tag: config.to_tag
        }
      })
      form.setFieldsValue({
        name: notifyInfo.name,
        notifyType: judgmentType(typeList.find(item => Object.hasOwn(notifyInfo, item))),
        emailConfigs,
        webhookConfigs,
        dingTalkConfigs,
        wechatConfigs
      })
    } else {
      form.resetFields()
    }
  }, [modalVisible, notifyInfo])

  const judgmentType = (type) => {
    switch (type) {
      case 'emailConfigs':
        return 'email'
      case 'webhookConfigs':
        return 'webhook'
      case 'dingTalkConfigs':
        return 'dingtalk'
      case 'wechatConfigs':
        return 'wechat'
    }
  }

  const typeList = [
    'emailConfigs',
    'webhookConfigs',
    'dingTalkConfigs',
    'wechatConfigs'
  ];

  return (
    <>
      <Modal
        title={'告警通知配置'}
        open={modalVisible}
        onCancel={closeModal}
        destroyOnClose
        centered
        okText={'保存'}
        cancelText="取消"
        maskClosable={false}
        onOk={saveRule}
        width="100vw"
        getContainer={false}
        classNames={{ content: 'h-screen', body: 'h-[90%] overflow-y-scroll' }}
      >
        <Form layout={'vertical'} form={form} preserve={false}>
          <Card title="基础配置">
            <Form.Item
              label="告警通知名"
              name="name"
              required
              rules={[
                {
                  validator: async (_, value) => {
                    if (!value) return Promise.reject(new Error('告警通知名不可为空'))
                  },
                },
              ]}
            >
              <Input placeholder="告警规则名" />
            </Form.Item>
            <Form.Item
              label="告警类型"
              name="notifyType"
              required
              rules={[
                {
                  validator: async (_, value) => {
                    if (!value) return Promise.reject(new Error('告警类型不可为空'))
                  },
                },
              ]}
            >
              <Select
                options={[
                  { label: '邮件通知', value: 'email' },
                  { label: 'Webhook通知', value: 'webhook' },
                  { label: '钉钉', value: 'dingtalk' },
                  { label: '企业微信', value: 'wechat' }
                ]}
                disabled={notifyInfo}
              />
            </Form.Item>
          </Card>
          <Form.Item
            noStyle
            shouldUpdate={(prevValues, curValues) => prevValues.notifyType !== curValues.notifyType}
          >
            {({ getFieldValue }) => {
              const notifyType = getFieldValue('notifyType')
              return (
                <>
                  {notifyType === 'email' && <EmailConfigsFormList />}
                  {notifyType === 'webhook' && <WebhookConfigsFormList />}
                  {notifyType === 'dingtalk' && <DingTalkConfigsFormList />}
                  {notifyType === 'wechat' && <WeChatConfigsFormList />}
                </>
              )
            }}
          </Form.Item>
        </Form>
      </Modal>
    </>
  )
}
