/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Form, Input, Modal, Select } from 'antd'
import { useEffect, useState } from 'react'
import { getDataGroupPermissionSubsApi, updateDataGroupSubsApi } from 'src/core/api/dataGroup'
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
  const [loading, setLoading] = useState(false)
  const [form] = Form.useForm()
  const [userList, setUserList] = useState([])

  const updateDataGroupSubs = (params: DataGroupSubsParams) => {
    updateDataGroupSubsApi(params).then((res) => {
      showToast({
        color: 'success',
        title: '数据组授权成功',
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
  const getDataGroupPermissionSubs = () => {
    getDataGroupPermissionSubsApi(groupInfo?.groupId).then((res) => {
      form.setFieldsValue({
        userList: res?.map((user) => user.userId),
      })
    })
  }
  useEffect(() => {
    if (open && groupInfo) {
      getUserList()
      getDataGroupPermissionSubs()
      form.setFieldValue('groupId', groupInfo.groupId)
    }
  }, [open])
  return (
    <Modal
      open={open}
      title={'数据组授权'}
      onCancel={closeModal}
      destroyOnClose
      centered
      okText={'保存'}
      cancelText={'取消'}
      maskClosable={false}
      onOk={savePermission}
      width={1000}
    >
      <LoadingSpinner loading={loading} />

      <Form form={form} labelCol={{ span: 3, offset: 1 }} wrapperCol={{ span: 18 }} colon={false}>
        <Form.Item name="groupId" hidden>
          <Input></Input>
        </Form.Item>
        <Form.Item label="数据组名">
          <Input readOnly defaultValue={groupInfo?.groupName} variant="borderless"></Input>
        </Form.Item>
        <Form.Item
          name="userList"
          label="授权用户"
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
      </Form>
    </Modal>
  )
}
export default PermissionModal
