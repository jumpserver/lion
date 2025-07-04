<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
import { useDebounceFn } from '@vueuse/core';
import { Delete, Undo2 } from 'lucide-vue-next';
import { computed, h, reactive, ref, watch } from 'vue';
import { NTag, useDialogReactiveList, useMessage } from 'naive-ui';
import type { Composer } from 'vue-i18n';
import CreateShareCard from './CreateShare.vue';

export type TranslateFunction = Composer['t'];

const props = defineProps<{
  session: string;
  users?: Array<{
    user_id: string;
    user: string;
    primary: boolean;
    writable: boolean;
  }>;
}>();

const { t } = useI18n();

const showModal = ref<boolean>(false);

const onlineUsers = ref<
  Array<{ user_id: string; user: string; primary: boolean; writable: boolean }>
>(props.users || []);

const handleRemoveShareUser = (user: {
  user_id: string;
  user: string;
  primary: boolean;
  writable: boolean;
}) => {
  onlineUsers.value = onlineUsers.value.filter((u) => u.user_id !== user.user_id);
};

const openModal = () => {
  showModal.value = true;
};
</script>

<template>
  <n-flex vertical align="center">
    <n-divider dashed title-placement="left" class="!mb-3 !mt-0">
      <n-text depth="2" class="text-sm opacity-70">
        {{ t('User') }} {{ onlineUsers?.length || 0 }}
      </n-text>
    </n-divider>

    <n-flex v-if="onlineUsers?.length" class="w-full mb-4">
      <n-list class="w-full" bordered hoverable>
        <n-list-item v-for="user in onlineUsers" :key="user.user_id">
          <template #suffix>
            <Delete
              v-if="!user.primary"
              :size="18"
              class="cursor-pointer hover:text-red-500 transition-all duration-200"
              @click="handleRemoveShareUser(user)"
            />
          </template>

          <n-flex vertical>
            <n-text>{{ user.user }}</n-text>
            <n-flex :size="8">
              <NTag :bordered="false" size="small" :type="user.primary ? 'info' : 'default'">
                {{ user.primary ? '主用户' : '共享用户' }}
              </NTag>
              <NTag :bordered="false" :type="user.writable ? 'warning' : 'success'" size="small">
                {{ user.writable ? t('Writable') : t('ReadOnly') }}
              </NTag>
            </n-flex>
          </n-flex>
        </n-list-item>
      </n-list>

      <n-button
        type="primary"
        size="small"
        class="!w-full mt-1"
        :disabled="false"
        @click="openModal"
      >
        <n-text class="text-white text-sm">
          {{ t('CreateLink') }}
        </n-text>
      </n-button>
    </n-flex>

    <CreateShareCard :session="session" v-model:show="showModal" />
  </n-flex>
</template>
