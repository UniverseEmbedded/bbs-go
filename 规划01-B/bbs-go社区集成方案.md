# bbs-go 社区集成方案

本文档定义 bbs-go 社区论坛与几何决斗账号系统的集成方案。

---

## 1. 定位

### 1.1 bbs-go 的角色

| 维度 | 说明 |
|------|------|
| 功能定位 | 独立社区论坛，提供帖子、评论、点赞等社区功能 |
| 数据定位 | 社区数据独立存储，不与游戏数据混合 |
| 认证定位 | 验证 Go 服务端签发的 JWT，不独立管理账号 |

### 1.2 与其他系统的关系

```
┌─────────────────────────────────────────────────────────────┐
│                    Go 服务端（认证中心）                      │
│  - 用户账号（权威）                                          │
│  - JWT 签发                                                  │
│  - 游戏数据                                                  │
│  - 内测资格                                                  │
└─────────────────────────────────────────────────────────────┘
                          │
          ┌───────────────┼───────────────┐
          │               │               │
          ▼               ▼               ▼
   ┌─────────────┐ ┌─────────────┐ ┌─────────────┐
   │  游戏客户端  │ │   bbs-go    │ │  星图系统   │
   │             │ │   (社区)    │ │   (官网)    │
   │ 验证 JWT    │ │ 验证 JWT    │ │ 验证 JWT    │
   └─────────────┘ └─────────────┘ └─────────────┘
```

---

## 2. 认证集成

### 2.1 核心原则

- **用户 ID 统一**：Go 服务端的 `users.id` 是全局唯一用户标识
- **JWT 统一**：所有系统验证同一个 JWT
- **不接入 bbs-go 登录**：不使用 bbs-go 的注册/短信/微信/Google 登录体系
- **首次访问幂等创建**：用户首次访问 bbs-go 时，按 `user_id` 幂等创建对应的 bbs-go 用户（并发安全）

### 2.2 认证流程

```
用户访问 bbs-go
       │
       ▼
bbs-go 尝试从请求中提取 JWT
       │
       ├── 无 JWT → 匿名访问（仅可访问公开内容）
       │
       └── 有 JWT → 验证 JWT 签名和有效期
              │
              ├── 无效 → 匿名访问
              │
              └── 有效 → 提取 user_id
                     │
                     ▼
              查询 bbs-go 用户表
                     │
                     ├── 存在 → 建立会话，继续访问
                     │
                     └── 不存在 → 幂等创建 bbs-go 用户（同步基本信息）
                              │
                              ▼
                         建立会话，继续访问
```

### 2.3 代码改动

#### 必须改动（以 bbs-go 新版目录结构为准）

- `internal/middleware/auth_middleware.go`：在请求入口处接入 Go 服务端 JWT（从 Header/Cookie 提取），校验后写入 `CurrentUser`。
- `internal/services/user_service.go`：新增“按 `user_id` 幂等创建/更新用户”的入口（保证并发安全与唯一约束）。
- `internal/services/user_token_service.go`：如 bbs-go 内部仍需要 token/cookie 维持会话，则在此处提供“为已认证用户生成/刷新会话 token”的封装。

#### 可选改动（按需）

- `internal/controllers/api/config_controller.go`（前端自维护时必须）：确保 `/api/config/configs` 下发 `baseURL` 与 `scriptInjections` 等字段，前端可消费。

### 2.4 配置项

在 bbs-go 的配置中增加“JWT 校验所需信息”，并明确不启用 bbs-go 自带登录入口：

```yaml
Auth:
  GoJwt:
    Secret: ${JWT_SECRET}
    Issuer: ${JWT_ISSUER}
```

### 2.5 静态资源（已定）

- 采用新版 `/res` 静态资源挂载（`res/images/...` 等由 bbs-go 服务端公开）。

### 2.6 配置下发必须适配项

- `baseURL`：用于生成绝对链接，避免部署在反代/多域名下出现链接漂移。
- `scriptInjections`：用于 `<head>` 脚本注入（统计/埋点等）。

### 2.7 任务中心范围（已定）

- 任务中心可以接入，但奖励仅积分（Score）；勋章与等级明确不做。
- 详见 [bbs-go任务中心范围.md](./bbs-go任务中心范围.md)。

---

## 3. 数据关联

### 3.1 关联层级

| 层级 | 关联程度 | 实现方式 | 示例 |
|------|----------|----------|------|
| 身份关联 | 强 | 共享 user_id（JWT） | 登录认证 |
| 展示关联 | 弱 | API 调用（实时） | 论坛显示战绩 |
| 状态同步 | 中 | 事件订阅（异步） | 资格状态同步 |
| 业务联动 | 强 | API 调用（事务） | 活动发放道具 |

### 3.2 API 调用（展示关联）

#### 获取用户游戏数据

bbs-go 调用 Go 服务端 API：

```
GET https://api.pama1234.tech/api/v1/player/{user_id}/game-stats
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

#### bbs-go 实现示例

新增文件：`internal/services/go_api_client.go`

```go
package services

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"
)

type GoAPIClient struct {
    baseURL    string
    httpClient *http.Client
}

var GoClient = &GoAPIClient{
    httpClient: &http.Client{Timeout: 10 * time.Second},
}

func (c *GoAPIClient) GetGameStats(userId int64) (*GameStats, error) {
    url := fmt.Sprintf("%s/api/v1/player/%d/game-stats", c.baseURL, userId)
    
    resp, err := c.httpClient.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var stats GameStats
    if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
        return nil, err
    }
    
    return &stats, nil
}

type GameStats struct {
    UserId       int64   `json:"user_id"`
    Wins         int     `json:"wins"`
    Losses       int     `json:"losses"`
    Rank         string  `json:"rank"`
    WinRate      float64 `json:"win_rate"`
    TotalMatches int     `json:"total_matches"`
}
```

#### 在用户渲染中使用

修改文件：`internal/controllers/render/user_render.go`

```go
func BuildUserProfile(user *models.User) *UserProfile {
    profile := &UserProfile{
        Id:          user.Id,
        Nickname:    user.Nickname,
        Avatar:      user.Avatar,
        Description: user.Description,
        TopicCount:  user.TopicCount,
        CommentCount: user.CommentCount,
    }
    
    if gameStats, err := services.GoClient.GetGameStats(user.Id); err == nil {
        profile.GameStats = &GameStatsInfo{
            Wins:    gameStats.Wins,
            Losses:  gameStats.Losses,
            Rank:    gameStats.Rank,
            WinRate: gameStats.WinRate,
        }
    }
    
    return profile
}
```

### 3.3 事件订阅（状态同步）

#### 事件类型

| 事件 | 说明 | bbs-go 处理 |
|------|------|-------------|
| `user.eligibility_changed` | 资格状态变更 | 更新用户权限 |
| `game.rank_updated` | 段位更新 | 更新展示数据 |
| `user.banned` | 用户封禁 | 禁用账号 |

#### 事件格式

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

#### bbs-go 事件处理

新增文件：`internal/services/event_handler.go`

```go
package services

import "encoding/json"

func HandleGoServerEvent(eventType string, payload []byte) error {
    switch eventType {
    case "user.eligibility_changed":
        return handleEligibilityChanged(payload)
    case "user.banned":
        return handleUserBanned(payload)
    default:
        return nil
    }
}

func handleEligibilityChanged(payload []byte) error {
    var data struct {
        UserId     int64  `json:"user_id"`
        NewStatus  string `json:"new_status"`
    }
    
    if err := json.Unmarshal(payload, &data); err != nil {
        return err
    }
    
    if data.NewStatus == "ineligible" {
        return UserService.RestrictUser(data.UserId, "lost_beta_eligibility")
    }
    
    return nil
}

func handleUserBanned(payload []byte) error {
    var data struct {
        UserId int64  `json:"user_id"`
        Reason string `json:"reason"`
    }
    
    if err := json.Unmarshal(payload, &data); err != nil {
        return err
    }
    
    return UserService.BanUser(data.UserId, data.Reason)
}
```

---

## 4. 业务联动

### 4.1 论坛活动发放游戏道具

#### 流程

```
论坛管理员创建活动
       │
       ▼
用户参与活动，获得奖励
       │
       ▼
bbs-go 调用 Go 服务端 API
POST /api/v1/admin/game/items/grant
       │
       ▼
Go 服务端验证权限，发放道具
       │
       ▼
返回结果给 bbs-go
```

#### API 定义

```
POST https://api.pama1234.tech/api/v1/admin/game/items/grant
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

### 4.2 游戏战绩分享到论坛

#### 流程

```
游戏客户端生成战绩截图/数据
       │
       ▼
调用 bbs-go API 创建帖子
POST /api/topics
Authorization: Bearer <jwt>
       │
       ▼
bbs-go 创建帖子，关联游戏数据
```

---

## 5. 改动清单

### 5.1 必须改动

| 文件 | 改动类型 | 说明 |
|------|----------|------|
| `internal/middleware/auth_middleware.go` | 修改 | 验证 Go 服务端 JWT |
| `internal/pkg/jwt/jwt.go` | 新增 | JWT 验证工具 |
| `internal/services/user_service.go` | 修改 | 新增 CreateFromGoServer 方法 |
| `bbs-go.yaml` | 修改 | 添加 JWT 配置 |

### 5.2 可选改动（按需）

| 文件 | 改动类型 | 说明 |
|------|----------|------|
| `internal/services/go_api_client.go` | 新增 | Go 服务端 API 客户端 |
| `internal/controllers/render/user_render.go` | 修改 | 渲染游戏数据 |
| `internal/services/event_handler.go` | 新增 | 事件处理 |

---

## 6. 部署说明

### 6.1 环境变量

```bash
JWT_SECRET=<与 Go 服务端相同的 JWT 密钥>
GO_SERVER_URL=https://api.pama1234.tech
```

### 6.2 启动顺序

1. Go 服务端（认证中心）先启动
2. bbs-go 后启动，配置指向 Go 服务端

### 6.3 登录页面配置

bbs-go 登录页面重定向到 Go 服务端：

```javascript
// bbs-go 登录页
function redirectToLogin() {
    const redirect = encodeURIComponent(window.location.href);
    window.location.href = `https://pama1234.tech/login?redirect=${redirect}`;
}
```

`redirect` 参数约束（必须满足，避免开放重定向）：

- Go 服务端只能接受站内 path（例如 `/topic/123`），或来自白名单域名的同源 URL
- 若不满足约束：服务端必须忽略 redirect，并回落到固定安全落地点（例如 `/`）
- 禁止通过 redirect 的 query/hash 传递 access_token/JWT

---

## 7. 安全考虑

### 7.1 JWT 验证

- 验证签名（使用与 Go 服务端相同的密钥）
- 验证过期时间
- 验证 issuer（可选）

### 7.2 服务间通信

- bbs-go 调用 Go 服务端 API 时使用 server-to-server token
- 该 token 由 Go 服务端签发，仅限服务间调用

### 7.3 OAuth 与 Token 落地

OAuth 路由与 token 落地策略以 [账号系统设计.md](./账号系统设计.md) 为准。约束：

- OAuth callback 结束后不得把 token 放到 URL（query/hash）
- bbs-go 侧只依赖浏览器正常携带的认证信息（例如 Cookie），不从 URL 中解析 token

### 7.4 用户权限

- bbs-go 用户权限独立于游戏权限
- 但可通过事件同步更新（如资格变更）

---

## 相关文档

- [账号系统设计.md](./账号系统设计.md) - 账号系统整体设计
- [Go服务端集成方案.md](./Go服务端集成方案.md) - Go 服务端架构
- [身份体系设计.md](./身份体系设计.md) - 身份主体设计
- [bbs-go工程集成约束.md](./bbs-go工程集成约束.md) - 子模块与前端维护策略
- [bbs-go任务中心范围.md](./bbs-go任务中心范围.md) - 任务中心（仅积分）范围
