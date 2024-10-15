import { Col, Form, Input, Row, Select } from 'antd'
import React from 'react'
import { IoIosRemoveCircleOutline, IoMdAddCircleOutline } from 'react-icons/io'
const routeKeyList = [
  { label: '_container_id_', value: '_container_id_' },
  { label: '_source_', value: '_source_' },
  { label: 'container.image.name', value: 'container.image.name' },
  { label: 'container.ip', value: 'container.ip' },
  { label: 'container.name', value: 'container.name' },
  { label: 'content', value: 'content' },
  { label: 'host.ip', value: 'host.ip' },
  { label: 'host.name', value: 'host.name' },
  { label: 'k8s.namespace.name', value: 'k8s.namespace.name' },
  { label: 'k8s.node.ip', value: 'k8s.node.ip' },
  { label: 'k8s.node.name', value: 'k8s.node.name' },
  { label: 'k8s.pod.name', value: 'k8s.pod.name' },
  { label: 'k8s.pod.uid', value: 'k8s.pod.uid' },
]

function isValidKey(str) {
  // 定义正则表达式，确保开头是字母或下划线
  const regex = /^[a-zA-Z_].*$/

  return regex.test(str)
}
export default function LogRouteRuleFormList() {
  return (
    <Form.List name={'routeRule'}>
      {(fields, { add, remove }, { errors }) => (
        <>
          <Form.Item
            required
            label={
              <>
                路由规则 <IoMdAddCircleOutline onClick={() => add()} size={20} className="mx-2" />
              </>
            }
          >
            {fields.map((field, index) => (
              <div key={field.key} className=" px-3 pt-3 pb-0 rounded relative">
                <Row gutter={12}>
                  <Col span={11} key={index}>
                    <Form.Item noStyle name={[field.name, 'key']} required>
                      <Select
                        options={routeKeyList}
                        labelInValue
                        placeholder="选择路由规则Key"
                        // onChange={(value) => changeGroupLabel('group', value?.key)}
                      />
                    </Form.Item>
                  </Col>
                  <Col span={11} key={index}>
                    <Form.Item noStyle name={[field.name, 'value']} required>
                      <Input placeholder="value" />
                    </Form.Item>
                  </Col>
                  <Col span={1}>
                    <IoIosRemoveCircleOutline
                      size={20}
                      className="mt-1"
                      onClick={() => remove(field.name)}
                    />
                  </Col>
                </Row>
              </div>
            ))}
          </Form.Item>
        </>
      )}
    </Form.List>
  )
}
