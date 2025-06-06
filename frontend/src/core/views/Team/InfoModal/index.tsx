/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Form, Input, Modal, Select } from 'antd'
import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { addTeamApi, updateTeamApi } from 'src/core/api/team'
import { getUserListApi } from 'src/core/api/user'
import LoadingSpinner from 'src/core/components/Spinner'
import { SaveTeamParams } from 'src/core/types/team'
import { notify } from 'src/core/utils/notify'
interface InfoModalProps {
  open: boolean
  closeModal: any
  teamInfo: any
  refresh: any
}
const InfoModal = ({ open, closeModal, teamInfo, refresh }: InfoModalProps) => {
  const { t } = useTranslation('core/team')
  const { t: ct } = useTranslation('common')

  const [form] = Form.useForm()
  const [userList, setUserList] = useState([])

  const [loading, setLoading] = useState(false)
  const getUserList = () => {
    getUserListApi({
      currentPage: 1,
      pageSize: 1000,
    }).then((res) => {
      setUserList(res?.users || [])
    })
  }
  const addTeam = (params: SaveTeamParams) => {
    addTeamApi(params)
      .then((res) => {
        notify({
          type: 'success',
          message: ct('addSuccess'),
        })
        refresh()
      })
      .finally(() => {
        setLoading(false)
      })
  }
  const updateTeam = (params: SaveTeamParams) => {
    updateTeamApi(params)
      .then((res) => {
        notify({
          type: 'success',
          message: ct('saveSuccess'),
        })
        refresh()
      })
      .finally(() => {
        setLoading(false)
      })
  }
  const saveTeam = () => {
    form.validateFields().then((values) => {
      if (values.teamId) {
        updateTeam(values)
      } else {
        addTeam(values)
      }
    })
  }
  useEffect(() => {
    if (open) {
      getUserList()
      //   form.setFieldValue('teamId', teamInfo.teamId)
    } else {
      form.resetFields()
    }
    if (teamInfo) {
      form.setFieldsValue(teamInfo)
      form.setFieldValue(
        'userList',
        teamInfo?.userList?.map((user) => user.userId),
      )
    }
  }, [open, teamInfo])
  return (
    <Modal
      open={open}
      title={(teamInfo ? ct('edit') : ct('add')) + ' ' + t('team')}
      onCancel={closeModal}
      destroyOnClose
      centered
      okText={ct('save')}
      cancelText={ct('cancel')}
      maskClosable={false}
      onOk={saveTeam}
      width={1000}
      styles={{ body: { height: '50vh', overflowY: 'hidden', overflowX: 'hidden' } }}
    >
      <LoadingSpinner loading={loading} />

      <Form form={form} labelCol={{ span: 4, offset: 1 }} wrapperCol={{ span: 15 }} colon={false}>
        <Form.Item name="teamId" hidden>
          <Input></Input>
        </Form.Item>
        <Form.Item name="teamName" label={t('teamName')} rules={[{ required: true }]}>
          <Input></Input>
        </Form.Item>
        <Form.Item name="description" label={t('description')}>
          <Input></Input>
        </Form.Item>
        <Form.Item name="userList" label={t('userList')}>
          <Select
            mode="multiple"
            showSearch
            allowClear
            options={userList}
            style={{ width: '100%' }}
            optionFilterProp={'username'}
            fieldNames={{ label: 'username', value: 'userId' }}
          />
        </Form.Item>
      </Form>
    </Modal>
  )
}
export default InfoModal
