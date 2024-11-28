import { Col, Form, Input, Row, Select } from 'antd'
import React from 'react'
import { AiOutlineInfoCircle } from 'react-icons/ai'
import { IoIosRemoveCircleOutline, IoMdAddCircleOutline } from 'react-icons/io'
import LogStructRuleFormList from './LogStructRuleFormList'
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
  const form = Form.useFormInstance()
  return (
    <Form.List name={'routeRule'}>
      {(fields, { add, remove }, { errors }) => (
        <>
          <Form.Item
            required
            label={
              <>
                {/* <div className="flex flex-row"> */}
                匹配规则{' '}
                <IoMdAddCircleOutline
                  onClick={() => add()}
                  size={20}
                  className="mx-2 cursor-pointer"
                />
                {/* </div>
                <div className="flex flex-row"> */}
                <AiOutlineInfoCircle size={16} className="ml-1 mr-1" />
                <span className="text-xs text-gray-400">解析规则只应用于满足匹配规则的日志</span>
                {/* </div> */}
              </>
            }
          >
            {fields.map((field, index) => (
              <div key={field.name} className=" px-3 pt-3 pb-0 rounded relative">
                <Row gutter={12}>
                  <Col span={11} key={index}>
                    <Form.Item
                      name={[field.name, 'key']}
                      required
                      rules={[
                        {
                          validator: async (_, value) => {
                            // 获取当前表单中所有的routeRule项
                            const routeRule = form.getFieldValue('routeRule') || []
                            // 检查是否有重复的key
                            if (!value) {
                              return Promise.reject('匹配规则key不可为空')
                            }
                            const duplicate = routeRule.filter(
                              (item, i) => item?.key?.key === value.key && i !== index,
                            )
                            if (duplicate.length) {
                              return Promise.reject('已存在相同的Key')
                            }
                          },
                        },
                      ]}
                    >
                      <Select options={routeKeyList} labelInValue placeholder="选择匹配规则Key" />
                    </Form.Item>
                  </Col>
                  <Col span={11} key={index}>
                    <Form.Item noStyle name={[field.name, 'value']} required>
                      <Input placeholder="匹配值，按照前缀匹配" />
                    </Form.Item>
                  </Col>
                  <Col span={1}>
                    <IoIosRemoveCircleOutline
                      size={20}
                      className="mt-1 cursor-pointer"
                      onClick={() => {
                        remove(field.name)
                      }}
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
