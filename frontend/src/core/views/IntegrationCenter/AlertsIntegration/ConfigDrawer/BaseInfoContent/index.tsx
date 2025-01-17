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
import ClusterSelector from './ClusterSelector'
import BaseInfoDescriptions from './BaseInfoDescriptions'
import { useAlertIntegrationContext } from 'src/core/contexts/AlertIntegrationContext'
import { useSearchParams } from 'react-router-dom'
import { AlertInputSourceParams, AlertKey } from 'src/core/types/alertIntegration'
interface BaseInfoContentProps {
  sourceId?: string | null
  sourceName?: string | null
  sourceType: AlertKey
  clusters?: any[]
  refreshDrawer: any
}
const BaseInfoContent = (props: BaseInfoContentProps) => {
  const [searchParams, setSearchParams] = useSearchParams()
  const { sourceId, sourceType, sourceName, clusters, refreshDrawer } = props
  const configDrawerVisible = useAlertIntegrationContext((ctx) => ctx.configDrawerVisible)
  const [type, setType] = useState<'view' | 'edit'>('view')
  const [form] = Form.useForm()
  const creatAlertsIntegration = (params: AlertInputSourceParams) => {
    creatAlertInputSourceApi(params)
      .then((res) => {
        showToast({
          title: '新增告警接入成功',
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
        title: '更新告警接入基础信息成功',
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
  return (
    <Card className="bg-[#202023] rounded-3xl" classNames={{ body: 'px-4 py-3' }}>
      {type === 'view' ? (
        sourceName && <BaseInfoDescriptions sourceName={sourceName} clusters={clusters} />
      ) : (
        <Form labelCol={{ span: 3, offset: 1 }} colon={false} form={form}>
          <Typography>
            <Title level={5}>基础信息</Title>
            <ConfigProvider
              theme={{
                components: {
                  Form: {
                    // labelColor: '#a6a6a6',
                  },
                },
              }}
            >
              <Form.Item name="sourceName" label="告警接入名" required rules={[{ required: true }]}>
                <Input></Input>
              </Form.Item>

              <ClusterSelector />
            </ConfigProvider>
          </Typography>
        </Form>
      )}
      <div className="w-full text-right">
        {type === 'view' ? (
          <Button type="primary" onClick={() => setType('edit')}>
            编辑
          </Button>
        ) : (
          <>
            <Button type="primary" onClick={saveBaseInfo} className="mr-2">
              保存
            </Button>
            <Button onClick={() => setType('view')}>取消</Button>
          </>
        )}
      </div>
    </Card>
  )
}
export default BaseInfoContent
