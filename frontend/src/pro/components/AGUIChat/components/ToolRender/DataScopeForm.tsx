import { Button, Checkbox, Form, Input, Radio, Select } from 'antd'
import DateTimeRangePickerCom from 'src/core/components/DateTime/DateTimeRangePickerCom'
import { useEffect } from 'react'

const DataScopeForm = ({ args, status }: { args: any; status: string }) => {
  const [form] = Form.useForm()
  const { fields } = args

  const initialValues = fields?.reduce((acc: Record<string, any>, item: any) => {
    acc[item.name] = item.defaultValue
    return acc
  }, {})
  const RenderFormItem = (item: any) => {
    switch (item.type) {
      case 'Input':
        return <Input defaultValue={item.defaultValue} />
      case 'Textarea':
        return <Input.TextArea defaultValue={item.defaultValue} />
      case 'Select':
        return <Select options={item.options} defaultValue={item.defaultValue} />
      case 'Radio':
        return <Radio.Group options={item.options} defaultValue={item.defaultValue} />
      case 'CheckBox':
        return <Checkbox.Group options={item.options} />
      case 'timerange':
        return <DateTimeRangePickerCom defaultTimerange={item.defaultValue} />
      default:
        return null
    }
  }
  useEffect(() => {
    form.setFieldsValue(initialValues)
  }, [initialValues])

  return (
    <>
      <Form layout="vertical" form={form} initialValues={initialValues}>
        {fields?.map((item) => (
          <Form.Item label={item.label} name={item.name} key={item.name}>
            {RenderFormItem(item)}
          </Form.Item>
        ))}
        <div className="text-right mt-2">
          <Button
            type="primary"
            onClick={() => {
              console.log(args)
            }}
          >
            5s 自动执行
          </Button>
        </div>
      </Form>
    </>
  )
}

export default DataScopeForm
