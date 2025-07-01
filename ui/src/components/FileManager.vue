<script lang="ts" setup>
import { ref, computed, onMounted, nextTick, h, onUnmounted } from 'vue';
import {
  NFlex,
  NButton,
  NInput,
  NText,
  NScrollbar,
  NDataTable,
  NCard,
  NModal,
  NUpload,
  NUploadTrigger,
  NDrawer,
  NDrawerContent,
  NPopover,
  NIcon,
  useMessage,
  type DropdownOption,
} from 'naive-ui';

import {
  ChevronLeft,
  ChevronRight,
  Download,
  Folder,
  ListTree,
  File,
  PenLine,
  Plus,
  RefreshCcw,
  Search,
  Trash,
  Upload,
} from 'lucide-vue-next';

import { useI18n } from 'vue-i18n';
const { t } = useI18n();

// 添加类型定义
interface RowData {
  name: string;
  is_dir: boolean;
  size?: number;
  [key: string]: any;
}

const props = defineProps({
  files: {
    type: Array,
    default: () => [],
  },
});

const handlePathBack = () => {};

const handlePathForward = () => {};

const handlePathClick = (item: any) => {
  // 处理路径点击事件
  console.log('Path clicked:', item);
};

const filePathList = ref([
  {
    id: 1,
    active: false,
    path: 'root',
    showArrow: false,
  },
  {
    id: 2,
    active: true,
    path: 'home',
    showArrow: true,
  },
]);

const searchValue = ref('');

const handleNewFolder = () => {
  // 处理新建文件夹事件
  console.log('New folder created');
};

const isShowUploadList = ref(false);
const uploadFileList = ref([]);
const customRequest = (options: any) => {
  // 自定义上传请求逻辑
  console.log('Custom upload request:', options);
  // 模拟上传成功
  setTimeout(() => {
    options.onSuccess({}, options.file);
  }, 1000);
};
const handleRemoveItem = (file: any) => {
  // 处理文件移除事件
  console.log('File removed:', file);
};
const handleUploadFileChange = (file: any) => {
  // 处理上传文件变化事件
  console.log('Upload file changed:', file);
  if (file.status === 'done') {
  } else if (file.status === 'error') {
  }
};

const scrollRef = ref(null);
const showInner = ref(false);
const drawerHeight = ref(300);
const handleShowInner = () => {
  showInner.value = !showInner.value;
  nextTick(() => {
    if (scrollRef.value) {
    }
  });
};
const handleRefresh = () => {
  // 刷新逻辑
  console.log('Refresh clicked');
};
const onClickOutside = () => {
  showDropdown.value = false;
};
const dataList = ref<RowData[]>(props.files as RowData[]);

const columns = [
  {
    title: t('Name'),
    key: 'name',
    ellipsis: {
      tooltip: true,
    },
    render(row: RowData) {
      const fileIcon = h(NIcon, {
        size: 18,
        component: row.is_dir ? Folder : File,
        style: { marginRight: '8px' },
      });

      const fileName = h(
        NPopover,
        {
          delay: 500,
          placement: 'top-start',
          style: { maxWidth: '485px' },
        },
        {
          trigger: () =>
            h(
              NText,
              {
                depth: 1,
                strong: true,
                style: {
                  cursor: 'pointer',
                  maxWidth: '200px',
                  overflow: 'hidden',
                  textOverflow: 'ellipsis',
                  whiteSpace: 'nowrap',
                },
              },
              { default: () => row.name },
            ),
          default: () =>
            h(
              NText,
              { style: { maxWidth: '300px', wordBreak: 'break-all' } },
              { default: () => row.name },
            ),
        },
      );

      return h(
        NFlex,
        {
          align: 'center',
          style: { gap: '0px' },
        },
        {
          default: () => [
            fileIcon,
            h(
              NFlex,
              {
                vertical: true,
                style: { gap: '0px' },
              },
              {
                default: () => [fileName].filter(Boolean),
              },
            ),
          ],
        },
      );
    },
  },
];

onMounted(() => {
  // dataList.value =  props.files
});

const currentRowData = ref<RowData | null>(null);
// const columns = [
//   { key: 'name', title: t('Name'), width: 200 },
// ];

// 动态设置 dropdown options
const options = computed(() => {
  if (!currentRowData.value) return [];

  const baseOptions = [];

  // 下载选项 - 对所有文件和文件夹都显示
  baseOptions.push({
    key: 'download',
    label: t('Download'),
    icon: () => h(Download, { size: 16 }),
    show: !currentRowData.value.is_dir,
  });

  baseOptions.push({
    key: 'open',
    label: t('Open'),
    icon: () => h(Folder, { size: 16 }),
    show: currentRowData.value.is_dir,
  });

  return baseOptions;
});

const showDropdown = ref(false);

const x = ref(0);
const y = ref(0);
const rowProps = (row: RowData) => {
  return {
    style: 'cursor: pointer',
    onContextmenu: (e: MouseEvent) => {
      currentRowData.value = row;
      console.log('Row data:', row);
      e.preventDefault();

      showDropdown.value = false;
      console.log('Row right-clicked:', e);
      nextTick().then(() => {
        showDropdown.value = true;
        x.value = e.clientX;
        y.value = e.clientY;
      });
    },
    onClick: () => {
      console.log('Row clicked:', row);
    },
  };
};

const handleSelect = (key: string) => {
  showDropdown.value = false;
  console.log('Selected option:', key);
  switch (key) {
    case 'rename': {
      // modalType.value = 'rename';
      // showModal.value = true;
      // modalTitle.value = t('Rename');

      break;
    }
    case 'delete': {
      // modalType.value = 'delete';
      // showModal.value = true;
      // modalTitle.value = t('ConfirmDelete');
      // modalContent.value = t('DangerWarning');
      break;
    }
    case 'download': {
      // mittBus.emit('download-file', {
      //   path: `${fileManageStore.currentPath}/${currentRowData?.value?.name as string}`,
      //   is_dir: currentRowData.value.is_dir!,
      //   size: currentRowData.value.size!,
      // });

      break;
    }
  }
};
</script>

<template>
  <n-flex align="center" justify="space-between" vertical class="!gap-x-6">
    <n-flex align="center" class="w-full !flex-nowrap">
      <n-flex class="controls-part !gap-x-6 h-full !flex-nowrap" align="center">
        <n-button text :disabled="true" @click="handlePathBack">
          <ChevronLeft :size="16" class="icon-hover" />
        </n-button>

        <n-button text :disabled="true" @click="handlePathForward">
          <ChevronRight :size="16" class="icon-hover" />
        </n-button>
      </n-flex>
      <n-scrollbar ref="scrollRef" x-scrollable :content-style="{ height: '40px' }">
        <n-flex class="file-part w-full h-full !flex-nowrap">
          <n-flex
            v-for="item of filePathList"
            :key="item.id"
            align="center"
            justify="flex-start"
            class="file-node !flex-nowrap"
          >
            <Folder :size="18" :color="item.active ? '#63e2b7' : ''" class="text-white" />
            <n-text
              depth="1"
              class="text-[16px] cursor-pointer whitespace-nowrap"
              :strong="item.active"
              @click="handlePathClick(item)"
            >
              {{ item.path }}
            </n-text>

            <ChevronRight v-if="item.showArrow" :size="16" class="text-white" />
          </n-flex>
        </n-flex>
      </n-scrollbar>
    </n-flex>

    <n-flex align="center" justify="space-between" class="w-full !flex-nowrap">
      <n-input v-model:value="searchValue" clearable size="small" :placeholder="t('PleaseInput')">
        <template #prefix>
          <Search :size="16" class="focus:outline-none" />
        </template>
      </n-input>
      <n-flex align="center" class="!flex-nowrap">
        <n-button secondary size="small" class="custom-button-text" @click="handleNewFolder">
          <template #icon>
            <Plus :size="12" />
          </template>
          {{ t('NewFolder') }}
        </n-button>
        <n-upload
          v-model:file-list="uploadFileList"
          abstract
          :multiple="false"
          :show-retry-button="false"
          :custom-request="customRequest"
          @remove="handleRemoveItem"
          @change="handleUploadFileChange"
        >
          <n-button-group>
            <n-upload-trigger #="{ handleClick }" abstract>
              <n-button
                secondary
                size="small"
                class="custom-button-text"
                @click="
                  () => {
                    handleClick();
                    isShowUploadList = !isShowUploadList;
                  }
                "
              >
                <template #icon>
                  <NIcon :component="Upload" :size="12" />
                </template>

                {{ t('UploadFile') }}
              </n-button>
              <n-drawer
                v-model:show="showInner"
                resizable
                placement="bottom"
                :default-height="drawerHeight"
                :max-height="drawerHeight"
                :show-mask="false"
                :trap-focus="false"
                :block-scroll="false"
                :native-scrollbar="false"
                :height="300"
                to="#drawer-inner-target"
              >
                <n-drawer-content
                  closable
                  :title="t('TransferHistory')"
                  :body-style="{
                    overflow: 'unset',
                    height: '100%',
                    display: 'flex',
                    flexDirection: 'column',
                  }"
                >
                  <n-scrollbar
                    v-if="uploadFileList"
                    :style="{ maxHeight: `${drawerHeight - 60}px`, flex: 1 }"
                  >
                    <n-upload-file-list />
                  </n-scrollbar>

                  <n-empty v-else class="w-full h-full justify-center" />
                </n-drawer-content>
              </n-drawer>
            </n-upload-trigger>
          </n-button-group>
        </n-upload>

        <!-- <n-popover>
          <template #trigger>
            <ListTree
              :size="16"
              class="icon-hover cursor-pointer !text-white focus:outline-none"
              @click="handleOpenTransferList"
            />
          </template>
          {{ t('TransferHistory') }}
        </n-popover> -->

        <n-popover>
          <template #trigger>
            <RefreshCcw
              :size="16"
              class="icon-hover cursor-pointer !text-white focus:outline-none"
              @click="handleRefresh"
            />
          </template>
          {{ t('Refresh') }}
        </n-popover>
      </n-flex>
    </n-flex>
    <n-flex class="mt-4">
      <n-card size="small">
        <n-data-table
          single-line
          flex-height
          virtual-scroll
          size="small"
          :bordered="false"
          :columns="columns"
          :row-props="rowProps"
          :data="dataList"
          :style="{ height: 'calc(100vh - 240px)' }"
        >
          <template #empty>
            <n-empty class="w-full h-full justify-center" :description="t('NoData')" />
          </template>
        </n-data-table>
        <n-dropdown
          size="small"
          trigger="manual"
          placement="bottom-start"
          :x="x"
          :y="y"
          :show-arrow="true"
          v-model:options="options"
          :show="showDropdown"
          @clickoutside="onClickOutside"
          @select="handleSelect"
        />
      </n-card>
    </n-flex>
  </n-flex>
</template>

<style scoped lang="scss">
:deep(.n-drawer .n-drawer-content .n-drawer-body) {
  overflow: unset !important;
}
</style>
