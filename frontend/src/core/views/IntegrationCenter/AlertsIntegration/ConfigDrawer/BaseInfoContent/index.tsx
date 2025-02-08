/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Button, Card, ConfigProvider, Form, Input } from 'antd'
import Title from 'antd/es/typography/Title'
import Typography from 'antd/es/typography/Typography'
import { creatAlertInputSourceApi, updateAlertsIntegrationApi } from 'src/core/api/alertInput'
import { useEffect, useState } from 'react'
import { showToast } from 'src/core/utils/toast'
import BaseInfoDescriptions from './BaseInfoDescriptions'
import { useAlertIntegrationContext } from 'src/core/contexts/AlertIntegrationContext'
import { useSearchParams } from 'react-router-dom'
import { AlertInputSourceParams, AlertKey } from 'src/core/types/alertIntegration'
import Text from 'antd/es/typography/Text'
import { v4 as uuidv4 } from 'uuid'
import { useTranslation } from 'react-i18next'

interface BaseInfoContentProps {
  sourceId?: string | null
  sourceName?: string | null
  sourceType: AlertKey
  clusters?: any[]
  refreshDrawer: any
  closeDrawer?: any
}
const BaseInfoContent = (props: BaseInfoContentProps) => {
  const { t } = useTranslation('core/alertsIntegration')
  const [searchParams, setSearchParams] = useSearchParams()
  const { sourceId, sourceType, sourceName, clusters, refreshDrawer, closeDrawer } = props
  const configDrawerVisible = useAlertIntegrationContext((ctx) => ctx.configDrawerVisible)
  const [type, setType] = useState<'view' | 'edit'>('view')
  const [form] = Form.useForm()
  const uuid = uuidv4()
  const creatAlertsIntegration = (params: AlertInputSourceParams) => {
    creatAlertInputSourceApi(params)
      .then((res) => {
        showToast({
          title: t('addSuccess'),
          color: 'success',
        })
        setType('view')
        if (!sourceId) {
          const newParams = new URLSearchParams(searchParams)
          newParams.set('sourceId', res?.sourceId)
          setSearchParams(newParams, { replace: true })
        }
      })
      .catch((error) => {
        console.error(error)
      })
  }
  const updateAlertsIntegration = (params: AlertInputSourceParams) => {
    updateAlertsIntegrationApi(params).then((res) => {
      showToast({
        title: t('updatedSuccess'),
        color: 'success',
      })
      refreshDrawer()
      setType('view')
    })
  }
  function saveBaseInfo() {
    form.validateFields().then((values) => {
      const params: AlertInputSourceParams = {
        sourceName: values.sourceName,
        sourceType: sourceType,
        clusters: values.clusters || [],
      }
      if (sourceId) {
        params.sourceId = sourceId
        updateAlertsIntegration(params)
      } else {
        params.sourceId = uuid
        creatAlertsIntegration(params)
      }
    })
  }
  useEffect(() => {
    if (sourceId) {
      setType('view')
    } else {
      setType('edit')
    }
  }, [sourceId])

  useEffect(() => {
    if (type === 'edit') {
      form.setFieldsValue({
        sourceName: sourceName,
        clusters: clusters,
      })
    }
  }, [type])

  useEffect(() => {
    form.resetFields()
  }, [configDrawerVisible])
  const getPublishUrl = () => {
    const baseUrl = window.location.origin + '/api/alertinput/event/source?sourceId='
    if (sourceId) {
      return baseUrl + sourceId
    } else {
      return baseUrl + uuid
    }
  }
  return (
    <Card className="bg-[#202023] rounded-3xl" classNames={{ body: 'px-4 py-3' }}>
      {type === 'view' ? (
        sourceName && (
          <BaseInfoDescriptions sourceName={sourceName} clusters={clusters} sourceId={sourceId} />
        )
      ) : (
        <Form labelCol={{ span: 5, offset: 1 }} wrapperCol={{ span: 15 }} colon={false} form={form}>
          <Typography>
            <Title level={5}>{t('basicInfo')}</Title>
            <ConfigProvider
              theme={{
                components: {
                  Form: {
                    // labelColor: '#a6a6a6',
                  },
                },
              }}
            >
              <Form.Item name="sourceId" hidden></Form.Item>
              <Form.Item
                name="sourceName"
                label={t('sourceName')}
                required
                rules={[{ required: true }]}
              >
                <Input></Input>
              </Form.Item>
              <Form.Item label={t('pushUrl')}>
                <Text copyable={{ text: getPublishUrl }}>{getPublishUrl()}</Text>
              </Form.Item>

              {/* <ClusterSelector /> */}
            </ConfigProvider>
          </Typography>
        </Form>
      )}
      <div className="w-full text-right">
        {type === 'view' ? (
          <Button type="primary" onClick={() => setType('edit')}>
            {t('edit')}
          </Button>
        ) : (
          <>
            <Button type="primary" onClick={saveBaseInfo} className="mr-2">
              {t('save')}
            </Button>
            <Button
              onClick={() => {
                if (sourceId) {
                  setType('view')
                } else {
                  closeDrawer()
                }
              }}
            >
              {t('cancel')}
            </Button>
          </>
        )}
      </div>
    </Card>
  )
}
export default BaseInfoContent
