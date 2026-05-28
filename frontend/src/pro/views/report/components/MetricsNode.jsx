/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react'
import { Handle, Position } from 'reactflow'
import { theme } from 'antd'
import { MetricsValue } from 'src/core/utils/metrics'
import { useTranslation } from 'react-i18next'
const MetricsNode = React.memo(({ data, isConnectable }) => {
  const { useToken } = theme
  const { token } = useToken()
  const { t } = useTranslation('common')
  const getMetricLabel = (key) => {
    const translated = t(`delayLineChartTitleMap.${key}`)
    return translated === `delayLineChartTitleMap.${key}` ? key : translated
  }
  const getMetricValue = (metric, key) => {
    const metricTypes = ['latency', 'errorRate', 'tps', 'logs']
    if (typeof metric.value === 'number' && metricTypes.includes(key)) {
      return MetricsValue(metric.value, key)
    }
    return metric.value
  }
  return (
    <div className="text-updater-node">
      <Handle
        type="target"
        position={Position.Left}
        isConnectable={isConnectable}
        className="invisible"
      />
      <div
        className="w-[200px] min-h-[60px] p-2 rounded-md border-2 border-solid overflow-hidden"
        style={{
          backgroundColor: token.colorBgLayout,
          color: token.colorPrimaryText,
          borderColor: token.colorPrimaryText,
        }}
      >
        <div className="text-center text-lg pt-2 px-2">
          {data.label}
          <div className="text-xs text-[#9e9e9e] text-left  break-all">{data.description}</div>
          <div className="w-full">
            {Object.entries(data.metrics || {})?.map(([key, metric]) => (
              <div key={key} className="flex text-xs text-[var(--ant-color-text)] w-full">
                <div className="shrink-0 w-[90px] text-left">
                  {getMetricLabel(key)}：
                </div>
                <div
                  className={
                    metric.value === 'N/A' ? 'text-slate-400' : 'text-[var(--ant-color-text)]'
                  }
                >
                  {getMetricValue(metric, key)}
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
      <Handle
        type="source"
        position={Position.Right}
        id="b"
        isConnectable={isConnectable}
        className="invisible"
      />
    </div>
  )
})

export default MetricsNode
