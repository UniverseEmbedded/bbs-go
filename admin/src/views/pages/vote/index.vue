<template>
  <div class="container">
    <div class="container-header">
      <a-form :model="filters" layout="inline" :size="appStore.table.size">
        <a-form-item>
          <a-input-number v-model="filters.id" placeholder="VoteId" :min="0" />
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
          <a-table-column title="ID" data-index="id" />
          <a-table-column title="Type" data-index="type" />
          <a-table-column title="Title" data-index="title" />
          <a-table-column title="TopicId" data-index="topicId" />
          <a-table-column title="UserId" data-index="userId" />
          <a-table-column title="VoteNum" data-index="voteNum" />
          <a-table-column title="OptionCount" data-index="optionCount" />
          <a-table-column title="VoteCount" data-index="voteCount" />
          <a-table-column title="ExpiredAt" data-index="expiredAt">
            <template #cell="{ record }">
              {{ record.expiredAt ? useFormatDate(record.expiredAt) : '-' }}
            </template>
          </a-table-column>
          <a-table-column title="Create" data-index="createTime">
            <template #cell="{ record }">
              {{ record.createTime ? useFormatDate(record.createTime) : '-' }}
            </template>
          </a-table-column>
        </template>
      </a-table>
    </div>
  </div>
</template>

<script setup lang="ts">
  const appStore = useAppStore();
  const loading = ref(false);

  const filters = reactive({
    limit: 20,
    page: 1,
    id: undefined as number | undefined,
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
</script>

<style scoped lang="less"></style>
