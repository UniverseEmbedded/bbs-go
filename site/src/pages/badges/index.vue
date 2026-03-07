<template>
  <section class="main badge-page">
    <div class="container main-container left-main size-360">
      <div class="left-container">
        <div class="main-content no-padding no-bg">
          <div class="widget">
            <div class="widget-content">
              <h1 class="badge-title">{{ $t("pages.badge.title") }}</h1>
              <div class="badge-grid">
                <div
                  v-for="badge in badges"
                  :key="badge.id"
                  class="badge-card"
                  :class="{ 'is-owned': badge.owned }"
                >
                  <img
                    class="badge-icon"
                    :src="resolveBadgeIcon(badge)"
                    :alt="badge.title || badge.name"
                  />
                  <div class="badge-name">{{ badge.title || badge.name }}</div>
                  <div class="badge-desc">{{ badge.description }}</div>
                  <div v-if="badge.owned" class="badge-time">
                    {{ $t("pages.badge.obtainedAt", { time: usePrettyDate(badge.obtainTime, $t) }) }}
                  </div>
                  <div v-else class="badge-time is-empty">{{ $t("pages.badge.notOwned") }}</div>
                </div>
                <div v-if="!badges.length" class="no-data">{{ $t("common.noData") }}</div>
              </div>
            </div>
          </div>
        </div>
      </div>
      <div class="right-container">
        <div class="widget">
          <div class="widget-content">
            <div class="media">
              <div class="media-left">
                <my-avatar :user="currentUser" :size="48" />
              </div>
              <div class="media-content">
                <p class="title is-6">{{ currentUser?.nickname || "-" }}</p>
                <p class="subtitle is-7">Lv.{{ currentUser?.level || 0 }}</p>
              </div>
            </div>
          </div>
        </div>
        <div class="widget">
          <div class="widget-content">
            <div class="mini-head">
              <h3 class="mini-title">{{ $t("pages.badge.myBadges") }}</h3>
              <nuxt-link class="mini-link" to="/badges">{{ $t("pages.badge.viewAll") }}</nuxt-link>
            </div>
            <div class="mini-badge-grid">
              <div v-for="badge in ownedBadges.slice(0, 4)" :key="badge.id" class="mini-badge">
                <img :src="resolveBadgeIcon(badge)" :alt="badge.title || badge.name" />
              </div>
              <div v-if="!ownedBadges.length" class="mini-empty">
                {{ $t("pages.badge.noBadges") }}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup>
const { t } = useI18n()
const userStore = useUserStore()
const currentUser = computed(() => userStore.user)
const badges = ref([])

try {
  badges.value = await fetchBadges()
} catch (e) {
  useCatchError(e)
}

const ownedBadges = computed(() => badges.value.filter((item) => item.owned))

useHead({
  title: useSiteTitle(t("pages.badge.title")),
})
</script>

<style scoped lang="scss">
.badge-title {
  font-size: 26px;
  margin-bottom: 18px;
}

.badge-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
}

.badge-card {
  background: var(--bg-color);
  border-radius: 8px;
  padding: 16px;
  text-align: center;
  border: 1px solid var(--border-color4);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}

.badge-icon {
  width: 72px;
  height: 72px;
  margin-bottom: 10px;
  filter: grayscale(100%);
  opacity: 0.6;
}

.badge-card.is-owned .badge-icon {
  filter: grayscale(0);
  opacity: 1;
}

.badge-name {
  font-weight: 600;
  font-size: 14px;
  margin-bottom: 5px;
}

.badge-desc {
  font-size: 12px;
  color: var(--text-color3);
  margin-bottom: 8px;
  min-height: 32px;
}

.badge-time {
  font-size: 12px;
  color: #67c23a;
  background: #f0f9eb;
  padding: 2px 6px;
  border-radius: 4px;
}

.badge-time.is-empty {
  color: var(--text-color3);
  background: var(--bg-color2);
}

.mini-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}

.mini-title {
  margin: 0;
  font-size: 15px;
}

.mini-link {
  font-size: 12px;
}

.mini-badge-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 10px;
}

.mini-badge {
  text-align: center;
}

.mini-badge img {
  width: 40px;
  height: 40px;
}

.mini-empty {
  grid-column: span 4;
  text-align: center;
  color: var(--text-color3);
  font-size: 12px;
}

.no-data {
  grid-column: span 3;
  text-align: center;
  color: var(--text-color3);
  padding: 16px 0;
}

@media (max-width: 1024px) {
  .badge-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
  .no-data {
    grid-column: span 2;
  }
}
</style>
