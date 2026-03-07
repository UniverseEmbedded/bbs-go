# Site前端现状与目标

本文档记录Site前端（用户端）的当前状态与目标，基于代码分析得出。

---

## 一、技术栈

| 项目 | 版本/名称 |
|------|-----------|
| 框架 | Nuxt 3 |
| UI库 | Bulma + Element Plus |
| 状态管理 | Pinia |
| HTTP封装 | useHttp / useMyFetch |
| 样式 | SCSS |

---

## 二、当前已实现的页面

### 页面清单

| 页面路径 | 说明 | 状态 |
|----------|------|------|
| / | 首页（话题列表） | 正常 |
| /about | 关于页面 | 正常 |
| /links | 友链页面 | 正常 |
| /search | 搜索页面 | 正常 |
| /install | 安装向导 | 正常 |
| /topic/create | 创建话题 | 正常 |
| /topic/edit/:id | 编辑话题 | 正常 |
| /topic/:id | 话题详情 | 正常 |
| /topics | 话题列表 | 正常 |
| /topics/node/:id | 节点话题列表 | 正常 |
| /topics/tag/:id | 标签话题列表 | 正常 |
| /article/create | 创建文章 | 正常 |
| /article/edit/:id | 编辑文章 | 正常 |
| /article/:id | 文章详情 | 正常 |
| /articles | 文章列表 | 正常 |
| /articles/tag/:id | 标签文章列表 | 正常 |
| /user/signin | 登录 | 正常 |
| /user/signup | 注册 | 正常 |
| /user/:userId | 用户主页 | 正常 |
| /user/:userId/articles | 用户文章 | 正常 |
| /user/:userId/fans | 用户粉丝 | 正常 |
| /user/:userId/followed | 用户关注 | 正常 |
| /user/profile | 个人资料 | 正常 |
| /user/profile/account | 账号设置 | 正常 |
| /user/favorites | 收藏列表 | 正常 |
| /user/messages | 消息列表 | 正常 |
| /user/scores | 积分记录 | 正常 |
| /user/email/verify | 邮箱验证 | 正常 |

---

## 三、当前已对接的API

### 按模块分类

| 模块 | 已对接的API |
|------|-------------|
| 登录注册 | /api/login/signup, /api/login/signin, /api/login/signout |
| 用户 | /api/user/current, /api/user/:id, /api/user/edit/:id, /api/user/score_logs, /api/user/favorites, /api/user/messages, /api/user/score/rank, /api/user/msg_recent, /api/user/send_verify_email, /api/user/verify_email, /api/user/set_username, /api/user/set_email, /api/user/set_password, /api/user/update_password, /api/user/set_background_image, /api/user/forbidden |
| 话题 | /api/topic/nodes, /api/topic/node_navs, /api/topic/create, /api/topic/edit/:id, /api/topic/:id, /api/topic/delete/:id, /api/topic/recommend/:id, /api/topic/sticky/:id, /api/topic/topics, /api/topic/user/topics, /api/topic/tag/topics, /api/topic/node, /api/topic/recentlikes/:id, /api/topic/hide_content |
| 文章 | /api/article/create, /api/article/edit/:id, /api/article/:id, /api/article/delete/:id, /api/article/articles, /api/article/user/articles, /api/article/tag/articles |
| 评论 | /api/comment/comments, /api/comment/replies, /api/comment/create |
| 标签 | /api/tag/:id, /api/tag/autocomplete |
| 点赞收藏 | /api/like/like, /api/like/unlike, /api/like/liked, /api/favorite/add, /api/favorite/delete |
| 关注 | /api/fans/follow, /api/fans/unfollow, /api/fans/is_followed, /api/fans/fans, /api/fans/followed, /api/fans/recent/fans, /api/fans/recent/follow |
| 签到 | /api/checkin/checkin, /api/checkin/rank |
| 上传 | /api/upload |
| 配置 | /api/config/configs |
| 链接 | /api/link/list, /api/link/top_links |
| 验证码 | /api/captcha/request_angle |
| 搜索 | /api/search/topic |
| 安装 | /api/install/status, /api/install/test_db_connection, /api/install/install |

---

## 四、缺失的功能

### 缺失的页面

| 页面 | 说明 | 对应后端API |
|------|------|-------------|
| 任务中心 | 用户任务列表与奖励 | /api/task/tasks, /api/task/groups |
| 投票列表 | 投票列表页 | /api/vote/* |
| 投票详情 | 投票参与页 | /api/vote/:id, /api/vote/cast |
| 勋章墙 | 徽章展示与佩戴状态 | /api/badge/badges |

### 缺失的登录方式

| 登录方式 | 说明 | 对应后端API |
|----------|------|-------------|
| 短信登录 | 手机号+验证码登录 | /api/login/login_sms_code, /api/login/login_sms |
| 微信登录 | 微信扫码登录 | /api/login/wx_login_config, /api/login/wx_login_submit |
| 微信绑定 | 绑定微信号 | /api/login/wx_bind, /api/login/wx_unbind, /api/user/wx_bind_info |
| 谷歌登录 | Google登录 | /api/login/google_login_config, /api/login/google_login_submit |
| 谷歌绑定 | 绑定Google账号 | /api/login/google_bind, /api/login/google_unbind, /api/user/google_bind_info |

### 登录页面现状

当前登录页面（/user/signin）仅包含：
- 用户名密码登录
- 注册入口

当前缺失：
- 短信登录入口
- 微信登录入口
- 谷歌登录入口

---

## 五、目标

基于「API对接缺口与对比.md」的分析，Site前端的目标为：

1. **任务系统**：新增任务中心页面，对接 /api/task/*
2. **投票系统**：新增投票页面，对接 /api/vote/*
3. **勋章系统**：新增勋章墙与个人佩戴展示，对接 /api/badge/badges
4. **短信登录**：新增短信登录功能，对接 /api/login/login_sms*
5. **微信登录**：新增微信登录、绑定功能，对接 /api/login/wx*, /api/user/wx_bind_info
6. **谷歌登录**：新增谷歌登录、绑定功能，对接 /api/login/google*, /api/user/google_bind_info

---

## 六、页面-API映射与优先级

| 优先级 | 页面/模块 | 关键API | 完成标准 |
|--------|-----------|---------|----------|
| P0 | 任务中心 | /api/task/tasks, /api/task/groups | 可按分组展示任务，完成态/奖励信息正确显示 |
| P0 | 投票详情 | /api/vote/{voteId}, /api/vote/cast | 可查看选项与投票状态，提交后结果即时刷新 |
| P0 | 登录页短信登录 | /api/login/login_sms_code, /api/login/login_sms | 可获取短信验证码并完成登录 |
| P0 | 账号设置第三方绑定 | /api/user/wx_bind_info, /api/user/google_bind_info, /api/login/*_bind, /api/login/*_unbind | 可查看绑定状态并完成绑定/解绑 |
| P1 | 勋章墙 | /api/badge/badges | 可展示徽章列表与拥有状态 |
| P1 | 登录页第三方登录 | /api/login/wx_login_config, /api/login/wx_login_submit, /api/login/google_login_config, /api/login/google_login_submit, /api/login/google_one_tap | 可触发登录流程并接收登录结果 |

---

## 七、验收标准（阶段二）

| 页面/模块 | 验收项 |
|-----------|--------|
| 任务中心 | 页面可访问；任务分组切换可用；任务状态与奖励字段展示正确 |
| 投票详情 | 可加载投票详情；可提交投票；提交后已投选项与统计同步更新 |
| 登录扩展 | 密码登录不回退；短信登录可用；微信与Google登录入口可触发并处理返回 |
| 账号绑定 | 微信/Google绑定状态可查询；可绑定；可解绑 |
| 勋章墙 | 徽章数据可加载；拥有/佩戴状态可区分展示；空数据有兜底状态 |
