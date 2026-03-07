<template>
  <div class="container">
    <div class="container-header">
      <a-form :model="filters" layout="inline" :size="appStore.table.size">
        <a-form-item>
          <a-select
            v-model="filters.groupName"
            placeholder="Group"
            allow-clear
            @change="list"
          >
            <a-option
              v-for="g in groups"
              :key="g.key"
              :label="g.name"
              :value="g.key"
            />
          </a-select>
        </a-form-item>
        <a-form-item>
          <a-input v-model="filters.title" placeholder="Title" />
        </a-form-item>
        <a-form-item>
          <a-input v-model="filters.eventType" placeholder="EventType" />
        </a-form-item>
        <a-form-item>
          <a-select
            v-model="filters.period"
            placeholder="Period"
            allow-clear
            @change="list"
          >
            <a-option :value="0" label="Lifetime" />
            <a-option :value="1" label="Daily" />
            <a-option :value="2" label="Weekly" />
            <a-option :value="3" label="Monthly" />
            <a-option :value="4" label="Yearly" />
          </a-select>
        </a-form-item>
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
            <template #icon> <icon-search /> </template>
            Search
          </a-button>
        </a-form-item>
      </a-form>

      <div class="action-btns">
        <a-button type="primary" :size="appStore.table.size" @click="showAdd">
          <template #icon>
            <icon-plus />
          </template>
          Add
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
          <a-table-column title="ID" data-index="id" />
          <a-table-column title="Group" data-index="groupName" />
          <a-table-column title="Title" data-index="title" />
          <a-table-column title="EventType" data-index="eventType" />
          <a-table-column title="Score" data-index="score" />
          <a-table-column title="Exp" data-index="exp" />
          <a-table-column title="BadgeId" data-index="badgeId" />
          <a-table-column title="Period" data-index="period" />
          <a-table-column title="MaxFinish" data-index="maxFinishCount" />
          <a-table-column title="EventCount" data-index="eventCount" />
          <a-table-column title="Sort" data-index="sortNo" />
          <a-table-column title="Status" data-index="status" />
          <a-table-column title="Update" data-index="updateTime">
            <template #cell="{ record }">
              {{ record.updateTime ? useFormatDate(record.updateTime) : '-' }}
            </template>
          </a-table-column>
          <a-table-column title="Edit" data-index="edit">
            <template #cell="{ record }">
              <a-button
                type="primary"
                :size="appStore.table.size"
                @click="showEdit(record.id)"
              >
                Edit
              </a-button>
            </template>
          </a-table-column>
        </template>
      </a-table>
    </div>

    <a-modal
      v-model:visible="edit.visible"
      :title="edit.title"
      :size="appStore.table.size"
      @cancel="handleCancel"
      @before-ok="handleBeforeOk"
    >
      <a-form ref="formRef" :model="form" :rules="rules">
        <a-form-item field="groupName" label="Group">
          <a-select v-model="form.groupName">
            <a-option
              v-for="g in groups"
              :key="g.key"
              :label="g.name"
              :value="g.key"
            />
          </a-select>
        </a-form-item>
        <a-form-item field="title" label="Title">
          <a-input v-model="form.title" />
        </a-form-item>
        <a-form-item field="description" label="Description">
          <a-textarea v-model="form.description" allow-clear />
        </a-form-item>
        <a-form-item field="eventType" label="EventType">
          <a-input v-model="form.eventType" />
        </a-form-item>
        <a-form-item field="score" label="Score">
          <a-input-number v-model="form.score" :min="0" />
        </a-form-item>
        <a-form-item field="exp" label="Exp">
          <a-input-number v-model="form.exp" :min="0" />
        </a-form-item>
        <a-form-item field="badgeId" label="BadgeId">
          <a-input-number v-model="form.badgeId" :min="0" />
        </a-form-item>
        <a-form-item field="period" label="Period">
          <a-select v-model="form.period">
            <a-option :value="0" label="Lifetime" />
            <a-option :value="1" label="Daily" />
            <a-option :value="2" label="Weekly" />
            <a-option :value="3" label="Monthly" />
            <a-option :value="4" label="Yearly" />
          </a-select>
        </a-form-item>
        <a-form-item field="maxFinishCount" label="MaxFinishCount">
          <a-input-number v-model="form.maxFinishCount" :min="1" />
        </a-form-item>
        <a-form-item field="eventCount" label="EventCount">
          <a-input-number v-model="form.eventCount" :min="1" />
        </a-form-item>
        <a-form-item field="btnName" label="BtnName">
          <a-input v-model="form.btnName" />
        </a-form-item>
        <a-form-item field="actionUrl" label="ActionUrl">
          <a-input v-model="form.actionUrl" />
        </a-form-item>
        <a-form-item field="sortNo" label="SortNo">
          <a-input-number v-model="form.sortNo" :min="0" />
        </a-form-item>
        <a-form-item field="startTime" label="StartTime">
          <a-input-number v-model="form.startTime" :min="0" />
        </a-form-item>
        <a-form-item field="endTime" label="EndTime">
          <a-input-number v-model="form.endTime" :min="0" />
        </a-form-item>
        <a-form-item field="status" label="Status">
          <a-select v-model="form.status">
            <a-option :value="0" label="Normal" />
            <a-option :value="1" label="Deleted" />
          </a-select>
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
  const appStore = useAppStore();
  const loading = ref(false);
  const formRef = ref();

  const groups = ref<any[]>([]);

  const filters = reactive({
    limit: 20,
    page: 1,
    groupName: undefined as string | undefined,
    title: undefined as string | undefined,
    eventType: undefined as string | undefined,
    period: undefined as number | undefined,
    status: undefined as number | undefined,
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

  const edit = reactive({
    visible: false,
    isCreate: false,
    title: '',
  });

  const emptyForm = () => ({
    id: undefined,
    groupName: 'newbie',
    eventType: '',
    title: '',
    description: '',
    score: 0,
    exp: 0,
    badgeId: 0,
    period: 0,
    maxFinishCount: 1,
    eventCount: 1,
    btnName: '',
    actionUrl: '',
    sortNo: 0,
    startTime: 0,
    endTime: 0,
    status: 0,
  });

  const form = ref<any>(emptyForm());
  const rules = {
    groupName: [{ required: true, message: 'Group required' }],
    title: [{ required: true, message: 'Title required' }],
    eventType: [{ required: true, message: 'EventType required' }],
    description: [{ required: true, message: 'Description required' }],
  };

  onMounted(() => {
    useTableHeight();
  });

  const loadGroups = async () => {
    try {
      const ret = await axios.get<any, any[]>('/api/admin/task-config/groups');
      groups.value = ret || [];
    } catch (e: any) {
      useHandleError(e);
    }
  };

  const list = async () => {
    loading.value = true;
    try {
      const ret = await axios.postForm<any>(
        '/api/admin/task-config/list',
        jsonToFormData(filters)
      );
      data.page = ret.page;
      data.results = ret.results;
    } finally {
      loading.value = false;
    }
  };

  loadGroups();
  list();

  const showAdd = () => {
    formRef.value?.resetFields?.();
    form.value = emptyForm();
    edit.isCreate = true;
    edit.title = 'New Task';
    edit.visible = true;
  };

  const showEdit = async (id: any) => {
    formRef.value?.resetFields?.();
    edit.isCreate = false;
    edit.title = 'Edit Task';
    try {
      form.value = await axios.get(`/api/admin/task-config/${id}`);
    } catch (e: any) {
      useHandleError(e);
    }
    edit.visible = true;
  };

  const handleCancel = () => {
    formRef.value?.resetFields?.();
  };

  const handleBeforeOk = async (done: (closed: boolean) => void) => {
    const validateErr = await formRef.value.validate();
    if (validateErr) {
      done(false);
      return;
    }
    try {
      const url = edit.isCreate
        ? '/api/admin/task-config/create'
        : '/api/admin/task-config/update';
      await axios.postForm<any>(url, jsonToFormData(form.value));
      useNotificationSuccess('Submit success');
      edit.visible = false;
      list();
      done(true);
    } catch (e: any) {
      useHandleError(e);
      done(false);
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
