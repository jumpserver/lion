<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
import { useDebounceFn } from '@vueuse/core';
import { Delete, Undo2 } from 'lucide-vue-next';
import { computed, h, reactive, ref, watch } from 'vue';
import { NTag, useDialogReactiveList, useMessage } from 'naive-ui';
import type { Composer } from 'vue-i18n';
import CreateShareCard from './CreateShare.vue';
import { removeShareUser } from '@/api';

export type TranslateFunction = Composer['t'];

const props = defineProps<{
  session: string;
  users?: Array<{
    user_id: string;
    user: string;
    primary: boolean;
    writable: boolean;
  }>;
  disableCreate?: boolean;
}>();

const { t } = useI18n();
const message = useMessage();

const showModal = ref<boolean>(false);

const handleRemoveShareUser = (user: {
  user_id: string;
  user: string;
  primary: boolean;
  writable: boolean;
}) => {
  console.log('Removing user:', user);
  removeShareUser(user)
    .then((res: any) => res.json())
    .then((response) => {
      if (response.message && !response.success) {
        message.error(response.message);
        return;
      }
    })
    .catch((error) => {
      console.error('Error removing share user:', error);
      message.error(t('ShareUserRemoveError'));
    });
};

const openModal = () => {
  showModal.value = true;
};
</script>

<template>
  <n-flex vertical align="center">
    <n-divider dashed title-placement="left" class="!mb-3 !mt-0">
      <n-text depth="2" class="text-sm opacity-70">
        {{ t('OnlineUser') }} {{ props.users?.length || 0 }}
      </n-text>
    </n-divider>

    <n-flex v-if="props.users?.length" class="w-full mb-4">
      <n-list class="w-full" bordered hoverable>
        <n-list-item v-for="user in props.users" :key="user.user_id">
          <template #suffix>
            <n-popconfirm v-if="!user.primary" @positive-click="handleRemoveShareUser(user)">
              <template #trigger>
                <Delete
                  :size="18"
                  class="cursor-pointer hover:text-red-500 transition-all duration-200"
                />
              </template>
              {{ t('RemoveShareUserConfirm') }}
            </n-popconfirm>
          </template>

          <n-flex vertical>
            <n-text>{{ user.user }}</n-text>
            <n-flex :size="8">
              <NTag :bordered="false" size="small" :type="user.primary ? 'info' : 'default'">
                {{ user.primary ? t('PrimaryUser') : t('ShareUser') }}
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
        :disabled="disableCreate"
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
