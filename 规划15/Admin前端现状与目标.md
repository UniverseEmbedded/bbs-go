# Admin前端现状与目标

本文档记录Admin管理后台的当前状态与目标，基于代码分析得出。

---

## 一、技术栈

| 项目 | 版本/名称 |
|------|-----------|
| 框架 | Vue 3 |
| UI库 | Arco Design |
| 路由 | Vue Router |
| HTTP | Axios |

---

## 二、当前已实现的页面

### 页面清单

| 页面路径 | 说明 | 对应API |
|----------|------|---------|
| /dashboard | 仪表盘 | /api/admin/common/system_info |
| /user | 用户管理 | /api/admin/user/* |
| /topic | 话题管理 | /api/admin/topic/* |
| /topic-node | 话题节点 | /api/admin/topic-node/* |
| /article | 文章管理 | /api/admin/article/* |
| /link | 链接管理 | /api/admin/link/* |
| /forbidden-word | 敏感词 | /api/admin/forbidden-word/* |
| /system/role | 角色管理 | /api/admin/role/* |
| /system/menu | 菜单管理 | /api/admin/menu/* |
| /system/api | 接口管理 | /api/admin/api/* |
| /system/permission | 权限管理 | /api/admin/role/*, /api/admin/menu/* |
| /system/dict | 字典管理 | /api/admin/dict/*, /api/admin/dict-type/* |
| /settings | 系统设置 | /api/admin/sys-config/* |

---

## 三、当前已对接的API

### 按模块分类

| 模块 | 已对接的API |
|------|-------------|
| 用户 | /api/admin/user/list, /api/admin/user/:id, /api/admin/user/create, /api/admin/user/update |
| 话题 | /api/admin/topic/list, /api/admin/topic/delete, /api/admin/topic/undelete, /api/admin/topic/recommend, /api/admin/topic/audit |
| 话题节点 | /api/admin/topic-node/list, /api/admin/topic-node/:id, /api/admin/topic-node/create, /api/admin/topic-node/update |
| 文章 | /api/admin/article/list, /api/admin/article/delete, /api/admin/article/audit |
| 链接 | /api/admin/link/list, /api/admin/link/:id, /api/admin/link/create, /api/admin/link/update |
| 敏感词 | /api/admin/forbidden-word/list, /api/admin/forbidden-word/:id, /api/admin/forbidden-word/create, /api/admin/forbidden-word/update |
| 角色 | /api/admin/role/list, /api/admin/role/roles, /api/admin/role/:id, /api/admin/role/create, /api/admin/role/update, /api/admin/role/update_sort, /api/admin/role/all_roles, /api/admin/_menus, /api/admin/role/role_menu_ids, /api/admin/role/save_role_menus |
| 菜单 | /api/admin/menu/list, /api/admin/menu/tree, /api/admin/menu/:id, /api/admin/menu/create, /api/admin/menu/update, /api/admin/menu/update_sort |
| 接口 | /api/admin/api/list, /api/admin/api/list_all, /api/admin/api/:id, /api/admin/api/create, /api/admin/api/update |
| 字典类型 | /api/admin/dict-type/list, /api/admin/dict-type/:id, /api/admin/dict-type/create, /api/admin/dict-type/update |
| 字典 | /api/admin/dict/list, /api/admin/dict/dicts, /api/admin/dict/:id, /api/admin/dict/create, /api/admin/dict/update, /api/admin/dict/update_sort |
| 系统配置 | /api/admin/sys-config/configs, /api/admin/sys-config/save |
| 仪表盘 | /api/admin/common/system_info |

---

## 四、缺失的页面

### 后端有但Admin前端无对应页面的模块（共16个）

| 模块 | 后端API | 说明 |
|------|---------|------|
| 任务配置 | /api/admin/task-config | 任务配置管理 |
| 徽章管理 | /api/admin/badge | 徽章配置 |
| 等级配置 | /api/admin/level-config | 等级经验配置 |
| 用户任务日志 | /api/admin/user-task-log | 用户完成任务记录 |
| 用户经验日志 | /api/admin/user-exp-log | 用户经验变动记录 |
| 用户徽章 | /api/admin/user-badge | 用户拥有的徽章 |
| 投票管理 | /api/admin/vote | 投票列表与编辑 |
| 投票选项 | /api/admin/vote-option | 投票选项管理 |
| 投票记录 | /api/admin/vote-record | 用户投票记录 |
| 第三方用户 | /api/admin/third-user | 微信/谷歌账号绑定 |
| 用户关注 | /api/admin/user-follow | 用户关注关系 |
| 用户动态 | /api/admin/user-feed | 用户发布的内容流 |
| 消息管理 | /api/admin/message | 系统消息管理 |
| 签到管理 | /api/admin/check-in | 签到记录 |
| 邮箱验证码 | /api/admin/email-code | 验证码记录 |
| 短信验证码 | /api/admin/sms-code | 短信验证码记录 |

---

## 五、目标

基于「API对接缺口与对比.md」的分析，Admin前端的目标为：

本文件仅维护 Admin 缺口台账，不属于阶段二（Site补全）的实施范围。

新增以下16个管理页面：

1. 任务配置管理 - 对接 /api/admin/task-config
2. 徽章管理 - 对接 /api/admin/badge
3. 等级配置 - 对接 /api/admin/level-config
4. 用户任务日志 - 对接 /api/admin/user-task-log
5. 用户经验日志 - 对接 /api/admin/user-exp-log
6. 用户徽章 - 对接 /api/admin/user-badge
7. 投票管理 - 对接 /api/admin/vote
8. 投票选项 - 对接 /api/admin/vote-option
9. 投票记录 - 对接 /api/admin/vote-record
10. 第三方用户 - 对接 /api/admin/third-user
11. 用户关注 - 对接 /api/admin/user-follow
12. 用户动态 - 对接 /api/admin/user-feed
13. 消息管理 - 对接 /api/admin/message
14. 签到管理 - 对接 /api/admin/check-in
15. 邮箱验证码 - 对接 /api/admin/email-code
16. 短信验证码 - 对接 /api/admin/sms-code
