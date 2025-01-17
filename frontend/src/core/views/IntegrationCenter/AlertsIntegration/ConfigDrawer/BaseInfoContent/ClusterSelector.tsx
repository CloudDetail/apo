/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Form, Select } from 'antd'
import { useEffect, useState } from 'react'
import { getClusterListApi } from 'src/core/api/alertInput'

const ClusterSelector = () => {
  const [clusterList, setClusterList] = useState([])
  const getgetClusterList = () => {
    getClusterListApi()
      .then((res) => {
        setClusterList(res?.clusters || [])
      })
      .catch(() => {
        setClusterList([])
      })
  }
  useEffect(() => {
    getgetClusterList()
  }, [])
  return (
    <Form.Item
      name="clusters"
      label="集群"
      normalize={(value) => {
        if (Array.isArray(value)) {
          return value.map((option) => ({
            id: option.value,
            name: option.label,
          }))
        }
        return []
      }}
      getValueProps={(value) => {
        if (Array.isArray(value)) {
          return {
            value: value.map((option) => ({
              value: option.id,
              label: option.name,
            })),
          }
        }
        return { value: [] }
      }}
    >
      <Select
        mode="multiple"
        allowClear
        style={{ width: '100%' }}
        placeholder="请选择集群"
        options={clusterList}
        fieldNames={{ label: 'name', value: 'id' }}
        labelInValue
      />
    </Form.Item>
  )
}
export default ClusterSelector
