/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Form, Input, Modal, Typography } from 'antd'
import { useEffect, useState } from 'react'
import DataGroupPermission from './DataGroupPermission'
import { SubsDataGroupParams } from 'src/core/types/dataGroup'
import { getSubsDataGroupApi, updateSubsDataGroupApi } from 'src/core/api/dataGroup'
import { useTranslation } from 'react-i18next'
import { notify } from 'src/core/utils/notify'

interface PermissionModalProps {
  open: boolean
  closeModal: any
  subjectId: string
  subjectName: string
  refresh: any
  type: 'user' | 'team'
}
const DataGroupAuthorizeModal = ({
  open,
  closeModal,
  subjectId,
  subjectName,
  type,
  refresh,
}: PermissionModalProps) => {
  const { t } = useTranslation('core/permission')
  const { t: ct } = useTranslation('common')
  const [loading, setLoading] = useState(false)
  const [form] = Form.useForm()
  const [permissionSourceTeam, setPermissionSourceTeam] = useState([])
  const authorizePermission = (params: SubsDataGroupParams) => {
    updateSubsDataGroupApi(params).then((res) => {
      notify({
        type: 'success',
        message: t('authorizedSuccess'),
      })
      refresh()
    })
  }
  const getUsersPermission = () => {
    getSubsDataGroupApi({
      subjectId: subjectId,
      subjectType: type,
    }).then((res) => {
      const dataGroupPermission: any[] = []
      const teamDataGroup: any[] = []
      ;(res || []).map((item) => {
        if (type === 'user') {
          if (item.source === type) {
            dataGroupPermission.push(item.groupId)
          } else {
            teamDataGroup.push(item.groupId)
          }
        } else {
          dataGroupPermission.push(item.groupId)
        }
      })
      form.setFieldsValue({
        dataGroupPermission: dataGroupPermission,
      })
      setPermissionSourceTeam(teamDataGroup)
      // form.setFieldsValue({
      //   dataGroupPermission: (res || []).map((item) => ({
      //     groupId: item.groupId,
      //     groupName: item.groupName,
      //     source: item.source,
      //   })),
      // })
    })
  }
  const saveAuthorize = () => {
    form.validateFields().then((values) => {
      console.log(values)
      authorizePermission({
        subjectId: subjectId,
        subjectType: type,
        dataGroupPermission: values?.dataGroupPermission.map((item) => ({
          groupId: item,
          type: 'view',
        })),
      })
    })
  }
  useEffect(() => {
    if (open && subjectId) {
      getUsersPermission()
    } else {
      form.resetFields()
    }
  }, [open, subjectId])
  return (
    <>
      <Modal
        open={open}
        title={t('authorized')}
        onCancel={closeModal}
        destroyOnClose
        centered
        okText={ct('save')}
        cancelText={ct('cancel')}
        maskClosable={false}
        onOk={saveAuthorize}
        width={1000}
      >
        <Form form={form} colon={false} layout="vertical">
          <Typography.Title level={5}>
            {t(type === 'team' ? 'teamName' : 'username')}
          </Typography.Title>
          <Form.Item noStyle>
            <Input readOnly defaultValue={subjectName} variant="borderless"></Input>
          </Form.Item>
          <Typography.Title level={5} className="mt-2">
            {t('permissions')}
          </Typography.Title>
          <Typography className="p-2">
            <Form.Item
              name="dataGroupPermission"
              label={t('permissionsFromUser')}
              valuePropName="dataGroupList"
              noStyle={type === 'team'}
            >
              <DataGroupPermission type={type} />
            </Form.Item>
            {type === 'user' && (
              <Form.Item label={t('permissionsFromTeam')}>
                <DataGroupPermission
                  type={type}
                  readOnly={true}
                  dataGroupList={permissionSourceTeam}
                />
              </Form.Item>
            )}
          </Typography>
        </Form>
      </Modal>
    </>
  )
}
export default DataGroupAuthorizeModal
