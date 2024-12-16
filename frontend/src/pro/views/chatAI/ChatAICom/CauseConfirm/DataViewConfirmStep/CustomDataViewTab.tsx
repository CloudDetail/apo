import { CAccordion, CAccordionHeader, CAccordionItem } from '@coreui/react'
import { Button, Form, Input, Transfer, TransferProps } from 'antd'
import { useEffect, useState } from 'react'
import { useChatAIContext } from 'src/pro/contexts/ChatAIContext'
import ComAccordionTransfer from './ComAccordionTransfer'
import { creatCustomDataView, updateCustomDataView } from 'src/pro/api/custom'
import { showToast } from 'src/core/utils/toast'

const CustomDataViewTab = ({ defaultValue }: any) => {
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
      <Form form={form} labelCol={{ span: 3 }} wrapperCol={{ span: 18 }}>
        <Form.Item
          name="customName"
          label="数据类型名"
          required
          rules={[
            {
              validator: async (_, value) => {
                if (!value) {
                  return Promise.reject('数据类型名不可为空')
                }
              },
            },
          ]}
        >
          <Input></Input>
        </Form.Item>
        <Form.Item
          name="type"
          label="展示数据"
          required
          rules={[
            {
              validator: async (_, value) => {
                if (!value || value?.length === 0) {
                  return Promise.reject('展示数据不可为空')
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
          确认根因并查看数据
        </Button>
      </div>
    </>
  )
}
export default CustomDataViewTab
