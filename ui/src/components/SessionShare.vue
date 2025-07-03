<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
import { useDebounceFn } from '@vueuse/core';
import { Delete, Undo2 } from 'lucide-vue-next';
import { computed, h, reactive, ref, watch } from 'vue';
import { NTag, useDialogReactiveList, useMessage } from 'naive-ui';
import type { Composer } from 'vue-i18n';

export type TranslateFunction = Composer['t'];


const { t } = useI18n();

const loading = ref<boolean>(false);
const searchLoading = ref<boolean>(false);
const showModal = ref<boolean>(false);
const showCreateForm = ref<boolean>(true);
const getMinuteLabel = (item: number, t: TranslateFunction): string => {
  let minuteLabel = t('Minute');
  if (item > 1) {
    minuteLabel = t('Minutes');
  }
  return `${item} ${minuteLabel}`;
}


export interface ShareUserOptions {
  id: string;

  name: string;

  username: string;
}


const expiredOptions = reactive([
  { label: getMinuteLabel(1, t), value: 1 },
  { label: getMinuteLabel(5, t), value: 5 },
  { label: getMinuteLabel(10, t), value: 10 },
  { label: getMinuteLabel(20, t), value: 20 },
  { label: getMinuteLabel(60, t), value: 60 },
]);

const shareLinkRequest = reactive({
  expiredTime: 10,
  actionPerm: 'writable',
  users: [] as ShareUserOptions[],
});

const actionsPermOptions = reactive([
  { label: t('Writable'), value: 'writable' },
  { label: t('ReadOnly'), value: 'readonly' },
]);

const onlineUsers = ref<Array<{ user_id: string; user: string; primary: boolean; writable: boolean }>>([
    { user_id: '1', user: 'User1', primary: true, writable: true },
    { user_id: '2', user: 'User2', primary: false, writable: false },
]);


const handleRemoveShareUser = (user: { user_id: string; user: string; primary: boolean; writable: boolean }) => {
  onlineUsers.value = onlineUsers.value.filter(u => u.user_id !== user.user_id);
};

const shareInfo = reactive({
  enableShare: true,
  shareCode: '',
});
const shareURL = computed(() => {
  return `https://example.com/share/${shareInfo.shareCode}`;
});
const handleShareURlCreated = () => {
  loading.value = true;
  setTimeout(() => {
    shareInfo.shareCode = '123456'; 
    showCreateForm.value = false;
    loading.value = false;
    }, 1000);
};
const handleBack = () => {
  showCreateForm.value = true;
};
const handleModalClose = () => {
  showModal.value = false;
  showCreateForm.value = true;
  shareInfo.shareCode = '';
};
const openModal = () => {
  showModal.value = true;
  shareLinkRequest.expiredTime = 10;
  shareLinkRequest.actionPerm = 'writable';
    shareLinkRequest.users = [];
};

const copyShareURLHandler = () => {
  navigator.clipboard.writeText(shareURL.value).then(() => {
    useMessage().success(t('CopySuccess'));
  }).catch(() => {
    useMessage().error(t('CopyFailed'));
  });
};
</script>

<template>

 <n-flex vertical align="center">
    <n-divider dashed title-placement="left" class="!mb-3 !mt-0">
      <n-text depth="2" class="text-sm opacity-70"> {{ t('OnLineUser') }}{{ onlineUsers?.length || 0 }}） </n-text>
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

      <n-button type="primary" size="small" class="!w-full mt-1" :disabled="!shareInfo.enableShare" @click="openModal">
        <n-text class="text-white text-sm">
          {{ t('CreateLink') }}
        </n-text>
      </n-button>
    </n-flex>

    <n-modal v-model:show="showModal" :auto-focus="false" @update:show="handleModalClose">
      <n-card style="width: 600px" bordered :title="cardTitle" role="dialog" size="large">
        <Transition name="fade" mode="out-in">
          <div v-if="showCreateForm" key="create-form" class="min-h-[305px] w-full">
            <n-form label-placement="top" :model="shareLinkRequest">
              <n-grid :cols="24">
                <n-form-item-gi :span="24" :label="t('ExpiredTime')">
                  <n-select
                    v-model:value="shareLinkRequest.expiredTime"
                    size="small"
                    :placeholder="t('SelectAction')"
                    :options="expiredOptions"
                  />
                </n-form-item-gi>
                <n-form-item-gi :span="24" :label="t('ActionPerm')">
                  <n-select
                    v-model:value="shareLinkRequest.actionPerm"
                    size="small"
                    :placeholder="t('ActionPerm')"
                    :options="actionsPermOptions"
                  />
                </n-form-item-gi>
                <n-form-item-gi :span="24" :label="t('ShareUser')">
                  <n-select
                    v-model:value="shareLinkRequest.users"
                    multiple
                    filterable
                    clearable
                    remote
                    size="small"
                    :loading="searchLoading"
                    :render-tag="renderTag"
                    :options="mappedUserOptions"
                    :clear-filter-after-select="false"
                    :placeholder="t('GetShareUser')"
                    @search="debounceSearch"
                  />
                </n-form-item-gi>
              </n-grid>

              <n-button
                type="primary"
                size="small"
                class="!w-full mt-1"
                :loading="loading"
                :disabled="!shareInfo.enableShare"
                @click="handleShareURlCreated"
              >
                <n-text class="text-white text-sm">
                  {{ t('CreateLink') }}
                </n-text>
              </n-button>
            </n-form>
          </div>

          <div v-else key="share-result" class="relative min-h-[305px] w-full">
            <n-result status="success" :description="t('CreateSuccess')" class="relative" />

            <n-tooltip size="small">
              <template #trigger>
                <Undo2
                  :size="16"
                  class="absolute top-0 right-0 focus:outline-none cursor-pointer"
                  @click="handleBack"
                />
              </template>
              <span>{{ t('Back') }}</span>
            </n-tooltip>

            <n-form label-placement="top">
              <n-grid :cols="24">
                <n-form-item-gi :label="t('LinkAddr')" :span="24">
                  <n-input readonly :value="shareURL" />
                </n-form-item-gi>
                <n-form-item-gi :label="t('VerifyCode')" :span="24">
                  <n-input readonly :loading="!shareInfo.shareCode" :value="shareInfo.shareCode" />
                </n-form-item-gi>
              </n-grid>

              <n-button type="primary" size="small" class="!w-full" @click="copyShareURLHandler">
                <n-text class="text-white text-sm">
                  {{ t('CopyLink') }}
                </n-text>
              </n-button>
            </n-form>
          </div>
        </Transition>
      </n-card>
    </n-modal>
  </n-flex>

</template>