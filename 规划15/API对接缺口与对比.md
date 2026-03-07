# API对接缺口与对比

本文档记录后端API与前端对接情况的对比，基于代码分析得出。

---

## 一、后端API模块清单

### 用户端API（/api/*）

| 模块 | Controller文件 | 说明 |
|------|----------------|------|
| /api/install | install_controller.go | 安装 |
| /api/topic | topic_controller.go | 话题 |
| /api/article | article_controller.go | 文章 |
| /api/login | login_controller.go | 登录 |
| /api/user | user_controller.go | 用户 |
| /api/tag | tag_controller.go | 标签 |
| /api/comment | comment_controller.go | 评论 |
| /api/favorite | favorite_controller.go | 收藏 |
| /api/like | like_controller.go | 点赞 |
| /api/checkin | checkin_controller.go | 签到 |
| /api/config | config_controller.go | 配置 |
| /api/upload | upload_controller.go | 上传 |
| /api/link | link_controller.go | 链接 |
| /api/captcha | captcha_controller.go | 验证码 |
| /api/search | search_controller.go | 搜索 |
| /api/fans | fans_controller.go | 粉丝/关注 |
| /api/user-report | user_report_controller.go | 用户举报 |
| /api/task | task_controller.go | 任务 |
| /api/badge | badge_controller.go | 徽章 |
| /api/vote | vote_controller.go | 投票 |

### 管理端API（/api/admin/*）

| 模块 | Controller文件 | 说明 |
|------|----------------|------|
| /api/admin/role | role_controller.go | 角色 |
| /api/admin/menu | menu_controller.go | 菜单 |
| /api/admin/api | api_controller.go | 接口 |
| /api/admin/dict-type | dict_type_controller.go | 字典类型 |
| /api/admin/dict | dict_controller.go | 字典 |
| /api/admin/email-log | email_log_controller.go | 邮件日志 |
| /api/admin/common | common_controller.go | 通用 |
| /api/admin/user | user_controller.go | 用户管理 |
| /api/admin/tag | tag_controller.go | 标签管理 |
| /api/admin/article | article_controller.go | 文章管理 |
| /api/admin/comment | comment_controller.go | 评论管理 |
| /api/admin/favorite | favorite_controller.go | 收藏管理 |
| /api/admin/article-tag | article_tag_controller.go | 文章标签 |
| /api/admin/topic | topic_controller.go | 话题管理 |
| /api/admin/topic-node | topic_node_controller.go | 话题节点 |
| /api/admin/sys-config | sys_config_controller.go | 系统配置 |
| /api/admin/link | link_controller.go | 链接管理 |
| /api/admin/user-score-log | user_score_log_controller.go | 用户积分日志 |
| /api/admin/task-config | task_config_controller.go | 任务配置 |
| /api/admin/badge | badge_controller.go | 徽章管理 |
| /api/admin/level-config | level_config_controller.go | 等级配置 |
| /api/admin/user-task-log | user_task_log_controller.go | 用户任务日志 |
| /api/admin/user-exp-log | user_exp_log_controller.go | 用户经验日志 |
| /api/admin/user-badge | user_badge_controller.go | 用户徽章 |
| /api/admin/operate-log | operate_log_controller.go | 操作日志 |
| /api/admin/user-report | user_report_controller.go | 用户举报管理 |
| /api/admin/forbidden-word | forbidden_word_controller.go | 敏感词 |
| /api/admin/vote | vote_controller.go | 投票管理 |
| /api/admin/vote-option | vote_option_controller.go | 投票选项 |
| /api/admin/vote-record | vote_record_controller.go | 投票记录 |
| /api/admin/user-task-event | user_task_event_controller.go | 用户任务事件 |
| /api/admin/third-user | third_user_controller.go | 第三方用户 |
| /api/admin/user-follow | user_follow_controller.go | 用户关注 |
| /api/admin/user-feed | user_feed_controller.go | 用户动态 |
| /api/admin/message | message_controller.go | 消息管理 |
| /api/admin/check-in | check_in_controller.go | 签到管理 |
| /api/admin/email-code | email_code_controller.go | 邮箱验证码 |
| /api/admin/sms-code | sms_code_controller.go | 短信验证码 |
| /api/admin/migration | migration_controller.go | 数据迁移 |

---

## 二、Site前端缺失的API对接

### Site阶段二专用缺口矩阵

| 模块 | Method | API路径 | 请求关键字段 | 响应关键字段 | 前端落点页面 | 后端文件位置 |
|------|--------|---------|--------------|--------------|--------------|--------------|
| 任务 | GET | /api/task/tasks | groupName(可选) | data[]: id,title,summary,rewardScore,rewardExp,status | 任务中心 | task_controller.go:20 |
| 任务 | GET | /api/task/groups | 无 | data[]: name,count | 任务中心 | task_controller.go:54 |
| 投票 | GET | /api/vote/{voteId} | voteId(路径参数) | data: voteId,title,multiple,options,userOptionIds | 投票详情 | vote_controller.go:18 |
| 投票 | POST | /api/vote/cast | voteId, optionIds[] | data: 最新VoteResponse | 投票详情 | vote_controller.go:26 |
| 徽章 | GET | /api/badge/badges | userId(可选) | data[]: id,name,icon,description,level,owned,worn,obtainTime | 勋章墙/个人侧边栏 | badge_controller.go:17 |
| 登录-短信 | POST | /api/login/login_sms_code | phone,captchaId,captchaCode | data: smsId | 登录页短信Tab | login_controller.go:153 |
| 登录-短信 | POST | /api/login/login_sms | smsId,smsCode,redirect | data: token,user,redirect | 登录页短信Tab | login_controller.go:182 |
| 登录-微信 | GET | /api/login/wx_login_config | 无 | data: appId,redirectUri,state | 登录页微信登录 | login_controller.go:218 |
| 登录-微信 | POST | /api/login/wx_login_submit | code,state,redirect | data: token,user,redirect | 登录回调处理 | login_controller.go:246 |
| 账号绑定-微信 | POST | /api/login/wx_bind | code,state | data: 绑定结果 | 账号设置-第三方绑定 | login_controller.go:267 |
| 账号绑定-微信 | POST | /api/login/wx_unbind | 无 | data: 解绑结果 | 账号设置-第三方绑定 | login_controller.go:287 |
| 登录-Google | GET | /api/login/google_login_config | 无 | data: clientId,redirectUri,state | 登录页Google登录 | login_controller.go:302 |
| 登录-Google | POST | /api/login/google_login_submit | code,state,redirect | data: token,user,redirect | 登录回调处理 | login_controller.go:334 |
| 账号绑定-Google | POST | /api/login/google_bind | code,state | data: 绑定结果 | 账号设置-第三方绑定 | login_controller.go:359 |
| 登录-Google | POST | /api/login/google_one_tap | credential,redirect | data: token,user,redirect | 登录页One Tap | login_controller.go:394 |
| 账号绑定-Google | POST | /api/login/google_unbind | 无 | data: 解绑结果 | 账号设置-第三方绑定 | login_controller.go:413 |
| 绑定信息 | GET | /api/user/wx_bind_info | 无 | data: bind,nickname,avatar | 账号设置-第三方绑定 | user_controller.go:418 |
| 绑定信息 | GET | /api/user/google_bind_info | 无 | data: bind,nickname,avatar | 账号设置-第三方绑定 | user_controller.go:436 |

---

## 三、Admin前端缺失的页面

### 后端有但Admin前端无对应页面的模块

| 后端API模块 | Controller文件 | 现状 |
|-------------|----------------|------|
| /api/admin/task-config | task_config_controller.go | 无页面 |
| /api/admin/badge | badge_controller.go | 无页面 |
| /api/admin/level-config | level_config_controller.go | 无页面 |
| /api/admin/user-task-log | user_task_log_controller.go | 无页面 |
| /api/admin/user-exp-log | user_exp_log_controller.go | 无页面 |
| /api/admin/user-badge | user_badge_controller.go | 无页面 |
| /api/admin/vote | vote_controller.go | 无页面 |
| /api/admin/vote-option | vote_option_controller.go | 无页面 |
| /api/admin/vote-record | vote_record_controller.go | 无页面 |
| /api/admin/third-user | third_user_controller.go | 无页面 |
| /api/admin/user-follow | user_follow_controller.go | 无页面 |
| /api/admin/user-feed | user_feed_controller.go | 无页面 |
| /api/admin/message | message_controller.go | 无页面 |
| /api/admin/check-in | check_in_controller.go | 无页面 |
| /api/admin/email-code | email_code_controller.go | 无页面 |
| /api/admin/sms-code | sms_code_controller.go | 无页面 |

---

## 四、登录模块详细对比

### 后端Login Controller已实现的API（16个）

| 方法 | 路径 | 功能 | Site前端状态 |
|------|------|------|--------------|
| POST | /api/login/signup | 注册 | 已对接 |
| POST | /api/login/signin | 密码登录 | 已对接 |
| POST | /api/login/send_reset_password_email | 找回密码邮件 | 已对接 |
| POST | /api/login/reset_password | 重置密码 | 已对接 |
| GET | /api/login/signout | 退出登录 | 已对接 |
| POST | /api/login/login_sms_code | 请求短信验证码 | **未对接** |
| POST | /api/login/login_sms | 短信登录 | **未对接** |
| GET | /api/login/wx_login_config | 微信登录配置 | **未对接** |
| POST | /api/login/wx_login_submit | 微信登录提交 | **未对接** |
| POST | /api/login/wx_bind | 微信绑定 | **未对接** |
| POST | /api/login/wx_unbind | 微信解绑 | **未对接** |
| GET | /api/login/google_login_config | 谷歌登录配置 | **未对接** |
| POST | /api/login/google_login_submit | 谷歌登录提交 | **未对接** |
| POST | /api/login/google_bind | 谷歌绑定 | **未对接** |
| POST | /api/login/google_one_tap | 谷歌一键登录 | **未对接** |
| POST | /api/login/google_unbind | 谷歌解绑 | **未对接** |

---

## 五、数据来源

- 后端路由定义：internal/server/router.go
- 后端Controller：internal/controllers/api/*.go, internal/controllers/admin/*.go
- 前端API调用：site/src/**/*.vue, admin/src/**/*.vue
- 分析时间：2026-03-06
