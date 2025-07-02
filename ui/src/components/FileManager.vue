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
  NCollapseItem,
  NCollapse,
  NDivider,
  NProgress,
} from 'naive-ui';

import { NAvatar, useNotification } from 'naive-ui';

import type {
  UploadCustomRequestOptions,
  NotificationOptions,
  NotificationReactive,
  UploadSettledFileInfo,
  UploadFileInfo,
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
import { defineEmits } from 'vue';
import { useDebounceFn } from '@vueuse/core';

const { t } = useI18n();

const emit = defineEmits(['open-folder', 'download-file', 'upload-file', 'remove-upload-file']);
const message = useMessage();

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
  name: {
    type: String,
    default: '',
  },
  folder: {
    type: Object,
    default: () => ({}),
  },
  loading: {
    type: Boolean,
    default: false,
  },
  displayUploadingFiles: {
    type: Array as () => UploadSettledFileInfo[],
    default: () => [],
  },
});

const handlePathBack = () => {
  if (!props.folder || !props.folder.parent) return;
  // 处理路径后退事件
  storeBackFolders.value.push(props.folder); // 保存当前文件夹到后退存储
  emit('open-folder', props.folder.parent);
};

const handlePathForward = () => {
  if (!storeBackFolders.value || storeBackFolders.value.length === 0) return;
  // 处理路径前进事件
  const nextFolder = storeBackFolders.value.pop(); // 获取第一个文件夹
  if (nextFolder) {
    emit('open-folder', nextFolder);
  }
};

const handlePathClick = (item: any) => {
  // 处理路径点击事件
  storeBackFolders.value.length = 0; // 清空后退存储
  emit('open-folder', item.row);
};

const filePathList = computed(() => {
  if (!props.folder) return [];
  const list = [];
  let currentFolder = props.folder;
  let parent = currentFolder?.parent;
  let index = 0;
  list.push({
    id: index,
    active: true,
    name: currentFolder.name,
    path: currentFolder.name,
    showArrow: false,
    row: currentFolder,
  });
  while (parent !== null) {
    currentFolder = parent;
    parent = currentFolder.parent;
    index++;
    list.unshift({
      id: index,
      active: false,
      name: currentFolder.name,
      path: currentFolder.name,
      showArrow: true,
      row: currentFolder,
    });
  }

  return list;
});
const searchValue = ref('');

const isShowUploadList = ref(false);
const uploadFileList = ref([]);

const customRequest = (options: UploadCustomRequestOptions) => {
  // 自定义上传请求逻辑
  const { onFinish, onError, file } = options;
  emit('upload-file', options, props.folder);
};

const handleUploadFileChange = useDebounceFn((file: any) => {
  // 处理上传文件变化事件
  // console.log('Upload file changed:', file);
}, 100);

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
  emit('open-folder', props.folder);
};
const onClickOutside = () => {
  showDropdown.value = false;
};
const dataList = computed(() => {
  return props.files.filter((file: any) => {
    // 这里可以添加搜索过滤逻辑
    return file.name.toLowerCase().includes(searchValue.value.toLowerCase());
  }) as RowData[];
});

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

onMounted(() => {});

const currentRowData = ref<RowData | null>(null);
const storeBackFolders = ref<any>([]);

const handleSearch = useDebounceFn(() => {
  console.log('Search value:', searchValue.value);
}, 300);

const disabledBack = computed(() => {
  // 禁用后退按钮的逻辑
  return !props.folder || !props.folder.parent;
});

const disabledForward = computed(() => {
  // 禁用前进按钮的逻辑
  return storeBackFolders.value === null || storeBackFolders.value.length === 0;
});

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
      e.preventDefault();

      showDropdown.value = false;
      nextTick().then(() => {
        showDropdown.value = true;
        x.value = e.clientX;
        y.value = e.clientY;
      });
    },
    onClick: () => {
      if (!row.is_dir) {
        return;
      }
      currentRowData.value = row;
      emit('open-folder', row);
    },
  };
};

const handleSelect = (key: string) => {
  showDropdown.value = false;
  switch (key) {
    case 'download': {
      // 处理下载逻辑
      if (currentRowData.value) {
        // 这里可以添加下载逻辑
        emit('download-file', currentRowData.value);
      }

      break;
    }
  }
};

const handleUploadFinish = (options: any) => {
  // 处理上传完成事件

  message.success(t('UploadSuccess') + ': ' + options.file.name);
};

const handleUploadError = (options: any) => {
  // 处理上传错误事件
  message.error(t('UploadError') + ': ' + options.file.name);
};

const removeUploadList = (options: any) => {
  // 处理移除上传文件列表
  const { file } = options;
  emit('remove-upload-file', file);
};

const tableHeight = computed(() => {
  if (!props.displayUploadingFiles || props.displayUploadingFiles.length === 0) {
    return 240;
  }
  return 300;
});
</script>

<template>
  <n-flex align="center" justify="space-between" vertical class="!gap-x-6">
    <n-flex align="center" class="w-full !flex-nowrap">
      <n-flex class="controls-part !gap-x-6 h-full !flex-nowrap" align="center">
        <n-button text :disabled="disabledBack" @click="handlePathBack">
          <ChevronLeft :size="16" class="icon-hover" />
        </n-button>

        <n-button text :disabled="disabledForward" @click="handlePathForward">
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
      <n-input
        v-model:value="searchValue"
        clearable
        size="small"
        @change="handleSearch"
        :placeholder="t('PleaseInput')"
      >
        <template #prefix>
          <Search :size="16" class="focus:outline-none" />
        </template>
      </n-input>
      <n-flex align="center" class="!flex-nowrap">
        <n-upload
          v-model:file-list="uploadFileList"
          abstract
          :multiple="false"
          :show-retry-button="false"
          @finish="handleUploadFinish"
          @error="handleUploadError"
          :custom-request="customRequest"
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
      <n-card
        size="small"
        :segmented="{
          content: true,
          footer: 'soft',
        }"
      >
        <n-data-table
          :loading="loading"
          single-line
          flex-height
          virtual-scroll
          size="small"
          :bordered="false"
          :columns="columns"
          :row-props="rowProps"
          :data="dataList"
          :style="{ height: 'calc(100vh - ' + tableHeight + 'px)' }"
        >
          <template #empty>
            <n-empty class="w-full h-full justify-center" :description="t('NoData')" />
          </template>
        </n-data-table>

        <template #footer v-if="displayUploadingFiles.length > 0">
          <n-upload
            abstract
            :file-list-class="'max-height-32'"
            :show-preview-button="false"
            :show-retry-button="false"
            @remove="removeUploadList"
            :file-list="props.displayUploadingFiles"
          >
            <n-upload-file-list />
          </n-upload>
        </template>
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
:deep(.n-drawer .n-drawer-content .n-drawer-body .n-upload-file-list) {
  overflow: unset !important;
}

:deep(.max-height-32) {
  max-height: 60px;
  overflow: scroll;
  scrollbar-width: none;
}
</style>
