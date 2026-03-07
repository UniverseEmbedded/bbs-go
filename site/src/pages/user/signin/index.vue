<template>
  <section class="main">
    <div class="container">
      <div class="main-body no-bg">
        <div class="widget signin">
          <div class="widget-content">
            <div class="tabs is-centered">
              <ul>
                <li
                  :class="{ 'is-active': activeTab === 'password' }"
                  @click="activeTab = 'password'"
                >
                  <a>{{ $t("user.signin.passwordLogin") }}</a>
                </li>
                <li
                  :class="{ 'is-active': activeTab === 'sms' }"
                  @click="activeTab = 'sms'"
                >
                  <a>{{ $t("user.signin.smsLogin") }}</a>
                </li>
              </ul>
            </div>

            <div>
              <login-password v-if="activeTab === 'password'" />
              <login-sms v-else />
              <login-oauth />
            </div>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup>
const route = useRoute()
const { t } = useI18n()
const activeTab = ref(route.query.tab === "sms" ? "sms" : "password")
useHead({
  title: useSiteTitle(t("user.signin.title")),
})
</script>

<style lang="scss" scoped>
.signin {
  max-width: 640px;
  margin: auto !important;
  padding: 20px;
}
</style>
