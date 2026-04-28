## ADDED Requirements

### Requirement: `/agent` 页面展示 creator 维度历史列表
系统 SHALL 在 `/agent` 页面内提供固定可见的历史对话列表，并且该列表 MUST 只展示当前登录账号自己创建的历史线程。每条列表记录 MUST 至少包含 `threadId`、`title` 和 `timeValue` 三个字段，用于后续跳转、展示和排序。对于 APO 当前能力，系统 MUST 将 creator 维度列表视为唯一受支持的列表契约，而不是“全部会话”语义。

#### Scenario: 当前用户只看到自己的历史线程
- **WHEN** 当前登录用户打开 `/agent`
- **THEN** 左侧历史列表 MUST 只返回 creator 为当前登录账号的线程

#### Scenario: 当前用户无历史时展示空状态
- **WHEN** 当前登录用户没有任何自己创建的历史线程
- **THEN** 左侧历史列表区域 MUST 展示空状态提示，且右侧仍可正常开始新对话

### Requirement: 历史列表使用统一时间基准进行排序与展示
系统 SHALL 为每条历史记录提供单一的 `timeValue` 原始时间字段，并且列表排序与时间展示 MUST 基于同一个 `timeValue` 语义。`timeValue` SHOULD 优先表示最近活跃时间；若一期后端暂不具备该字段，则 MUST 在整个版本中统一使用创建时间。若使用最近活跃时间，则该字段 MUST 表示线程级最近持久化活动时间，而不是仅限 creator 本人的最近活动时间。

#### Scenario: 时间显示遵循相对时间与绝对时间规则
- **WHEN** 前端基于 `timeValue` 渲染历史列表时间
- **THEN** 1 小时内 MUST 显示为 `N 分钟前`，当天且超过 1 小时 MUST 显示为 `N 小时前`，昨天 MUST 显示为 `昨天 HH:mm`，7 天内 MUST 显示为 `N 天前`，超过 7 天 MUST 显示为 `YYYY-MM-DD HH:mm`

#### Scenario: 列表排序与展示语义保持一致
- **WHEN** 用户查看历史列表的排序结果和每条记录时间文案
- **THEN** 系统 MUST 保证两者都基于同一个 `timeValue` 字段，而不是分别使用不同时间来源

#### Scenario: 分享访问更新线程活跃时间但不改变归属
- **WHEN** 当前线程采用最近活跃时间作为 `timeValue`，且其他用户通过已知 `threadId` 继续同一个共享线程
- **THEN** creator 视角下的历史列表 MAY 因新的线程活动而刷新排序，但该线程的 creator MUST 保持不变，且该线程 MUST NOT 自动进入访问者的历史列表

### Requirement: 选择历史线程后保持在 `/agent` 并同步 URL
系统 SHALL 在用户点击左侧历史记录后继续停留在 `/agent` 页面，并将所选线程写入 URL `threadId`。页面刷新后 MUST 能根据 URL 中的 `threadId` 重新打开同一线程。

#### Scenario: 点击历史记录打开对应线程
- **WHEN** 用户点击左侧某条历史记录
- **THEN** 右侧聊天区域 MUST 回放该 `threadId` 对应的历史消息，左侧当前项 MUST 高亮，且浏览器 URL MUST 更新为带有该 `threadId` 的 `/agent` 地址

#### Scenario: 刷新页面后恢复选中的历史线程
- **WHEN** 用户打开带 `threadId` 的 `/agent` 地址并刷新页面
- **THEN** 系统 MUST 重新加载该 `threadId` 对应的历史线程，而不是退回到新对话状态

#### Scenario: 清空当前对话后返回新对话状态
- **WHEN** 用户在已打开历史线程的状态下执行“新对话”或清空当前对话
- **THEN** 系统 MUST 清空当前消息上下文，并允许移除 URL 中的 `threadId` 以回到新对话状态

### Requirement: 新对话从首条用户提问开始就必须形成可直达线程
系统 SHALL 在用户处于新对话状态时，于首条用户消息发送时建立新的 `threadId`，并且 MUST 在首条消息发出时或不晚于发出后立即把该 `threadId` 写入 URL。该线程在后端持久化后 MUST 能出现在当前用户的历史列表中，而不要求整页刷新。

#### Scenario: 首条用户消息生成可刷新的线程地址
- **WHEN** 用户在无 `threadId` 的新对话状态下发送第一条用户消息
- **THEN** 系统 MUST 将该消息关联到新的 `threadId`，并更新 URL 以保证刷新后仍能回到同一线程

#### Scenario: 新线程在持久化后进入当前用户历史列表
- **WHEN** 新线程的首条用户消息已经被后端接受并持久化
- **THEN** 当前用户的历史列表 MUST 能在不整页刷新的情况下显示该线程，并使用一期标题规则展示标题

### Requirement: 已知 `threadId` 的直达访问在一期继续保留
系统 SHALL 在一期继续保留基于已知 `threadId` 的 URL 直达能力，即使当前用户不是该线程的 creator，只要 `threadId` 有效，系统 MUST 允许打开对应历史线程。本条直达能力与左侧“我的历史对话”列表范围相互独立。

#### Scenario: 通过 URL 打开其他 creator 的线程
- **WHEN** 当前用户访问一个有效但 creator 不属于自己的 `threadId`
- **THEN** 系统 MUST 允许回放该线程历史内容

#### Scenario: 非 creator 共享线程续聊不属于 APO 一期正式 contract
- **WHEN** 当前用户通过 URL 或分享链接打开一个不属于自己的 `threadId`
- **THEN** 系统 MUST NOT 将该线程上的继续追问视为 APO 一期正式支持能力；若现状中仍存在可行路径，也只视为兼容行为，且本次 change MUST NOT 额外要求新增主动拦截

#### Scenario: 外部线程不会自动进入当前用户历史列表
- **WHEN** 当前用户通过 URL 或分享链接打开一个不属于自己的 `threadId`
- **THEN** 该线程 MUST NOT 因为这次访问而自动出现在当前用户左侧“我的历史对话”列表中

#### Scenario: 无效 `threadId` 返回可恢复错误
- **WHEN** 用户访问一个不存在或已删除的 `threadId`
- **THEN** 页面 MUST 以右上角短暂提示的方式展示可恢复错误，并且 MUST NOT 卡死在不可恢复的加载状态

### Requirement: 会话线程必须持久化 creator 与一期标题
系统 SHALL 为每个 `threadId` 持久化 creator 元数据，并将 creator 语义定义为线程创建者。新线程标题 MUST 默认持久化为首条用户提问；已有历史数据若缺少可用标题，系统 MUST 使用确定性占位文案兜底，而不是隐藏该线程。

#### Scenario: 新线程创建时写入真实 creator 与首问标题
- **WHEN** 当前登录用户首次创建一个新线程并发送首条用户消息
- **THEN** 系统 MUST 将该用户写入线程 creator，并将首条用户提问作为该线程的一期标题持久化

#### Scenario: 分享访问不会改写线程 creator
- **WHEN** 其他用户通过已知 `threadId` 打开并继续同一个线程
- **THEN** 系统 MUST 保持原始 creator 不变，而不是把 creator 改写为当前访问者

#### Scenario: 旧数据缺失标题时仍可展示
- **WHEN** 某条历史线程没有可用持久化标题
- **THEN** 系统 MUST 使用统一的占位标题兜底，并继续在历史列表中展示该线程

### Requirement: 从自己的历史列表打开历史线程后继续追问必须沿用同一 `threadId`
系统 SHALL 在用户通过左侧历史列表打开自己的历史线程后，继续沿用该线程原有的 `threadId` 发送后续消息，而不是重新生成新的线程标识。

#### Scenario: 历史线程中的追问复用原始线程
- **WHEN** 用户已经从自己的历史列表打开某个历史线程并继续输入新的问题
- **THEN** 系统 MUST 使用当前已打开的 `threadId` 继续对话，并将新消息归入同一线程

### Requirement: 切换线程时必须隔离运行中回答与线程级诊断状态
系统 SHALL 防止一个线程中尚未结束的流式回答，或与线程绑定的 findings / evidence / detail 状态，在用户切换到其他历史线程或回到新对话状态后继续渲染或残留到错误线程中。用户发起切换动作时，如果当前 run 仍在进行，系统 MUST 先停止、解绑或显式隔离该 run，并清理、重置或重新分区当前线程以外的诊断状态，再加载目标线程。

#### Scenario: 运行中切换到另一条历史线程
- **WHEN** 当前线程仍在生成回答，且用户点击左侧另一条历史记录
- **THEN** 系统 MUST 先停止、解绑或隔离当前 run，再加载目标 `threadId` 的历史消息

#### Scenario: 运行中返回新对话状态
- **WHEN** 当前线程仍在生成回答，且用户执行“新对话”或清空当前对话
- **THEN** 系统 MUST 先停止、解绑或隔离当前 run，再清空当前线程上下文并返回新对话状态

#### Scenario: 切换后不保留上一个线程的诊断状态
- **WHEN** 用户已经从线程 A 切换到线程 B，或从线程 A 返回新对话状态
- **THEN** 系统 MUST 清理、重建或按 `threadId` 重新分区 findings / evidence / detail 相关状态与缓存，并且 MUST NOT 在线程 B 或新对话界面展示线程 A 的诊断结果

### Requirement: 历史列表在桌面与窄屏场景下都必须可访问
系统 SHALL 在桌面端以固定左侧 `Sider` 形态承载历史入口，并允许默认收起；在窄屏设备上，系统 MAY 退化为可展开历史面板，但 MUST 保持查看历史、打开历史线程和在历史线程中继续追问三项核心能力。

#### Scenario: 桌面端默认收起固定历史侧栏但保留常驻入口
- **WHEN** 用户在桌面端打开 `/agent`
- **THEN** 系统 MUST 保持固定左侧 `Sider` + 右侧聊天区域的双栏结构，且 MAY 默认收起该 `Sider`，但 MUST 提供 header 常驻开关重新展开历史入口

#### Scenario: 窄屏退化后仍保留核心能力
- **WHEN** 用户在窄屏设备打开 `/agent`
- **THEN** 系统 MAY 将历史列表退化为可展开面板，但 MUST 仍允许用户查看自己的历史、打开目标线程，并在打开后的历史线程中继续追问

### Requirement: 用户必须能够最终访问自己账号下的全部历史
系统 SHALL 允许当前用户最终访问自己账号下的全部历史线程。实现 MAY 采用分页、游标或增量加载以避免首屏无界拉取，但 MUST 保持跨页或跨批次的一致排序、时间语义和选中行为。

#### Scenario: 首屏列表不做无界全量拉取
- **WHEN** 当前用户拥有大量历史线程
- **THEN** 系统 MAY 只加载首批历史记录，但 MUST 提供继续访问后续历史的机制，直到用户能够遍历全部自己的历史线程

#### Scenario: 增量加载后仍保持稳定排序与选中
- **WHEN** 用户通过分页、游标或“加载更多”继续浏览历史列表
- **THEN** 系统 MUST 保持与首批列表一致的排序规则、时间语义和当前选中线程行为

### Requirement: 一期不暴露历史删除或重命名入口
系统 SHALL 在一期仅提供历史查看与打开能力。历史列表及相关界面 MUST NOT 暴露删除或重命名入口，也 MUST NOT 依赖现有删除接口的权限语义来界定本次能力是否成立。

#### Scenario: 用户浏览历史列表时不存在删除或重命名操作
- **WHEN** 用户查看一期历史列表能力
- **THEN** 系统 MUST NOT 显示删除、重命名等历史治理操作入口
