# QQ机器人集成方案

本文档定义使用NapCat + Koishi实现内测资格管理的方案。

---

## 技术选型

| 组件 | 选择 | 理由 |
|------|------|------|
| 协议层 | NapCat | 基于NTQQ的OneBot 11实现 |
| 框架 | Koishi | 插件化、有OneBot适配器、Web控制台 |
| 通信方式 | reverse ws | 部署安全更好 |

---

## 门禁规则

### 硬规则

```
gate_ok = in_group1 AND NOT in_group2
```

| 条件 | 说明 |
|------|------|
| 群1 | 内测群（必须在内） |
| 群2 | 大群（必须不在内） |

### 状态转换

| 条件 | 用户状态 |
|------|----------|
| gate_ok = true | eligible（允许注册/登录/使用） |
| gate_ok = false | suspended（禁止登录/使用） |
| 后台封禁 | banned |

---

## 业务场景

### 邀请发放流程

1. Koishi定时上报群快照到服务端
2. 服务端计算可邀请候选（in_group1 && !in_group2 && 未注册）
3. 服务端生成invite_token，写入bot_outbox任务
4. Koishi拉取任务，私聊发送邀请链接
5. Koishi回传发送结果

### 门禁检查流程

1. 用户登录时，服务端检查gate_ok
2. Koishi定时上报群快照
3. 服务端更新所有用户的门禁状态
4. 连续2次不满足门禁才suspended

### 2FA流程

1. 用户密码登录成功
2. 服务端创建OTP任务（写入bot_outbox）
3. Koishi拉取任务，私聊发送验证码
4. 用户输入验证码
5. 服务端验证，签发token

---

## 架构设计

### 分工原则

- **服务端（权威）**：决定资格状态、存储数据、审计日志、生成任务
- **机器人（执行者）**：群成员上报、私聊发送、回传结果

### 数据流

```
QQ群 → NapCat → Koishi → Bot API → 服务端数据库
                    ↓
              私聊发送消息
```

---

## Bot API设计

详见 [Koishi内测门禁与私聊邀请插件设计v2.md](./Koishi内测门禁与私聊邀请插件设计v2.md)

### 认证方式

使用HMAC签名：

```
X-Bot-Id: koishi-bot
X-Timestamp: 1730000000
X-Signature: hmac-sha256签名值
```

### 核心端点

| 端点 | 方法 | 说明 |
|------|------|------|
| /api/v1/bot/qq/snapshot | POST | 上报群快照 |
| /api/v1/bot/qq/events | POST | 上报增量事件 |
| /api/v1/bot/outbox/pull | POST | 拉取待发送任务 |
| /api/v1/bot/outbox/ack | POST | 回传发送结果 |

---

## Koishi插件结构

详细实现见 [Koishi内测门禁与私聊邀请插件设计v2.md](./Koishi内测门禁与私聊邀请插件设计v2.md)

### 插件拆分

```
plugins/
├── duel-gate/              # 门禁与邀请插件
│   ├── src/
│   │   ├── index.ts        # 主入口
│   │   ├── snapshot.ts     # 群快照上报
│   │   ├── outbox.ts       # 任务拉取与执行
│   │   ├── commands.ts     # 管理员命令
│   │   └── webhooks.ts     # 事件监听
│   └── package.json
```

---

## 管理员命令

| 命令 | 说明 |
|------|------|
| /gate status <qq> | 查询用户门禁状态 |
| /gate check <qq> | 立即检查指定用户 |
| /gate sync | 立即同步群快照 |
| /gate resend <qq> | 重新发送邀请 |

---

## 防误封策略

### 连续检测确认

- 连续2次对账都显示不满足门禁才从 eligible → suspended
- 对"加入群2"可立刻触发 suspended（可配置）

### 申诉机制

- 用户可通过网页申诉
- 管理员可手动恢复

### 日志追溯

- 所有资格变更写入audit_logs
- 包含检测时间、原因

---

## 部署配置

### NapCat配置

1. 安装NapCat并登录NTQQ
2. 配置reverse ws：
```json
{
  "reverseWs": {
    "enable": true,
    "urls": ["ws://127.0.0.1:5140/onebot/v11/ws"]
  }
}
```

### Koishi配置

1. 安装adapter-onebot插件
2. 配置连接：
```yaml
plugins:
  adapter-onebot:
    protocol: ws-reverse
    selfId: "机器人QQ号"
```

3. 安装duel-gate插件并配置

---

## 风险与应对

| 风险 | 应对 |
|------|------|
| NapCat协议变更 | 关注更新、预留降级方案 |
| QQ风控 | 控制发送频率、避免敏感词 |
| 群成员接口不稳定 | 缓存结果、重试机制 |
| 误封用户 | 连续检测确认、申诉入口 |
| 机器人掉线 | 服务端检测心跳、告警 |

---

## 相关文档

- [账号系统设计.md](./账号系统设计.md) - 账号系统整体设计
- [身份体系设计.md](./身份体系设计.md) - 身份主体设计
- [Koishi内测门禁与私聊邀请插件设计v2.md](./Koishi内测门禁与私聊邀请插件设计v2.md) - Koishi插件详细设计
