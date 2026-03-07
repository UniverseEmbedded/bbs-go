<template>
  <div class="oauth-login">
    <div class="split">{{ $t("user.signin.oauthDivider") }}</div>
    <button class="oauth-btn wx-btn" @click="wxLogin">
      {{ $t("user.signin.weixinLogin") }}
    </button>
    <button class="oauth-btn google-btn" @click="googleLogin">
      {{ $t("user.signin.googleLogin") }}
    </button>
    <button class="oauth-btn one-tap-btn" @click="googleOneTapLogin">
      {{ $t("user.signin.googleOneTap") }}
    </button>
  </div>
</template>

<script setup>
const route = useRoute()
const { t } = useI18n()
const userStore = useUserStore()

const wxLogin = async () => {
  try {
    const payload = { redirect: route.query.redirect || "" }
    const config = await fetchWxLoginConfig(payload)
    const query = new URLSearchParams({
      appid: config.appid,
      redirect_uri: config.redirect_uri,
      response_type: "code",
      scope: config.scope,
      state: config.state,
    })
    location.href = `https://open.weixin.qq.com/connect/qrconnect?${query.toString()}#wechat_redirect`
  } catch (e) {
    useCatchError(e)
  }
}

const googleLogin = async () => {
  try {
    const config = await fetchGoogleLoginConfig({
      redirect: route.query.redirect || "",
    })
    if (config.authUrl) {
      location.href = config.authUrl
      return
    }
    useMsgError(t("user.signin.googleConfigError"))
  } catch (e) {
    useCatchError(e)
  }
}

const googleOneTapLogin = async () => {
  const credential = prompt(t("user.signin.googleCredentialPrompt"))
  if (!credential) {
    return
  }
  try {
    const { user, redirect } = await userStore.signinGoogleOneTap({
      credential,
      redirect: route.query.redirect || "",
    })
    if (redirect) {
      useLinkTo(redirect)
    } else {
      useLinkTo(`/user/${user.id}`)
    }
  } catch (e) {
    useCatchError(e)
  }
}
</script>

<style scoped lang="scss">
.oauth-login {
  margin-top: 12px;
  .split {
    margin-top: 12px;
    display: flex;
    align-items: center;
    gap: 12px;
    color: var(--text-color3);
    font-size: 14px;
  }
  .split::before,
  .split::after {
    content: "";
    flex: 1;
    height: 1px;
    background: var(--border-color4);
  }
  .oauth-btn {
    margin-top: 12px;
    height: 42px;
    width: 100%;
    border-radius: 10px;
    border: 1px solid var(--border-color);
    background: var(--bg-color);
    font-weight: 700;
    color: var(--text-color);
    cursor: pointer;
  }
  .wx-btn {
    color: #18a058;
  }
  .google-btn {
    color: #db4437;
  }
}
</style>
