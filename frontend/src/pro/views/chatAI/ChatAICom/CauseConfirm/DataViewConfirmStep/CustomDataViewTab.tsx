import { CAccordion, CAccordionHeader, CAccordionItem } from '@coreui/react'
import { Button, Form, Input, Transfer, TransferProps } from 'antd'
import { useEffect, useState } from 'react'
import { useChatAIContext } from 'src/pro/contexts/ChatAIContext'
import ComAccordionTransfer from './ComAccordionTransfer'
import { creatCustomDataView, updateCustomDataView } from 'src/pro/api/custom'
import { showToast } from 'src/core/utils/toast'
import { useTranslation } from 'react-i18next'
const i18nKey = 'causeConfirm'

const CustomDataViewTab = ({ defaultValue }: any) => {
  const { t } = useTranslation('pro/chatAI')
  const analysisNode = useChatAIContext((ctx) => ctx.analysisNode)
  const [form] = Form.useForm()
  const addNew = (params) => {
    creatCustomDataView(params)
      .then((res) => {
        analysisNode(params.type)
      })
      .catch((error) => {
        console.log(error)
      })
  }
  const updateCustom = (params) => {
    updateCustomDataView(params)
      .then((res) => {
        analysisNode(params.type)
      })
      .catch((error) => {
        console.log(error)
      })
  }
  const saveForm = () => {
    form
      .validateFields()
      .then((formState) => {
        if (defaultValue) {
          updateCustom(formState)
        } else {
          addNew(formState)
        }
      })
      .catch((error) => {
        console.log(error)
      })
  }
  useEffect(() => {
    if (defaultValue) {
      form.setFieldValue('customName', defaultValue.name)
      form.setFieldValue('type', defaultValue.type)
    }
  }, [defaultValue])
  return (
    <>
      <Form form={form} labelCol={{ span: 3 }}>
        <Form.Item
          name="customName"
          label={t(`${i18nKey}.dataTypeName`)}
          required
          rules={[
            {
              validator: async (_, value) => {
                if (!value) {
                  return Promise.reject(t(`${i18nKey}.dataTypeNameRequired`))
                }
              },
            },
          ]}
        >
          <Input />
        </Form.Item>
        <Form.Item
          name="type"
          label={t(`${i18nKey}.displayData`)}
          required
          rules={[
            {
              validator: async (_, value) => {
                if (!value || value?.length === 0) {
                  return Promise.reject(t(`${i18nKey}.displayDataRequired`))
                }
              },
            },
          ]}
        >
          <ComAccordionTransfer />
        </Form.Item>
      </Form>
      <div className="text-right">
        <Button type="primary" onClick={saveForm}>
          {t(`${i18nKey}.confirmRootCauseAndViewData`)}
        </Button>
      </div>
    </>
  )
}
export default CustomDataViewTab
