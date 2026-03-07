# Site阶段二实施与验收

本文档定义阶段二执行范围、实施顺序与验收标准，配合以下文档使用：

- 里程碑.md
- API对接缺口与对比.md
- Site前端现状与目标.md
- Site阶段二API契约表.md

---

## 一、范围

### In Scope

- 任务中心
- 投票详情交互
- 短信登录
- 微信登录与微信绑定
- Google 登录与 Google 绑定
- 勋章墙展示

### Out of Scope

- Admin 缺失页面补全
- Loomio 风格投票重构（阶段四）
- 阶段二之外的后端改造

---

## 二、实施顺序

1. 任务中心（/api/task/tasks, /api/task/groups）
2. 投票详情（/api/vote/{voteId}, /api/vote/cast）
3. 登录扩展（短信、微信、Google）
4. 账号绑定（wx_bind_info/google_bind_info + bind/unbind）
5. 勋章墙（/api/badge/badges）

---

## 三、页面与原型映射

| 页面能力 | 原型文件 | 核心接口 |
|----------|----------|----------|
| 任务中心 | task_center.html | /api/task/tasks, /api/task/groups |
| 投票模块 | voting_component.html | /api/vote/{voteId}, /api/vote/cast |
| 登录扩展 | login_modal.html | /api/login/login_sms_code, /api/login/login_sms, /api/login/wx_login_config, /api/login/google_login_config |
| 账号设置绑定 | account_settings.html | /api/user/wx_bind_info, /api/user/google_bind_info, /api/login/*_bind, /api/login/*_unbind |
| 勋章墙 | badge_wall.html | /api/badge/badges |

---

## 四、联调检查清单

| 检查项 | 通过标准 |
|--------|----------|
| 任务列表加载 | 分组切换正常，任务状态与奖励字段正确渲染 |
| 投票详情加载 | 选项与已投状态显示正确 |
| 投票提交 | 提交成功后结果刷新，无需手动刷新页面 |
| 短信验证码 | 可获取 smsId，异常提示可见 |
| 短信登录 | 登录成功后获得 token 与用户信息 |
| 微信登录 | 可拿到配置并完成回调提交 |
| Google登录 | 可拿到配置并完成回调提交 |
| One Tap | 可提交 credential 并得到登录结果 |
| 微信绑定 | 可查询、绑定、解绑 |
| Google绑定 | 可查询、绑定、解绑 |
| 勋章墙 | 可展示拥有与佩戴状态，空态可用 |

---

## 五、验收步骤

1. 按模块执行联调检查清单并记录结果。
2. 对每个模块保留成功截图与失败截图各一组。
3. 对照原型文件检查结构与交互是否一致。
4. 通过项全部满足后，标记阶段二完成。

---

## 六、交付物

- 更新后的 Site 前端页面代码
- 联调记录
- 验收截图
- 文档回填（本文件与 Site前端现状与目标.md）
