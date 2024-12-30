/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Form, Input, InputNumber, Table, Typography } from 'antd'
import React, { useState } from 'react'
import { setSingleTableTTLApi } from 'core/api/config'
import { showToast } from 'src/core/utils/toast'
import { useTranslation } from 'react-i18next'

const EditableCell = ({
  editing,
  dataIndex,
  title,
  inputType,
  record,
  index,
  children,
  ...restProps
}) => {
  const { t } = useTranslation('oss/config')
  const inputNode =
    inputType === 'number' ? (
      <InputNumber
        min={1}
        addonAfter={t('configTTL.days')}
        controls={false}
        className="w-28"
        changeOnBlur={true}
        precision={0}
      />
    ) : (
      <Input />
    )
  return (
    <td {...restProps}>
      {editing ? (
        <Form.Item
          name={dataIndex}
          style={{
            margin: 0,
          }}
          rules={[
            {
              required: true,
              message: t('TTLTable.validationMessage', { title }),
            },
          ]}
        >
          {inputNode}
        </Form.Item>
      ) : (
        children
      )}
    </td>
  )
}

export default function TTLTable(props) {
  const { t } = useTranslation('oss/config')
  const [form] = Form.useForm()
  const { list = [], refreshPage } = props
  const [editingKey, setEditingKey] = useState('')
  const isEditing = (record) => record.name === editingKey
  const confirmSingleTableTTL = () => {
    setSingleTableTTLApi({ name: editingKey, day: form.getFieldValue('originalDays') })
      .then((res) => {
        showToast({
          title: t('TTLTable.updateInfo'),
          color: 'info',
        })
      })
      .finally(() => {
        setEditingKey('')
        refreshPage()
      })
  }
  const edit = (record) => {
    console.log(record)
    form.setFieldsValue({
      name: record.name,
      originalDays: record.originalDays,
    })
    setEditingKey(record.name)
  }
  const columns = [
    {
      title: t('TTLTable.tableName'),
      dataIndex: 'name',
      width: '50%',
      editable: false,
    },
    {
      title: t('TTLTable.dataRetention'),
      dataIndex: 'originalDays',
      width: '30%',
      editable: true,
    },
    {
      title: t('TTLTable.operation'),
      dataIndex: 'operation',
      render: (_, record) => {
        const editable = isEditing(record)
        console.log(record)
        return editable ? (
          <span>
            <Typography.Link
              onClick={confirmSingleTableTTL}
              style={{
                marginInlineEnd: 8,
              }}
            >
              {t('TTLTable.update')}
            </Typography.Link>
            <Typography.Link
              onClick={() => setEditingKey('')}
              style={{
                marginInlineEnd: 8,
              }}
            >
              {t('TTLTable.cancel')}
            </Typography.Link>
          </span>
        ) : (
          <Typography.Link
            disabled={editingKey !== '' || !record.originalDays}
            onClick={() => edit(record)}
          >
            {t('TTLTable.edit')}
          </Typography.Link>
        )
      },
    },
  ]
  const mergedColumns = columns.map((col) => {
    if (!col.editable) {
      return col
    }
    return {
      ...col,
      onCell: (record) => ({
        record,
        inputType: 'number',
        dataIndex: col.dataIndex,
        title: col.title,
        editing: isEditing(record),
      }),
    }
  })
  return (
    <>
      <Form form={form} component={false}>
        <Table
          size="middle"
          components={{
            body: {
              cell: EditableCell,
            },
          }}
          bordered
          dataSource={list}
          columns={mergedColumns}
          rowClassName="editable-row"
          rowKey="name"
          pagination={false}
        />
      </Form>
    </>
  )
}
