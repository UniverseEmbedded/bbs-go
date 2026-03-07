<template>
  <section class="main task-page">
    <div class="container main-container left-main size-360">
      <div class="left-container">
        <div class="main-content no-padding no-bg">
          <div class="widget task-widget">
            <div class="widget-content">
              <div class="task-head">
                <div>
                  <h1 class="task-title">{{ $t("pages.tasks.title") }}</h1>
                  <p class="task-sub">{{ $t("pages.tasks.subtitle") }}</p>
                </div>
                <div class="top-stats">
                  <div class="top-pill">
                    {{ $t("pages.tasks.total", { total: taskTotal }) }}
                  </div>
                  <div class="top-pill done">
                    {{ $t("pages.tasks.done", { done: doneTotal }) }}
                  </div>
                </div>
              </div>
              <div class="tabs-wrap">
                <button
                  v-for="group in groups"
                  :key="group.key"
                  class="tab-btn"
                  :class="{ active: activeGroup === group.key }"
                  @click="activeGroup = group.key"
                >
                  {{ group.name }}
                </button>
              </div>
              <div class="task-grid">
                <div v-for="task in tasks" :key="task.id" class="task-card">
                  <div class="task-row">
                    <h3 class="task-card-title">{{ task.title }}</h3>
                    <span class="status-tag" :class="{ done: task.status === 1 }">
                      {{
                        task.status === 1
                          ? $t("pages.tasks.finished")
                          : $t("pages.tasks.pending")
                      }}
                    </span>
                  </div>
                  <div class="task-desc">{{ task.description }}</div>
                  <div class="reward-row">
                    <span class="reward-chip">+{{ task.score }} {{ $t("pages.tasks.score") }}</span>
                    <span class="reward-chip">+{{ task.exp }} {{ $t("pages.tasks.exp") }}</span>
                  </div>
                  <div class="progress-meta">
                    <span>{{
                      $t("pages.tasks.progress", {
                        current: task.userProgress?.finishedCount || 0,
                        target: task.userProgress?.maxFinishCount || task.maxFinishCount || 1,
                      })
                    }}</span>
                    <span>{{ getTaskPercent(task) }}%</span>
                  </div>
                  <div class="bar">
                    <div
                      class="bar-fill"
                      :style="{ width: `${getTaskPercent(task)}%` }"
                    />
                  </div>
                  <div class="period-meta">
                    {{
                      $t("pages.tasks.eventProgress", {
                        current: task.userProgress?.eventProgress || 0,
                        target: task.userProgress?.eventTarget || task.eventCount || 1,
                      })
                    }}
                  </div>
                </div>
                <div v-if="!tasks.length" class="no-task">{{ $t("common.noData") }}</div>
              </div>
            </div>
          </div>
        </div>
      </div>
      <div class="right-container">
        <div class="widget side-widget">
          <div class="widget-content">
            <div class="user-top">
              <my-avatar :user="currentUser" :size="40" />
              <div class="user-info">
                <div class="name-row">
                  <span class="name">{{ currentUser?.nickname || "-" }}</span>
                  <span class="lv">Lv.{{ currentUser?.level || 0 }}</span>
                </div>
                <div class="user-sub">{{ currentUser?.description || "-" }}</div>
              </div>
            </div>
            <div class="exp-line">
              <span>{{ $t("pages.tasks.userExp") }}</span>
              <span
                >{{ currentUser?.expProgress?.expInCurrentLevel || 0 }}/{{
                  currentUser?.expProgress?.expNeedForNextLevel || 0
                }}</span
              >
            </div>
            <div class="bar side-exp">
              <div
                class="bar-fill"
                :style="{
                  width: `${currentUser?.expProgress?.expProgressPercent || 0}%`,
                }"
              />
            </div>
            <div class="side-stats">
              <div class="cell">
                <div class="label">{{ $t("pages.tasks.score") }}</div>
                <div class="num">{{ currentUser?.score || 0 }}</div>
              </div>
              <div class="cell">
                <div class="label">{{ $t("pages.tasks.topic") }}</div>
                <div class="num">{{ currentUser?.topicCount || 0 }}</div>
              </div>
              <div class="cell">
                <div class="label">{{ $t("pages.tasks.comment") }}</div>
                <div class="num">{{ currentUser?.commentCount || 0 }}</div>
              </div>
              <div class="cell">
                <div class="label">{{ $t("pages.tasks.fans") }}</div>
                <div class="num">{{ currentUser?.fansCount || 0 }}</div>
              </div>
            </div>
          </div>
        </div>
        <div class="widget checkin-widget">
          <div class="widget-content">
            <div class="checkin-top">
              <h3 class="checkin-title">{{ $t("component.checkIn.title") }}</h3>
              <span class="checkin-tag">{{
                checkIn?.checkIn
                  ? $t("component.checkIn.alreadyCheckedIn")
                  : $t("pages.tasks.notCheckedIn")
              }}</span>
            </div>
            <div class="checkin-desc">
              {{
                $t("component.checkIn.consecutiveDays", {
                  days: checkIn?.consecutiveDays || 0,
                })
              }}
            </div>
            <button class="checkin-btn" :disabled="checkIn?.checkIn" @click="doCheckIn">
              {{
                checkIn?.checkIn
                  ? $t("component.checkIn.alreadyCheckedIn")
                  : $t("component.checkIn.checkInNow")
              }}
            </button>
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

const groups = ref([])
const activeGroup = ref("")
const tasks = ref([])

const { data: checkIn, refresh: refreshCheckIn } = await useMyFetch(
  "/api/checkin/checkin"
)

const taskTotal = computed(() => tasks.value.length)
const doneTotal = computed(
  () => tasks.value.filter((item) => item.status === 1).length
)

const getTaskPercent = (task) => {
  if (task.status === 1) {
    return 100
  }
  const progress = task.userProgress
  if (!progress) {
    return 0
  }
  const target = Math.max(progress.maxFinishCount || 1, 1)
  const percent = Math.floor(((progress.finishedCount || 0) / target) * 100)
  return Math.max(0, Math.min(percent, 100))
}

const loadTasks = async () => {
  try {
    tasks.value = await fetchTasks(activeGroup.value)
  } catch (e) {
    useCatchError(e)
    tasks.value = []
  }
}

const loadGroups = async () => {
  try {
    groups.value = await fetchTaskGroups()
    if (!groups.value.length) {
      return
    }
    activeGroup.value = groups.value[0].key
    await loadTasks()
  } catch (e) {
    useCatchError(e)
  }
}

const doCheckIn = async () => {
  try {
    checkIn.value = await useHttpPost("/api/checkin/checkin")
    useMsgSuccess(t("component.checkIn.checkInSuccess"))
    refreshCheckIn()
  } catch (e) {
    useCatchError(e)
  }
}

watch(activeGroup, async () => {
  if (!activeGroup.value) {
    return
  }
  await loadTasks()
})

await loadGroups()

useHead({
  title: useSiteTitle(t("pages.tasks.title")),
})
</script>

<style scoped lang="scss">
.task-widget,
.side-widget,
.checkin-widget {
  border: 1px solid var(--border-color4);
}

.task-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.task-title {
  margin: 0;
  font-size: 28px;
  font-weight: 800;
  color: var(--text-color2);
}

.task-sub {
  margin-top: 6px;
  color: var(--text-color3);
  font-size: 14px;
}

.top-stats {
  display: flex;
  gap: 8px;
}

.top-pill {
  font-size: 13px;
  border-radius: 999px;
  padding: 6px 10px;
  background: var(--bg-color2);
  color: var(--text-color2);
  border: 1px solid var(--border-color);
}

.top-pill.done {
  background: #ecfff4;
  color: #18a058;
  border-color: #d7f5e5;
}

.tabs-wrap {
  display: inline-flex;
  gap: 8px;
  margin-top: 14px;
}

.tab-btn {
  border: 1px solid var(--border-color);
  background: var(--bg-color);
  color: var(--text-color2);
  border-radius: 8px;
  padding: 6px 14px;
  font-size: 15px;
  cursor: pointer;
}

.tab-btn.active {
  border-color: var(--text-link-color);
  color: var(--text-link-color);
  background: var(--bg-color2);
}

.task-grid {
  margin-top: 14px;
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.task-card {
  border: 1px solid var(--border-color4);
  border-radius: 12px;
  padding: 14px;
  background: var(--bg-color);
}

.task-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 10px;
}

.task-card-title {
  margin: 0;
  font-size: 20px;
  line-height: 1.2;
  font-weight: 800;
  color: var(--text-color2);
}

.task-desc {
  margin-top: 8px;
  color: var(--text-color3);
  font-size: 14px;
  line-height: 1.45;
  min-height: 44px;
}

.status-tag {
  font-size: 12px;
  font-weight: 700;
  padding: 2px 10px;
  border-radius: 999px;
  border: 1px solid var(--border-color);
  color: var(--text-color3);
  background: var(--bg-color2);
}

.status-tag.done {
  color: #18a058;
  border-color: #baf1d2;
  background: #ecfff4;
}

.reward-row {
  display: flex;
  gap: 8px;
  margin-top: 10px;
}

.reward-chip {
  border-radius: 999px;
  padding: 4px 10px;
  background: var(--bg-color2);
  color: var(--text-color2);
  border: 1px solid var(--border-color4);
  font-size: 13px;
  font-weight: 700;
}

.progress-meta {
  margin-top: 10px;
  display: flex;
  justify-content: space-between;
  font-size: 13px;
  color: var(--text-color3);
}

.bar {
  margin-top: 6px;
  width: 100%;
  height: 8px;
  background: var(--bg-color2);
  border-radius: 999px;
  overflow: hidden;
}

.bar-fill {
  height: 100%;
  background: var(--text-link-color);
  border-radius: 999px;
}

.period-meta {
  margin-top: 8px;
  color: var(--text-color3);
  font-size: 12px;
}

.no-task {
  grid-column: span 2;
  text-align: center;
  color: var(--text-color3);
  padding: 20px 0;
}

.user-top {
  display: flex;
  align-items: center;
  gap: 10px;
}

.user-info {
  flex: 1;
}

.name-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.name {
  font-size: 20px;
  font-weight: 800;
  color: var(--text-color2);
}

.lv {
  font-size: 12px;
  padding: 2px 6px;
  border-radius: 6px;
  color: #f69c28;
  background: #fff3dc;
  border: 1px solid #ffe2b6;
  font-weight: 700;
}

.user-sub {
  margin-top: 2px;
  color: var(--text-color3);
  font-size: 13px;
}

.exp-line {
  margin-top: 12px;
  display: flex;
  justify-content: space-between;
  color: var(--text-color2);
  font-size: 13px;
}

.side-exp {
  margin-top: 8px;
}

.side-stats {
  margin-top: 12px;
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 8px;
}

.cell {
  background: var(--bg-color2);
  border: 1px solid var(--border-color4);
  border-radius: 10px;
  text-align: center;
  padding: 8px 4px;
}

.label {
  font-size: 12px;
  color: var(--text-color3);
}

.num {
  font-size: 16px;
  font-weight: 800;
  color: var(--text-color2);
  margin-top: 2px;
}

.checkin-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.checkin-title {
  font-size: 22px;
  margin: 0;
  font-weight: 800;
  color: var(--text-color2);
}

.checkin-tag {
  font-size: 12px;
  border-radius: 999px;
  padding: 2px 8px;
  border: 1px solid var(--border-color);
  color: var(--text-color3);
  background: var(--bg-color2);
}

.checkin-desc {
  margin-top: 8px;
  color: var(--text-color3);
  font-size: 14px;
}

.checkin-btn {
  margin-top: 12px;
  width: 100%;
  height: 40px;
  border: none;
  border-radius: 10px;
  background: var(--text-link-color);
  color: #fff;
  font-size: 15px;
  font-weight: 700;
  cursor: pointer;
}

.checkin-btn:disabled {
  background: #88b5fa;
  cursor: default;
}

@media (max-width: 1180px) {
  .task-grid {
    grid-template-columns: 1fr;
  }
  .no-task {
    grid-column: span 1;
  }
}
</style>
