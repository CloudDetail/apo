/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Form, Input, Modal } from 'antd'
import { useEffect, useState } from 'react'
import DatasourceSelector from './DatasourceSelector'
import { creatDataGroupApi, updateDataGroupApi } from 'src/core/api/dataGroup'
import { SaveDataGroupParams } from 'src/core/types/dataGroup'
import { showToast } from 'src/core/utils/toast'
import LoadingSpinner from 'src/core/components/Spinner'

interface InfoModalProps {
  open: boolean
  closeModal: any
  groupInfo: SaveDataGroupParams | null
  refresh: any
}
const InfoModal = ({ open, closeModal, groupInfo, refresh }: InfoModalProps) => {
  const [form] = Form.useForm()
  const [loading, setLoading] = useState(false)
  const saveDataGroup = (params: SaveDataGroupParams) => {
    let api
    if (params.groupId) {
      api = updateDataGroupApi
    } else {
      api = creatDataGroupApi
    }
    api(params)
      .then((res) => {
        showToast({
          color: 'success',
          title: '保存数据组成功',
        })
        refresh()
      })
      .finally(() => {
        setLoading(false)
      })
  }
  const saveInfo = () => {
    setLoading(true)
    form
      .validateFields()
      .then((values) => {
        saveDataGroup(values)
      })
      .catch(() => {
        setLoading(false)
      })
  }
  useEffect(() => {
    if (groupInfo) {
      form.setFieldsValue(groupInfo)
    } else {
      form.resetFields()
    }
  }, [open, groupInfo])
  return (
    <>
      <Modal
        open={open}
        title={groupInfo ? '编辑数据组' : '新建数据组'}
        onCancel={closeModal}
        destroyOnClose
        centered
        okText={'保存'}
        cancelText={'取消'}
        maskClosable={false}
        onOk={saveInfo}
        width={1000}
        styles={{ body: { height: '80vh', overflowY: 'hidden', overflowX: 'hidden' } }}
      >
        <LoadingSpinner loading={loading} />

        <Form form={form} labelCol={{ span: 3, offset: 1 }} wrapperCol={{ span: 18 }} colon={false}>
          <Form.Item name="groupId" hidden>
            <Input></Input>
          </Form.Item>
          <Form.Item name="groupName" label="数据组名" rules={[{ required: true }]}>
            <Input></Input>
          </Form.Item>
          <Form.Item name="description" label="数据组描述">
            <Input></Input>
          </Form.Item>
          <Form.Item
            name="datasourceList"
            label="数据源"
            // rules={[{ required: true, message: '请选择至少一个数据源' }]}
            valuePropName="datasourceList"
          >
            <DatasourceSelector />
          </Form.Item>
        </Form>
      </Modal>
    </>
  )
}
export default InfoModal
