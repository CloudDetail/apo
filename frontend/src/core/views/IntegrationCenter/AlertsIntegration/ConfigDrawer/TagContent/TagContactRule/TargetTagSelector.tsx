/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Form, Select } from 'antd'
import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useAlertIntegrationContext } from 'src/core/contexts/AlertIntegrationContext'
import { TargetTag } from 'src/core/views/IntegrationCenter/types'

interface TargetTagSelectorProps {
  fieldName: string | number
  noLabel?: boolean
}
const TargetTagSelector = ({ fieldName, noLabel = false }: TargetTagSelectorProps) => {
  const { t } = useTranslation('core/alertsIntegration')
  const targetTags = useAlertIntegrationContext((ctx) => ctx.targetTags)
  const [options, setOptions] = useState<TargetTag[]>([])
  useEffect(() => {
    setOptions(targetTags)
  }, [targetTags])

  const handleChange = (newValue: string) => {
    if (newValue && !targetTags.some((option: TargetTag) => option.tagName === newValue)) {
      // 如果输入的值不存在于 options 中，则添加
      setOptions([...targetTags, { tagName: newValue, id: 0 }])
    } else {
      setOptions(targetTags)
    }
    // setValue(newValue)
  }
  return (
    <Form.Item
      name={[fieldName, 'targetTag']}
      label={!noLabel && t('target')}
      style={noLabel ? { marginBottom: 0, overflow: 'hidden', flexGrow: 1 } : {}}
      required
      rules={[
        {
          validator: async (_, value) => {
            console.log(value)
            if (value === undefined || !value.hasOwnProperty('targetTagId')) {
              return Promise.reject(new Error(t('schemaFieldsRequired')))
            }
          },
        },
      ]}
      normalize={(value) => {
        return {
          targetTagId: value.value,
          customTag: value.label,
        }
      }}
      getValueProps={(value) => {
        return {
          id: value?.targetTagId,
          tagName:
            value?.targetTagId === 0
              ? value.customTag
              : options.find((item) => item.id === value?.targetTagId)?.tagName,
          value: value?.targetTagId === 0 ? value.customTag : value?.targetTagId,
          label:
            value?.targetTagId === 0
              ? value.customTag
              : options.find((item) => item.id === value?.targetTagId)?.tagName,
        }
      }}
    >
      <Select
        // allowClear
        className="overflow-hidden"
        placeholder={t('schemaFieldsRequired')}
        options={options}
        fieldNames={{ label: 'tagName', value: 'id' }}
        onSearch={handleChange}
        labelInValue
        filterOption={(input, option) =>
          (option?.tagName ?? '').toLowerCase().includes(input.toLowerCase())
        }
        showSearch
      />
    </Form.Item>
  )
}
export default TargetTagSelector
