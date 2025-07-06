<script lang="ts" setup>
import { nextTick, onMounted, onUnmounted, ref, useTemplateRef, watch, h, computed } from 'vue';
import { useWindowSize } from '@vueuse/core';
import { useDebounceFn } from '@vueuse/core';
// @ts-ignore
// import Guacamole from 'guacamole-common-js';
import Guacamole from '@dushixiang/guacamole-common-js';
import type { UploadCustomRequestOptions, UploadFileInfo, UploadSettledFileInfo } from 'naive-ui';
import { NSpin, useMessage, NTabPane } from 'naive-ui';
import { useI18n } from 'vue-i18n';
import { getCurrentConnectParams, BaseAPIURL } from '@/utils/common';
import { getSupportedGuacMimeTypes } from '@/utils/guacamole_helper';
import type { GuacamoleDisplay } from '@/types/guacamole.type';
import { lunaCommunicator } from '@/utils/lunaBus.ts';
import { LUNA_MESSAGE_TYPE } from '@/types/postmessage.type';
import { ErrorStatusCodes, ConvertGuacamoleError } from '@/utils/status';
import { LanguageCode } from '@/locales';
import ClipBoardText from '@/components/ClipBoardText.vue';
import SessionShare from '@/components/SessionShare.vue';
import FileManager from '@/components/FileManager.vue';
import { readClipboardText } from '@/utils/clipboard';

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
  NProgress,
  useNotification,
} from 'naive-ui';

const message = useMessage();
const { t } = useI18n();
const { width, height } = useWindowSize();

import { useGuacamoleClient } from '@/hooks/useGuacamoleClient';

const {
  guaDisplay,
  connectToGuacamole,
  onlineUsersMap,
  disconnectGuaclient,
  sendTextToRemote,
  sendKeyEvent,
  uploadFile,
  clientFileReceived,
  resizeGuaScale,
  sendGuaSize,
  scale,
  handleFolderOpen,
  driverName,
  loading,
  registerMouseAndKeyboardHanlder,
  sessionObject,
  currentFolder,
  currentFolderFiles,
  hasClipboardPermission,
} = useGuacamoleClient(t);

const apiPrefix = ref('');
const wsPrefix = ref('');

const drawShow = ref(false);
const connectStatus = ref('Connecting');
const fileFsloading = ref(false);

const remoteClipboardText = ref<string>('');
const autoFit = ref<boolean>(true);
const debouncedResize = useDebounceFn(() => {
  resizeGuaScale(width.value, height.value);
  if (!autoFit.value) {
    return;
  }
  sendGuaSize(width.value, height.value);
}, 300);

watch(
  [width, height],
  ([newWidth, newHeight]) => {
    debouncedResize();
  },
  { immediate: true },
);

interface GuacamoleFile {
  mimetype?: any;
  streamName?: any;
  type: 'DIRECTORY' | 'FILE';
  name: string;
  parent?: GuacamoleFile | null;
  is_dir?: boolean;
}

interface UploadItem {
  uploadOptions: UploadCustomRequestOptions;
  folder: GuacamoleFile;
}

const uploadingFiles = ref<Array<UploadItem>>([]);
const isUploading = ref(false);

const displayUploadingFiles = ref<Array<UploadSettledFileInfo>>([]);

const handleUploadFile = (options: UploadCustomRequestOptions, folder: any) => {
  const item = {
    uploadOptions: options,
    folder: folder || currentFolder.value,
  };
  displayUploadingFiles.value.push(options.file);
  uploadingFiles.value.push(item);
  if (isUploading.value) {
    console.warn('Already uploading files, skipping new upload:', options.file.name);
    return;
  }
  isUploading.value = true;

  processUploadQueue().then(() => {
    handleFolderOpen(currentFolder.value);
  });
};

const handleRemoveFile = (file: any) => {
  if (file.status === 'uploading') {
    message.warning(t('FileUploadingWarning'));
    return;
  }
  const newDisplayFiles = displayUploadingFiles.value.filter((f) => {
    return f.name !== file.name;
  });
  nextTick(() => {
    displayUploadingFiles.value = newDisplayFiles;
  });
};

const processUploadQueue = async () => {
  while (isUploading.value && uploadingFiles.value.length > 0) {
    const UploadItem = uploadingFiles.value.shift();
    if (!UploadItem || !UploadItem.uploadOptions) {
      continue;
    }
    const { uploadOptions, folder } = UploadItem;

    try {
      uploadOptions.file.status = 'uploading';
      await uploadFile(uploadOptions, folder);
      uploadOptions.file.status = 'finished';
      setTimeout(() => {
        handleRemoveFile(uploadOptions.file);
      }, 1000 * 5); // 延迟5秒后移除上传文件
    } catch (error) {
      console.error('Error processing upload queue:', error);
      message.error(t('FileUploadError') + ': ' + error);
      uploadOptions.file.status = 'error';
    }
  }
  isUploading.value = false;
};

const showOsk = ref<boolean>(false);

const fileDrop = (event: any) => {
  event.stopPropagation();
  event.preventDefault();
  const files = event.dataTransfer.files;
  if (files.length === 0) {
    return;
  }
  console.log('Files dropped:', files);
  // Handle file drop logic here
  // For example, you can upload the files or process them as needed
  // This is a placeholder for actual file handling logic
};

const debouncedSendClipboardToRemote = useDebounceFn(async () => {
  const text = await readClipboardText();
  if (!text || !text.trim()) {
    return;
  }
  sendTextToRemote(text);
}, 300);

onMounted(async () => {
  loading.value = true;
  const handLunaOpen = (message: any) => {
    console.log('Received Luna command:', message);
    nextTick(() => {
      drawShow.value = !drawShow.value;
    });
  };
  lunaCommunicator.onLuna(LUNA_MESSAGE_TYPE.OPEN, handLunaOpen);
  const params = getCurrentConnectParams();
  wsPrefix.value = params.ws || '';
  apiPrefix.value = params.api || '';
  const token = params['data'].token || '';
  const param = {
    TOKEN_ID: encodeURIComponent(token),
  };
  connectToGuacamole(wsPrefix.value, param, window.innerWidth, window.innerHeight, true);
  const displayEl = document.getElementById('display');
  if (!displayEl) {
    console.error('Display element not found');
    return;
  }
  displayEl.appendChild(guaDisplay.value.getElement());

  displayEl.addEventListener(
    'dragenter',
    function (e: any) {
      e.stopPropagation();
      e.preventDefault();
    },
    false,
  );
  displayEl.addEventListener(
    'dragover',
    function (e: any) {
      e.stopPropagation();
      e.preventDefault();
    },
    false,
  );
  displayEl.addEventListener('drop', fileDrop, false);
  registerMouseAndKeyboardHanlder();
  window.addEventListener('focus', () => {
    console.log('Window focused, sending clipboard text to remote');
    debouncedSendClipboardToRemote();
  });
});

onUnmounted(() => {
  disconnectGuaclient();
  lunaCommunicator.offLuna(LUNA_MESSAGE_TYPE.OPEN);
  lunaCommunicator.sendLuna(LUNA_MESSAGE_TYPE.CLOSE, '');
  window.removeEventListener, 'focus';
});

const ClipBoardTextChange = (text: string) => {
  if (!text || !text.trim()) {
    return;
  }
  console.log('ClipBoardTextChange:', text);
  sendTextToRemote(text);
};

document.addEventListener(
  'contextmenu',
  (e: MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
  },
  false,
);
import Osk from '@/components/Osk.vue';
import KeyboardOption from '@/components/KeyboardOption.vue';
import OtherOption from '@/components/OtherOption.vue';
const keyboardLayout = ref<string>('en-us-qwerty');

const handleScreenKeyboard = (name: string, keysym: any) => {
  console.log('Screen keyboard change:', name, keysym);
  switch (name) {
    case 'keydown':
      sendKeyEvent(1, keysym);
      break;
    case 'keyup':
      sendKeyEvent(0, keysym);
      break;
    default:
      console.warn('Unknown screen keyboard event:', name);
  }
};

const handleDownloadFile = (file: GuacamoleFile) => {
  if (!file || !file.streamName) {
    console.warn('Cannot download file, file is not valid:', file);
    return;
  }
  console.log('Downloading file:', file.name);
  clientFileReceived(file.streamName, file.name, file.mimetype);
};

const fitPercentage = computed(() => {
  return Math.floor(scale.value * 100);
});

watch(
  [autoFit],
  ([newAutoFit]) => {
    if (newAutoFit) {
      debouncedResize();
    }
  },
  { immediate: true },
);

const handleCombineKeys = (keys: string[]) => {
  keys.forEach((keysym: any) => {
    sendKeyEvent(1, keysym);
  });
  setTimeout(() => {
    keys.forEach((keysym: any) => {
      sendKeyEvent(0, keysym);
    });
  }, 100);
};

const scaleGuacDisplay = (value: number) => {
  if (value <= 0) {
    console.warn('Invalid scale value:', scale);
    return;
  }
  console.log('Scaling Guacamole display to:', value);
  const newScale = value / 100; // 限制缩放范围在0.1到5之间

  guaDisplay.value.scale(newScale);
  scale.value = newScale;
};

const onlineUsers = computed(() => {
  const users: any = [];
  for (const userId in onlineUsersMap.value) {
    const user = onlineUsersMap.value[userId];
    if (user) {
      users.push(user);
    }
  }
  return users;
});
</script>

<template>
  <div class="w-full h-full justify-center flex flex-col">
    <div v-if="loading" class="flex justify-center items-center w-screen h-screen">
      <n-spin :show="loading" size="large" :description="`${t('Connecting')}: ${connectStatus}`">
      </n-spin>
    </div>
    <div
      id="display"
      v-show="!loading"
      class="w-screen h-screen flex justify-center relative"
    ></div>
    <Osk v-if="showOsk" :keyboard="keyboardLayout" @keyboard-change="handleScreenKeyboard" />
  </div>

  <n-drawer v-model:show="drawShow" :min-width="502" :default-width="502" resizable>
    <n-drawer-content>
      <n-tabs default-value="settings" justify-content="space-evenly" type="line">
        <n-tab-pane name="settings" tab="设置">
          <ClipBoardText
            :disabled="!hasClipboardPermission"
            :remote-text="remoteClipboardText"
            @update:text="ClipBoardTextChange"
          />
          <br />
          <KeyboardOption v-model:opened="showOsk" v-model:keyboard="keyboardLayout" />
          <br />
          <OtherOption
            v-model:auto-fit="autoFit"
            :fit-percentage="fitPercentage"
            @combine-keys="handleCombineKeys"
            @update-scale="scaleGuacDisplay"
            :is-remote-app="false"
          />
        </n-tab-pane>
        <n-tab-pane name="file-manager" tab="文件管理">
          <FileManager
            :loading="fileFsloading"
            :files="currentFolderFiles"
            :name="driverName"
            :folder="currentFolder"
            :display-uploading-files="displayUploadingFiles"
            @open-folder="handleFolderOpen"
            @download-file="handleDownloadFile"
            @upload-file="handleUploadFile"
            @remove-upload-file="handleRemoveFile"
          />
        </n-tab-pane>
        <n-tab-pane name="share-collaboration" tab="分享会话" v-if="sessionObject">
          <SessionShare :session="sessionObject.id" :users="onlineUsers" />
        </n-tab-pane>
      </n-tabs>
    </n-drawer-content>
  </n-drawer>
</template>

<style scoped></style>
