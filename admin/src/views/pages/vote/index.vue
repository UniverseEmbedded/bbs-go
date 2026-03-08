<template>
  <div class="container">
    <div class="container-header">
      <a-form :model="filters" layout="inline" :size="appStore.table.size">
        <a-form-item>
          <a-input-number v-model="filters.id" placeholder="VoteId" :min="0" />
        </a-form-item>
        <a-form-item>
          <a-select v-model="filters.pollType" placeholder="Poll Type" allow-clear style="width: 150px">
            <a-option value="proposal">Proposal</a-option>
            <a-option value="poll">Poll</a-option>
            <a-option value="count">Count</a-option>
            <a-option value="ranked_choice">Ranked Choice</a-option>
          </a-select>
        </a-form-item>
        <a-form-item>
          <a-button type="primary" html-type="submit" @click="list">
            <template #icon> <icon-search /> </template>
            Search
          </a-button>
        </a-form-item>
      </a-form>
    </div>

    <div class="container-main">
      <a-table
        :loading="loading"
        :data="data.results"
        :size="appStore.table.size"
        :bordered="appStore.table.bordered"
        :pagination="pagination"
        :sticky-header="true"
        style="height: 100%"
        column-resizable
        @page-change="onPageChange"
        @page-size-change="onPageSizeChange"
      >
        <template #columns>
          <a-table-column title="ID" data-index="id" :width="80" />
          <a-table-column title="Poll Type" data-index="pollType" :width="120">
            <template #cell="{ record }">
              <a-tag :color="getPollTypeColor(record.pollType)">
                {{ getPollTypeLabel(record.pollType) }}
              </a-tag>
            </template>
          </a-table-column>
          <a-table-column title="Title" data-index="title" :width="200" ellipsis />
          <a-table-column title="Options" data-index="optionCount" :width="80" />
          <a-table-column title="Votes" data-index="voteCount" :width="80" />
          <a-table-column title="Hide Results" data-index="hideResults" :width="100">
            <template #cell="{ record }">
              <a-tag v-if="record.hideResults === 0" color="green">Off</a-tag>
              <a-tag v-else-if="record.hideResults === 1" color="orange">Until Vote</a-tag>
              <a-tag v-else-if="record.hideResults === 2" color="red">Until Closed</a-tag>
            </template>
          </a-table-column>
          <a-table-column title="Anonymous" data-index="anonymous" :width="90">
            <template #cell="{ record }">
              <a-tag v-if="record.anonymous" color="purple">Yes</a-tag>
              <a-tag v-else color="gray">No</a-tag>
            </template>
          </a-table-column>
          <a-table-column title="Status" :width="100">
            <template #cell="{ record }">
              <a-tag v-if="record.closedAt" color="red">Closed</a-tag>
              <a-tag v-else-if="isExpired(record)" color="orange">Expired</a-tag>
              <a-tag v-else color="green">Open</a-tag>
            </template>
          </a-table-column>
          <a-table-column title="ExpiredAt" data-index="expiredAt" :width="150">
            <template #cell="{ record }">
              {{ record.expiredAt ? useFormatDate(record.expiredAt) : '-' }}
            </template>
          </a-table-column>
          <a-table-column title="Create" data-index="createTime" :width="150">
            <template #cell="{ record }">
              {{ record.createTime ? useFormatDate(record.createTime) : '-' }}
            </template>
          </a-table-column>
          <a-table-column title="Actions" :width="150" fixed="right">
            <template #cell="{ record }">
              <a-space>
                <a-button type="text" size="small" @click="showDetail(record)">
                  Detail
                </a-button>
                <a-button type="text" size="small" @click="showOutcomeDialog(record)">
                  Outcome
                </a-button>
              </a-space>
            </template>
          </a-table-column>
        </template>
      </a-table>
    </div>

    <a-modal v-model:visible="detailVisible" title="Vote Detail" :footer="false" width="800px">
      <div v-if="currentVote" class="vote-detail">
        <a-descriptions :column="2" bordered>
          <a-descriptions-item label="ID">{{ currentVote.id }}</a-descriptions-item>
          <a-descriptions-item label="Poll Type">{{ getPollTypeLabel(currentVote.pollType) }}</a-descriptions-item>
          <a-descriptions-item label="Title" :span="2">{{ currentVote.title }}</a-descriptions-item>
          <a-descriptions-item label="Vote Count">{{ currentVote.voteCount }}</a-descriptions-item>
          <a-descriptions-item label="Option Count">{{ currentVote.optionCount }}</a-descriptions-item>
          <a-descriptions-item label="Hide Results">{{ getHideResultsLabel(currentVote.hideResults) }}</a-descriptions-item>
          <a-descriptions-item label="Anonymous">{{ currentVote.anonymous ? 'Yes' : 'No' }}</a-descriptions-item>
          <a-descriptions-item label="Reason Required">{{ getReasonRequiredLabel(currentVote.stanceReasonRequired) }}</a-descriptions-item>
          <a-descriptions-item label="Voter Can Add Options">{{ currentVote.voterCanAddOptions ? 'Yes' : 'No' }}</a-descriptions-item>
        </a-descriptions>

        <h4 style="margin-top: 16px">Options</h4>
        <a-table :data="currentVoteOptions" :pagination="false" size="small">
          <template #columns>
            <a-table-column title="ID" data-index="id" :width="60" />
            <a-table-column title="Content" data-index="content" />
            <a-table-column title="Vote Count" data-index="voteCount" :width="100" />
            <a-table-column title="Total Score" data-index="totalScore" :width="100" />
          </template>
        </a-table>
      </div>
    </a-modal>

    <a-modal v-model:visible="outcomeVisible" title="Create Outcome" @ok="submitOutcome" @cancel="outcomeVisible = false">
      <a-form :model="outcomeForm" layout="vertical">
        <a-form-item label="Statement" required>
          <a-textarea v-model="outcomeForm.statement" placeholder="Enter outcome statement" :rows="4" />
        </a-form-item>
        <a-form-item label="Winning Option">
          <a-select v-model="outcomeForm.pollOptionId" placeholder="Select winning option" allow-clear>
            <a-option v-for="opt in currentVoteOptions" :key="opt.id" :value="opt.id">
              {{ opt.content }}
            </a-option>
          </a-select>
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
  import { Message } from '@arco-design/web-vue';

  const appStore = useAppStore();
  const loading = ref(false);
  const detailVisible = ref(false);
  const outcomeVisible = ref(false);
  const currentVote = ref<any>(null);
  const currentVoteOptions = ref<any[]>([]);

  const filters = reactive({
    limit: 20,
    page: 1,
    id: undefined as number | undefined,
    pollType: undefined as string | undefined,
  });

  const outcomeForm = reactive({
    pollId: 0,
    statement: '',
    pollOptionId: undefined as number | undefined,
  });

  const data = reactive({
    page: {
      page: 1,
      limit: 20,
      total: 0,
    },
    results: [] as any[],
  });

  const pagination = computed(() => {
    return {
      total: data.page.total,
      current: data.page.page,
      pageSize: data.page.limit,
      showTotal: true,
      showJumper: true,
      showPageSize: true,
      pageSizeOptions: [20, 50, 100, 200, 300, 500],
    };
  });

  onMounted(() => {
    useTableHeight();
  });

  const list = async () => {
    loading.value = true;
    try {
      const ret = await axios.postForm<any>(
        '/api/admin/vote/list',
        jsonToFormData(filters)
      );
      data.page = ret.page;
      data.results = ret.results;
    } finally {
      loading.value = false;
    }
  };

  list();

  const onPageChange = (page: number) => {
    filters.page = page;
    list();
  };

  const onPageSizeChange = (pageSize: number) => {
    filters.limit = pageSize;
    list();
  };

  const showDetail = async (record: any) => {
    currentVote.value = record;
    try {
      const ret = await axios.get<any>(`/api/admin/vote-option/list?voteId=${record.id}`);
      currentVoteOptions.value = ret.data?.results || [];
    } catch (e) {
      currentVoteOptions.value = [];
    }
    detailVisible.value = true;
  };

  const showOutcomeDialog = async (record: any) => {
    currentVote.value = record;
    outcomeForm.pollId = record.id;
    outcomeForm.statement = '';
    outcomeForm.pollOptionId = undefined;
    try {
      const ret = await axios.get<any>(`/api/admin/vote-option/list?voteId=${record.id}`);
      currentVoteOptions.value = ret.data?.results || [];
    } catch (e) {
      currentVoteOptions.value = [];
    }
    outcomeVisible.value = true;
  };

  const submitOutcome = async () => {
    if (!outcomeForm.statement.trim()) {
      Message.warning('Please enter outcome statement');
      return;
    }
    try {
      await axios.post('/api/outcome/create', outcomeForm);
      Message.success('Outcome created successfully');
      outcomeVisible.value = false;
    } catch (e: any) {
      Message.error(e.message || 'Failed to create outcome');
    }
  };

  const isExpired = (record: any) => {
    if (!record.expiredAt) return false;
    return Date.now() > record.expiredAt;
  };

  const getPollTypeLabel = (type: string) => {
    const labels: Record<string, string> = {
      proposal: 'Proposal',
      poll: 'Poll',
      count: 'Count',
      ranked_choice: 'Ranked Choice',
    };
    return labels[type] || type || 'Poll';
  };

  const getPollTypeColor = (type: string) => {
    const colors: Record<string, string> = {
      proposal: 'red',
      poll: 'blue',
      count: 'green',
      ranked_choice: 'orange',
    };
    return colors[type] || 'gray';
  };

  const getHideResultsLabel = (val: number) => {
    const labels: Record<number, string> = {
      0: 'Off',
      1: 'Until Vote',
      2: 'Until Closed',
    };
    return labels[val] || 'Unknown';
  };

  const getReasonRequiredLabel = (val: number) => {
    const labels: Record<number, string> = {
      0: 'Disabled',
      1: 'Optional',
      2: 'Required',
    };
    return labels[val] || 'Unknown';
  };
</script>

<style scoped lang="less">
.vote-detail {
  h4 {
    font-weight: 600;
    margin-bottom: 8px;
  }
}
</style>
