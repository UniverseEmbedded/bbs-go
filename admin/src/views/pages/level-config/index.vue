<template>
  <div class="container">
    <div class="container-header">
      <a-form :model="filters" layout="inline" :size="appStore.table.size">
        <a-form-item>
          <a-select
            v-model="filters.status"
            placeholder="Status"
            allow-clear
            @change="list"
          >
            <a-option :value="0" label="Normal" />
            <a-option :value="1" label="Deleted" />
          </a-select>
        </a-form-item>
        <a-form-item>
          <a-button type="primary" html-type="submit" @click="list">
            <template #icon> <icon-refresh /> </template>
            Refresh
          </a-button>
        </a-form-item>
      </a-form>

      <div class="action-btns">
        <a-button :size="appStore.table.size" @click="addRow">Add Row</a-button>
        <a-button
          type="primary"
          :loading="saving"
          :size="appStore.table.size"
          @click="saveAll"
        >
          Save All
        </a-button>
      </div>
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
          <a-table-column title="Level" data-index="level">
            <template #cell="{ record }">
              <a-input-number v-model="record.level" :min="1" />
            </template>
          </a-table-column>
          <a-table-column title="NeedExp" data-index="needExp">
            <template #cell="{ record }">
              <a-input-number v-model="record.needExp" :min="0" />
            </template>
          </a-table-column>
          <a-table-column title="Title" data-index="title">
            <template #cell="{ record }">
              <a-input v-model="record.title" />
            </template>
          </a-table-column>
          <a-table-column title="Status" data-index="status">
            <template #cell="{ record }">
              <a-select v-model="record.status" style="width: 120px">
                <a-option :value="0" label="Normal" />
                <a-option :value="1" label="Deleted" />
              </a-select>
            </template>
          </a-table-column>
          <a-table-column title="Update" data-index="updateTime">
            <template #cell="{ record }">
              {{ record.updateTime ? useFormatDate(record.updateTime) : '-' }}
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
  const saving = ref(false);

  const filters = reactive({
    limit: 200,
    page: 1,
    status: undefined as number | undefined,
  });

  const data = reactive({
    page: {
      page: 1,
      limit: 200,
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
        '/api/admin/level-config/list',
        jsonToFormData(filters)
      );
      data.page = ret.page;
      data.results = ret.results || [];
    } finally {
      loading.value = false;
    }
  };

  list();

  const addRow = () => {
    data.results.unshift({
      id: undefined,
      level: 1,
      needExp: 0,
      title: '',
      status: 0,
    });
  };

  const saveAll = async () => {
    saving.value = true;
    try {
      const items = (data.results || []).map((x) => ({
        id: x.id,
        level: Number(x.level || 0),
        needExp: Number(x.needExp || 0),
        title: x.title || '',
        status: Number(x.status || 0),
      }));
      await axios.post('/api/admin/level-config/save_all', items);
      useNotificationSuccess('Saved');
      list();
    } catch (e: any) {
      useHandleError(e);
    } finally {
      saving.value = false;
    }
  };

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
