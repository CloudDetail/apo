/**
 * Copyright 2026 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react'
import { Select } from 'antd'
import {
  isStandardResourceVariable,
  variableDisplayLabel,
  variableSelectMode,
  variableSelectOptions,
} from './behavior.mjs'

const CustomDashboardResourceFilter = ({
  t,
  variables = [],
  values = {},
  candidates = {},
  candidateErrors = {},
  onChange,
  onDropdownVisibleChange,
}) => {
  const resourceVariables = variables.filter(isStandardResourceVariable)
  if (!resourceVariables.length) return null

  return (
    <>
      {resourceVariables.map((variable) => (
        <span className="custom-dashboard-toolbar-parameter" key={variable.name}>
          <span className="custom-dashboard-toolbar-parameter-label">
            {variableDisplayLabel(t, variable)}
          </span>
          <Select
            placeholder={variable.name}
            value={values[variable.name] || undefined}
            onChange={(value) => onChange?.(variable, value)}
            onDropdownVisibleChange={(open) => onDropdownVisibleChange?.(variable, open)}
            style={{ width: 180 }}
            allowClear
            mode={variableSelectMode(variable)}
            maxTagCount={2}
            maxTagPlaceholder={(omittedValues) => `+${omittedValues.length}...`}
            popupMatchSelectWidth={false}
            notFoundContent={candidateErrors[variable.name] || t('noAvailableValues')}
            options={variableSelectOptions(t, variable, candidates[variable.name] || [])}
          />
        </span>
      ))}
    </>
  )
}

export default React.memo(CustomDashboardResourceFilter)
