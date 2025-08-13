import { Button, Flex, Form, Select, Tag } from 'antd'
import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { AiOutlineLine } from 'react-icons/ai'
import { useSelector } from 'react-redux'
import { useDebounce } from 'react-use'
import { selectProcessedTimeRange } from 'src/core/store/reducers/timeRangeReducer'
import { getServiceNameRuleApi, getServiceValuesRuleApi } from 'src/pro/api/dataplane'
const ConditionsFormListItem = ({
  subField,
  index,
  subOpt,
  keys = [],
}: {
  subField: any
  index: number
  subOpt: any
  keys: string[]
}) => {
  const { t } = useTranslation('pro/dataplane')
  const [keyOptions, setKeyOptions] = useState<{ label: string; value: string }[]>([])
  const operationOptions = [
    {
      value: 'equals',
      label: 'equals',
    },
    {
      value: 'startsWith',
      label: 'startsWith',
    },
    {
      value: 'endsWith',
      label: 'endsWith',
    },
    {
      value: 'contains',
      label: 'contains',
    },
    {
      value: 'has',
      label: <span>{t('has')}</span>,
    },
  ]
  const [keySearchValue, setKeySearchValue] = useState('')
  const [valueSearchValue, setValueSearchValue] = useState('')
  const [valueOptions, setValueOptions] = useState<{ label: string; value: string }[]>([])
  const form = Form.useFormInstance()
  const keyValue = Form.useWatch(['conditions', subField.name, 'key'], form)
  const { startTime, endTime } = useSelector(selectProcessedTimeRange)
  const getValuesOptions = () => {
    getServiceValuesRuleApi({
      key: keyValue,
      startTime: startTime,
      endTime: endTime,
    }).then((res) => {
      setValueOptions(res.values?.map((value) => ({ label: value, value: value })) || [])
    })
  }
  useEffect(() => {
    if (keyValue && startTime && endTime) {
      getValuesOptions()
    }
  }, [keyValue, startTime, endTime])
  useEffect(() => {
    setKeyOptions(keys?.map((key) => ({ label: key, value: key })) || [])
  }, [keys])
  useEffect(() => {
    // 先过滤已有的 keys，生成初始选项
    let filteredOptions =
      keys
        ?.filter((key) => key.toLowerCase().includes(keySearchValue.toLowerCase()))
        ?.map((key) => ({
          label: key,
          value: key,
        })) || []

    // 判断当前搜索值是否已经存在于 options 中
    const hasSearchValue = keys?.some((key) => key.toLowerCase() === keySearchValue.toLowerCase())

    // 如果不存在，并且有输入，则添加为新选项
    if (keySearchValue && !hasSearchValue) {
      filteredOptions = [
        {
          label: keySearchValue,
          value: keySearchValue,
        },
        ...filteredOptions,
      ]
    }

    setKeyOptions(filteredOptions)
  }, [keySearchValue, keys])

  useEffect(() => {
    // 先过滤已有的 keys，生成初始选项
    let filteredOptions = valueOptions.filter((value) =>
      value.label.toLowerCase().includes(valueSearchValue.toLowerCase()),
    )

    // 判断当前搜索值是否已经存在于 options 中
    const hasSearchValue = valueOptions.some(
      (value) => value.label.toLowerCase() === valueSearchValue.toLowerCase(),
    )

    // 如果不存在，并且有输入，则添加为新选项
    if (valueSearchValue && !hasSearchValue) {
      filteredOptions = [
        {
          label: valueSearchValue,
          value: valueSearchValue,
        },
        ...filteredOptions,
      ]
    }

    setValueOptions(filteredOptions)
  }, [valueSearchValue])
  return (
    <Form.Item key={subField.key} rules={[{ required: true }]} style={{ marginBottom: 0 }}>
      <Flex justify="center" align="flex-start" gap={5}>
        {index > 0 && (
          <Form.Item style={{ marginBottom: 0 }}>
            <Tag color="processing">{t('and')}</Tag>
          </Form.Item>
        )}

        <Form.Item name={[subField.name, 'conditionId']} hidden></Form.Item>

        <Form.Item
          name={[subField.name, 'key']}
          rules={[{ required: true, message: t('keyRequired') }]}
          style={{ marginBottom: 0, flex: 1, overflow: 'hidden' }}
        >
          <Select
            options={keyOptions}
            showSearch
            searchValue={keySearchValue}
            onSearch={setKeySearchValue}
          ></Select>
        </Form.Item>
        <Form.Item
          name={[subField.name, 'matchType']}
          style={{ marginBottom: 0 }}
          initialValue="equals"
        >
          <Select options={operationOptions} style={{ width: 120 }}></Select>
        </Form.Item>
        <Form.Item
          name={[subField.name, 'value']}
          className="flex-1"
          rules={[{ required: true, message: t('valueRequired') }]}
          style={{ marginBottom: 0, flex: 1, overflow: 'hidden' }}
        >
          <Select
            options={valueOptions}
            showSearch
            searchValue={valueSearchValue}
            onSearch={setValueSearchValue}
          ></Select>
        </Form.Item>
        <Form.Item style={{ marginBottom: 0 }}>
          <Button
            size="small"
            danger
            icon={<AiOutlineLine />}
            className="flex-grow-0 flex-shrink-0"
            onClick={() => subOpt.remove(subField.name)}
          >
            {/* <span className="text-xs"></span> */}
          </Button>
        </Form.Item>
      </Flex>
    </Form.Item>
  )
}

const ConditionsFormList = () => {
  const { t } = useTranslation('pro/serviceMatcher')
  const { startTime, endTime } = useSelector(selectProcessedTimeRange)

  const [keys, setKeys] = useState<string[]>([])

  useDebounce(
    () => {
      getServiceNameRuleApi({
        startTime: startTime,
        endTime: endTime,
      }).then((res) => {
        setKeys(res.labels)
      })
    },
    500,
    [],
  )
  return (
    <Form.List
      name={'conditions'}
      rules={[
        {
          validator: (_, value) => {
            console.log(value)
            if (value.length === 0) {
              return Promise.reject(new Error(t('conditionsRequired')))
            }
            return Promise.resolve()
          },
        },
      ]}
    >
      {(subFields, subOpt, { errors }) => (
        <div style={{ display: 'flex', flexDirection: 'column', rowGap: 16 }}>
          {subFields.map((subField, index) => (
            <ConditionsFormListItem subField={subField} index={index} subOpt={subOpt} keys={keys} />
          ))}
          <Button color="primary" variant="filled" onClick={() => subOpt.add()} block>
            + {t('addConditions')}
          </Button>
          <Form.ErrorList errors={errors} />
        </div>
      )}
    </Form.List>
  )
}
export default ConditionsFormList
