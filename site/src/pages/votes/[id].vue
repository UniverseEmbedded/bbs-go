<template>
  <section class="main">
    <div class="container main-container">
      <div class="main-content">
        <h1 class="vote-page-title">{{ vote?.title || $t("pages.vote.title") }}</h1>
        <vote-vote-panel v-if="vote" :vote="vote" @updated="vote = $event" />
      </div>
    </div>
  </section>
</template>

<script setup>
const route = useRoute()
const { t } = useI18n()

const vote = ref(null)

try {
  vote.value = await fetchVote(route.params.id)
} catch (e) {
  useCatchError(e)
}

useHead({
  title: useSiteTitle(vote.value?.title || t("pages.vote.title")),
})
</script>

<style scoped lang="scss">
.vote-page-title {
  font-size: 24px;
  margin-bottom: 12px;
  color: var(--text-color2);
}
</style>
