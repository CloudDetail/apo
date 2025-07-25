import { Button, Checkbox, Form, Input } from 'antd'
import { createDataplaneApi, updateDataplaneApi } from 'src/pro/api/dataplane'
import { useEffect } from 'react'
import { DataplaneItem, DataplaneTypeItem } from 'src/pro/types/datasource'
import { notify } from 'src/core/utils/notify'

const DataplaneConfigForm = ({
  dataplaneParams,
  onSave,
  readonly = false,
}: {
  dataplaneParams: DataplaneTypeItem | DataplaneItem
  onSave: () => void
  readonly: boolean
}) => {
  const { paramSpec, typeName, capabilitySpec, capability, params, name, id } = dataplaneParams
  const [form] = Form.useForm()
  const saveDataplane = (values: any) => {
    console.log(values)
    values.params = JSON.stringify(values.params)
    console.log(values)
    const api = id ? updateDataplaneApi : createDataplaneApi
    api(values)
      .then((res) => {
        console.log(res)
        notify({
          message: '接入成功',
          type: 'success',
        })
      })
      .finally(() => {
        onSave()
      })
  }
  const saveForm = () => {
    form.validateFields().then((values) => {
      console.log(values)
      saveDataplane(values)
      onSave()
    })
  }
  useEffect(() => {
    form.resetFields()
    form.setFieldsValue({
      id: id,
      typ: typeName,
      capability: capability || [],
      params: JSON.parse(params || '{}'),
      name: name || '',
    })
  }, [typeName, dataplaneParams])
  return (
    <div className="flex flex-col h-full">
      {dataplaneParams && (
        <Form
          layout="vertical"
          className="mx-4 flex-1 h-0 overflow-y-auto"
          initialValues={{
            typ: typeName,
          }}
          form={form}
          disabled={readonly}
        >
          <Form.Item name="id" hidden>
            <Input />
          </Form.Item>
          <Form.Item name="typ" hidden rules={[{ required: true }]}>
            <Input defaultValue={typeName} />
          </Form.Item>
          <Form.Item name="name" label="接入数据源名称" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          {paramSpec?.children?.map((item) => (
            <Form.Item
              key={item.name}
              name={['params', item.name]}
              label={item.name}
              rules={[{ required: true }]}
            >
              <Input />
            </Form.Item>
          ))}

          <Form.Item name="capability" label="capability" rules={[{ required: true }]}>
            <Checkbox.Group
              options={capabilitySpec?.map((item) => ({ label: item, value: item }))}
            />
          </Form.Item>
        </Form>
      )}
      {!readonly && (
        <div className="flex justify-end  flex-0">
          <Button type="primary" onClick={saveForm}>
            接入
          </Button>
        </div>
      )}
    </div>
  )
}

export default DataplaneConfigForm
