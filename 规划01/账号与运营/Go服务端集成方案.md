# Go 服务端集成方案

本文档定义 Go 服务端如何与 Duel 联机系统对接，包括账号系统改造、QQ 机器人集成、OAuth 配置等。

---

## 1. 整体架构

### 1.1 服务拓扑

```
┌─────────────────────────────────────────────────────────────┐
│                         客户端                              │
│  (Tauri 桌面 / 移动端 / WebXR)                              │
└─────────────────┬───────────────────────────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────────────────────────┐
│                    Go Web 服务端（认证中心）                │
│  - 静态网站托管                                             │
│  - 账号/认证 API                                            │
│  - 房间目录服务                                             │
│  - 管理后台                                                 │
│  - SQLite 数据库                                            │
│  - 游戏数据 API（供 bbs-go 调用）                           │
│  - 事件发布（资格变更、段位更新等）                         │
└─────────────────┬───────────────────────────────────────────┘
                  │
        ┌─────────┼─────────┬─────────────┬─────────────┐
        ▼         ▼         ▼             ▼             ▼
┌───────────┐ ┌───────────┐ ┌───────────┐ ┌───────────┐ ┌───────────┐
│ Rust      │ │ 本地      │ │ PocketBase│ │ bbs-go    │ │ 星图系统  │
│ duel-     │ │ Koishi    │ │ (查看台)  │ │ (社区)    │ │ (官网)    │
│ server    │ │ + NapCat  │ │           │ │           │ │           │
│ (权威模拟)│ │ (QQ机器人)│ │ 只读展示  │ │ 验证JWT   │ │ 验证JWT   │
└───────────┘ └───────────┘ └───────────┘ └───────────┘ └───────────┘
```

### 1.2 职责划分

| 组件 | 职责 |
|------|------|
| Go Web 服务端 | 账号、认证、资格、房间目录、管理后台、静态站点、游戏数据 API、事件发布 |
| Rust duel-server | 游戏权威模拟、tick 推进、快照广播 |
| 本地 Koishi/NapCat | QQ 群成员同步、邀请发放、资格检查执行 |
| PocketBase | 运营查看台（只读，不作为权威数据源） |
| bbs-go | 社区论坛（验证 Go 服务端 JWT，独立存储社区数据） |
| 星图系统 | 官网（验证 Go 服务端 JWT） |

### 1.3 数据库选择

| 项目 | 选择 | 说明 |
|------|------|------|
| 主数据库 | SQLite | 先跑通，后续可迁 PostgreSQL |
| PocketBase | 内置 SQLite | 只做查看（默认不同步，可选投影同步） |

---

## 2. Go 服务端功能模块

### 2.1 核心模块（必须）

| 模块 | 功能 |
|------|------|
| auth | JWT 认证、登录、注册、token 刷新 |
| user | 用户资料、状态管理 |
| eligibility | 内测资格管理、邀请码、门禁 |
| room | 房间目录、匹配入口 |
| admin | 管理后台、封禁、审计日志 |
| bot | Bot API（HMAC认证） |

### 2.2 扩展模块（保留）

| 模块 | 功能 |
|------|------|
| llm | 大语言模型服务 |
| proxy | API 代理服务 |
| oauth | 第三方登录（仅Google启用） |

### 2.3 静态站点

| 页面 | 功能 |
|------|------|
| Landing | 产品介绍、下载入口 |
| Login | 登录/注册（Google OAuth + 账号密码） |
| Dashboard | 资格状态、邀请链接、战报 |
| Admin | 管理后台 |

---

## 3. 账号系统改造

### 3.1 与联机对接的改造点

#### A. WS 鉴权扩展

当远程联机启用账号后，WebSocket 握手需要携带 token：

```json
{
  "t": "hello",
  "protocol": 1,
  "client": { "app": "duel-tauri", "version": "0.3.0" },
  "auth": {
    "kind": "bearer",
    "token": "eyJhbGciOiJIUzI1NiIs..."
  }
}
```

#### B. 鉴权流程

1. Go 服务端验证 JWT
2. 检查 user.status（active/disabled/banned）
3. 检查 beta_eligibility.status（eligible/ineligible）
4. 返回鉴权结果或错误

#### C. 错误码扩展

| 错误 | 说明 |
|------|------|
| UNAUTHORIZED | token 无效或过期 |
| BANNED | 用户被封禁 |
| NOT_ELIGIBLE | 无内测资格 |

### 3.2 资格检查与 QQ 机器人联动

```
┌─────────────┐    群成员快照     ┌─────────────┐
│ 本地 Koishi │ ───────────────> │ Go 服务端   │
│ + NapCat    │                  │             │
│             │ <─────────────── │             │
│             │   邀请发放任务    │             │
└─────────────┘                  └─────────────┘
```

#### 流程说明

1. **群成员同步**：Koishi 定时（每 10 分钟）获取两群成员列表，上报到 Go 服务端
2. **资格更新**：Go 服务端更新所有用户的 `is_in_beta_group` / `is_in_main_group`
3. **邀请发放**：Go 服务端计算需要发邀请的用户列表，返回给 Koishi 执行
4. **结果回传**：Koishi 将发送结果回传，Go 服务端更新邀请码状态

---

## 4. OAuth 配置

### 4.1 Google OAuth（启用）

#### A. Google Developer Console 配置

1. 创建 OAuth 2.0 客户端 ID
2. 配置回调 URL：`https://pama1234.tech/api/auth/oauth/callback/google`
3. 获取客户端 ID 和客户端密钥

#### B. 需要配置的内容

| 配置项 | 说明 |
|--------|------|
| 客户端 ID | 公开，用于识别应用 |
| 客户端密钥 | 保密，用于后端交换 token |
| 回调 URL | Google 重定向的地址 |
| OAuth 同意屏幕 | 应用名称、隐私政策等 |

#### C. Go 后端处理流程

```
用户点击登录
    │
    ▼
重定向到 Google 授权页面
（带 client_id + redirect_uri + scope）
    │
    ▼
用户授权
    │
    ▼
Google 重定向到回调 URL
（带 authorization code）
    │
    ▼
Go 后端用 code 换 access_token
    │
    ▼
Go 后端用 access_token 获取用户信息（sub, email）
    │
    ▼
按 google_sub 查找用户；仅当本次 OAuth 会话绑定了有效 invite_token 时允许创建用户
    │
    ▼
登录与会话续期必须校验门禁 gate_ok(bound_qq)
    │
    ▼
签发 JWT
```

OAuth 的公网路由规范（begin/callback/回跳、安全约束、错误码）以 [账号系统设计.md](./账号系统设计.md) 为准。

> **审核注意**：本项目使用 Google OAuth 仅用于登录，不申请敏感权限。如需了解审核要求与配置详情，详见 [GoogleOAuth配置与审核指南.md](./GoogleOAuth配置与审核指南.md)。

### 4.2 冻结的 OAuth Provider

以下 OAuth provider 已冻结（代码保留但不对外）：

| Provider | 状态 | 说明 |
|----------|------|------|
| GitHub OAuth | 冻结 | 代码保留，不启用 |
| QQ OAuth | 冻结 | 不使用QQ互联，只用QQ机器人 |
| 抖音 OAuth | 冻结 | 代码保留，不启用 |

---

## 5. 房间目录服务

### 5.1 API 设计

#### 创建房间

```
POST /api/v1/rooms
Authorization: Bearer <token>

Request:
{
  "visibility": "private",
  "max_players": 2
}

Response:
{
  "room_id": "room_xxx",
  "room_code": "ABCD",
  "ws_url": "wss://game.pama1234.tech/ws"
}
```

#### 加入房间

```
POST /api/v1/rooms/join
Authorization: Bearer <token>

Request:
{
  "room_code": "ABCD"
}

Response:
{
  "room_id": "room_xxx",
  "ws_url": "wss://game.pama1234.tech/ws"
}
```

### 5.2 与 Rust duel-server 的对接

两种方案：

#### 方案 A：Go 服务端做目录，Rust 做游戏服务

- Go 服务端：房间创建、匹配、目录
- Rust duel-server：实际游戏对局
- Go 服务端返回 Rust 服务端的 WS 地址

#### 方案 B：Go 服务端代理 WS

- 客户端连接 Go 服务端
- Go 服务端代理到 Rust duel-server
- 便于统一鉴权和日志

v1 建议使用方案 A，更简单。

---

## 6. QQ 机器人集成

### 6.1 架构

```
┌─────────────────────────────────────────────────────────────┐
│                        本地机器                              │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐     │
│  │   Koishi    │<──>│   NapCat    │<──>│    QQ       │     │
│  │  (Bot框架)  │    │ (OneBot)    │    │  (客户端)   │     │
│  └──────┬──────┘    └─────────────┘    └─────────────┘     │
│         │                                                   │
│         │ HTTPS + HMAC签名                                  │
│         ▼                                                   │
│  ┌─────────────────────────────────────────────────────┐   │
│  │              Go 服务端（远程）                       │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

### 6.2 认证方式

使用 HMAC 签名（详见 [身份体系设计.md](./身份体系设计.md)）：

```
X-Bot-Id: koishi-bot
X-Timestamp: 1730000000
X-Signature: hmac-sha256签名值
```

### 6.3 Bot API 设计

详见 [Koishi内测门禁与私聊邀请插件设计v2.md](./Koishi内测门禁与私聊邀请插件设计v2.md)

| 端点 | 方法 | 说明 |
|------|------|------|
| /api/v1/bot/qq/snapshot | POST | 上报群快照 |
| /api/v1/bot/qq/events | POST | 上报增量事件 |
| /api/v1/bot/outbox/pull | POST | 拉取待发送任务 |
| /api/v1/bot/outbox/ack | POST | 回传发送结果 |

### 6.4 Outbox 模式

服务端生成任务 → Koishi 拉取 → 执行 → 回传结果

```
服务端                           Koishi
   │                               │
   │  创建任务（invite/otp）        │
   │  写入 bot_outbox              │
   │                               │
   │<────── pull（拉取任务）───────│
   │                               │
   │─────── 返回任务列表 ──────────>│
   │                               │
   │                               │ 执行发送
   │                               │
   │<────── ack（回传结果）────────│
   │                               │
   │  更新任务状态                  │
```

### 6.5 Koishi 插件开发

详细实现见 [Koishi内测门禁与私聊邀请插件设计v2.md](./Koishi内测门禁与私聊邀请插件设计v2.md)

| 功能 | 说明 |
|------|------|
| 群快照上报 | 定时获取群成员列表，上报到 Go 服务端 |
| 任务拉取 | 从 outbox 拉取待发送任务 |
| 私聊发送 | 执行邀请/OTP发送 |
| 结果回传 | 回传发送结果 |

---

## 7. 部署说明

### 7.1 服务端部署（无 Docker）

#### Go 服务端

```bash
# 编译
go build -o duel-web-server

# 使用 systemd 管理
sudo systemctl start duel-web
sudo systemctl enable duel-web
```

#### systemd 配置示例

```ini
[Unit]
Description=Duel Web Server
After=network.target

[Service]
Type=simple
ExecStart=/path/to/duel-web-server
WorkingDirectory=/path/to/app
Restart=always
User=duel
Group=duel

[Install]
WantedBy=multi-user.target
```

### 7.2 本地服务

#### Koishi + NapCat

- 运行在本地机器
- 使用 pm2 或 systemd 管理
- 通过 HTTPS 与远程 Go 服务端通信

---

## 8. 安全考虑

### 8.1 通信安全

- 所有外部通信使用 HTTPS/WSS
- Bot 与服务端通信使用 shared_secret

### 8.2 Token 安全

- access_token 有效期 15 分钟
- refresh_token 有效期 7 天
- 敏感操作需要重新验证

### 8.3 资格安全

- 退出内测群后，资格自动撤销
- 连续两次检测才禁用（防误封）
- 提供申诉入口

---

## 9. 实施顺序

### 9.1 阶段 1：基础账号

1. SQLite 数据库初始化
2. Go 服务端：auth/user 模块
3. 静态站点：登录/注册页面
4. JWT 鉴权

### 9.2 阶段 2：资格系统

1. eligibility 模块
2. 邀请码功能
3. 门禁规则实现
4. QQ 机器人基础集成

### 9.3 阶段 3：联机对接

1. 房间目录服务
2. WS 鉴权扩展
3. Rust duel-server 对接

### 9.4 阶段 4：扩展功能

1. Google OAuth 登录
2. PocketBase 运营查看台
3. 管理后台

---

## 10. 游戏数据 API（供 bbs-go 调用）

### 10.1 API 设计

#### 获取用户游戏数据

```
GET /api/v1/player/{user_id}/game-stats
Authorization: Bearer <server-to-server-token>

Response:
{
  "user_id": 10086,
  "wins": 42,
  "losses": 10,
  "rank": "gold",
  "win_rate": 0.808,
  "total_matches": 52
}
```

#### 获取用户段位信息

```
GET /api/v1/player/{user_id}/rank
Authorization: Bearer <server-to-server-token>

Response:
{
  "user_id": 10086,
  "current_rank": "gold",
  "rank_points": 1250,
  "season": "S1",
  "updated_at": 1730000000
}
```

#### 发放游戏道具（管理员）

```
POST /api/v1/admin/game/items/grant
Authorization: Bearer <admin-token>

Request:
{
  "user_id": 10086,
  "item_id": "skin_golden",
  "quantity": 1,
  "reason": "论坛活动奖励",
  "event_id": "forum_event_001"
}

Response:
{
  "success": true,
  "grant_id": "grant_xxx",
  "granted_at": 1730000000
}
```

### 10.2 服务间认证

bbs-go 调用 Go 服务端 API 时，使用 server-to-server token：

```
Authorization: Bearer <server-to-server-token>
```

该 token 由 Go 服务端签发，仅限服务间调用，与用户 JWT 分离。

---

## 11. 事件发布

### 11.1 事件类型

| 事件 | 说明 | 订阅方 |
|------|------|--------|
| `user.eligibility_changed` | 资格状态变更 | bbs-go、PocketBase |
| `game.rank_updated` | 段位更新 | bbs-go |
| `game.match_completed` | 对局完成 | bbs-go、星图系统 |
| `user.banned` | 用户封禁 | bbs-go |

### 11.2 事件格式

```json
{
  "event": "user.eligibility_changed",
  "timestamp": 1730000000,
  "data": {
    "user_id": 10086,
    "old_status": "eligible",
    "new_status": "ineligible",
    "reason": "left_beta_group"
  }
}
```

### 11.3 订阅方式

#### 方式 A：Webhook

bbs-go 提供 webhook 端点，Go 服务端主动推送：

```
POST https://bbs.pama1234.tech/api/webhook/events
X-Event-Signature: hmac-sha256签名
```

#### 方式 B：消息队列

使用 Redis Pub/Sub 或 NATS：

```
Go 服务端 → Redis/NATS → bbs-go 订阅
```

v1 建议使用方式 A（Webhook），更简单。

---

## 相关文档

- [账号系统设计.md](./账号系统设计.md) - 账号系统整体设计
- [身份体系设计.md](./身份体系设计.md) - 身份主体设计
- [QQ机器人集成方案.md](./QQ机器人集成方案.md) - QQ机器人集成
- [Koishi内测门禁与私聊邀请插件设计v2.md](./Koishi内测门禁与私聊邀请插件设计v2.md) - Koishi插件详细设计
- [数据库迁移计划.md](./数据库迁移计划.md) - 文件存储到SQLite迁移
- [bbs-go社区集成方案.md](./bbs-go社区集成方案.md) - bbs-go 社区集成详细方案
