// 使用AG-UI官方SDK的类型
export { EventType } from '@ag-ui/client';

// AG-UI协议类型定义 - 基于官方SDK
export interface AGUIMessage {
    id: string;
    role: 'user' | 'assistant' | 'system';
    content: string;
    timestamp: Date;
    metadata?: {
        model?: string;
        tokens?: number;
        latency?: number;
    };
}

export interface AGUIToolCall {
    id: string;
    type: 'function';
    function: {
        name: string;
        arguments: string;
    };
}

export interface AGUIToolCallResult {
    id: string;
    result: any;
    error?: string;
}

// ID生成工具函数
export function generateThreadId(): string {
    return `thread_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
}

export function generateRunId(): string {
    return `run_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
}

export function generateMessageId(): string {
    return `msg_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
}

// 使用官方Event类型
export type Event = any; // 使用官方SDK的Event类型

// 具体事件类型定义 (基于官方SDK)
export type BaseEvent = {
    type: any;
    timestamp?: number;
    rawEvent?: any;
}

// 生命周期事件
export type RunStartedEvent = BaseEvent & {
    type: 'RUN_STARTED';
    threadId: string;
    runId: string;
}

export type RunFinishedEvent = BaseEvent & {
    type: 'RUN_FINISHED';
    threadId: string;
    runId: string;
    result?: any;
}

export type RunErrorEvent = BaseEvent & {
    type: 'RUN_ERROR';
    message: string;
    code?: string;
}

export type StepStartedEvent = BaseEvent & {
    type: 'STEP_STARTED';
    stepName: string;
}

export type StepFinishedEvent = BaseEvent & {
    type: 'STEP_FINISHED';
    stepName: string;
}

// 文本消息事件
export type TextMessageStartEvent = BaseEvent & {
    type: 'TEXT_MESSAGE_START';
    messageId: string;
    role: "assistant";
}

export type TextMessageContentEvent = BaseEvent & {
    type: 'TEXT_MESSAGE_CONTENT';
    messageId: string;
    delta: string;
}

export type TextMessageEndEvent = BaseEvent & {
    type: 'TEXT_MESSAGE_END';
    messageId: string;
}

// 工具调用事件
export type ToolCallStartEvent = BaseEvent & {
    type: 'TOOL_CALL_START';
    toolCallId: string;
    toolCallName: string;
    parentMessageId?: string;
}

export type ToolCallArgsEvent = BaseEvent & {
    type: 'TOOL_CALL_ARGS';
    toolCallId: string;
    delta: string;
}

export type ToolCallEndEvent = BaseEvent & {
    type: 'TOOL_CALL_END';
    toolCallId: string;
}

export type ToolCallResultEvent = BaseEvent & {
    type: 'TOOL_CALL_RESULT';
    messageId: string;
    toolCallId: string;
    content: string;
    role?: "tool";
}

// 状态管理事件
export type StateSnapshotEvent = BaseEvent & {
    type: 'STATE_SNAPSHOT';
    snapshot: any;
}

export type StateDeltaEvent = BaseEvent & {
    type: 'STATE_DELTA';
    delta: any[];
}

export type MessagesSnapshotEvent = BaseEvent & {
    type: 'MESSAGES_SNAPSHOT';
    messages: AGUIMessage[];
}

// 特殊事件
export type RawEvent = BaseEvent & {
    type: 'RAW';
    event: any;
    source?: string;
}

export type CustomEvent = BaseEvent & {
    type: 'CUSTOM';
    name: string;
    value: any;
}

// AG-UI会话和状态
export interface AGUISession {
    id: string;
    title: string;
    messages: AGUIMessage[];
    state: Record<string, any>;
    createdAt: Date;
    updatedAt: Date;
}

// 根据AG-UI协议定义的RunAgentInput结构
export interface AGUIRunRequest {
    threadId: string;
    runId: string;
    state: any;
    messages: AGUIMessage[];
    tools: any[];
    context: any[];
    forwardedProps: any;
}

export interface AGUIRunResponse {
    runId: string;
    threadId: string;
    status: 'running' | 'completed' | 'error';
}

// AG-UI客户端配置
export interface AGUIClientConfig {
    baseUrl: string;
    apiKey?: string;
    timeout?: number;
    retryCount?: number;
}

// AG-UI客户端接口
export interface IAGUIClient {
    startRun(request: AGUIRunRequest): Promise<AGUIRunResponse>;
    subscribeToRun(request: AGUIRunRequest, callback: (event: Event) => void): () => void;
    getSession(sessionId: string): Promise<AGUISession>;
    updateSessionState(sessionId: string, state: Record<string, any>): Promise<void>;
    cancelRun(runId: string): Promise<void>;
}

// 为了向后兼容，保留旧的类型别名
export type AGUIEventType = any;


export type AnalysisMessageStartEvent = BaseEvent & {
    type: 'CUSTOM';
    name: 'ANALYSIS_MESSAGE_START',
    value: {
        id: string
        parentId: string
    }
}

export type AnalysisMessageContentEvent = BaseEvent & {
    type: 'CUSTOM';
    name: 'ANALYSIS_MESSAGE_CONTENT',
    value: {
        id: string,
        delta: string
    }
}

export type AnalysisMessageEndEvent = BaseEvent & {
    type: 'CUSTOM';
    name: 'ANALYSIS_MESSAGE_END';
    value: {
        id: string;
    };
};

export type AnalysisResultEvent = BaseEvent & {
    type: 'CUSTOM';
    name: 'ANALYSIS_RESULT';
    value: {
        parentId: string;
        content: any;
        id: string
    };
};