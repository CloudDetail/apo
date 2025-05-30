/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Card, Form, Input, Modal, Select, message } from 'antd'
import React, { useEffect, useState } from 'react'
import { addAlertNotifyApi, updateAlertNotifyApi } from 'core/api/alerts'
import { notify } from 'src/core/utils/notify'
import EmailConfigsFormList from './compontents/EmailConfigsFormList'
import WebhookConfigsFormList from './compontents/WebhookConfigsFormList'
import DingTalkConfigsFormList from './compontents/DingTalkConfigsFormList'
import WeChatConfigsFormList from './compontents/WeChatConfigsFormList'
import { useTranslation } from 'react-i18next' // 引入i18n

export default function ModifyAlertNotifyModal({
  modalVisible,
  notifyInfo = null,
  closeModal,
  refresh,
}) {
  const [messageApi, contextHolder] = message.useMessage()
  const [form] = Form.useForm()
  const [tip, setTip] = useState('')
  const { t } = useTranslation('oss/alert')

  const updateAlertNotify = (amConfigReceiver, type) => {
    let api = addAlertNotifyApi
    let params = typeof type === 'undefined' ? { amConfigReceiver } : { amConfigReceiver, type }

    if (notifyInfo) {
      api = updateAlertNotifyApi
      params.oldName = notifyInfo.name
    }

    api(params).then(() => {
      notify({
        message: t('modifyAlertNotifyModal.saveSuccess'),
        type: 'success',
      })
      closeModal()
      refresh()
    })
  }

  const saveRule = () => {
    let config
    if (form.getFieldsValue(true).wechatConfigs) {
      config = form.getFieldsValue(true).wechatConfigs[0]
      if (config.to_user || config.to_party || config.to_tag) {
        setTip('')
      } else {
        console.log(config.to_user)
        setTip(t('modifyAlertNotifyModal.invalidNotifyType'))
      }
    }
    form
      .validateFields()
      .then(() => {
        const formState = form.getFieldsValue(true)
        let amConfigReceiver = {
          name: formState.name,
        }

        if (formState.notifyType === 'email') {
          amConfigReceiver.emailConfigs = formState.emailConfigs
            ?.map((item) => {
              let config = Object.keys(item).reduce((acc, key) => {
                if (item[key] !== '' && !['smarthost', 'smarthostPort'].includes(key)) {
                  acc[key] = item[key]
                }
                return acc
              }, {})
              if (item.smarthost && item.smarthostPort) {
                config.smarthost = item.smarthost + ':' + item.smarthostPort
              }
              config.tlsConfig = {
                insecureSkipVerify: true,
              }
              config.requireTls = item.requireTls
              return config
            })
            .filter(Boolean)
        } else if (formState.notifyType === 'webhook') {
          amConfigReceiver.webhookConfigs = formState.webhookConfigs
            ?.map((item) => {
              let config = {}
              if (item.url) config.url = item.url
              config.httpConfig = {}
              if (item.authType === 'user' && item.basicAuthUsername && item.basicAuthPassword) {
                config.httpConfig.basicAuth = {
                  username: item.basicAuthUsername,
                  password: item.basicAuthPassword,
                }
              } else if (item.authType === 'token' && item.bearerToken) {
                config.httpConfig.bearerToken = item.bearerToken
              }

              if (item.webhookConfigsHeader?.length > 0) {
                let headers = {}
                item.webhookConfigsHeader?.forEach((header) => {
                  headers[header?.key] = { values: [header.value] }
                })
                config.httpConfig.httpHeaders = headers
              }
              config.httpConfig.tlsConfig = {
                insecureSkipVerify: true,
              }
              return config
            })
            .filter(Boolean)
        } else if (formState.notifyType === 'dingtalk') {
          amConfigReceiver.dingTalkConfigs = formState.dingTalkConfigs
            ?.map((item) => {
              let config = {}
              if (item.url) config.url = item.url
              if (item.secret) config.secret = item.secret
              return config
            })
            .filter(Boolean)
        } else if (formState.notifyType === 'wechat') {
          amConfigReceiver.wechatConfigs = formState.wechatConfigs
            ?.map((item) => {
              const config = {
                apiUrl: item.api_url,
                apiSecret: item.api_secret,
                corpId: item.corp_id,
                agentId: item.agent_id,
                toUser: item.to_user,
                toParty: item.to_party,
                toTag: item.to_tag,
              }

              // 过滤掉没有任何接收者 (toUser, toParty, toTag) 的配置
              const hasReceiver = config.toUser || config.toParty || config.toTag
              return hasReceiver ? config : null
            })
            .filter(Boolean)
        }

        if (
          (formState.notifyType === 'email' && !amConfigReceiver.emailConfigs?.length) ||
          (formState.notifyType === 'webhook' && !amConfigReceiver.webhookConfigs?.length) ||
          (formState.notifyType === 'dingtalk' && !amConfigReceiver.dingTalkConfigs?.length) ||
          (formState.notifyType === 'wechat' && !amConfigReceiver.wechatConfigs?.length)
        ) {
          return
        }
        updateAlertNotify(amConfigReceiver, formState.notifyType)
      })
      .catch((error) => console.log(error))
  }

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
          secret: config.secret,
        }
      })
      const wechatConfigs = notifyInfo?.wechatConfigs?.map((config) => {
        return {
          api_secret: config.apiSecret,
          corp_id: config.corpId,
          agent_id: config.agentId,
          to_user: config.toUser,
          to_party: config.toParty,
          to_tag: config.toTag,
        }
      })
      form.setFieldsValue({
        name: notifyInfo.name,
        notifyType: judgmentType(typeList.find((item) => Object.hasOwn(notifyInfo, item))),
        emailConfigs,
        webhookConfigs,
        dingTalkConfigs,
        wechatConfigs,
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

  const typeList = ['emailConfigs', 'webhookConfigs', 'dingTalkConfigs', 'wechatConfigs']

  return (
    <>
      {contextHolder}
      <Modal
        title={t('modifyAlertNotifyModal.title')}
        open={modalVisible}
        onCancel={closeModal}
        destroyOnClose
        centered
        okText={t('modifyAlertNotifyModal.save')}
        cancelText={t('modifyAlertNotifyModal.cancel')}
        maskClosable={false}
        onOk={saveRule}
        width="100vw"
        getContainer={false}
        classNames={{ content: 'h-screen', body: 'h-[90%] overflow-y-scroll' }}
        afterClose={() => setTip('')}
      >
        <Form layout={'vertical'} form={form} preserve={false}>
          <Card title={t('modifyAlertNotifyModal.basicConfig')}>
            <Form.Item
              label={t('modifyAlertNotifyModal.alertNotifyName')}
              name="name"
              required
              rules={[
                {
                  validator: async (_, value) => {
                    if (!value)
                      return Promise.reject(
                        new Error(t('modifyAlertNotifyModal.invalidNotifyName')),
                      )
                  },
                },
              ]}
            >
              <Input placeholder={t('modifyAlertNotifyModal.alertNotifyName')} />
            </Form.Item>
            <Form.Item
              label={t('modifyAlertNotifyModal.notifyType')}
              name="notifyType"
              required
              rules={[
                {
                  validator: async (_, value) => {
                    if (!value)
                      return Promise.reject(
                        new Error(t('modifyAlertNotifyModal.invalidNotifyType')),
                      )
                  },
                },
              ]}
            >
              <Select
                options={[
                  { label: t('modifyAlertNotifyModal.email'), value: 'email' },
                  { label: t('modifyAlertNotifyModal.webhook'), value: 'webhook' },
                  { label: t('modifyAlertNotifyModal.dingtalk'), value: 'dingtalk' },
                  { label: t('modifyAlertNotifyModal.wechat'), value: 'wechat' },
                ]}
                disabled={notifyInfo}
                placeholder={t('modifyAlertNotifyModal.typePlaceholder')}
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
                  {notifyType === 'wechat' && <WeChatConfigsFormList tip={tip} setTip={setTip} />}
                </>
              )
            }}
          </Form.Item>
        </Form>
      </Modal>
    </>
  )
}
