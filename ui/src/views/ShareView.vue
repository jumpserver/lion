<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
import { useDebounceFn } from '@vueuse/core';
import { Delete, Undo2 } from 'lucide-vue-next';
import { NTag, useDialogReactiveList, useMessage } from 'naive-ui';
import type { Composer } from 'vue-i18n';

import { nextTick, onMounted, ref } from 'vue';

const { t } = useI18n();

const verifyValue = ref<string[]>([]);
const showModal = ref<boolean>(true);
const shareCode = ref<string>('');
const onFinish = () => {
  (shareCode.value = verifyValue.value.join('')),
    nextTick(() => {
      showModal.value = false;
    });
};

onMounted(() => {});
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
    <n-input-otp
      v-model:value="verifyValue"
      :length="4"
      size="large"
      class="justify-center pb-3"
      @finish="onFinish"
    />
  </n-modal>
</template>
