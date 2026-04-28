import { HttpAgent, EventType } from '@ag-ui/client';

// 使用我们自定义的类型定义，但底层使用官方SDK
import {
    AGUISession,
    AGUIClientConfig,
    AGUIMessage,
    Event,
    AGUIRunRequest,
    AGUIRunResponse,
} from '../types/ag-ui';
import i18next from 'i18next';
import dayjs from 'dayjs';

// 包装HttpAgent以提供我们需要的接口
export class AGUIClient {
    private agent: HttpAgent;
    private config: AGUIClientConfig;
    private abortController?: AbortController;

    constructor(config: AGUIClientConfig) {
        this.config = {
            timeout: 30000,
            retryCount: 3,
            ...config
        };

        // 创建HttpAgent实例
        const headers = {
            ...(this.config.headers || {}),
            ...(this.config.apiKey ? {
                'Authorization': `Bearer ${this.config.apiKey}`,
            } : {}),
        };
        this.agent = new HttpAgent({
            url: this.config.baseUrl,
            headers,
        });
        this.agent.abortController = this.abortController;
    }

    private getHeaders(): Record<string, string> {
        return {
            ...(this.config.headers || {}),
            ...(this.config.apiKey ? {
                'Authorization': `Bearer ${this.config.apiKey}`,
            } : {}),
        };
    }

    // 开始运行Agent (适配我们的接口到官方SDK)
    async startRun(request: AGUIRunRequest): Promise<AGUIRunResponse> {
        console.log('startRun', request)
        try {
            Object.assign(this.agent.headers, this.getHeaders());
            // 转换我们的请求格式到官方格式
            const runInput = {
                threadId: request.threadId,
                runId: request.runId,
                messages: request.messages.map(msg => ({
                    role: msg.role,
                    content: msg.content,
                    id: msg.id,
                })),
                state: request.state,
                tools: request.tools || [],
                context: request.context || [],
                ...request.forwardedProps
            };

            await this.agent.runAgent(runInput);

            return {
                runId: request.runId,
                threadId: request.threadId,
                status: 'running'
            };
        } catch (error) {
            console.error('启动Agent运行失败:', error);
            throw new Error(`启动运行失败: ${error instanceof Error ? error.message : '未知错误'}`);
        }
    }

    // 订阅运行事件 (使用流式API)
    subscribeToRun(request: AGUIRunRequest, callback: (event: Event) => void): () => void {
        try {
            Object.assign(this.agent.headers, this.getHeaders());
            this.agent.headers['Accept-Language'] = i18next.language;
            // 创建AbortController用于取消
            this.abortController = new AbortController();
            this.agent.abortController = this.abortController;
            // 转换请求格式
            // request.context ??= []
            // request.context.push({
            //     clientDate: dayjs().format('YYYY-MM-DDTHH:mm:ssZ')
            // })
            request.forwardedProps ??= {}
            request.forwardedProps.clientDate = dayjs().format('YYYY-MM-DDTHH:mm:ssZ')
            const runInput = {
                threadId: request.threadId,
                runId: request.runId,
                messages: request.messages.map(msg => ({
                    role: msg.role,
                    content: msg.content,
                    id: msg.id,
                })),
                state: request.state,
                tools: request.tools || [],
                context: request.context || [],
                forwardedProps: request.forwardedProps,

            };

            // 使用HttpAgent的run方法进行流式处理
            const runObservable = this.agent.run(runInput);

            // 订阅Observable事件流
            const subscription = runObservable.subscribe({
                next: (event: any) => {
                    // 直接转发事件
                    console.log('Agent运行事件:', event);
                    callback(event);
                },
                error: (error: any) => {
                    console.error('Agent运行错误:', error);
                    const errorEvent: Event = {
                        type: EventType.RUN_ERROR,
                        timestamp: Date.now(),
                        message: error instanceof Error ? error.message : '运行时错误',
                        code: 'RUNTIME_ERROR'
                    } as any;
                    callback(errorEvent);
                },
                complete: () => {
                    console.log('Agent运行完成');
                }
            });

            // 返回取消函数
            return () => {
                if (this.abortController) {
                    this.abortController.abort();
                }
                subscription.unsubscribe();
            };
        } catch (error) {
            console.error('订阅Agent事件失败:', error);
            return () => { };
        }
    }



    // 取消运行
    async cancelRun(runId: string): Promise<void> {
        try {
            console.log(`取消运行: ${runId}`);

            // 先中止当前的AbortController
            if (this.abortController) {
                this.abortController.abort();
                this.abortController = undefined; // 重置abortController
            }

            // 然后调用HttpAgent的abortRun方法
            try {
                await this.agent.abortRun();
            } catch (abortError) {
                // 忽略abortRun的错误，因为可能已经被AbortController中止了
                console.log('HttpAgent abortRun被忽略:', abortError);
            }
        } catch (error) {
            console.error('取消运行失败:', error);
            // 不要抛出错误，因为取消操作本身不应该被视为错误
            // throw new Error(`取消运行失败: ${error instanceof Error ? error.message : '未知错误'}`);
        }
    }

    // 会话管理方法 (保持原有的本地存储实现)
    async getSession(sessionId: string): Promise<AGUISession> {
        const stored = localStorage.getItem('agui_sessions') || '[]';
        const sessions = JSON.parse(stored);
        const session = sessions.find((s: any) => s.id === sessionId);
        if (!session) {
            throw new Error(`会话不存在: ${sessionId}`);
        }
        return session;
    }

    async updateSessionState(sessionId: string, state: Record<string, any>): Promise<void> {
        const stored = localStorage.getItem('agui_sessions') || '[]';
        const sessions = JSON.parse(stored);
        const sessionIndex = sessions.findIndex((s: any) => s.id === sessionId);
        if (sessionIndex === -1) {
            throw new Error(`会话不存在: ${sessionId}`);
        }
        sessions[sessionIndex] = { ...sessions[sessionIndex], ...state, updatedAt: new Date() };
        localStorage.setItem('agui_sessions', JSON.stringify(sessions));
    }

    async deleteSession(sessionId: string): Promise<void> {
        const stored = localStorage.getItem('agui_sessions') || '[]';
        const sessions = JSON.parse(stored);
        const filtered = sessions.filter((s: any) => s.id !== sessionId);
        localStorage.setItem('agui_sessions', JSON.stringify(filtered));
    }

    async getAllSessions(): Promise<AGUISession[]> {
        const stored = localStorage.getItem('agui_sessions') || '[]';
        return JSON.parse(stored);
    }

    async createSession(session: AGUISession): Promise<void> {
        const stored = localStorage.getItem('agui_sessions') || '[]';
        const sessions = JSON.parse(stored);
        sessions.unshift(session);
        localStorage.setItem('agui_sessions', JSON.stringify(sessions));
    }

    async addMessageToSession(sessionId: string, message: AGUIMessage): Promise<void> {
        const session = await this.getSession(sessionId);
        const messages = session.messages || [];
        messages.push(message);
        await this.updateSessionState(sessionId, { messages });
    }
}
