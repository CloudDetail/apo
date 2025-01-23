/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Button, Form } from 'antd'
import { getAlertEnrichApi, saveAlertEnrichApi } from 'src/core/api/alertInput'
import { showToast } from 'src/core/utils/toast'
import { useAlertIntegrationContext } from 'src/core/contexts/AlertIntegrationContext'
import { useEffect, useState } from 'react'
import TagRuleFormListCom from './TagRuleFormListCom'
import LoadingSpinner from 'src/core/components/Spinner'
import {
  EnrichRuleConfigItem,
  SaveAlertEnrichParams,
  SchemaTargetItem,
} from 'src/core/types/alertIntegration'
interface TagContactRuleProps {
  sourceId: string
}
const TagContactRule = ({ sourceId }: TagContactRuleProps) => {
  const [form] = Form.useForm()
  const configDrawerVisible = useAlertIntegrationContext((ctx) => ctx.configDrawerVisible)
  const [readOnly, setReadOnly] = useState(true)
  const [loading, setLoading] = useState(false)

  const [isFormEmpty, setIsFormEmpty] = useState(false)

  const saveAlertEnrich = (params: SaveAlertEnrichParams) => {
    setLoading(true)
    saveAlertEnrichApi(params)
      .then((res) => {
        showToast({
          title: '保存关联应用规则规则成功',
          color: 'success',
        })
        refreshReadOnly()
      })
      .catch(() => {
        setLoading(false)
      })
  }
  function saveForm() {
    form.validateFields().then((values) => {
      const enrichRuleConfigs: EnrichRuleConfigItem[] = []
      values.enrichRuleConfigs.map((config) => {
        let newConfig: EnrichRuleConfigItem = {
          rType: config.rType,
          fromField: config.fromField,
          fromRegex: config.fromRegex,
          conditions: config.conditions,
          enrichRuleId: config.enrichRuleId,
        }
        if (newConfig.rType === 'tagMapping') {
          newConfig.targetTagId = config.targetTag.targetTagId
          if (config.targetTag.targetTagId === 0) {
            newConfig.customTag = config.targetTag.customTag
          }
        } else {
          newConfig.schema = config.schemaObject[0]
          newConfig.schemaSource = config.schemaObject[1]
          //   targetTagId,   customTag,  schemaField
          const schemaTargets: SchemaTargetItem[] = []
          config.schemaTargets.map((target) => {
            let newTarget = {
              schemaField: target.schemaField,
              targetTagId: target.targetTag.targetTagId,
            }
            if (target.targetTag.targetTagId === 0) {
              newTarget.customTag = target.targetTag.customTag
            }
            schemaTargets.push(newTarget)
          })
          newConfig.schemaTargets = schemaTargets
        }
        enrichRuleConfigs.push(newConfig)
      })

      const params = {
        sourceId: sourceId,
        enrichRuleConfigs: enrichRuleConfigs,
      }
      saveAlertEnrich(params)
    })
  }
  useEffect(() => {
    if (configDrawerVisible) {
      form.resetFields()
    } else {
      setReadOnly(true)
    }
  }, [configDrawerVisible])
  const getAlertEnrichInfo = async () => {
    try {
      setLoading(true)
      const res = await getAlertEnrichApi({ sourceId })

      if (!res?.enrichRuleConfigs || res.enrichRuleConfigs.length === 0) {
        setIsFormEmpty(true)
        return
      }
      setIsFormEmpty(false)

      const enrichRuleConfigs = res.enrichRuleConfigs.map((config) => {
        // 创建一个浅拷贝对象
        const newConfig = { ...config }

        if (config.rType === 'tagMapping') {
          // 处理 tagMapping 类型
          newConfig.targetTag = {
            targetTagId: config.targetTagId,
            customTag: config.targetTagId === 0 ? config.customTag : '',
          }
        } else {
          // 处理其他类型
          newConfig.schemaObject = [config.schema, config.schemaSource]
          newConfig.schemaTargets = config.schemaTargets.map((target) => ({
            schemaField: target.schemaField,
            targetTag: {
              targetTagId: target.targetTagId,
              customTag: target.targetTagId === 0 ? target.customTag : '',
            },
          }))
          newConfig.schemaFields = newConfig.schemaTargets.map((target) => target.schemaField)
        }

        return newConfig
      })

      form.setFieldsValue({ enrichRuleConfigs })
    } catch (error) {
      console.error('Failed to fetch alert enrich info:', error)
    } finally {
      setLoading(false)
    }
  }
  useEffect(() => {
    if (sourceId) getAlertEnrichInfo()
  }, [sourceId])

  const refreshReadOnly = async () => {
    try {
      await getAlertEnrichInfo()
      setReadOnly(true)
      // 下一步操作放在这里
    } catch (error) {
      console.error('Error in refreshReadOnly:', error)
    }
  }
  const creatFirstRuleConfig = () => {
    setReadOnly(false)
    form.setFieldValue(['enrichRuleConfigs', 0], {})
  }
  const editFormList = (ref) => {
    setReadOnly(false)
    setTimeout(() => {
      if (ref?.current) {
        ref.current.scrollIntoView({ behavior: 'smooth', block: 'center' })
      }
    }, 100)
  }
  return (
    <>
      {isFormEmpty && readOnly ? (
        <>
          <Button
            color="primary"
            variant="outlined"
            block
            onClick={() => creatFirstRuleConfig()}
            className="mb-3"
          >
            新增标签规则
          </Button>
        </>
      ) : (
        <>
          <Form
            form={form}
            labelCol={{ span: 3, offset: 1 }}
            wrapperCol={{ span: 18 }}
            colon={false}
          >
            <LoadingSpinner loading={loading} />
            <Form.List name="enrichRuleConfigs">
              {(fields, { add, remove }) => (
                <div style={{ display: 'flex', rowGap: 16, flexDirection: 'column' }}>
                  {fields.map((field, index) => {
                    return (
                      <TagRuleFormListCom
                        field={field}
                        remove={remove}
                        readOnly={readOnly}
                        onEdit={editFormList}
                      />
                    )
                  })}
                  {readOnly || (
                    <Button
                      color="primary"
                      variant="outlined"
                      onClick={() => add()}
                      block
                      className="mb-3"
                    >
                      + 新增标签规则
                    </Button>
                  )}
                </div>
              )}
            </Form.List>
          </Form>
          <div className="w-full text-right mt-3">
            {readOnly ? (
              <>
                <Button type="primary" className="mr-2" onClick={() => setReadOnly(false)}>
                  编辑
                </Button>
              </>
            ) : (
              <>
                <Button type="primary" className="mr-2" onClick={saveForm}>
                  保存
                </Button>
                <Button onClick={refreshReadOnly}>取消</Button>
              </>
            )}
          </div>
        </>
      )}
    </>
  )
}
export default TagContactRule
