# Koishi 内测门禁与私聊邀请插件设计 v2

本文档定义 NapCat + Koishi 插件如何实现“QQ群门禁（群1 ∧ 非群2）+ 私聊邀请链接发放 + QQ Bot 参与 2FA”的完整方案，并定义与 Go 服务端的 API、鉴权与数据契约。

本文件为 v2 口径，用于覆盖并替代本目录中旧文档里与以下决策冲突的内容：

- 用户主键为自增 ID（而非邮箱）
- 服务端先使用 SQLite
- 玩家登录方式同时支持：账号密码 + Google OAuth
- QQ 机器人负责：门禁信号源（群1/群2）、私聊邀请链接发放、以及可选的登录 2FA 通知通道
- “获得邀请链接 / 注册 / 使用账号”的硬门槛：必须在群1且不在群2
- GoAdmin 与 PocketBase 并排部署；PocketBase 仅用于查看/运营视图，不作为身份权威

---

## 1. 目标与非目标

### 1.1 目标

- 以 QQ 群成员关系为信号，持续计算玩家门禁状态：`eligible := in_group1 && !in_group2`
- 仅对满足门禁的人发放“私聊邀请链接”（不得群聊发放）
- 服务端在关键路径强校验门禁：
  - 发放邀请链接前校验
  - 注册/绑定 QQ 时校验
  - 登录与 token 刷新时校验（不满足即拒绝或吊销）
- 支持两类玩家登录方式并行存在：
  - 账号 + 密码
  - Google OAuth
- 支持 QQ Bot 参与 2FA（第一阶段实现为：服务端下发“发码任务”，Koishi 私聊发一次性验证码；网页/客户端输入验证码完成登录）
- 机器人配置可在 Koishi 控制台 UI 填写（群号等运行期可变）

### 1.2 非目标（本阶段不做）

- 远程联机/房间目录/匹配
- 玩家自助合并账号（仅允许后台执行合并）
- QQ OAuth 登录（不使用 QQ 互联 OAuth）

---

## 2. 术语与身份模型（最小共识）

### 2.1 身份与主体

- PlayerUser（玩家）：业务用户，主键 `user_id`（SQLite 自增整数），可绑定 0..N 种登录凭证
- SystemUser（后台人员）：用于 GoAdmin 等后台操作与审计（不在本插件范围内）
- ServiceAccount（机器人）：Koishi 插件作为服务账号与服务端通讯

### 2.2 门禁与资格

- 门禁（gate）：由 QQ 群关系推导的硬门槛：`in_group1 && !in_group2`
- 资格（eligibility）：服务端存储的玩家状态（可包含：eligible / suspended / banned 等）
- 邀请链接（invite link）：服务端生成的带 token 的注册入口，只能通过私聊发给 QQ 号

门禁与登录方式解耦：门禁依赖 QQ 号；登录可用密码或 Google。玩家无论用哪种登录方式，想“注册/使用账号”都必须通过门禁。

---

## 3. 总体架构与职责

### 3.1 拓扑

```
QQ群事件/成员列表
   │
   ▼
NapCat (OneBot 11)
   │ reverse ws
   ▼
Koishi + gate-invite 插件（ServiceAccount）
   │ HTTPS（HMAC 签名）
   ▼
Go 服务端（权威：账号/门禁/邀请/会话/审计） ── SQLite
   │
   ├─ GoAdmin（后台入口）
   └─ PocketBase（运营查看台：镜像/只读为主）
```

### 3.2 分工原则（硬约束）

- Go 服务端是唯一权威：决定门禁结果、决定是否发放邀请链接、决定是否允许注册/登录/继续使用
- Koishi 插件只做：
  - 观测群成员关系（事件 + 对账）
  - 执行私聊发送（邀请链接/2FA 一次性码）
  - 回传发送结果（用于幂等与审计）

---

## 4. Koishi 插件设计（gate-invite）

### 4.1 插件职责清单

- 群成员观测：
  - 监听“成员加入/退出”事件，形成增量信号
  - 周期性对账：拉取群1/群2成员列表，生成快照并上报
- 任务拉取与执行：
  - 定时从服务端拉取 Outbox 任务（邀请链接私聊、2FA 码私聊）
  - 执行发送并回传 ack（sent/failed + message_id + error）
- 可靠性与风控：
  - 私聊发送节流与排队（避免 QQ 风控）
  - 失败重试（指数退避 + 最大次数）
  - 幂等：同一任务不重复发送；重复请求不产生重复副作用

### 4.2 配置项（需在 Koishi 控制台 UI 可编辑）

建议配置（名称示例，可按 Koishi Schema 表达）：

- `apiBaseUrl`：Go 服务端地址（例如 `https://example.com`）
- `botId`：机器人实例 ID（用于多实例区分与审计）
- `botSecret`：HMAC 密钥（仅 Koishi 与服务端共享）
- `group1Id`：内测群号
- `group2Id`：大群号（门禁要求“不在此群”）
- `syncIntervalMs`：对账间隔（默认 10 分钟）
- `pollIntervalMs`：Outbox 拉取间隔（默认 30 秒）
- `sendRateLimitPerMinute`：私聊发送限额（默认 12/min，具体按风控调整）
- `sendMinIntervalMs`：两次私聊最小间隔（默认 1500ms）
- `enableEventDriven`：是否启用事件驱动增量上报（默认 true）
- `dryRun`：只打印不发送（用于联调）

### 4.3 “机器人非管理员”的能力假设与降级

机器人将以“普通群成员”身份存在。实现上分两层：

- 优先路径：能够调用 OneBot `get_group_member_list` 拉取群成员列表完成对账（推荐）
- 若 OneBot 实现限制导致无法拉取全量成员：
  - 插件仍可监听成员增减事件维持近似集合
  - 但在机器人离线/掉线期间会出现不可恢复的盲区

结论：若无法对账拉取全量成员，则只能做到“尽力而为”，无法满足“必须在群1且不在群2”的严格门禁。此时需要二选一：

- 提升机器人为群管理员（你当前不选）
- 或接受门禁在短时间窗口内可能不严格（不推荐）

---

## 5. 门禁规则与状态机（服务端权威）

### 5.1 门禁硬规则（必须满足）

```
gate_ok(qq) := in_group1(qq) && !in_group2(qq)
```

### 5.2 资格状态（建议最小集合）

- `eligible`：门禁满足，允许发邀请/注册/登录/使用
- `suspended`：门禁不满足，立即禁止登录与使用（可恢复）
- `banned`：封禁（后台操作），不可恢复或需人工恢复

推荐：一旦发现 `gate_ok=false`，服务端将玩家置为 `suspended`，并撤销其 refresh/session（如果实现了 refresh）。

### 5.3 防误判策略（必须可配置）

QQ群事件与成员列表可能短暂不一致（例如缓存、延迟、机器人短暂掉线）。建议：

- 对账快照为主，事件为辅
- “撤回资格”可采用二次确认：
  - 连续 N 次（例如 2 次）对账都显示不满足门禁才从 `eligible` → `suspended`
  - 但对“加入群2”这种高风险信号可立刻触发 `suspended`（可选策略）

---

## 6. 私聊邀请链接发放流程（Outbox 拉取式）

### 6.1 流程概述

1) Koishi 上报群快照/增量
2) Go 服务端计算“可邀请候选”（满足 `gate_ok` 且尚未发放/尚未注册）
3) Go 服务端生成 Outbox 任务（invite）
4) Koishi 拉取任务，私聊发送邀请链接
5) Koishi 回传发送结果（sent/failed）
6) 玩家打开邀请链接进入注册（密码或 Google），注册时绑定该邀请 token 对应的 QQ

### 6.2 邀请链接形式（建议）

```
https://your.domain/register?invite_token=xxxxx
```

约束：

- `invite_token` 必须一次性或短有效期（例如 24h）
- `invite_token` 必须绑定到目标 QQ 号（`issued_to_qq`），并在注册/绑定时二次校验当前门禁仍满足
- 发送文案必须私聊，不得群聊泄露链接

### 6.3 私聊不可达的回退（会话前置）

现实约束：
- 在 NapCat/QQNT 的实现中，机器人对“未建立会话”的目标主动私聊可能失败（常见表现为发送超时），而目标先主动发过消息后成功率明显上升。

策略：
- 私聊发送失败时，不将“邀请链接/一次性码”等秘密内容降级到群聊发送；群聊只用于引导。
- 插件/服务端对失败原因进行分类：若为“会话不可达/发送超时”类，则进入“待用户动作”状态。

引导文案建议（群内 @目标）：
- 提示用户先私聊机器人发送指定口令（例如“邀请”/“绑定”），以建立会话；建立会话后再进行下一次私聊投递。

### 6.4 失败回退与重试（可操作）

目标：
- 私聊失败不应导致流程永久中断；应产出用户可执行的动作，并支持重试。

建议：
- outbox 任务 `failed` 应区分“永久失败”与“可重试失败”：
  - 可重试失败：发送超时、会话不可达、临时网络错误等
  - 永久失败：门禁不满足、目标 QQ 不存在/已注销、账号被封禁等
- 可重试失败进入冷却期后自动重试，或在用户完成“建立会话”动作后触发重试。

### 6.5 备用绑定路径：用户提交网页 token 到机器人

动机：
- 如果机器人主动私聊受限，可改为“用户 → 机器人”的挑战-应答模式完成绑定，仍能实现“网站账号 ↔ QQ 号”强关联。

流程建议：
1) 网站注册页生成一次性 `register_token`（短期有效、未绑定）
2) 用户私聊机器人发送：`bind <register_token>`（或统一命令）
3) 服务端收到 `QQ号 + register_token` 后完成绑定，并使 token 作废

服务端约束（必须）：
- token 一次性、短期有效
- token 首次提交即锁定提交者 QQ，后续同 token 的提交必须为同一 QQ
- 对提交接口做频控与失败次数限制，防止暴力猜测
 结论：此方案可作为私聊投递失败时的备用路径，也可作为默认绑定入口

---

## 7. QQ Bot 参与 2FA（第一阶段：私聊 OTP）

### 7.1 适用场景

当用户使用“账号密码”登录（或某些高风险操作）时，服务端可要求 2FA：

- 服务端创建 outbox 任务（otp），Koishi 私聊发送一次性码
- 用户在网页/客户端输入一次性码完成登录

### 7.2 与门禁的关系

2FA 不替代门禁；它只是“确认该 QQ 号仍由用户掌控”的额外安全层。

门禁依然由 `gate_ok` 决定，且服务端应在：

- 登录第一阶段（密码正确）后立即校验 `gate_ok`
- 通过 2FA 后再次校验 `gate_ok`（避免窗口期变化）

---

## 8. Koishi ⇄ 服务端 API 设计

本节定义 HTTP API（由 Koishi 调用 Go 服务端）。所有接口仅供 ServiceAccount 使用，不对公网用户开放。

### 8.1 鉴权：HMAC 签名（必选）

请求头：

- `X-Bot-Id`: string
- `X-Timestamp`: unix_ms
- `X-Nonce`: string（随机，至少 96 bit）
- `X-Signature`: base64(hmac_sha256(secret, canonical_string))

canonical_string 建议：

```
METHOD \n
PATH \n
X-Timestamp \n
X-Nonce \n
SHA256_HEX(body)
```

服务端校验：

- 时间窗（例如 ±5 分钟）
- nonce 去重（窗口内）
- 签名匹配
- botId 是否启用/是否被吊销

### 8.2 幂等与审计（必选）

所有写入接口请求体都应包含 `request_id`（uuid/ulid）：

- 服务端以 `(bot_id, request_id)` 做幂等去重
- 审计日志必须落：bot_id、request_id、action、目标 QQ、结果、时间

### 8.3 群快照上报

`POST /api/v1/bot/qq/snapshot`

请求：

```json
{
  "request_id": "01J...",
  "observed_at_ms": 1730000000000,
  "group1_id": "123456",
  "group2_id": "234567",
  "group1_members": ["10001", "10002"],
  "group2_members": ["10002", "10003"],
  "snapshot_hash": "sha256-hex-of-normalized"
}
```

响应：

```json
{
  "ok": true,
  "snapshot_id": "01J...",
  "group1_count": 2,
  "group2_count": 2
}
```

说明：

- `snapshot_hash` 用于服务端快速判重；成员列表建议排序后再 hash
- 当成员列表过大时可扩展为“分片上报”（本阶段可不做）

### 8.4 增量事件上报（可选但推荐）

`POST /api/v1/bot/qq/events`

```json
{
  "request_id": "01J...",
  "events": [
    {
      "type": "group_member_added",
      "group_id": "123456",
      "qq": "10001",
      "event_at_ms": 1730000000000
    }
  ]
}
```

### 8.5 Outbox 拉取（邀请/2FA）

`POST /api/v1/bot/outbox/pull`

请求：

```json
{
  "request_id": "01J...",
  "cursor": "",
  "limit": 50,
  "types": ["invite", "otp"]
}
```

响应：

```json
{
  "ok": true,
  "tasks": [
    {
      "task_id": "01J...",
      "type": "invite",
      "to_qq": "10001",
      "content": "你已获得内测资格，请使用此链接注册：https://.../register?invite_token=...",
      "expires_at_ms": 1730000000000
    }
  ],
  "next_cursor": ""
}
```

规则：

- `content` 由服务端生成，Koishi 只负责发送
- Koishi 必须只用私聊发送（OneBot `send_private_msg`）

### 8.6 Outbox 回执

`POST /api/v1/bot/outbox/ack`

```json
{
  "request_id": "01J...",
  "task_id": "01J...",
  "status": "sent",
  "message_id": "123456",
  "sent_at_ms": 1730000000000,
  "error": ""
}
```

`status` 可选：`sent` / `failed`

---

## 9. 服务端侧对接要点（登录、注册、绑定、使用）

本节不是完整的账号系统规格，但给出与 Koishi 门禁联动必须具备的行为。

### 9.1 注册入口与激活（邀请链接优先，但允许无邀请受限登录）

推荐做法：

- 携带 `invite_token` 时：允许创建账号/首登并完成激活
  - `invite_token` 绑定 `issued_to_qq`
  - 兑换后在用户档案中记录 `bound_qq`，并二次校验门禁
- 未携带 `invite_token` 时：允许创建账号/首登，但账号为受限态（例如 `pending_invite`）
  - 允许进入站内与完成激活流程
  - 禁止访问需要内测资格/门禁的资源
  - 直到用户提交 `invite_token` 激活并绑定 `bound_qq` 为止

### 9.2 登录与刷新必须二次校验门禁

无论使用密码还是 Google：

- 登录成功前校验 `gate_ok(bound_qq)`，否则返回 `SUSPENDED_BY_GATE`
- refresh/会话延续同样校验，避免“先登录后退群仍可用”

### 9.3 账号密码 + QQ Bot 2FA（推荐第一阶段实现）

建议接口语义（示意）：

- `POST /api/v1/auth/login`：
  - 密码正确且门禁满足 → 返回 `2fa_required` 与 `challenge_id`
  - 同时服务端创建 outbox 任务（otp）给该用户绑定 QQ
- `POST /api/v1/auth/2fa/verify`：提交 `challenge_id + code` 完成登录

约束：

- 2FA code 必须短有效期（例如 5 分钟）
- 对同一用户/同一 IP 限流，防止爆破

### 9.4 账号绑定与合并（最小规则）

- 允许同一 `user_id` 同时拥有：
  - password credential
  - google identity
- 账号合并只允许后台触发，并记录审计（本阶段不开放玩家自助）

---

## 10. 可靠性、安全与风控

### 10.1 QQ 风控应对（私聊发送）

- 发送频率必须可配置并默认保守（例如每分钟 ≤ 12 条私聊）
- 发送失败应指数退避，不要紧密重试
- 文案避免敏感词与“批量营销”特征；每条消息尽量短且稳定

### 10.2 服务端安全边界

- Bot API 独立限流与独立日志
- 任何来自 bot 的写入必须幂等（request_id）
- Bot secret 必须可轮换；轮换期间支持双密钥窗口

### 10.3 审计要求（最少字段）

- 谁触发：bot_id
- 对谁：target qq / user_id（若已绑定）
- 做了什么：snapshot_ingested / invite_sent / otp_sent / eligibility_changed
- 结果：ok/failed + 错误码
- request_id 与时间戳

---

## 11. 实施顺序（建议）

1) 服务端实现 Bot API：鉴权、幂等、outbox、快照写入、门禁计算
2) Koishi 插件实现：
   - 配置项
   - outbox 拉取 + 私聊发送 + ack
   - 群快照对账上报
3) 服务端实现邀请链接注册：
   - invite_token 发行、兑换与绑定 QQ
4) 服务端实现登录态：
   - 密码登录 + Google 登录并行
   - 登录/刷新门禁校验
5) 加入 QQ Bot 2FA（otp outbox 类型）

