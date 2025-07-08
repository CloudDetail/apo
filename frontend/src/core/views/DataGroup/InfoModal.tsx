/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Form, Input, Modal } from 'antd'
import React, { useEffect, useState } from 'react'
import DatasourceSelector from './component/DatasourceSelector'
import { addDataGroupApi, updateDataGroupApiV2 } from 'src/core/api/dataGroup'
import { SaveDataGroupParams } from 'src/core/types/dataGroup'
import LoadingSpinner from 'src/core/components/Spinner'
import { useTranslation } from 'react-i18next'
import { notify } from 'src/core/utils/notify'

interface InfoModalProps {
  open: boolean
  closeModal: () => void
  groupInfo: SaveDataGroupParams | null
  groupId: number | null
  refresh: () => void
}

const InfoModal: React.FC<InfoModalProps> = ({ open, closeModal, groupInfo, refresh, groupId }) => {
  const { t } = useTranslation('core/dataGroup')

  const [form] = Form.useForm()
  const [loading, setLoading] = useState<boolean>(false)
  console.log('parent', groupId)
  const saveDataGroup = (params: SaveDataGroupParams) => {
    let api
    let apiParams: any = params
    if (params.groupId) {
      api = updateDataGroupApiV2
    } else {
      if (groupId === null) {
        notify({
          type: 'error',
          message: t('addGroupError'),
        })
        setLoading(false)
        return
      }
      api = addDataGroupApi
      apiParams = { ...params, parentGroupId: groupId }
    }
    api(apiParams)
      .then(() => {
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
  }, [open, groupInfo, form])

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

        <Form form={form} labelCol={{ span: 4, offset: 0 }} wrapperCol={{ span: 20 }} colon={false}>
          <Form.Item name="groupId" hidden>
            <Input />
          </Form.Item>
          <Form.Item name="groupName" label={t('dataGroupName')} rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="description" label={t('dataGroupDes')}>
            <Input />
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
