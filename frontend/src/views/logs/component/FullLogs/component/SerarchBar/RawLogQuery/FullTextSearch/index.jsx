import { Button, Col, Form, Input, Popover, Row } from 'antd'
import React, { useEffect, useState } from 'react'

const FullTextSearch = ({ searchValue, setSearchValue }) => {
  const [form] = Form.useForm()
  const [open, setOpen] = useState(false)

  const hide = () => {
    setOpen(false)
  }

  const handleOpenChange = (newOpen) => {
    setOpen(newOpen)
  }

  const clickSubmit = () => {
    form.validateFields().then(() => {
      const formState = form.getFieldsValue(true)
      let newQuery = searchValue
      if (newQuery.length > 0) {
        newQuery += ' AND '
      }
      newQuery += '`' + formState.key + '` LIKE ' + `'%` + formState.value + `%'`
      setSearchValue(newQuery)
      hide()
    })
  }
  useEffect(() => {
    form.resetFields()
  }, [open])
  return (
    <Popover
      destroyTooltipOnHide
      content={
        <div>
          <Form layout="vertical" className="px-2" form={form} initialValues={{ key: 'content' }}>
            <Row gutter={10}>
              <Col span={7}>
                <Form.Item
                  label="全文检索字段"
                  name="key"
                  rules={[
                    {
                      required: true,
                      message: '请输入全文检索字段',
                    },
                  ]}
                >
                  <Input placeholder="请输入全文检索字段" />
                </Form.Item>
              </Col>
              <Col span={13}>
                <Form.Item
                  label="全文检索内容"
                  name="value"
                  rules={[
                    {
                      required: true,
                      message: '请输入全文检索内容',
                    },
                  ]}
                >
                  <Input placeholder="请输入全文检索内容" />
                </Form.Item>
              </Col>
              <Col>
                <Form.Item label="  ">
                  <Button type="primary" htmlType="submit" onClick={clickSubmit}>
                    确认
                  </Button>
                </Form.Item>
              </Col>
            </Row>
          </Form>
        </div>
      }
      title="全文检索"
      trigger="click"
      open={open}
      onOpenChange={handleOpenChange}
    >
      <Button type="primary">全文检索</Button>
    </Popover>
  )
}
export default FullTextSearch
