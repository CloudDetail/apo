import { Checkbox, Form, Input, Radio, Select } from 'antd'
import DateTimeRangePickerCom from 'src/core/components/DateTime/DateTimeRangePickerCom'
import { useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'

const DataScopeForm = ({ args }: { args: any }) => {
  const { t } = useTranslation('pro/agent')
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
        return <DateTimeRangePickerCom defaultTimerange={item.defaultValue} readOnly={executed} />
      default:
        return null
    }
  }
  useEffect(() => {
    form.setFieldsValue(initialValues)
  }, [initialValues])
  const [countdown, setCountdown] = useState<number>(5)
  const [executed, setExecuted] = useState(false)
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  useEffect(() => {
    // 初始化后自动开始倒计时
    if (countdown === 0) {
      console.log('自动执行:', args)
      setExecuted(true)
      return
    }

    timerRef.current = setTimeout(() => {
      setCountdown((prev) => prev - 1)
    }, 1000)

    return () => {
      if (timerRef.current) clearTimeout(timerRef.current)
    }
  }, [countdown, args])
  return (
    <>
      <Form layout="vertical" form={form} initialValues={initialValues} disabled={true}>
        {fields?.map((item) => (
          <Form.Item label={item.label} name={item.name} key={item.name}>
            {RenderFormItem(item)}
          </Form.Item>
        ))}
        {/* {!executed && (
          <div className="text-right mt-2">
            <Button type="primary">
              {countdown}s {t('autoExecute')}
            </Button>
          </div>
        )} */}
      </Form>
    </>
  )
}

export default DataScopeForm
