# 阶段四：基于Loomio设计改造投票系统

## 目标

参考Loomio的投票系统设计，改造BBS-Go现有的投票功能

## 依据

- 规划01-B/Loomio设计分析.md
- 规划01-B/Loomio源码分析.md

---

## 4.1 数据模型重构

### 4.1.1 扩展Vote表

| 字段名 | 类型 | 说明 |
|--------|------|------|
| poll_type | string | 投票类型：proposal/poll/count等 |
| hide_results | int | 结果可见性：0=off, 1=until_vote, 2=until_closed |
| anonymous | bool | 是否匿名投票 |
| opening_at | *time.Time | 开始时间 |
| opened_at | *time.Time | 实际开启时间 |
| closing_at | *time.Time | 截止时间 |
| closed_at | *time.Time | 实际关闭时间 |
| voter_can_add_options | bool | 投票者是否能添加选项 |
| specified_voters_only | bool | 是否仅限指定用户投票 |
| stance_reason_required | int | 是否必须填写理由：0=禁用, 1=可选, 2=必填 |
| quorum_pct | *int | 法定人数百分比 |
| versions_count | int | 版本数 |

### 4.1.2 扩展VoteOption表

| 字段名 | 类型 | 说明 |
|--------|------|------|
| icon | *string | 选项图标 |
| meaning | *string | 选项语义 |
| prompt | *string | 选择时的提示 |
| priority | int | 排序优先级 |
| total_score | int | 总得分（缓存） |
| voter_count | int | 投票人数（缓存） |

### 4.1.3 新建Stance表（替代VoteRecord）

| 字段名 | 类型 | 说明 |
|--------|------|------|
| poll_id | bigint | 关联的投票 |
| participant_id | bigint | 投票用户 |
| reason | text | 投票理由 |
| reason_format | string | 理由格式 |
| latest | bool | 是否最新立场 |
| cast_at | *time.Time | 投票时间 |
| revoked_at | *time.Time | 撤销时间 |
| revoker_id | bigint | 撤销者ID |
| admin | bool | 是否管理员邀请 |
| guest | bool | 是否被邀请的访客 |
| inviter_id | bigint | 邀请者ID |
| option_scores | jsonb | 选项得分 |
| none_of_the_above | bool | 是否选择"以上都不是" |

### 4.1.4 新建StanceChoice表

| 字段名 | 类型 | 说明 |
|--------|------|------|
| stance_id | bigint | 关联的立场 |
| poll_option_id | bigint | 关联的选项 |
| score | int | 得分 |

### 4.1.5 新建Outcome表（结果声明）

| 字段名 | 类型 | 说明 |
|--------|------|------|
| poll_id | bigint | 关联的投票 |
| statement | text | 声明内容 |
| statement_format | string | 内容格式 |
| author_id | bigint | 发布者 |
| poll_option_id | *bigint | 关联的获胜选项 |
| latest | bool | 是否最新声明 |
| review_on | *date | 复查日期 |
| custom_fields | jsonb | 扩展字段 |

### 4.1.6 后端API

| API | 方法 | 说明 |
|-----|------|------|
| /api/stance | POST | 创建立场 |
| /api/stance/:id | PUT | 修改立场 |
| /api/stance/:id/revoke | POST | 撤销立场 |
| /api/stance/latest | GET | 获取当前用户最新立场 |
| /api/stance/options | GET | 获取投票所有立场 |
| /api/outcome | POST | 创建结果声明 |
| /api/outcome/:id | PUT | 修改结果声明 |
| /api/outcome/:pollId | GET | 获取投票结果声明 |

---

## 4.2 四象限投票实现

### 4.2.1 投票类型配置

- 配置文件：internal/config/poll_types.go
- 支持类型：proposal（四象限）、poll（普通）、count、ranked_choice等

### 4.2.2 立场可修改逻辑

- 首次投票 → 创建Stance，latest=true
- 再次投票 → 旧Stance的latest=false，新建Stance，latest=true
- 撤销 → revoked_at=当前时间，latest=false

### 4.2.3 结果可见性策略

| hide_results值 | 名称 | 规则 |
|----------------|------|------|
| 0 | off | 始终可见 |
| 1 | until_vote | 投票后或关闭后可见 |
| 2 | until_closed | 关闭后可见 |

---

## 4.3 前端与Admin补全

### 4.3.1 Site前端新增

| 组件 | 说明 |
|------|------|
| VoteProposal.vue | 四象限投票组件 |
| VotePoll.vue | 普通投票组件 |
| VoteResult.vue | 投票结果展示 |
| StanceHistory.vue | 投票历史/修改记录 |
| OutcomeStatement.vue | 结果声明展示 |

### 4.3.2 Admin新增

| 页面 | 说明 |
|------|------|
| /vote | 投票列表（扩展） |
| /vote/:id | 投票详情编辑 |
| /outcome | 结果声明管理 |

---

## 涉及文件清单

### 后端

- internal/models/models.go（扩展Vote/VoteOption，新增Stance/StanceChoice/Outcome）
- internal/install/install.go（建表脚本）
- internal/config/poll_types.go（投票类型配置）
- internal/server/router.go（注册路由）
- internal/controllers/api/stance_controller.go（新建）
- internal/controllers/api/outcome_controller.go（新建）
- internal/services/stance_service.go（新建）
- internal/services/outcome_service.go（新建）
- internal/controllers/render/vote_render.go（扩展）

### 前端

- site/components/vote/VoteProposal.vue（新建）
- site/components/vote/VotePoll.vue（新建）
- site/components/vote/VoteResult.vue（新建）
- site/components/vote/StanceHistory.vue（新建）
- site/components/vote/OutcomeStatement.vue（新建）
- admin/src/views/vote/index.vue（扩展）
- admin/src/views/outcome/index.vue（新建）

---

## 待补充（阶段四启动时细化）

- 实施步骤
- 验收检查项
