/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Button, Flex, Popconfirm, Table } from 'antd'
import { getSubGroupsApiV2 } from 'src/core/api/dataGroup'
import { MdOutlineAdd, MdOutlineEdit } from 'react-icons/md'
import { RiDeleteBin5Line } from 'react-icons/ri'
import { LuShieldCheck } from 'react-icons/lu'
import DatasourceTag from './component/DatasourceTag'
import Paragraph from 'antd/es/typography/Paragraph'
import { useTranslation } from 'react-i18next'
import { useEffect, useState } from 'react'
import React from 'react'

export default function DataGroupTable({
  parentGroupInfo,
  openAddModal,
  openEditModal,
  openPermissionModal,
  deleteDataGroup,
  key,
}) {
  const { t } = useTranslation('core/dataGroup')
  const { t: ct } = useTranslation('common')
  const [subGroups, setSubGroups] = useState([])

  const columns = [
    {
      title: 'groupId',
      dataIndex: 'groupId',
      key: 'groupId',
      hidden: true,
    },
    {
      title: t('dataGroupName'),
      dataIndex: 'groupName',
      width: 150,
      key: 'groupName',
    },
    {
      title: t('dataGroupDes'),
      width: 200,
      dataIndex: 'description',
      key: 'description',
    },
    {
      title: t('datasource'),
      dataIndex: 'datasources',
      key: 'datasources',
      render: (value) => {
        return (
          <Paragraph
            className="m-0 items-center flex flex-wrap"
            ellipsis={{
              expandable: true,
              rows: 3,
            }}
          >
            {value
              ?.sort((a, b) => {
                const typeOrder = ['system', 'cluster', 'namespace', 'service']
                const aIndex = typeOrder.indexOf(a.type)
                const bIndex = typeOrder.indexOf(b.type)
                return aIndex - bIndex
              })
              ?.map((item) => <DatasourceTag {...item} block={false} />)}
          </Paragraph>
        )
      },
    },
    {
      title: ct('operation'),
      dataIndex: 'operation',
      key: 'operation',
      width: 250,
      render: (_, record) => {
        return (
          <Flex align="center" justify="space-evenly">
            {record?.permissionType === 'edit' && (
              <>
                <Button
                  type="text"
                  size="small"
                  onClick={() => {
                    openEditModal(record)
                  }}
                  icon={
                    <MdOutlineEdit className="!text-[var(--ant-color-primary-text)] !hover:text-[var(--ant-color-primary-text-active)]" />
                  }
                >
                  <span className="text-[var(--ant-color-primary-text)] hover:text-[var(--ant-color-primary-text-active)]">
                    {t('edit')}
                  </span>
                </Button>
                <Popconfirm
                  title={t('confirmDelete', {
                    groupName: record.groupName,
                  })}
                  onConfirm={() => deleteDataGroup(record)}
                  okText={ct('confirm')}
                  cancelText={ct('cancel')}
                >
                  <Button type="text" size="small" icon={<RiDeleteBin5Line />} danger>
                    {ct('delete')}
                  </Button>
                </Popconfirm>
                <Button
                  color="primary"
                  variant="outlined"
                  size="small"
                  icon={<LuShieldCheck />}
                  onClick={() => {
                    openPermissionModal(record)
                  }}
                >
                  {t('authorize')}
                </Button>
              </>
            )}
          </Flex>
        )
      },
    },
  ]
  useEffect(() => {
    if (parentGroupInfo?.groupId != null) {
      getSubGroupsApiV2(parentGroupInfo.groupId).then((res) => {
        setSubGroups(res.subGroups)
      })
    }
  }, [parentGroupInfo, key])
  return (
    <>
      <div className="w-full flex justify-between h-[40px]">
        <div className="text-lg font-bold ml-2">{parentGroupInfo?.groupName}</div>
        <div>
          {parentGroupInfo?.permissionType !== 'known' && (
            <Button type="primary" icon={<MdOutlineAdd />} onClick={openAddModal} className="mr-2">
              {t('add')}
            </Button>
          )}
          {parentGroupInfo?.permissionType === 'edit' && (
            <Button
              color="primary"
              variant="outlined"
              icon={<LuShieldCheck />}
              onClick={() => {
                openPermissionModal(parentGroupInfo)
              }}
            >
              {t('authorize')}
            </Button>
          )}
        </div>
      </div>
      <Table
        dataSource={subGroups}
        columns={columns}
        scroll={{ y: 'calc(100vh - 240px)' }}
        className="overflow-auto text-xs"
        size="small"
      />
    </>
  )
}
