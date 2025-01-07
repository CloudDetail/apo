/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

export interface SourceId {
  sourceId: string
}

export interface ClustersItem {
  id: string
  name: string
}
export type AlertKey = 'json' | 'zabbix' | 'prometheus'

export interface SourceInfo {
  sourceId: string
  sourceType: AlertKey
  sourceName: string
  clusters?: ClustersItem[] | null // 可选属性，允许为 null 或 undefined
}

export interface AlertInputSourceParams extends Partial<SourceId> {
  sourceType?: AlertKey
  sourceName?: string
  clusters?: ClustersItem[] | null
}

export type AlertInputBaseInfo = Required<AlertInputSourceParams>

export type RType = 'tagMapping' | 'schemaMapping'
export type Operation = 'match' | 'notMatch'
interface ConditionItem {
  fromField: string
  operation: Operation
  expr: string
}
interface EnrichRuleBasicInfo {
  enrichRuleId?: string
  fromField: string
  fromRegex?: string
  conditions?: ConditionItem[]
}
export interface SchemaTargetItem {
  targetTagId: string
  customTag?: string
  schemaField?: string
}
interface TagMappingConfig extends EnrichRuleBasicInfo {
  rType: 'tagMapping'
  targetTagId: number
  customTag?: string
}

interface SchemaMappingConfig extends EnrichRuleBasicInfo {
  rType: 'schemaMapping'
  schema: string
  schemaSource: string
  schemaTargets: SchemaTargetItem[]
}
export type EnrichRuleConfigItem = TagMappingConfig | SchemaMappingConfig
export interface SaveAlertEnrichParams {
  sourceId: string
  setAsDefault?: boolean
  enrichRuleConfigs: EnrichRuleConfigItem[]
}
