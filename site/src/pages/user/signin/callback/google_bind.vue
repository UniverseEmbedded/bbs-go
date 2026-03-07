<template>
  <login-o-auth-callback provider="google" :bind="true" />
</template>

<script setup>
const { t } = useI18n()
useHead({
  title: useSiteTitle(t("user.signin.callback.title")),
})
</script>
