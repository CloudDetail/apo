/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { createContext, useContextSelector } from '@fluentui/react-context-selector'
import { ReactNode, useEffect, useMemo, useState } from 'react'
import { Schemas, TargetTag } from '../views/IntegrationCenter/types'
import { getAllSchemaApi, getTargetTagsListApi } from '../api/alertInput'

interface AlertIntegrationType {
  configDrawerVisible: boolean
  setConfigDrawerVisible: any
  schemas: Schemas
  targetTags: TargetTag[]
}
const AlertIntegrationContext = createContext<AlertIntegrationType>({} as AlertIntegrationType)

export const useAlertIntegrationContext = <T,>(selector: (context: any) => T): T =>
  useContextSelector(AlertIntegrationContext, selector)
export const AlertIntegrationProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const [configDrawerVisible, setConfigDrawerVisible] = useState(false)
  const [schemas, setSchemas] = useState<Schemas>({})
  const [targetTags, setTargetTags] = useState<TargetTag[]>([])
  const getAllSchema = () => {
    getAllSchemaApi()
      .then((res) => {
        setSchemas(res?.schemas || {})
      })
      .catch(() => {
        setSchemas({})
      })
  }
  const getTargetTagsList = () => {
    getTargetTagsListApi()
      .then((res) => {
        const filteredTags = (res.targetTags || []).filter((tag) => tag.id !== 0)
        setTargetTags(filteredTags)
      })
      .catch((error) => {
        setTargetTags([])
      })
  }
  useEffect(() => {
    if (configDrawerVisible) {
      getAllSchema()
      getTargetTagsList()
    }
  }, [configDrawerVisible])

  const drawerMemo = useMemo(() => ({ configDrawerVisible }), [configDrawerVisible])
  const schemasMemo = useMemo(() => ({ schemas }), [schemas])
  const targetTagsMemo = useMemo(() => ({ targetTags }), [targetTags])
  const finalValue = {
    ...drawerMemo,
    ...schemasMemo,
    ...targetTagsMemo,
    setConfigDrawerVisible,
  }
  return (
    <AlertIntegrationContext.Provider value={finalValue}>
      {children}
    </AlertIntegrationContext.Provider>
  )
}
