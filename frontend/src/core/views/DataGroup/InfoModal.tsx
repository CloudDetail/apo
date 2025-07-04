/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Form, Input, Modal } from 'antd'
import { useEffect, useState } from 'react'
import DatasourceSelector from './component/DatasourceSelector'
import { addDataGroupApi, updateDataGroupApiV2 } from 'src/core/api/dataGroup'
import { SaveDataGroupParams } from 'src/core/types/dataGroup'
import LoadingSpinner from 'src/core/components/Spinner'
import { useTranslation } from 'react-i18next'
import { notify } from 'src/core/utils/notify'

interface InfoModalProps {
  open: boolean
  closeModal: any
  groupInfo: SaveDataGroupParams | null
  groupId: string | null
  refresh: any
}
const InfoModal = ({ open, closeModal, groupInfo, refresh, groupId }: InfoModalProps) => {
  const { t } = useTranslation('core/dataGroup')

  const [form] = Form.useForm()
  const [loading, setLoading] = useState(false)
  const saveDataGroup = (params: SaveDataGroupParams) => {
    let api
    if (params.groupId) {
      api = updateDataGroupApiV2
    } else {
      api = addDataGroupApi
      params.parentGroupId = groupId
    }
    api(params)
      .then((res) => {
        notify({
          type: 'success',
          message: t('saveSuccess'),
        })
        refresh()
      })
      .finally(() => {
        setLoading(false)
      })
  }
  const saveInfo = () => {
    // setLoading(true)
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
        title={(groupInfo ? t('edit') : t('add')) + t('group')}
        onCancel={closeModal}
        destroyOnClose
        centered
        okText={t('save')}
        cancelText={t('cancel')}
        maskClosable={false}
        onOk={saveInfo}
        width={'80vw'}
        styles={{ body: { height: '80vh', overflowY: 'hidden', overflowX: 'hidden' } }}
      >
        <LoadingSpinner loading={loading} />

        <Form form={form} labelCol={{ span: 3, offset: 1 }} wrapperCol={{ span: 18 }} colon={false}>
          <Form.Item name="groupId" hidden>
            <Input></Input>
          </Form.Item>
          <Form.Item name="groupName" label={t('dataGroupName')} rules={[{ required: true }]}>
            <Input></Input>
          </Form.Item>
          <Form.Item name="description" label={t('dataGroupDes')}>
            <Input></Input>
          </Form.Item>
          <Form.Item name="datasources" label={t('datasource')} valuePropName="datasources">
            <DatasourceSelector groupId={groupInfo?.groupId || groupId} isAdd={!groupInfo} />
          </Form.Item>
        </Form>
      </Modal>
    </>
  )
}
export default InfoModal
