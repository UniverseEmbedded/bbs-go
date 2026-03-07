export async function fetchTaskGroups() {
  return await useHttpGet("/api/task/groups")
}

export async function fetchTasks(groupName = "") {
  const params = {}
  if (groupName) {
    params.groupName = groupName
  }
  return await useHttpGet("/api/task/tasks", { params })
}

export async function fetchVote(voteId) {
  return await useHttpGet(`/api/vote/${voteId}`)
}

export async function castVote(voteId, optionIds) {
  return await useHttpPost("/api/vote/cast", {
    voteId,
    optionIds,
  })
}

export async function fetchBadges(userId = "") {
  const params = {}
  if (userId) {
    params.userId = userId
  }
  return await useHttpGet("/api/badge/badges", { params })
}

export async function requestLoginSmsCode(payload) {
  return await useHttpPost("/api/login/login_sms_code", useJsonToForm(payload))
}

export async function loginBySms(payload) {
  return await useHttpPost("/api/login/login_sms", useJsonToForm(payload))
}

export async function fetchWxLoginConfig(payload = {}) {
  return await useHttpGet("/api/login/wx_login_config", { params: payload })
}

export async function submitWxLogin(payload) {
  return await useHttpPost("/api/login/wx_login_submit", useJsonToForm(payload))
}

export async function wxBind(payload) {
  return await useHttpPost("/api/login/wx_bind", useJsonToForm(payload))
}

export async function wxUnbind() {
  return await useHttpPost("/api/login/wx_unbind", useJsonToForm({}))
}

export async function fetchGoogleLoginConfig(payload = {}) {
  return await useHttpGet("/api/login/google_login_config", { params: payload })
}

export async function submitGoogleLogin(payload) {
  return await useHttpPost(
    "/api/login/google_login_submit",
    useJsonToForm(payload)
  )
}

export async function googleBind(payload) {
  return await useHttpPost("/api/login/google_bind", useJsonToForm(payload))
}

export async function googleUnbind() {
  return await useHttpPost("/api/login/google_unbind", useJsonToForm({}))
}

export async function googleOneTap(payload) {
  return await useHttpPost("/api/login/google_one_tap", useJsonToForm(payload))
}

export async function fetchWxBindInfo() {
  return await useHttpGet("/api/user/wx_bind_info")
}

export async function fetchGoogleBindInfo() {
  return await useHttpGet("/api/user/google_bind_info")
}

export function resolveBadgeIcon(badge) {
  if (badge?.icon) {
    return badge.icon
  }
  if (badge?.name) {
    return `/res/images/badges/${badge.name}.svg`
  }
  return "/res/images/badges/template.svg"
}
