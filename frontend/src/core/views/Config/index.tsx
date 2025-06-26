import React, { useState, useEffect } from 'react'
import { Table, Button, Modal, Form, Input, message, Space, Popconfirm } from 'antd'
import { PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons'
import { useTranslation } from 'react-i18next'

const ServiceAliasConfig = () => {
  const { t } = useTranslation()
  const [data, setData] = useState([])
  const [loading, setLoading] = useState(false)
  const [modalVisible, setModalVisible] = useState(false)
  const [editingRecord, setEditingRecord] = useState(null)
  const [form] = Form.useForm()

  const mockData = [
    { id: 1, serviceName: 'user-service', businessAlias: '用户服务' },
    { id: 2, serviceName: 'order-service', businessAlias: '订单服务' },
    { id: 3, serviceName: 'payment-service', businessAlias: '支付服务' },
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
    message.success(t('删除成功'))
  }

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields()
      if (editingRecord) {
        // 更新
        setData(data.map((item) => (item.id === editingRecord.id ? { ...item, ...values } : item)))
        message.success(t('更新成功'))
      } else {
        // 新增
        const newRecord = {
          id: Date.now(),
          ...values,
        }
        setData([...data, newRecord])
        message.success(t('添加成功'))
      }
      setModalVisible(false)
    } catch (error) {
      console.error('表单验证失败:', error)
    }
  }

  const columns = [
    {
      title: t('服务名'),
      dataIndex: 'serviceName',
      key: 'serviceName',
    },
    {
      title: t('业务别名'),
      dataIndex: 'businessAlias',
      key: 'businessAlias',
    },
    {
      title: t('操作'),
      key: 'action',
      render: (_, record) => (
        <Space>
          <Button type="link" icon={<EditOutlined />} onClick={() => handleEdit(record)}>
            {t('编辑')}
          </Button>
          <Popconfirm
            title={t('确定删除吗？')}
            onConfirm={() => handleDelete(record.id)}
            okText={t('确定')}
            cancelText={t('取消')}
          >
            <Button type="link" danger icon={<DeleteOutlined />}>
              {t('删除')}
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ]

  return (
    <div>
      <div style={{ marginBottom: 16 }}>
        <Button type="primary" icon={<PlusOutlined />} onClick={handleAdd}>
          {t('添加服务别名')}
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
          showTotal: (total) => t('共 {{total}} 条', { total }),
        }}
      />

      <Modal
        title={editingRecord ? t('编辑服务别名') : t('添加服务别名')}
        open={modalVisible}
        onOk={handleSubmit}
        onCancel={() => setModalVisible(false)}
        okText={t('确定')}
        cancelText={t('取消')}
      >
        <Form form={form} layout="vertical">
          <Form.Item
            name="serviceName"
            label={t('服务名')}
            rules={[{ required: true, message: t('请输入服务名') }]}
          >
            <Input placeholder={t('请输入服务名')} />
          </Form.Item>
          <Form.Item
            name="businessAlias"
            label={t('业务别名')}
            rules={[{ required: true, message: t('请输入业务别名') }]}
          >
            <Input placeholder={t('请输入业务别名')} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default ServiceAliasConfig
