/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState, useEffect } from 'react'
import { Table, Button, Modal, Form, Input, message, Space, Popconfirm } from 'antd'
import { PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import GlassCard from 'src/core/components/GlassCard'

const ServiceAliasConfig = () => {
  const { t } = useTranslation('core/config')
  const [data, setData] = useState([])
  const [loading, setLoading] = useState(false)
  const [modalVisible, setModalVisible] = useState(false)
  const [editingRecord, setEditingRecord] = useState(null)
  const [form] = Form.useForm()

  const mockData = [
    { id: 1, serviceName: 'user-service', businessAlias: t('serviceAlias.userService') },
    { id: 2, serviceName: 'order-service', businessAlias: t('serviceAlias.orderService') },
    { id: 3, serviceName: 'payment-service', businessAlias: t('serviceAlias.paymentService') },
  ]

  useEffect(() => {
    loadData()
  }, [])

  const loadData = () => {
    setLoading(true)
    // 模拟API调用
    setTimeout(() => {
      setData(mockData)
      setLoading(false)
    }, 500)
  }

  const handleAdd = () => {
    setEditingRecord(null)
    form.resetFields()
    setModalVisible(true)
  }

  const handleEdit = (record) => {
    setEditingRecord(record)
    form.setFieldsValue(record)
    setModalVisible(true)
  }

  const handleDelete = (id) => {
    setData(data.filter((item) => item.id !== id))
    message.success(t('serviceAlias.deleteSuccess'))
  }

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields()
      if (editingRecord) {
        setData(data.map((item) => (item.id === editingRecord.id ? { ...item, ...values } : item)))
        message.success(t('serviceAlias.updateSuccess'))
      } else {
        const newRecord = {
          id: Date.now(),
          ...values,
        }
        setData([...data, newRecord])
        message.success(t('serviceAlias.addSuccess'))
      }
      setModalVisible(false)
    } catch (error) {
      console.error('表单验证失败:', error)
    }
  }

  const columns = [
    {
      title: t('serviceAlias.serviceName'),
      dataIndex: 'serviceName',
      key: 'serviceName',
    },
    {
      title: t('serviceAlias.businessAlias'),
      dataIndex: 'businessAlias',
      key: 'businessAlias',
    },
    {
      title: t('serviceAlias.operation'),
      key: 'action',
      render: (_, record) => (
        <Space>
          <Button type="link" icon={<EditOutlined />} onClick={() => handleEdit(record)}>
            {t('serviceAlias.edit')}
          </Button>
          <Popconfirm
            title={t('serviceAlias.confirmDelete')}
            onConfirm={() => handleDelete(record.id)}
            okText={t('serviceAlias.confirm')}
            cancelText={t('serviceAlias.cancel')}
          >
            <Button type="link" danger icon={<DeleteOutlined />}>
              {t('serviceAlias.delete')}
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ]

  return (
    <div style={{ position: 'relative', height: 'calc(100vh - 120px)' }}>
      <GlassCard content={<p>{t('core/mask:comingSoon')}</p>} />
      <div style={{ marginBottom: 16 }}>
        <Button type="primary" icon={<PlusOutlined />} onClick={handleAdd}>
          {t('serviceAlias.add')}
        </Button>
      </div>

      <Table
        columns={columns}
        dataSource={data}
        loading={loading}
        rowKey="id"
        pagination={{
          showSizeChanger: true,
          showQuickJumper: true,
          showTotal: (total) => t('serviceAlias.totalItems', { total }),
        }}
      />

      <Modal
        title={editingRecord ? t('serviceAlias.edit') : t('serviceAlias.add')}
        open={modalVisible}
        onOk={handleSubmit}
        onCancel={() => setModalVisible(false)}
        okText={t('serviceAlias.confirm')}
        cancelText={t('serviceAlias.cancel')}
      >
        <Form form={form} layout="vertical">
          <Form.Item
            name="serviceName"
            label={t('serviceAlias.serviceName')}
            rules={[{ required: true, message: t('serviceAlias.pleaseInputServiceName') }]}
          >
            <Input placeholder={t('serviceAlias.pleaseInputServiceName')} />
          </Form.Item>
          <Form.Item
            name="businessAlias"
            label={t('serviceAlias.businessAlias')}
            rules={[{ required: true, message: t('serviceAlias.pleaseInputBusinessAlias') }]}
          >
            <Input placeholder={t('serviceAlias.pleaseInputBusinessAlias')} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default ServiceAliasConfig
