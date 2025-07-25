export interface Session {
    id: string
    title: string
    updatedAt: Date | string
}

export interface Message {
    id: string
    role: 'user' | 'assistant' | 'system'
    content: string
    timestamp?: Date | string
    metadata?: {
        model?: string
    }
    toolCalls?: any[]
}

export interface ClientConfig {
    baseUrl: string
    apiKey: string
}

export interface ToolCall {
    id: string
    name: string
    args: any
    status: 'pending' | 'completed' | 'error'
    result?: any
    error?: string
} 