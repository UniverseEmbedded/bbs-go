<template>
  <div class="sms-login">
    <div class="login-field">
      <input
        v-model="form.phone"
        type="text"
        :placeholder="$t('user.signin.sms.phonePlaceholder')"
      />
    </div>

    <div class="login-field">
      <input
        v-model="form.smsCode"
        type="text"
        :placeholder="$t('user.signin.sms.smsCodePlaceholder')"
      />
      <a class="sms-code-btn" @click="requestCode">
        {{ countdown > 0 ? `${countdown}s` : $t("user.signin.sms.getSmsCode") }}
      </a>
    </div>

    <div class="login-btn">
      <el-button type="primary" @click="signin">
        {{ $t("user.signin.sms.loginBtn") }}
      </el-button>
    </div>

    <div class="login-bottom">
      <a @click="toSignup">{{ $t("user.signin.password.noAccount") }}</a>
    </div>

    <CaptchaDialog ref="captchaDialog" />
  </div>
</template>

<script setup>
const route = useRoute()
const { t } = useI18n()
const userStore = useUserStore()

const form = reactive({
  phone: "",
  smsCode: "",
  smsId: "",
  redirect: route.query.redirect || "",
})

const countdown = ref(0)
let timer = null
const captchaDialog = ref(null)

const requestCode = async () => {
  if (countdown.value > 0) {
    return
  }
  const phone = (form.phone || "").trim()
  if (!/^1\d{10}$/.test(phone)) {
    useMsgError(t("user.signin.sms.phoneError"))
    return
  }
  const captcha = await captchaDialog.value.show()
  if (!captcha) {
    return
  }
  try {
    const data = await requestLoginSmsCode({
      phone,
      captchaId: captcha.captchaId,
      captchaCode: captcha.captchaCode,
    })
    form.smsId = data.smsId || ""
    countdown.value = 60
    timer = setInterval(() => {
      if (countdown.value <= 1) {
        clearInterval(timer)
        timer = null
        countdown.value = 0
      } else {
        countdown.value--
      }
    }, 1000)
    useMsgSuccess(t("user.signin.sms.smsSent"))
  } catch (e) {
    useCatchError(e)
  }
}

const signin = async () => {
  const phone = (form.phone || "").trim()
  if (!/^1\d{10}$/.test(phone)) {
    useMsgError(t("user.signin.sms.phoneError"))
    return
  }
  if (!form.smsCode) {
    useMsgError(t("user.signin.sms.smsCodeRequired"))
    return
  }
  if (!form.smsId) {
    useMsgError(t("user.signin.sms.needSmsRequest"))
    return
  }
  try {
    const { user, redirect } = await userStore.signinSms({
      smsId: form.smsId,
      smsCode: form.smsCode,
      redirect: form.redirect,
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

const toSignup = async () => {
  if (form.redirect) {
    useLinkTo(`/user/signup?redirect=${encodeURIComponent(form.redirect)}`)
  } else {
    useLinkTo("/user/signup")
  }
}

onUnmounted(() => {
  if (timer) {
    clearInterval(timer)
  }
})
</script>

<style lang="scss" scoped>
.sms-login {
  max-width: 400px;
  margin: auto;
  .login-field {
    width: 100%;
    height: 39px;
    margin: 40px 0;
    display: flex;
    align-items: center;
    background-color: var(--bg-color2);
    border: 1px solid var(--border-color);
    border-radius: 3px;
    &:has(input:focus) {
      background-color: var(--bg-color3);
      border: 1px solid var(--border-hover-color);
      input {
        background-color: var(--bg-color3);
      }
    }
    input {
      padding: 0 15px;
      width: 100%;
      height: 37px;
      border: none;
      outline: none;
      background-color: var(--bg-color2);
      border-radius: 3px;
    }
    .sms-code-btn {
      min-width: max-content;
      padding: 0 12px;
      font-size: 13px;
      color: var(--text-link-color);
      white-space: nowrap;
    }
  }
  .login-btn {
    width: 100%;
    button {
      width: 100%;
      height: 40px;
    }
  }
  .login-bottom {
    margin: 20px 0;
    font-size: 13px;
    display: flex;
    justify-content: center;
    a {
      color: var(--text-color3);
    }
  }
}
</style>
