/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Alert, Form, Input, Modal, Select, Tag, Typography } from 'antd'
import { useEffect, useState } from 'react'
import DataGroupPermission from './DaraGroupPermission'
import { SubsDataGroupParams } from 'src/core/types/dataGroup'
import { getSubsDataGroupApi, updateSubsDataGroupApi } from 'src/core/api/dataGroup'
import { showToast } from 'src/core/utils/toast'
import { useTranslation } from 'react-i18next'

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
      showToast({
        color: 'success',
        title: t('authorizedSuccess'),
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
            dataGroupPermission.push({
              groupId: item.groupId,
              groupName: item.groupName,
            })
          } else {
            teamDataGroup.push({
              groupId: item.groupId,
              groupName: item.groupName,
            })
          }
        } else {
          dataGroupPermission.push({
            groupId: item.groupId,
            groupName: item.groupName,
          })
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
      authorizePermission({
        subjectId: subjectId,
        subjectType: type,
        dataGroupPermission: values?.dataGroupPermission.map((item) => ({
          groupId: item.groupId,
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
              <DataGroupPermission type={type} permissionSourceTeam={permissionSourceTeam} />
            </Form.Item>
            {type === 'user' && (
              <Form.Item label={t('permissionsFromTeam')} valuePropName="dataGroupList">
                {permissionSourceTeam?.map((item) => <Tag>{item.groupName}</Tag>)}
              </Form.Item>
            )}
          </Typography>
        </Form>
      </Modal>
    </>
  )
}
export default DataGroupAuthorizeModal
