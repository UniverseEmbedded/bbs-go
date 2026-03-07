# PocketBase集成方案

PocketBase 在本项目中的定位是“运营查看台”：只读展示与检索，不作为账号/资格等权威数据源，也不作为写入口。权威写入路径仅允许通过 Go 服务端 API。

相关文档：
- [`账号系统设计.md`](账号系统设计.md)
- [`Go服务端集成方案.md`](Go服务端集成方案.md)
- [`GoAdmin运营后台方案.md`](GoAdmin运营后台方案.md)
- [`后台与查看台验收清单.md`](后台与查看台验收清单.md)

---

## 目标与非目标

### 目标

- 为运营提供快速检索、筛选、导出与聚合视图
- 支持用最小工程量建设“只读报表/看板”，减少自研后台页面压力
- 通过可控的数据投影机制让 PocketBase 拥有可查询数据，但不破坏权威边界

### 非目标

- PocketBase 不参与登录、注册、JWT 签发与验证
- PocketBase 不直接写入权威业务表（users/beta_eligibility/invite_codes/audit_log 等）

---

## 数据来源与投影策略

PocketBase 想要“能查到数据”，需要一个投影来源。推荐只支持以下两类，并明确“最终一致”语义。

| 策略 | 触发 | 一致性 | 推荐场景 |
|------|------|--------|----------|
| 定时投影 | cron 定时拉取/导出 | 最终一致 | 早期先跑通，数据量中小 |
| 事件驱动投影 | Go 服务端发布事件，投影器消费 | 更接近实时（仍最终一致） | 后期追求更低延迟、更可控回放 |

约束：
- 投影器是唯一写 PocketBase 数据的路径
- 投影数据必须可重放与幂等（以 `source_id`/`updated_at`/版本号作为幂等键）
- 投影字段必须遵守脱敏白名单（见下）

---

## Collection 规划（只读）

PocketBase 的 collection 仅用于展示与检索。建议以“权威表的只读镜像 + 少量聚合视图”组织。

| collection | 来源 | 主键 | 说明 |
|-----------|------|------|------|
| users_view | Go.users | user_id | 只读镜像；不包含敏感字段 |
| eligibility_view | Go.beta_eligibility | user_id | 资格状态展示 |
| invites_view | Go.invite_codes | code_id | 邀请码生命周期展示（脱敏） |
| audit_view | Go.audit_log | audit_id | 审计日志只读查询与导出 |

聚合视图（可选）：
- `eligibility_summary_daily`：每日 eligible/ineligible 统计
- `invites_summary`：邀请码发放/兑换转化率

---

## 脱敏与字段白名单

PocketBase 作为查看台必须默认最小化数据，建议采用“白名单字段”而不是“黑名单剔除”。

最低白名单建议：
- users_view：`user_id`、`display_name`、`status`、`created_at`、`last_login_at`、`qq_bound`（如需要）
- eligibility_view：`user_id`、`status`、`reason_code`、`updated_at`
- invites_view：`code_id`、`status`、`created_at`、`redeemed_at`、`redeemed_user_id`（若展示需脱敏）
- audit_view：`audit_id`、`actor_user_id`、`action`、`target_user_id`、`created_at`、`request_id`

明确禁止：
- 密码 hash、OAuth refresh token、HMAC secret、任何可用于登录/冒用的凭证
- 完整邀请码明文（如需展示，必须做部分脱敏）

---

## 访问与部署约束

- 默认只允许内网或管理员域名访问
- PocketBase 的对外暴露仅限“读查询与管理 UI”，且必须有访问控制（反代 BasicAuth/SSO 等）
- 任何对 PocketBase 的写入只允许投影器（网络隔离 + 凭证隔离）

---

## 验收

以 [`后台与查看台验收清单.md`](后台与查看台验收清单.md) 为准。

