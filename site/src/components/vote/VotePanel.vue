<template>
  <div v-if="voteData" class="vote-box">
    <div class="vote-title">
      <span class="vote-type-tag">
        {{ voteData.type === 2 ? $t("pages.vote.multi") : $t("pages.vote.single") }}
      </span>
      <span>{{ voteData.title }}</span>
    </div>

    <div v-if="!voteData.voted && !voteData.expired">
      <div class="vote-options">
        <div
          v-for="option in voteData.options"
          :key="option.id"
          class="vote-option"
          :class="{ 'is-selected': selectedOptions.includes(option.id) }"
          @click="selectOption(option.id)"
        >
          <label class="checkbox" v-if="voteData.type === 2">
            <input type="checkbox" :checked="selectedOptions.includes(option.id)" />
            {{ option.content }}
          </label>
          <label class="radio" v-else>
            <input type="radio" :checked="selectedOptions.includes(option.id)" />
            {{ option.content }}
          </label>
        </div>
      </div>
      <div class="mt-4">
        <button class="button is-primary" :disabled="!selectedOptions.length" @click="submitVote">
          {{ $t("pages.vote.submit") }}
        </button>
      </div>
    </div>
    <div v-else>
      <div v-for="option in voteData.options" :key="option.id" class="vote-result">
        <div class="vote-text">
          <span>
            {{ option.content }}
            <span v-if="option.voted" class="tag is-primary is-light is-rounded ml-2">
              {{ $t("pages.vote.selected") }}
            </span>
          </span>
          <span>{{ option.voteCount }}{{ $t("pages.vote.ticket") }} ({{ option.percent }}%)</span>
        </div>
        <div class="vote-progress-container">
          <div class="vote-progress-bar">
            <div
              class="vote-progress-fill"
              :class="{ 'is-winner': isWinner(option) }"
              :style="{ width: `${option.percent}%` }"
            />
          </div>
        </div>
      </div>
    </div>
    <div class="vote-meta">
      <span>{{ voteData.voteCount }}{{ $t("pages.vote.participated") }}</span>
      <span class="mx-2">·</span>
      <span>
        {{
          $t("pages.vote.expiredAt", {
            time: voteData.expiredAt ? usePrettyDate(voteData.expiredAt, $t) : "-",
          })
        }}
      </span>
      <span class="is-pulled-right">
        {{ voteData.expired ? $t("pages.vote.closed") : $t("pages.vote.open") }}
      </span>
    </div>
  </div>
</template>

<script setup>
const props = defineProps({
  vote: {
    type: Object,
    required: true,
  },
})

const emits = defineEmits(["updated"])

const voteData = ref(props.vote)
const selectedOptions = ref([])
const { t } = useI18n()

watch(
  () => props.vote,
  (val) => {
    voteData.value = val
    selectedOptions.value = []
  },
  { deep: true }
)

const selectOption = (id) => {
  if (voteData.value.type === 2) {
    const idx = selectedOptions.value.indexOf(id)
    if (idx > -1) {
      selectedOptions.value.splice(idx, 1)
    } else {
      selectedOptions.value.push(id)
    }
  } else {
    selectedOptions.value = [id]
  }
}

const submitVote = async () => {
  try {
    const result = await castVote(voteData.value.id, selectedOptions.value)
    voteData.value = result
    selectedOptions.value = []
    emits("updated", result)
    useMsgSuccess(t("pages.vote.submitSuccess"))
  } catch (e) {
    useCatchError(e)
  }
}

const isWinner = (option) => {
  const max = Math.max(...(voteData.value?.options || []).map((o) => o.voteCount))
  return option.voteCount === max && max > 0
}
</script>

<style scoped lang="scss">
.vote-box {
  background-color: var(--bg-color2);
  border: 1px solid var(--border-color4);
  border-radius: 6px;
  padding: 20px;
  margin-top: 20px;
}

.vote-title {
  font-weight: 600;
  margin-bottom: 15px;
  display: flex;
  align-items: center;
}

.vote-type-tag {
  font-size: 12px;
  background: #ecf5ff;
  color: #409eff;
  padding: 2px 6px;
  border-radius: 4px;
  margin-right: 8px;
  border: 1px solid #d9ecff;
}

.vote-option {
  margin-bottom: 10px;
  padding: 10px;
  border-radius: 4px;
  cursor: pointer;
  transition: background 0.2s;
  border: 1px solid transparent;
}

.vote-option:hover {
  background-color: var(--bg-color3);
}

.vote-option.is-selected {
  background-color: #ecf5ff;
  border-color: #409eff;
}

.vote-progress-container {
  margin-bottom: 15px;
}

.vote-progress-bar {
  height: 20px;
  background-color: var(--border-color4);
  border-radius: 10px;
  overflow: hidden;
}

.vote-progress-fill {
  height: 100%;
  background-color: #409eff;
  border-radius: 10px;
}

.vote-progress-fill.is-winner {
  background-color: #67c23a;
}

.vote-text {
  display: flex;
  justify-content: space-between;
  font-size: 14px;
  margin-bottom: 5px;
}

.vote-meta {
  font-size: 12px;
  color: var(--text-color3);
  margin-top: 15px;
  border-top: 1px solid var(--border-color4);
  padding-top: 10px;
}
</style>
