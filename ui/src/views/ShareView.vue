<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
import { useMessage } from 'naive-ui';
import { getShareSession } from '@/api/index';
import { nextTick, onMounted, ref, computed, watch } from 'vue';
import Osk from '@/components/Osk.vue';
import { useGuacamoleClient } from '@/hooks/useGuacamoleClient';
import SessionShare from '@/components/SessionShare.vue';

import { useWindowSize, useDebounceFn } from '@vueuse/core';

const { width, height } = useWindowSize();
const message = useMessage();
const { t } = useI18n();
const wsUrl = '/lion/ws/share/';
const verifyValue = ref<string>('');
const showModal = ref<boolean>(true);
const shareCode = ref<string>('');
const onFinish = () => {
  shareCode.value = verifyValue.value;
  nextTick(() => {
    showModal.value = false;
    connectShareSession(shareCode.value);
  });
};

const debouncedResize = useDebounceFn(() => {
  resizeGuaScale(width.value, height.value);
}, 300);

watch(
  [width, height],
  ([newWidth, newHeight]) => {
    debouncedResize();
  },
  { immediate: true },
);

const params = ref<Record<string, string>>({
  type: 'share',
  SESSION_ID: '',
  SHARE_ID: '',
  RECORD_ID: '',
});
const readonly = ref<boolean>(false);

const handleKeyUp = (event: KeyboardEvent) => {
  console.log('Key pressed:', event.key);
  if (event.key === 'Enter') {
    onFinish();
  }
};
import { useRoute } from 'vue-router';

const route = useRoute();

const shareId = route.params.id as string;
// 从 route 中获取 id
const recordObj = ref<Record<string, any>>({});

const errMessage = ref<string>('');

const connectShareSession = (code: string) => {
  const data = {
    code: code,
  };
  getShareSession(shareId, data)
    .then((response: any) => response.json())
    .then((res) => {
      console.log('Share session response:', res);

      if (res.message && !res.success) {
        message.error(res.message || t('ShareSessionError'));
        loading.value = false;
        errMessage.value = res.message || t('ShareSessionError');
        return;
      }

      const actionPerm = res['action_permission'];
      if (actionPerm['value'] === 'readonly') {
        readonly.value = true;
      }
      const recordId = res.id;
      const sessionId = res.session.id;
      recordObj.value = res;
      const shareParams = {
        type: 'share',
        SESSION_ID: sessionId,
        SHARE_ID: shareId,
        RECORD_ID: recordId,
        Writable: readonly.value ? 'false' : 'true',
      };

      connectToGuacamole(wsUrl, shareParams, window.innerWidth, window.innerHeight);
      const displayEl = document.getElementById('display');
      if (displayEl) {
        displayEl.appendChild(guaDisplay.value.getElement());
      }
      if (readonly.value) {
        return;
      }
      registerMouseAndKeyboardHanlder();
    })
    .catch((error) => {
      console.error('Error fetching share session:', error);
      message.error(error.message || t('ShareSessionError'));
    });
};

const showOsk = ref<boolean>(false);
const keyboardLayout = ref<string>('default');
const handleScreenKeyboard = (layout: string) => {
  keyboardLayout.value = layout;
  showOsk.value = true;
};
const connectStatus = ref<string>('Connecting...');

const {
  connectToGuacamole,
  guaDisplay,
  loading,
  onlineUsersMap,
  registerMouseAndKeyboardHanlder,
  resizeGuaScale,
} = useGuacamoleClient(t);

const drawShow = ref<boolean>(false);
const sessionObject = ref<Record<string, any>>({});

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
onMounted(() => {
  if (route.query.code) {
    shareCode.value = route.query.code as string;
    showModal.value = false;
    nextTick(() => {
      connectShareSession(shareCode.value);
    });
  }
});
</script>

<template>
  <n-modal
    v-model:show="showModal"
    preset="dialog"
    :title="t('VerifyCode')"
    :closable="false"
    :show-icon="false"
    :close-on-esc="false"
    :mask-closable="false"
  >
    <n-input
      v-model:value="verifyValue"
      :length="4"
      size="large"
      class="justify-center pb-3"
      :placeholder="t('VerifyCodePlaceholder')"
      show-password-on="mousedown"
      @keyup="handleKeyUp"
    />
  </n-modal>
  <div v-if="!showModal" class="w-full h-full justify-center flex flex-col">
    <div v-if="loading" class="flex justify-center items-center w-screen h-screen">
      <n-spin :show="loading && !errMessage" size="large" :description="`${t('Connecting')}: ${connectStatus}`">
      </n-spin>
    </div>
    <div
      id="display"
      v-show="!loading && !errMessage"
      class="w-screen h-screen flex justify-center relative"
    ></div>
    <Osk v-if="showOsk" :keyboard="keyboardLayout" @keyboard-change="handleScreenKeyboard" />
    <div v-if="errMessage" class="text-red-900 font-bold text-xl"> {{ errMessage }}</div>
  </div>
 


  <n-drawer v-model:show="drawShow" :min-width="502" :default-width="502" resizable>
    <n-drawer-content>
      <n-tabs default-value="settings" justify-content="space-evenly" type="line">
        <n-tab-pane name="share-collaboration" tab="分享会话" v-if="sessionObject">
          <SessionShare :session="sessionObject.id" :users="onlineUsers" :disable-create="true" />
        </n-tab-pane>
      </n-tabs>
    </n-drawer-content>
  </n-drawer>
</template>
