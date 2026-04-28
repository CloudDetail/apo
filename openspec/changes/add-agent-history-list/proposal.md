## Why

`/agent` 已经支持通过 URL 中的 `threadId` 回放历史线程，但当前缺少可见入口，用户只有在已知具体 `threadId` 的前提下才能查看历史内容。这使历史对话可发现性和复用效率都过低，也无法满足“按账号查看自己的历史对话”这一基础体验。

## What Changes

- 在现有 `/agent` 页面内新增左侧固定历史对话列表，展示当前登录账号自己创建的历史线程，右侧继续承载现有聊天主区域。
- 桌面端保留左侧固定 `Sider` 形态，但一期默认收起，并通过 header 常驻开关重新展开；窄屏设备允许退化为可展开历史面板，但必须保留“查看历史、打开历史、继续追问”三项核心能力。
- 历史列表每条记录展示标题和时间，并统一约束排序语义与时间展示语义使用同一时间基准字段；一期优先使用最近活跃时间，若后端暂不具备该字段，则允许统一退回创建时间。
- 点击左侧历史记录后，页面继续停留在 `/agent`，同步 URL `threadId`，并在右侧回放对应线程；用户从自己的历史列表打开历史线程后，必须允许继续追问且沿用同一 `threadId`。
- 用户在当前线程仍处于流式回答过程中切换到其他历史线程，或返回新对话状态时，系统必须先停止或隔离当前 run，并清理、重建或重新分区 findings / evidence / 诊断详情等线程级状态，避免旧线程事件或残留诊断状态串入新线程视图。
- 新对话从首条用户提问开始就必须形成可直达的 `threadId` 地址，并在会话持久化后让该线程出现在当前用户历史列表中，无需整页刷新。
- 保留基于 `threadId` 的 URL 直达能力，作为一期轻量分享行为；即使当前用户不是 creator，只要知道有效 `threadId`，仍可打开并回放该线程，但该线程不会自动出现在当前用户左侧“我的历史对话”列表中。至于非 creator 在共享线程上的续聊，本期不把它定义为 APO 侧正式支持 contract，但本次也不要求为此额外增加主动拦截。
- 聊天主区域顶部不再保留运行状态条；对于无效 `threadId` 等可恢复错误，系统必须改为在聊天主区域内联展示错误提示，而不是静默吞掉。
- 一期历史标题默认取会话的首条用户提问，不要求补齐旧数据；若老数据缺少可用标题，则使用确定性占位文案兜底。
- 明确会话 creator 语义：每个 `threadId` 都必须绑定创建者，且写入链路必须使用真实当前登录账号，而不是固定伪用户。
- APO 侧将 `GET /ai/api/v1/conversations/list` 的 creator 维度返回视为本能力唯一受支持契约，实施时需确认当前仓没有其他受支持页面依赖旧的“全部会话”语义。
- 本期不暴露删除或重命名 UI，也不依赖现有删除接口的权限语义。
- “查看全部自己的历史对话”指能力上最终可达全部历史；实现层允许采用分页或增量加载，避免 `/agent` 首屏做无界全量拉取。

## Capabilities

### New Capabilities
- `agent-history-list`: 定义 `/agent` 页面内历史对话列表、creator 维度归属、URL `threadId` 直达兼容分享、标题与时间展示规则，以及空状态与异常处理要求。

### Modified Capabilities

None.

## Impact

- API:
  - `GET /ai/api/v1/conversations/list`
  - `GET /ai/api/v1/conversations/messages`
  - 外部 AgentSRE 会话创建 / 续聊写入链路
- Code:
  - 本仓库 `frontend/src/pro/views/agent/index.tsx`
  - 本仓库 `frontend/src/pro/components/AGUIChat/**`
  - 本仓库 `frontend/src/pro/api/conversation.ts`
  - 本仓库 `frontend/public/locales/*/pro/**`
  - AgentSRE 仓 `api/controllers/chat.py`
  - AgentSRE 仓 `api/controllers/conversation.py`
  - AgentSRE 仓 `api/models/conversation.py`
- Repo Scope:
  - 本仓库 `apo` 商业仓负责 `/agent` 前端与 `pro` 侧文案、线程切换时的 findings / evidence / detail 状态隔离，以及 creator 维度列表契约在当前仓内的消费收口；这些路径属于商业仓独有范围，本次不预期开 OSS `feat/agent-ce` 变更。
  - AgentSRE 仓负责会话归属、列表过滤、`timeValue` 输出和历史回放错误语义。
  - 其他仓或网关默认不在主范围内；仅当现有 APO -> AgentSRE 调用链无法透传真实用户身份时，才需要在拥有该透传链路的仓库补充改动。
- Dependencies:
  - 外部 AgentSRE conversation history 服务及其 `conversations` 持久化模型
  - 当前登录账号上下文
- Systems:
  - APO 商业版 `/agent` 页面交互
  - AgentSRE 会话历史查询、回放与线程归属
  - 基于 `threadId` 的轻量分享 / 直达访问行为
- Compatibility:
  - APO 侧对 `GET /ai/api/v1/conversations/list` 的支持语义收敛为 creator 维度历史列表，不再将其视为“全部会话”接口；apply 阶段需确认当前仓无其他受支持页面依赖旧语义。
- Explicitly out of scope:
  - 管理员查看所有用户历史对话
  - 搜索、筛选、收藏、置顶、重命名、删除等增强能力，以及基于现有删除接口权限语义的能力扩展
  - 对历史存量数据批量补齐标题
  - 一期强制引入独立 share token / share link
