# AGUIChat 组件

一个集成了AI权限检查功能的SRE Agent聊天组件，提供完整的AI对话体验和权限管理。

## 功能特性

- 🔐 **自动权限检查**: 组件加载时自动检查用户是否有AI使用权限
- 🎨 **美观界面**: 专业的聊天界面设计，支持会话管理和消息交互
- 🚫 **权限提示**: 无权限时显示友好的提示界面，引导用户配置
- 🔄 **错误处理**: 完善的错误处理和重试机制
- ⚙️ **配置管理**: 内置设置对话框，支持API配置
- 🌐 **多语言支持**: 支持中英文界面

## 权限检查流程

1. **加载检查**: 组件挂载时自动检查AI权限
2. **权限验证**: 检查用户是否有访问AI相关路由的权限
3. **界面展示**: 
   - 有权限 → 显示正常聊天界面
   - 无权限 → 显示权限提示界面
   - 检查中 → 显示加载状态
   - 错误 → 显示错误信息和重试按钮

## 使用方法

### 基本使用

```tsx
import React from 'react'
import { AGUIProvider } from '../../contexts/AGUIContext'
import AGUIChat from './components/AGUIChat'

const App: React.FC = () => {
  return (
    <AGUIProvider
      initialConfig={{
        baseUrl: 'http://localhost:8081/mock',
        apiKey: 'your-api-key',
      }}
    >
      <AGUIChat />
    </AGUIProvider>
  )
}
```

### 带初始会话ID

```tsx
<AGUIChat initialSessionId="session-123" />
```

## 组件结构

```
AGUIChat/
├── index.tsx                    # 主组件（已集成权限检查）
├── AGUIChat.css                # 样式文件
├── components/                 # 子组件
│   ├── ChatHeader.tsx         # 聊天头部
│   ├── ChatMain.tsx           # 聊天主体
│   ├── MessageList.tsx        # 消息列表
│   ├── InputArea.tsx          # 输入区域
│   ├── SettingsModal.tsx      # 设置对话框
│   └── StatusIndicators.tsx   # 状态指示器
├── types/                     # 类型定义
└── README.md                  # 说明文档
```

## 权限检查逻辑

组件会检查以下路由的权限：
- `/agent` - SRE Agent代理页面
- `/chat` - 聊天页面
- `/ai-chat` - AI聊天页面

如果用户有其中任何一个路由的访问权限，则认为有SRE Agent使用权限。

## 权限提示界面

当用户没有AI使用权限时，会显示专业的SRE Agent权限提示界面，包含：

- 🤖 **功能介绍**: 展示SRE Agent的主要功能
- ⚠️ **权限提示**: 说明需要配置大语言模型API
- 🔧 **配置引导**: 提供配置按钮和手册链接
- 📋 **功能列表**: 展示性能分析、告警分析等功能

## 配置选项

### AGUIChat Props

| 属性 | 类型 | 默认值 | 描述 |
|------|------|--------|------|
| `initialSessionId` | `string` | - | 初始会话ID |

### AGUIProvider Config

| 属性 | 类型 | 默认值 | 描述 |
|------|------|--------|------|
| `baseUrl` | `string` | `http://localhost:3001` | AGUI服务地址 |
| `apiKey` | `string` | `''` | API密钥 |
| `timeout` | `number` | `30000` | 请求超时时间 |
| `retryCount` | `number` | `3` | 重试次数 |

## 环境变量

组件支持以下环境变量：

```bash
# AGUI服务配置
VITE_AGUI_BASE_URL=http://localhost:8081/mock
VITE_AGUI_API_KEY=your-api-key
```

## 样式定制

可以通过修改 `AGUIChat.css` 文件来自定义样式：

```css
.agui-chat-container {
  /* 自定义容器样式 */
  height: 100vh;
  background: #f5f5f5;
}

.chat-content {
  /* 自定义聊天内容区域样式 */
  background: white;
  border-radius: 8px;
  margin: 16px;
}
```

## 多语言支持

组件支持中英文界面，翻译文件位置：
- 中文：`frontend/public/locales/zh/pro/chatAI.json`
- 英文：`frontend/public/locales/en/pro/chatAI.json`

## 错误处理

组件内置了完善的错误处理机制：

1. **权限检查失败**: 显示错误信息和重试按钮
2. **网络错误**: 自动重试，显示友好的错误提示
3. **API错误**: 在SRE Agent聊天界面中显示错误状态

## 性能优化

- 使用 `useCallback` 优化函数引用
- 权限检查结果缓存，避免重复请求
- 组件懒加载，减少初始包大小

## 注意事项

1. **权限系统**: 确保在权限系统中正确配置了SRE Agent相关的路由权限
2. **API配置**: 用户需要配置正确的AGUI服务地址和API密钥
3. **网络环境**: 确保网络环境能够访问AGUI服务
4. **浏览器兼容**: 建议使用现代浏览器以获得最佳体验

## 示例

完整的使用示例请参考 `AGUIChatWithPermission.example.tsx` 文件。

## 更新日志

### v1.0.0
- 初始版本，集成SRE Agent权限检查功能
- 支持完整的聊天界面和会话管理
- 添加权限提示界面和配置引导 