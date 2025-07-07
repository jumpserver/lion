<script lang="ts" setup>
import { c, type UploadFileInfo } from 'naive-ui';
import { useMessage, NTag } from 'naive-ui';
import { computed, h, reactive, ref, watch } from 'vue';
import type { Composer } from 'vue-i18n';
import { useDebounceFn } from '@vueuse/core';
import type { SelectRenderTag } from 'naive-ui';
import { Crown, Delete, Lock, PenLine, Undo2, UserRound } from 'lucide-vue-next';
export type TranslateFunction = Composer['t'];

import { GetUsersInfo, createShareURL } from '@/api';
import { BASE_URL } from '@/utils/config';
import { writeToClipboard } from '@/utils/clipboard';
const message = useMessage();
const props = defineProps<{
  session: string;
  show: boolean;
}>();

const emit = defineEmits(['update:show']);
const getMinuteLabel = (item: number, t: TranslateFunction): string => {
  let minuteLabel = t('Minute');
  if (item > 1) {
    minuteLabel = t('Minutes');
  }
  return `${item} ${minuteLabel}`;
};
import { useI18n } from 'vue-i18n';
const { t } = useI18n();

export interface UserInfo {
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
  users: [] as string[],
});

const actionsPermOptions = reactive([
  { label: t('Writable'), value: 'writable' },
  { label: t('ReadOnly'), value: 'readonly' },
]);

const createTitle = t('CreateLink');
const shareLinkTitle = t('LinkAddr');

const cardTitle = ref<string>(createTitle);
const loading = ref<boolean>(false);
const showCreateForm = ref<boolean>(true);
const searchLoading = ref<boolean>(false);

const renderTag: SelectRenderTag = ({ option, handleClose }) => {
  return h(
    NTag,
    {
      closable: true,
      size: 'small',
      type: 'primary',
      onMousedown: (e: FocusEvent) => {
        e.preventDefault();
      },
      onClose: (e: MouseEvent) => {
        e.stopPropagation();
        handleClose();
      },
    },
    {
      default: () => option.label,
    },
  );
};

const userOptions = ref<UserInfo[]>([]);
const shareInfo = ref({
  shareCode: '',
  sessionId: props.session,
  shareId: '',
});

const mappedUserOptions = computed(() => {
  if (userOptions.value && userOptions.value.length > 0) {
    return userOptions.value.map((item: UserInfo) => ({
      label: `${item.name}(${item.username})`,
      value: item.id,
    }));
  } else {
    return [];
  }
});

const searchUsers = useDebounceFn(async (value: string) => {
  if (value === '') {
    return;
  }
  userOptions.value = [];
  searchLoading.value = true;
  try {
    const response = await GetUsersInfo(value).then((res: any) => res.json());
    userOptions.value = response.filter((user: UserInfo) => {
      const query = value.toLowerCase();
      const caseInsensitiveMatch = (name: any, query: string) => name.toLowerCase().includes(query);
      return caseInsensitiveMatch(user.name, query) || caseInsensitiveMatch(user.username, query);
    });
  } catch (error) {
    message.error(t('NoUserFound'));
  } finally {
    searchLoading.value = false;
  }
}, 300);

const handleShareURlCreated = () => {
  if (!props.session) {
    return message.error(t('SessionNotFound'));
  }
  loading.value = true;
  const data = {
    session_id: props.session,
    expired_time: shareLinkRequest.expiredTime,
    users: shareLinkRequest.users,
    action_perm: shareLinkRequest.actionPerm,
  };
  createShareURL(data)
    .then((response: any) => response.json())
    .then((res: any) => {
      if (res.success && !res.success) {
        console.log('Error creating share URL:', res);
        return message.error(t('CreateLinkFailed') + `: ${res?.message || ''}`);
      }
      shareInfo.value.shareId = res.id;
      shareInfo.value.shareCode = res.verify_code;
      showCreateForm.value = false;
      cardTitle.value = shareLinkTitle;
    })
    .catch((error: any) => {
      message.error(t('CreateLinkFailed'));
    })
    .finally(() => {
      loading.value = false;
    });
};

const shareURL = computed(() => {
  if (!shareInfo.value.shareId || !shareInfo.value.shareCode) {
    return t('NoLink');
  }
  const shareCode = shareInfo.value.shareCode;
  // url encode the share code to ensure it is safe for URLs
  const encodedShareCode = encodeURIComponent(shareCode);
  return `${BASE_URL}/luna/share/${shareInfo.value.shareId}?type=lion&code=${encodedShareCode}`;
});

const handleBack = () => {
  loading.value = false;
  showCreateForm.value = true;
  shareLinkRequest.expiredTime = 10;
  shareLinkRequest.actionPerm = 'writable';
  shareLinkRequest.users = [];
};

const copyShareURLHandler = () => {
  console.log('shareURL', shareURL.value);
  const url = shareURL.value;
  const shareCode = shareInfo.value.shareCode;
  if (!url || !shareCode) {
    return message.error(t('NoLink'));
  }
  const linkTitle = t('LinkAddr');
  const codeTitle = t('VerifyCode');
  const text = `${linkTitle}: ${url}\n${codeTitle}: ${shareCode}`;
  writeToClipboard(text);
  message.info(t('CopyShareURLSuccess'));
};

const handleModalClose = () => {
  emit('update:show', false);
};
</script>

<template>
  <n-modal :show="props.show" :auto-focus="false" draggable transform-origin="center">
    <n-card
      style="max-width: 600px"
      bordered
      :title="cardTitle"
      size="medium"
      closable="true"
      @close="handleModalClose"
    >
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
                  @search="searchUsers"
                />
              </n-form-item-gi>
            </n-grid>

            <n-button
              type="primary"
              size="small"
              class="!w-full mt-1"
              :loading="loading"
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
          <!-- 
          <n-tooltip size="small">
            <template #trigger>
              <Undo2
                :size="16"
                class="absolute top-0 right-0 focus:outline-none cursor-pointer"
                @click="handleBack"
              />
            </template>
            <span>{{ t('Back') }}</span>
          </n-tooltip> -->

          <n-form label-placement="top">
            <n-grid :cols="24">
              <n-form-item-gi :label="t('LinkAddr')" :span="24">
                <n-input readonly :value="shareURL" />
              </n-form-item-gi>
              <n-form-item-gi :label="t('VerifyCode')" :span="24">
                <n-input readonly :loading="!shareInfo.shareCode" :value="shareInfo.shareCode" />
              </n-form-item-gi>
            </n-grid>
            <n-flex justify="space-between" class="mt-2">
              <n-button type="primary" size="small" @click="copyShareURLHandler">
                <n-text class="text-white text-sm">
                  {{ t('CopyLink') }}
                </n-text>
              </n-button>
              <n-button type="primary" size="small" @click="handleBack">
                <n-text class="text-white text-sm">
                  {{ t('CreateLink') }}
                </n-text>
              </n-button>
            </n-flex>
          </n-form>
        </div>
      </Transition>
    </n-card>
  </n-modal>
</template>
