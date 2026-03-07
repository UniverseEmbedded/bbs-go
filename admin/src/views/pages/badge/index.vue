<template>
  <div class="container">
    <div class="container-header">
      <a-form :model="filters" layout="inline" :size="appStore.table.size">
        <a-form-item>
          <a-input v-model="filters.name" placeholder="Name" />
        </a-form-item>
        <a-form-item>
          <a-input v-model="filters.title" placeholder="Title" />
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
          <a-table-column title="Name" data-index="name" />
          <a-table-column title="Title" data-index="title" />
          <a-table-column title="Icon" data-index="icon">
            <template #cell="{ record }">
              <a-image
                v-if="record.icon"
                :src="record.icon"
                :width="32"
                :height="32"
                :preview="true"
              />
              <span v-else>-</span>
            </template>
          </a-table-column>
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
        <a-form-item field="name" label="Name">
          <a-input v-model="form.name" />
        </a-form-item>
        <a-form-item field="title" label="Title">
          <a-input v-model="form.title" />
        </a-form-item>
        <a-form-item field="description" label="Description">
          <a-textarea v-model="form.description" allow-clear />
        </a-form-item>
        <a-form-item field="icon" label="Icon">
          <ImageUpload v-model="form.icon" />
        </a-form-item>
        <a-form-item field="sortNo" label="SortNo">
          <a-input-number v-model="form.sortNo" :min="0" />
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
  import ImageUpload from '@/components/ImageUpload.vue';

  const appStore = useAppStore();
  const loading = ref(false);
  const formRef = ref();

  const filters = reactive({
    limit: 20,
    page: 1,
    name: undefined as string | undefined,
    title: undefined as string | undefined,
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
    name: '',
    title: '',
    description: '',
    icon: '',
    sortNo: 0,
    status: 0,
  });

  const form = ref<any>(emptyForm());
  const rules = {
    name: [{ required: true, message: 'Name required' }],
    title: [{ required: true, message: 'Title required' }],
  };

  onMounted(() => {
    useTableHeight();
  });

  const list = async () => {
    loading.value = true;
    try {
      const ret = await axios.postForm<any>(
        '/api/admin/badge/list',
        jsonToFormData(filters)
      );
      data.page = ret.page;
      data.results = ret.results;
    } finally {
      loading.value = false;
    }
  };

  list();

  const showAdd = () => {
    formRef.value?.resetFields?.();
    form.value = emptyForm();
    edit.isCreate = true;
    edit.title = 'New Badge';
    edit.visible = true;
  };

  const showEdit = async (id: any) => {
    formRef.value?.resetFields?.();
    edit.isCreate = false;
    edit.title = 'Edit Badge';
    try {
      form.value = await axios.get(`/api/admin/badge/${id}`);
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
        ? '/api/admin/badge/create'
        : '/api/admin/badge/update';
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
