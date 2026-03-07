# Loomio 源码分析

> 【需要详细讨论】

本文档记录对 Loomio 源码的技术分析，为社区平台投票功能设计提供参考。

---

## 一、项目概述

- **仓库地址**：https://github.com/loomio/loomio
- **技术栈**：Ruby on Rails + Vue 3 + PostgreSQL
- **许可证**：AGPL-3.0
- **本地路径**：`temp/loomio-master/`

---

## 二、核心数据模型

### 2.1 模型关系图

```
Group (群组)
├── Membership (成员关系) × N
│   └── User (用户)
├── Discussion (讨论) × N
│   ├── Comment (评论) × N
│   ├── Poll (投票) × N
│   │   ├── PollOption (选项) × N
│   │   ├── Stance (立场) × N
│   │   │   └── StanceChoice (选项选择) × N
│   │   └── Outcome (结果声明)
│   └── Event (事件) × N
└── Event (事件) × N
```

### 2.2 核心模型详解

#### Group（群组）

路径：`app/models/group.rb`

```
字段：
- name: 名称
- description: 描述
- group_privacy: 隐私设置 (open/secret/closed)
- parent_id: 父群组（支持层级）

关联：
- memberships: 成员关系
- discussions: 讨论
- polls: 投票
```

#### Discussion（讨论）

路径：`app/models/discussion.rb`

```
字段：
- title: 标题
- description: 描述
- closed_at: 关闭时间

关联：
- group: 所属群组
- comments: 评论
- polls: 投票
- events: 事件流
- discussion_readers: 阅读状态
```

#### Poll（投票）

路径：`app/models/poll.rb`

```
字段：
- title: 标题
- details: 详情
- poll_type: 投票类型
- closing_at: 截止时间
- closed_at: 关闭时间
- opened_at: 开启时间
- anonymous: 是否匿名
- hide_results: 隐藏结果策略
- stance_reason_required: 是否要求填写理由

状态判断：
- wip? — 草稿（无截止时间）
- scheduled? — 已计划（有开启时间但未到）
- active? — 进行中
- closed? — 已关闭
```

#### Stance（立场）

路径：`app/models/stance.rb`

```
字段：
- reason: 理由
- cast_at: 投票时间
- latest: 是否最新（支持修改投票）
- option_scores: 选项得分（JSON）

关联：
- poll: 所属投票
- participant: 投票者
- stance_choices: 选项选择
```

#### Event（事件）

路径：`app/models/event.rb`

```
字段：
- kind: 事件类型
- eventable_type: 主体类型（多态）
- eventable_id: 主体ID
- parent_id: 父事件
- sequence_id: 序列号
- position: 位置
- position_key: 位置键（用于嵌套）
- depth: 深度

事件类型（40+种）：
- new_discussion, discussion_edited, discussion_closed
- new_comment, comment_edited
- poll_created, poll_edited, poll_closed_by_user
- stance_created, stance_updated
- outcome_created, outcome_announced
- user_joined_group, membership_created
- user_mentioned
```

---

## 三、投票类型定义

配置文件：`config/poll_types.yml`

### 3.1 类型列表

| 类型 | 图标 | 说明 |
|------|------|------|
| proposal | mdi-thumbs-up-down | 提案投票（四象限） |
| count | mdi-human-greeting-variant | 简单计数 |
| check | mdi-head-question | 检查投票 |
| question | chat-question | 问题收集 |
| meeting | mdi-calendar-multiselect | 会议时间选择 |
| poll | mdi-check | 多选投票 |
| dot_vote | mdi-chart-bar-stacked | 点数投票 |
| score | mdi-chart-gantt | 评分投票 |
| ranked_choice | mdi-format-list-numbered | 排名投票 |

### 3.2 四象限投票选项

`proposal` 类型的 `common_poll_options`：

| 选项 | 图标 | 含义 |
|------|------|------|
| agree | agree | 同意 |
| abstain | abstain | 弃权 |
| disagree | disagree | 不同意 |
| block | block | 阻止 |
| consent | agree | 认可 |
| objection | disagree | 反对 |
| veto | block | 否决 |

---

## 四、权限系统

### 4.1 架构

路径：`app/models/ability/base.rb`

使用 CanCanCan 库，模块化设计：

```ruby
class Ability::Base
  include CanCan::Ability
  prepend Ability::Comment
  prepend Ability::Discussion
  prepend Ability::Poll
  prepend Ability::Group
  prepend Ability::User
  # ... 20+ 模块
end
```

### 4.2 权限模块示例

投票权限（`app/models/ability/poll.rb`）：

| 权限 | 条件 |
|------|------|
| :vote_in | 登录 + 进行中 + 是成员或被邀请者 |
| :show | 可见性检查 |
| :create | 管理员或成员（如果群组允许） |
| :update | 管理员 + 未关闭 |
| :close | 进行中 + 管理员 |
| :reopen | 已关闭 + 非匿名 + 管理员 |

群组权限（`app/models/ability/group.rb`）：

| 权限 | 条件 |
|------|------|
| :show | 公开群组或成员 |
| :update | 管理员 |
| :add_members | 管理员或成员（如果群组允许） |
| :join | 公开 + 邮箱验证 + 允许申请 |

---

## 五、服务层

### 5.1 PollService

路径：`app/services/poll_service.rb`

核心方法：

| 方法 | 功能 |
|------|------|
| create | 创建投票，初始化立场 |
| update | 更新投票，处理选项变更 |
| invite | 邀请投票者 |
| remind | 发送提醒 |
| close | 关闭投票 |
| reopen | 重开投票 |
| calculate_results | 计算结果 |
| create_stances | 批量创建立场 |

### 5.2 结果计算逻辑

```ruby
def self.calculate_results(poll, poll_options)
  # 按得分排序
  sorted_poll_options = poll_options.sort_by { |o| -(o.total_score) }
  
  # 计算各项指标
  sorted_poll_options.map do |option|
    {
      score_percent: (option.total_score / poll.total_score) * 100,
      voter_percent: (option.voter_count / poll.voters_count) * 100,
      average: option.average_score,
      test_result: 是否达到阈值
    }
  end
end
```

---

## 六、事件系统

### 6.1 事件发布流程

```ruby
# 1. 创建事件
event = Event.build(eventable, **args)

# 2. 保存并异步发布
event.save!
PublishEventWorker.perform_async(event.id)

# 3. 触发回调
def trigger!
  EventBus.broadcast("#{kind}_event", self)
end
```

### 6.2 通知触发

事件通过 Concern 模块触发通知：

```
app/models/concerns/events/notify/
├── author.rb      # 通知作者
├── mentions.rb    # 通知被提及者
├── subscribers.rb # 通知订阅者
├── by_email.rb    # 邮件通知
├── in_app.rb      # 站内通知
└── chatbots.rb    # Chatbot 通知
```

---

## 七、前端架构

### 7.1 技术栈

- Vue 3
- Vuetify（UI 组件库）
- vue-i18n（国际化）
- TipTap（富文本编辑器）

### 7.2 目录结构

```
vue/src/
├── components/           # 组件
│   ├── poll/            # 投票组件
│   ├── discussion/      # 讨论组件
│   ├── group/           # 群组组件
│   └── strand/          # 时间线组件
├── shared/
│   ├── models/          # 前端模型
│   ├── record_store/    # 数据存储
│   └── services/        # 服务
└── i18n.js              # 国际化配置
```

### 7.3 RecordStore 模式

路径：`vue/src/shared/record_store/record_store.js`

```javascript
class RecordStore {
  // RESTful 客户端
  remote = new RestfulClient
  
  // 数据集合
  users, groups, discussions, polls, stances, ...
  
  // 导入 API 响应
  importJSON(json) {
    // 自动反序列化并存入对应集合
  }
  
  // 视图查询
  view({name, collections, query}) {
    // 创建派生视图
  }
}
```

---

## 八、API 设计

### 8.1 路由结构

```
/api/v1/
├── boot/
│   ├── site              # 站点配置
│   └── user              # 当前用户
├── groups/               # 群组 CRUD
├── discussions/          # 讨论 CRUD
├── polls/                # 投票 CRUD
│   ├── announce          # 发布通知
│   ├── remind            # 提醒
│   ├── close             # 关闭
│   └── reopen            # 重开
├── stances/              # 立场 CRUD
├── outcomes/             # 结果声明 CRUD
├── events/               # 事件流
└── notifications/        # 通知
```

### 8.2 版本化

支持多版本 API 共存：
- `/api/b1/`
- `/api/b2/`
- `/api/b3/`
- `/api/v1/`

---

## 九、设计模式

### 9.1 Concern 模块化

可复用能力模块：

| Concern | 功能 |
|---------|------|
| HasEvents | 事件关联 |
| HasMentions | @提及功能 |
| Reactable | 反应（点赞等） |
| Translatable | 翻译支持 |
| Searchable | 搜索索引 |
| HasVolume | 通知音量设置 |
| HasTimeframe | 时间范围 |
| HasRichText | 富文本 |

### 9.2 Null Object 模式

避免 nil 检查：

```ruby
NullGroup.new      # 空群组
NullDiscussion.new # 空讨论
NullPoll.new       # 空投票
AnonymousUser.new  # 匿名用户
```

### 9.3 多态关联

广泛使用多态：

```ruby
# 事件主体
belongs_to :eventable, polymorphic: true

# 反应主体
belongs_to :reactable, polymorphic: true

# 评论父级
belongs_to :parent, polymorphic: true
```

---

## 十、可借鉴要点

| 模块 | 借鉴点 |
|------|--------|
| 投票类型 | 多种投票类型、YAML 配置驱动 |
| 四象限投票 | 同意/需要修改/弃权/阻止 |
| 权限系统 | CanCanCan 模块化设计 |
| 事件系统 | 事件驱动、异步发布、嵌套结构 |
| 前端架构 | RecordStore 模式、组件模块化 |
| Concern | 可复用能力模块 |
| Null Object | 避免 nil 检查 |

---

## 相关文档

- [Loomio设计分析.md](./Loomio设计分析.md) — 设计哲学解读
- [社区平台设计.md](./社区平台设计.md) — 社区平台整体设计
