# Site阶段二API契约表

本文档用于阶段二（补全 Site 前端缺失功能）联调，所有接口均采用统一响应协议：

- success: boolean
- message: string
- data: any

---

## 一、任务系统

### 1) 获取任务列表

- Method: GET
- Path: /api/task/tasks
- 鉴权: 需要登录
- Query: groupName(可选)
- 成功响应 data 关键字段：
  - id
  - title
  - summary
  - rewardScore
  - rewardExp
  - status
  - completeCondition

### 2) 获取任务分组

- Method: GET
- Path: /api/task/groups
- 鉴权: 需要登录
- 成功响应 data 关键字段：
  - name
  - count

---

## 二、投票系统

### 1) 获取投票详情

- Method: GET
- Path: /api/vote/{voteId}
- 鉴权: 需要登录
- 成功响应 data 关键字段：
  - voteId
  - title
  - multiple
  - options[]
  - userOptionIds[]

### 2) 提交投票

- Method: POST
- Path: /api/vote/cast
- 鉴权: 需要登录
- Body(JSON):
  - voteId: number
  - optionIds: number[]
- 成功响应 data:
  - 最新 VoteResponse（用于刷新已投状态与计数）

---

## 三、勋章系统

### 1) 获取勋章列表

- Method: GET
- Path: /api/badge/badges
- 鉴权: 可匿名（携带登录态可返回当前用户拥有状态）
- Query:
  - userId(可选)
- 成功响应 data 关键字段：
  - id
  - name
  - icon
  - description
  - level
  - owned
  - worn
  - obtainTime

---

## 四、登录扩展

### 1) 请求短信验证码

- Method: POST
- Path: /api/login/login_sms_code
- 鉴权: 不需要
- Body(Form):
  - phone
  - captchaId
  - captchaCode
- 成功响应 data:
  - smsId

### 2) 短信登录

- Method: POST
- Path: /api/login/login_sms
- 鉴权: 不需要
- Body(Form):
  - smsId
  - smsCode
  - redirect
- 成功响应 data 关键字段：
  - token
  - user
  - redirect

### 3) 微信登录配置

- Method: GET
- Path: /api/login/wx_login_config
- 鉴权: 不需要
- 成功响应 data 关键字段：
  - appId
  - redirectUri
  - state

### 4) 微信登录提交

- Method: POST
- Path: /api/login/wx_login_submit
- 鉴权: 不需要
- Body(Form):
  - code
  - state
  - redirect
- 成功响应 data 关键字段：
  - token
  - user
  - redirect

### 5) Google 登录配置

- Method: GET
- Path: /api/login/google_login_config
- 鉴权: 不需要
- 成功响应 data 关键字段：
  - clientId
  - redirectUri
  - state

### 6) Google 登录提交

- Method: POST
- Path: /api/login/google_login_submit
- 鉴权: 不需要
- Body(Form):
  - code
  - state
  - redirect
- 成功响应 data 关键字段：
  - token
  - user
  - redirect

### 7) Google One Tap

- Method: POST
- Path: /api/login/google_one_tap
- 鉴权: 不需要
- Body(Form):
  - credential
  - redirect
- 成功响应 data 关键字段：
  - token
  - user
  - redirect

---

## 五、账号绑定

### 1) 微信绑定状态

- Method: GET
- Path: /api/user/wx_bind_info
- 鉴权: 需要登录
- 成功响应 data:
  - bind
  - nickname
  - avatar

### 2) Google 绑定状态

- Method: GET
- Path: /api/user/google_bind_info
- 鉴权: 需要登录
- 成功响应 data:
  - bind
  - nickname
  - avatar

### 3) 微信绑定/解绑

- Method: POST
- Path: /api/login/wx_bind
- 鉴权: 需要登录
- Body(Form): code, state

- Method: POST
- Path: /api/login/wx_unbind
- 鉴权: 需要登录

### 4) Google 绑定/解绑

- Method: POST
- Path: /api/login/google_bind
- 鉴权: 需要登录
- Body(Form): code, state

- Method: POST
- Path: /api/login/google_unbind
- 鉴权: 需要登录

---

## 六、对接约束

- 登录相关接口采用表单参数提交，投票提交采用 JSON。
- 所有鉴权接口需携带前端当前 token 策略。
- 联调时以本文档路径命名为准，避免使用历史别名。
