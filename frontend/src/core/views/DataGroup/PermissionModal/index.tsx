/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Form, Input, Modal, Select } from 'antd'
import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { getDataGroupPermissionSubsApi, updateDataGroupSubsApi } from 'src/core/api/dataGroup'
import { getTeamsApi } from 'src/core/api/team'
import { getUserListApi } from 'src/core/api/user'
import LoadingSpinner from 'src/core/components/Spinner'
import { DataGroupSubsParams, SaveDataGroupParams } from 'src/core/types/dataGroup'
import { showToast } from 'src/core/utils/toast'

interface PermissionModalProps {
  open: boolean
  closeModal: any
  groupInfo: SaveDataGroupParams | null
  refresh: any
}
const PermissionModal = ({ open, closeModal, groupInfo, refresh }: PermissionModalProps) => {
  const { t } = useTranslation('core/dataGroup')
  const { t: ct } = useTranslation('common')
  const [loading, setLoading] = useState(false)
  const [form] = Form.useForm()
  const [userList, setUserList] = useState([])
  const [teamList, setTeamList] = useState([])

  const updateDataGroupSubs = (params: DataGroupSubsParams) => {
    updateDataGroupSubsApi(params).then((res) => {
      showToast({
        color: 'success',
        title: t('savePermissionSuccess'),
      })
      refresh()
    })
  }
  const savePermission = () => {
    form.validateFields().then((values) => {
      updateDataGroupSubs({
        groupId: values.groupId,
        userList: values.userList.map((user) => ({
          subjectId: user,
          type: 'view',
        })),
        teamList: values.teamList.map((team) => ({
          subjectId: team,
          type: 'view',
        })),
      })
    })
  }
  const getUserList = () => {
    getUserListApi({
      currentPage: 1,
      pageSize: 1000,
    }).then((res) => {
      setUserList(res?.users || [])
    })
  }
  const getTeamList = () => {
    getTeamsApi({
      currentPage: 1,
      pageSize: 1000,
    }).then((res) => {
      setTeamList(res?.teamList || [])
    })
  }
  const getDataGroupPermissionSubs = () => {
    getDataGroupPermissionSubsApi(groupInfo?.groupId).then((res) => {
      form.setFieldsValue({
        userList: res?.filter((user) => user.userId).map((user) => user.userId),
        teamList: res?.filter((team) => team.teamId).map((team) => team.teamId),
      })
    })
  }
  useEffect(() => {
    if (open && groupInfo) {
      getUserList()
      getTeamList()
      getDataGroupPermissionSubs()
      form.setFieldValue('groupId', groupInfo.groupId)
    }
  }, [open])
  return (
    <Modal
      open={open}
      title={t('dataGroupAuthorize')}
      onCancel={closeModal}
      destroyOnClose
      centered
      okText={t('save')}
      cancelText={t('cancel')}
      maskClosable={false}
      onOk={savePermission}
      width={1000}
    >
      <LoadingSpinner loading={loading} />

      <Form form={form} labelCol={{ span: 4, offset: 1 }} wrapperCol={{ span: 18 }} colon={false}>
        <Form.Item name="groupId" hidden>
          <Input></Input>
        </Form.Item>
        <Form.Item label={t('dataGroupName')}>
          <Input readOnly defaultValue={groupInfo?.groupName} variant="borderless"></Input>
        </Form.Item>
        <Form.Item
          name="userList"
          label={t('authorizeToUser')}
          //   normalize={(value) => {
          //     if (Array.isArray(value)) {
          //       return value.map((option) => ({
          //         userId: option.value,
          //         username: option.label,
          //       }))
          //     }
          //     return []
          //   }}
          //   getValueProps={(value) => {
          //     if (Array.isArray(value)) {
          //       return {
          //         value: value.map((option) => ({
          //           value: option.userId,
          //           label: option.username,
          //         })),
          //       }
          //     }
          //     return { value: [] }
          //   }}
        >
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
        <Form.Item name="teamList" label={t('authorizeToTeam')}>
          <Select
            mode="multiple"
            showSearch
            allowClear
            options={teamList}
            style={{ width: '100%' }}
            optionFilterProp={'teamName'}
            fieldNames={{ label: 'teamName', value: 'teamId' }}
          />
        </Form.Item>
      </Form>
    </Modal>
  )
}
export default PermissionModal
