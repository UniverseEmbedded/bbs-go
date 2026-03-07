<template>
  <section class="main">
    <div class="container">
      <div class="main-body no-bg">
        <div class="widget callback-widget">
          <div class="widget-content">
            <p class="callback-title">{{ message }}</p>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup>
const props = defineProps({
  provider: {
    type: String,
    required: true,
  },
  bind: {
    type: Boolean,
    default: false,
  },
})

const route = useRoute()
const { t } = useI18n()
const userStore = useUserStore()
const message = ref(t("user.signin.callback.processing"))

const doBind = async () => {
  const payload = {
    code: route.query.code || "",
    state: route.query.state || "",
  }
  if (!payload.code || !payload.state) {
    throw new Error(t("user.signin.callback.invalidParams"))
  }
  if (props.provider === "weixin") {
    await wxBind(payload)
  } else {
    await googleBind(payload)
  }
  useMsgSuccess(t("user.signin.callback.bindSuccess"))
  useLinkTo("/user/profile/account")
}

const doSignin = async () => {
  const payload = {
    code: route.query.code || "",
    state: route.query.state || "",
  }
  if (!payload.code || !payload.state) {
    throw new Error(t("user.signin.callback.invalidParams"))
  }
  const result =
    props.provider === "weixin"
      ? await userStore.signinWx(payload)
      : await userStore.signinGoogle(payload)
  if (result.redirect) {
    useLinkTo(result.redirect)
  } else {
    useLinkTo(`/user/${result.user.id}`)
  }
}

onMounted(async () => {
  try {
    if (props.bind) {
      await doBind()
    } else {
      await doSignin()
    }
  } catch (e) {
    message.value = t("user.signin.callback.failed")
    useCatchError(e)
    setTimeout(() => {
      useLinkTo("/user/signin")
    }, 1200)
  }
})
</script>

<style scoped lang="scss">
.callback-widget {
  max-width: 560px;
  margin: auto;
}

.callback-title {
  text-align: center;
  font-size: 16px;
  color: var(--text-color2);
  padding: 36px 0;
}
</style>
