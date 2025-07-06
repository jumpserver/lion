<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
import { useDebounceFn } from '@vueuse/core';
import { Delete, Undo2 } from 'lucide-vue-next';
import { NTag, useDialogReactiveList, useMessage } from 'naive-ui';
import type { Composer } from 'vue-i18n';
import { getShareSession } from '@/api/index';
import { nextTick, onMounted, ref, computed } from 'vue';
import Osk from '@/components/Osk.vue';
import { useGuacamoleClient } from '@/hooks/useGuacamoleClient';
const message = useMessage();
import { useRoute } from 'vue-router';

const route = useRoute();
const { t } = useI18n();

const sessionId = route.query.session as string;
const wsUrl = '/lion/ws/monitor/';
const params = {
  type: 'monitor',
  SESSION_ID: sessionId,
};
const readonly = ref<boolean>(true);


const connectStatus = ref<string>('Connecting...');

const { connectToGuacamole, guaDisplay, loading } = useGuacamoleClient(t);

onMounted(() => {
      connectToGuacamole(wsUrl, params, window.innerWidth, window.innerHeight);
      const displayEl = document.getElementById('display');
      if (displayEl) {
        displayEl.appendChild(guaDisplay.value.getElement());
      }
});
</script>

<template>
  <div class="w-full h-full justify-center flex flex-col">
    <div v-if="loading" class="flex justify-center items-center w-screen h-screen">
      <n-spin :show="loading" size="large" :description="`${t('Connecting')}: ${connectStatus}`">
      </n-spin>
    </div>
    <div id="display" v-show="!loading" class="w-screen h-screen"></div>
  </div>
</template>
