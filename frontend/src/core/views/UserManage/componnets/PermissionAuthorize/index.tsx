import { Form, Input, Modal, Select } from 'antd'
import { useEffect, useState } from 'react'
import DataGroupPermission from './DaraGroupPermission'
import { SubsDataGroupParams } from 'src/core/types/dataGroup'
import { getSubsDataGroupApi, updateSubsDataGroupApi } from 'src/core/api/dataGroup'
import { showToast } from 'src/core/utils/toast'

interface PermissionModalProps {
  open: boolean
  closeModal: any
  userInfo: any
  refresh: any
}
const DataGroupAuthorizeModal = ({ open, closeModal, userInfo, refresh }: PermissionModalProps) => {
  const [loading, setLoading] = useState(false)
  const [form] = Form.useForm()
  const authorizePermission = (params: SubsDataGroupParams) => {
    updateSubsDataGroupApi(params).then((res) => {
      showToast({
        color: 'success',
        title: '授权数据组成功',
      })
      refresh()
    })
  }
  const getUsersPermission = () => {
    getSubsDataGroupApi({
      subjectId: userInfo.userId,
      subjectType: 'user',
    }).then((res) => {
      form.setFieldsValue({
        dataGroupPermission: (res || []).map((item) => ({
          groupId: item.groupId,
          groupName: item.groupName,
        })),
      })
    })
  }
  const saveAuthorize = () => {
    form.validateFields().then((values) => {
      authorizePermission({
        subjectId: userInfo.userId,
        subjectType: 'user',
        dataGroupPermission: values.dataGroupPermission.map((item) => ({
          groupId: item.groupId,
          type: 'view',
        })),
      })
    })
  }
  useEffect(() => {
    if (open) {
      getUsersPermission()
    } else {
      form.resetFields()
    }
  }, [open, userInfo])
  return (
    <>
      <Modal
        open={open}
        title={'授权数据组'}
        onCancel={closeModal}
        destroyOnClose
        centered
        okText={'保存'}
        cancelText={'取消'}
        maskClosable={false}
        onOk={saveAuthorize}
        width={1000}
      >
        <Form form={form} labelCol={{ span: 3, offset: 1 }} wrapperCol={{ span: 18 }} colon={false}>
          <Form.Item label="用户名">
            <Input readOnly defaultValue={userInfo?.username} variant="borderless"></Input>
          </Form.Item>
          <Form.Item name="dataGroupPermission" label="数据组权限" valuePropName="dataGroupList">
            <DataGroupPermission />
          </Form.Item>
        </Form>
      </Modal>
    </>
  )
}
export default DataGroupAuthorizeModal
