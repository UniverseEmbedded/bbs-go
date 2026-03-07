import test from "node:test"
import assert from "node:assert/strict"
import fs from "node:fs"
import path from "node:path"

const siteRoot = path.resolve(process.cwd())
const read = (p) => fs.readFileSync(path.join(siteRoot, p), "utf-8")

test("阶段二页面已创建", () => {
  const requiredFiles = [
    "src/pages/tasks/index.vue",
    "src/pages/badges/index.vue",
    "src/pages/votes/[id].vue",
    "src/pages/user/signin/callback/weixin.vue",
    "src/pages/user/signin/callback/google.vue",
    "src/pages/user/signin/callback/weixin_bind.vue",
    "src/pages/user/signin/callback/google_bind.vue",
  ]
  for (const file of requiredFiles) {
    assert.equal(fs.existsSync(path.join(siteRoot, file)), true, `${file} 不存在`)
  }
})

test("阶段二API契约已在前端封装", () => {
  const apiCode = read("src/composables/stage2Api.js")
  const requiredApis = [
    "/api/task/groups",
    "/api/task/tasks",
    "/api/vote/",
    "/api/vote/cast",
    "/api/badge/badges",
    "/api/login/login_sms_code",
    "/api/login/login_sms",
    "/api/login/wx_login_config",
    "/api/login/wx_login_submit",
    "/api/login/wx_bind",
    "/api/login/wx_unbind",
    "/api/login/google_login_config",
    "/api/login/google_login_submit",
    "/api/login/google_bind",
    "/api/login/google_unbind",
    "/api/login/google_one_tap",
    "/api/user/wx_bind_info",
    "/api/user/google_bind_info",
  ]
  for (const api of requiredApis) {
    assert.equal(apiCode.includes(api), true, `${api} 未封装`)
  }
})

test("登录页已接入短信与第三方登录", () => {
  const signinCode = read("src/pages/user/signin/index.vue")
  assert.equal(signinCode.includes("<login-sms"), true)
  assert.equal(signinCode.includes("<login-oauth"), true)
})

test("账号设置页已接入绑定状态与绑定动作", () => {
  const accountCode = read("src/pages/user/profile/account.vue")
  assert.equal(accountCode.includes("fetchWxBindInfo"), true)
  assert.equal(accountCode.includes("fetchGoogleBindInfo"), true)
  assert.equal(accountCode.includes("wxUnbind"), true)
  assert.equal(accountCode.includes("googleUnbind"), true)
})

test("话题页已接入投票组件", () => {
  const topicCode = read("src/pages/topic/[id].vue")
  assert.equal(topicCode.includes("<vote-vote-panel"), true)
})
