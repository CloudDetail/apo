
export type ReportType = 'slow' | 'error' | 'infra'
export type TagKey = 'serviceName' | 'pod' | 'containerId' | 'nodeName' | 'pid'
export type TagsParams = {
    [key in TagKey]: string
}
export type RootCauseTyoe = 'cpu' | 'code' | 'net' | 'storage' | 'downstream'
export type SuggestLevel = 'high' | 'medium' | 'low'