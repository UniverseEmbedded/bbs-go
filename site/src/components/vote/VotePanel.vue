<template>
  <div v-if="voteData" class="vote-box">
    <div class="vote-title">
      <span class="vote-type-tag" :class="pollTypeClass">
        {{ pollTypeLabel }}
      </span>
      <span>{{ voteData.title }}</span>
    </div>

    <div v-if="outcomeData" class="vote-outcome">
      <div class="outcome-header">
        <span class="icon"><i class="fas fa-gavel"></i></span>
        <span>{{ $t("pages.vote.outcome") }}</span>
      </div>
      <div class="outcome-statement">{{ outcomeData.statement }}</div>
      <div v-if="outcomeData.pollOption" class="outcome-option">
        {{ $t("pages.vote.winningOption") }}: {{ outcomeData.pollOption.content }}
      </div>
    </div>

    <div v-if="!voteData.voted && !voteData.expired && canVote">
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
            <span class="option-content">
              <span v-if="option.icon" class="option-icon">{{ option.icon }}</span>
              {{ option.content }}
            </span>
            <span v-if="option.prompt" class="option-prompt">{{ option.prompt }}</span>
          </label>
          <label class="radio" v-else>
            <input type="radio" :checked="selectedOptions.includes(option.id)" />
            <span class="option-content">
              <span v-if="option.icon" class="option-icon">{{ option.icon }}</span>
              {{ option.content }}
            </span>
            <span v-if="option.prompt" class="option-prompt">{{ option.prompt }}</span>
          </label>
        </div>
      </div>

      <div v-if="showReasonInput" class="vote-reason">
        <textarea
          v-model="reason"
          class="textarea"
          :placeholder="reasonPlaceholder"
          :rows="3"
        ></textarea>
      </div>

      <div class="mt-4">
        <button class="button is-primary" :disabled="!canSubmit" @click="submitVote">
          {{ $t("pages.vote.submit") }}
        </button>
      </div>
    </div>

    <div v-else-if="showResults">
      <div v-for="option in voteData.options" :key="option.id" class="vote-result">
        <div class="vote-text">
          <span>
            <span v-if="option.icon" class="option-icon">{{ option.icon }}</span>
            {{ option.content }}
            <span v-if="option.voted" class="tag is-primary is-light is-rounded ml-2">
              {{ $t("pages.vote.selected") }}
            </span>
          </span>
          <span>{{ option.voteCount }}{{ $t("pages.vote.ticket") }} ({{ option.percent.toFixed(1) }}%)</span>
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

    <div v-else-if="!voteData.voted && !voteData.expired && !canVote" class="vote-restricted">
      <p>{{ $t("pages.vote.restricted") }}</p>
    </div>

    <div v-else-if="!showResults && voteData.hideResults !== 0" class="vote-hidden-results">
      <p>{{ $t("pages.vote.resultsHidden") }}</p>
    </div>

    <div v-if="voteData.voted && !voteData.expired" class="vote-actions">
      <button class="button is-small is-light" @click="showModifyDialog = true">
        {{ $t("pages.vote.modifyStance") }}
      </button>
      <button class="button is-small is-light is-danger ml-2" @click="handleRevoke">
        {{ $t("pages.vote.revokeStance") }}
      </button>
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

    <div v-if="showModifyDialog" class="modal is-active">
      <div class="modal-background" @click="showModifyDialog = false"></div>
      <div class="modal-content">
        <div class="box">
          <h3 class="title is-5">{{ $t("pages.vote.modifyStance") }}</h3>
          <div class="vote-options">
            <div
              v-for="option in voteData.options"
              :key="option.id"
              class="vote-option"
              :class="{ 'is-selected': modifyOptions.includes(option.id) }"
              @click="selectModifyOption(option.id)"
            >
              <label class="checkbox" v-if="voteData.type === 2">
                <input type="checkbox" :checked="modifyOptions.includes(option.id)" />
                {{ option.content }}
              </label>
              <label class="radio" v-else>
                <input type="radio" :checked="modifyOptions.includes(option.id)" />
                {{ option.content }}
              </label>
            </div>
          </div>
          <div v-if="showReasonInput" class="vote-reason mt-4">
            <textarea
              v-model="modifyReason"
              class="textarea"
              :placeholder="reasonPlaceholder"
              :rows="3"
            ></textarea>
          </div>
          <div class="mt-4">
            <button class="button is-primary" :disabled="!canModifySubmit" @click="submitModify">
              {{ $t("pages.vote.submit") }}
            </button>
            <button class="button is-light ml-2" @click="showModifyDialog = false">
              {{ $t("common.cancel") }}
            </button>
          </div>
        </div>
      </div>
      <button class="modal-close is-large" @click="showModifyDialog = false"></button>
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
const reason = ref("")
const modifyOptions = ref([])
const modifyReason = ref("")
const showModifyDialog = ref(false)
const outcomeData = ref(null)
const currentStance = ref(null)
const { t } = useI18n()

watch(
  () => props.vote,
  (val) => {
    voteData.value = val
    selectedOptions.value = []
    reason.value = ""
    loadOutcome()
    loadCurrentStance()
  },
  { deep: true, immediate: true }
)

const pollTypeLabel = computed(() => {
  const types = {
    proposal: t("pages.vote.proposal"),
    poll: t("pages.vote.poll"),
    count: t("pages.vote.count"),
    ranked_choice: t("pages.vote.rankedChoice"),
  }
  return types[voteData.value?.pollType] || t("pages.vote.poll")
})

const pollTypeClass = computed(() => {
  return `poll-type-${voteData.value?.pollType || "poll"}`
})

const showReasonInput = computed(() => {
  return voteData.value?.stanceReasonRequired !== 0
})

const reasonPlaceholder = computed(() => {
  if (voteData.value?.stanceReasonRequired === 2) {
    return t("pages.vote.reasonRequired")
  }
  return t("pages.vote.reasonOptional")
})

const canSubmit = computed(() => {
  if (selectedOptions.value.length === 0) return false
  if (voteData.value?.stanceReasonRequired === 2 && !reason.value.trim()) return false
  return true
})

const canModifySubmit = computed(() => {
  if (modifyOptions.value.length === 0) return false
  if (voteData.value?.stanceReasonRequired === 2 && !modifyReason.value.trim()) return false
  return true
})

const canVote = computed(() => {
  return !voteData.value?.specifiedVotersOnly || voteData.value?.voted
})

const showResults = computed(() => {
  if (!voteData.value) return false
  if (voteData.value.hideResults === 0) return true
  if (voteData.value.hideResults === 1) return voteData.value.voted || voteData.value.expired
  if (voteData.value.hideResults === 2) return voteData.value.expired
  return false
})

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

const selectModifyOption = (id) => {
  if (voteData.value.type === 2) {
    const idx = modifyOptions.value.indexOf(id)
    if (idx > -1) {
      modifyOptions.value.splice(idx, 1)
    } else {
      modifyOptions.value.push(id)
    }
  } else {
    modifyOptions.value = [id]
  }
}

const submitVote = async () => {
  try {
    const form = {
      pollId: voteData.value.id,
      optionIds: selectedOptions.value,
      reason: reason.value,
    }
    const result = await createStance(form)
    currentStance.value = result
    const updatedVote = await fetchVote(voteData.value.id)
    voteData.value = updatedVote
    selectedOptions.value = []
    reason.value = ""
    emits("updated", updatedVote)
    useMsgSuccess(t("pages.vote.submitSuccess"))
  } catch (e) {
    useCatchError(e)
  }
}

const submitModify = async () => {
  try {
    const form = {
      pollId: voteData.value.id,
      optionIds: modifyOptions.value,
      reason: modifyReason.value,
    }
    const result = await createStance(form)
    currentStance.value = result
    const updatedVote = await fetchVote(voteData.value.id)
    voteData.value = updatedVote
    showModifyDialog.value = false
    modifyOptions.value = []
    modifyReason.value = ""
    emits("updated", updatedVote)
    useMsgSuccess(t("pages.vote.modifySuccess"))
  } catch (e) {
    useCatchError(e)
  }
}

const handleRevoke = async () => {
  if (!currentStance.value?.id) return
  try {
    await revokeStance(currentStance.value.id)
    currentStance.value = null
    const updatedVote = await fetchVote(voteData.value.id)
    voteData.value = updatedVote
    emits("updated", updatedVote)
    useMsgSuccess(t("pages.vote.revokeSuccess"))
  } catch (e) {
    useCatchError(e)
  }
}

const loadOutcome = async () => {
  if (voteData.value?.id) {
    try {
      outcomeData.value = await fetchOutcome(voteData.value.id)
    } catch (e) {
      outcomeData.value = null
    }
  }
}

const loadCurrentStance = async () => {
  if (voteData.value?.id) {
    try {
      currentStance.value = await fetchStance(voteData.value.id)
      if (currentStance.value) {
        const choices = currentStance.value.choices || []
        modifyOptions.value = choices.map(c => c.pollOptionId)
        modifyReason.value = currentStance.value.reason || ""
      }
    } catch (e) {
      currentStance.value = null
    }
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

  &.poll-type-proposal {
    background: #fef0f0;
    color: #f56c6c;
    border-color: #fde2e2;
  }

  &.poll-type-count {
    background: #f0f9eb;
    color: #67c23a;
    border-color: #e1f3d8;
  }

  &.poll-type-ranked_choice {
    background: #fdf6ec;
    color: #e6a23c;
    border-color: #faecd8;
  }
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

.option-content {
  display: flex;
  align-items: center;
}

.option-icon {
  margin-right: 8px;
  font-size: 1.2em;
}

.option-prompt {
  display: block;
  font-size: 12px;
  color: var(--text-color3);
  margin-top: 4px;
  margin-left: 20px;
}

.vote-reason {
  margin-top: 15px;
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

.vote-outcome {
  background-color: #f0f9eb;
  border: 1px solid #e1f3d8;
  border-radius: 4px;
  padding: 15px;
  margin-bottom: 15px;
}

.outcome-header {
  font-weight: 600;
  color: #67c23a;
  margin-bottom: 10px;

  .icon {
    margin-right: 8px;
  }
}

.outcome-statement {
  font-size: 14px;
  margin-bottom: 8px;
}

.outcome-option {
  font-size: 13px;
  color: var(--text-color2);
}

.vote-actions {
  margin-top: 15px;
  padding-top: 10px;
  border-top: 1px solid var(--border-color4);
}

.vote-restricted,
.vote-hidden-results {
  padding: 20px;
  text-align: center;
  color: var(--text-color3);
}
</style>
