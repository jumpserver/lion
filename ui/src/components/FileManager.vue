<script lang="ts" setup>
import { ref, computed, onMounted, nextTick, h } from 'vue';
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
} from 'naive-ui';

import {
  ChevronLeft,
  ChevronRight,
  Download,
  Folder,
  ListTree,
  PenLine,
  Plus,
  RefreshCcw,
  Search,
  Trash,
  Upload,
} from 'lucide-vue-next';

import { useI18n } from 'vue-i18n';
const { t } = useI18n();

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
const dataList = [
  // 模拟数据列表
  { id: 1, name: 'File1.txt', size: '1MB', type: 'Text File' },
  { id: 2, name: 'File2.jpg', size: '2MB', type: 'Image File' },
  { id: 3, name: 'File3.mp4', size: '5MB', type: 'Video File' },
];



const columns = [
  { key: 'name', title: t('Name'), width: 200 },
  { key: 'size', title: t('Size'), width: 100 },
];
const options: DropdownOption[] = [
  {
    key: 'rename',
    label: t('Rename'),
    icon: () => h(PenLine, { size: 16 }),
  },
  {
    key: 'download',
    label: t('Download'),
    icon: () => h(Download, { size: 16 }),
  },
  {
    type: 'divider',
    key: 'd1',
  },
  {
    key: 'delete',
    icon: () => h(Trash, { size: 16, color: '#ff6b6b' }),
    label: () =>
      h(NText, { depth: 1, style: { color: '#ff6b6b' } }, { default: () => t('Delete') }),
  },
];
const showDropdown = ref(false);
const currentRowData = ref(null);
const x = ref(0);
const y = ref(0);
const rowProps = (row: RowData) => {
  return {
    style: 'cursor: pointer',
    onContextmenu: (e: MouseEvent) => {
      currentRowData.value = row;

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
          :options="options"
          :show="showDropdown"
          @clickoutside="onClickOutside"
          @select="handleSelect"
        />
      </n-card>
    </n-flex>
  </n-flex>
</template>
