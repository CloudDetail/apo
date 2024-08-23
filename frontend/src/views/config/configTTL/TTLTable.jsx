import { Form, Input, InputNumber, Table, Typography } from 'antd'
import React, { useState } from 'react'
import { setSingleTableTTLApi } from 'src/api/config'
import { showToast } from 'src/utils/toast'

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
  const inputNode =
    inputType === 'number' ? (
      <InputNumber
        min={1}
        addonAfter="天"
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
              message: `${title}必须为大于0 的数字`,
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
  const [form] = Form.useForm()
  const { list = [], refreshPage } = props
  const [editingKey, setEditingKey] = useState('')
  const isEditing = (record) => record.name === editingKey
  const confirmSingleTableTTL = () => {
    setSingleTableTTLApi({ name: editingKey, day: form.getFieldValue('originalDays') })
      .then((res) => {
        showToast({
          title: '配置数据保留周期可能需一定时间生效，请稍后刷新页面查看结果',
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
      title: '表名',
      dataIndex: 'name',
      width: '50%',
      editable: false,
    },
    {
      title: '数据保留（天）',
      dataIndex: 'originalDays',
      width: '30%',
      editable: true,
    },
    {
      title: '操作',
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
              更新
            </Typography.Link>
            <Typography.Link
              onClick={() => setEditingKey('')}
              style={{
                marginInlineEnd: 8,
              }}
            >
              取消
            </Typography.Link>
          </span>
        ) : (
          <Typography.Link
            disabled={editingKey !== '' || !record.originalDays}
            onClick={() => edit(record)}
          >
            编辑
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
