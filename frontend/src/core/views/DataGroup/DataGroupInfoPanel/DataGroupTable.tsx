/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Button, Flex, Popconfirm, Table } from 'antd'
import { MdOutlineEdit } from 'react-icons/md'
import { RiDeleteBin5Line } from 'react-icons/ri'
import { LuShieldCheck } from 'react-icons/lu'
import DatasourceTag from '../component/DatasourceTag'
import Paragraph from 'antd/es/typography/Paragraph'
import { useTranslation } from 'react-i18next'
import React, { useCallback, useMemo } from 'react'
import { DataGroupPermissionInfo } from 'src/core/types/dataGroup'

interface DataGroupTableProps {
  openEditModal: (record: DataGroupPermissionInfo) => void
  deleteDataGroup: (record: DataGroupPermissionInfo) => void
  openPermissionModal: (record: DataGroupPermissionInfo) => void
  subGroups: any[]
  scrolllHeight: number | string
  parentGroupName: string
}

export default function DataGroupTable({
  openEditModal,
  deleteDataGroup,
  openPermissionModal,
  subGroups,
  scrolllHeight,
  parentGroupName,
}: DataGroupTableProps) {
  const { t } = useTranslation('core/dataGroup')
  const { t: ct } = useTranslation('common')

  const handleEdit = useCallback(
    (record: DataGroupPermissionInfo) => {
      openEditModal(record)
    },
    [openEditModal],
  )
  const handlePermission = useCallback(
    (record: DataGroupPermissionInfo) => {
      openPermissionModal(record)
    },
    [openPermissionModal],
  )

  const handleDelete = useCallback(
    (record: DataGroupPermissionInfo) => {
      deleteDataGroup(record)
    },
    [deleteDataGroup],
  )

  const columns = useMemo(
    () => [
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
        render: (value: any[]) => {
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
                ?.map((item) => <DatasourceTag key={item.id} {...item} block={false} />)}
            </Paragraph>
          )
        },
      },
      {
        title: ct('operation'),
        dataIndex: 'operation',
        key: 'operation',
        width: 250,
        render: (_: any, record: DataGroupPermissionInfo) => {
          return (
            <Flex align="center" justify="space-evenly">
              {record?.permissionType === 'edit' && (
                <>
                  <Button
                    type="text"
                    size="small"
                    onClick={() => handleEdit(record)}
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
                    onConfirm={() => handleDelete(record)}
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
                    onClick={() => handlePermission(record)}
                  >
                    {t('authorize')}
                  </Button>
                </>
              )}
            </Flex>
          )
        },
      },
    ],
    [t, ct, handleEdit, handleDelete, handlePermission],
  )

  return (
    <>
      <Table
        title={() => (
          <span className="text-sm font-bold">
            {t('subGroups', { groupName: parentGroupName })}
          </span>
        )}
        dataSource={subGroups}
        columns={columns}
        scroll={{ y: scrolllHeight }}
        className="overflow-auto text-xs h-full"
        size="small"
        rowKey="groupId"
      />
    </>
  )
}
