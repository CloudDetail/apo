

export type DataType = 'RED Metrics' | 'Service Topology' | 'Basic Metrics'
    | 'Logs' | 'Traces' | 'Alerts' | 'Issues' | 'Problems'

export type DataplaneType = 'APO' | 'Datadog' | 'NewRelic' | 'Dynatrace'
export interface DataplaneTypeItem {
    typeName: string
    typeDisplayName: string
    paramSpec: any
    desc: string
    capabilitySpec: string[]
}
export interface DataplaneItem extends DataplaneTypeItem {
    id: string
    capability: string[]
    desc: string
    params: string
    config: any
}