export interface Model {
    name: string
}

export interface Config {
    name: string
    label: string
    placeholder: string
    required?: boolean
    value?: string
}

export interface Provider {
    name: string
    description: string
    models: Model[]
    config: Config[]
    setUp: boolean
}

export interface ModelOption {
    provider: string
    model: string
    label?: string
    value?: string
} 