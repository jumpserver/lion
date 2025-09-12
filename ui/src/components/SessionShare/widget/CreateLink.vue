<script setup lang="ts">
import type { SelectRenderTag } from 'naive-ui';

import { useI18n } from 'vue-i18n';
import { ArrowLeft, Copy, Link } from 'lucide-vue-next';
import { useMessage, NTag } from 'naive-ui';
import { computed, h, reactive, ref, watch } from 'vue';
import type { Composer } from 'vue-i18n';
import { useDebounceFn } from '@vueuse/core';
export type TranslateFunction = Composer['t'];

import { useColor } from '@/hooks/useColor';
import { createShareURL } from '@/api';
import { writeToClipboard } from '@/utils/clipboard.ts';
import { BASE_URL } from '@/utils/config.ts';
const props = defineProps<{
  session: string;
  disabledCreateLink: boolean;
}>();
const getMinuteLabel = (item: number, t: TranslateFunction): string => {
  let minuteLabel = t('Minute');
  if (item > 1) {
    minuteLabel = t('Minutes');
  }
  return `${item} ${minuteLabel}`;
};

export interface ShareUserOptions {
  id: string;

  name: string;

  username: string;
}

export interface UserInfo {
  id: string;
  name: string;
  username: string;
}

interface ExpiredOption {
  label: string;
  value: number;
  checked: boolean;
}

interface ActionPermOption {
  label: string;
  value: string;
  checked: boolean;
}

const { t } = useI18n();
const { lighten } = useColor();
const message = useMessage();
const shareInfo = ref({
  shareCode: '',
  sessionId: props.session,
  shareId: '',
  shareURL: '',
});
const userOptions = ref<UserInfo[]>([]);
const currentQuery = ref<string>('');
const currentPage = ref<number>(1);
const hasMore = ref<boolean>(true);

const searchUsers = useDebounceFn(async (value: string, isLoadMore: boolean = false) => {
  if (value === '' && !isLoadMore) {
    searchLoading.value = false;
    return;
  }

  // 如果是新搜索，重置状态
  if (!isLoadMore || value !== currentQuery.value) {
    currentQuery.value = value;
    currentPage.value = 1;
    userOptions.value = [];
    hasMore.value = true;
  }

  searchLoading.value = true;

  try {
    // 修改API调用以支持分页参数
    const params = new URLSearchParams({
      action: 'suggestion',
      search: currentQuery.value,
      page: currentPage.value.toString(),
      limit: '20', // 每页加载20条数据
    });

    const response = await fetch(`${BASE_URL}/api/v1/users/users/?${params}`).then((res: any) =>
      res.json(),
    );

    // 假设分页响应格式为：{ results: [...], count: number, next: string|null }
    const newUsers = response.results || response; // 兼容非分页格式

    if (isLoadMore && currentPage.value > 1) {
      // 加载更多时追加数据
      const filteredUsers = newUsers.filter((user: UserInfo) => {
        const query = currentQuery.value.toLowerCase();
        const caseInsensitiveMatch = (name: any, query: string) =>
          name.toLowerCase().includes(query);
        return caseInsensitiveMatch(user.name, query) || caseInsensitiveMatch(user.username, query);
      });
      userOptions.value = [...userOptions.value, ...filteredUsers];
    } else {
      // 新搜索时替换数据
      userOptions.value = newUsers.filter((user: UserInfo) => {
        const query = currentQuery.value.toLowerCase();
        const caseInsensitiveMatch = (name: any, query: string) =>
          name.toLowerCase().includes(query);
        return caseInsensitiveMatch(user.name, query) || caseInsensitiveMatch(user.username, query);
      });
    }

    // 检查是否还有更多数据
    hasMore.value = response.next !== null && response.next !== undefined;
  } catch (error) {
    console.error('Search users error:', error);
    message.error(t('NoUserFound'));
  } finally {
    searchLoading.value = false;
  }
}, 300);

const searchLoading = ref<boolean>(false);
const showLinkResult = ref<boolean>(false);
watch(
  () => shareInfo.value.shareCode,
  (nv) => {
    if (nv) {
      showLinkResult.value = true;
    } else {
      showLinkResult.value = false;
    }
  },
);

const mappedUserOptions = computed(() => {
  if (userOptions.value && userOptions.value.length > 0) {
    return userOptions.value.map((item: ShareUserOptions) => ({
      label: item.username,
      value: item.id,
    }));
  } else {
    return [];
  }
});

// 装饰器模式：创建单选处理器
const createSingleSelectHandler = <T, K extends keyof T>(
  options: T[],
  valueKey: K,
  checkedKey: keyof T,
  onSelect?: (value: T[K]) => void,
) => {
  return (selectedValue: T[K]) => {
    options.forEach((item) => {
      (item as any)[checkedKey] = item[valueKey] === selectedValue;
    });

    // 执行回调函数，更新 shareLinkRequest
    if (onSelect) {
      onSelect(selectedValue);
    }
  };
};

const shareLinkRequest = reactive({
  expiredTime: 10,
  actionPerm: 'writable',
  users: [] as ShareUserOptions[],
});

const expiredOptions = reactive<ExpiredOption[]>([
  { label: getMinuteLabel(1, t), value: 1, checked: false },
  { label: getMinuteLabel(5, t), value: 5, checked: false },
  { label: getMinuteLabel(10, t), value: 10, checked: true },
  { label: getMinuteLabel(20, t), value: 20, checked: false },
  { label: getMinuteLabel(60, t), value: 60, checked: false },
]);

const actionsPermOptions = reactive<ActionPermOption[]>([
  { label: t('Writable'), value: 'writable', checked: true },
  { label: t('ReadOnly'), value: 'readonly', checked: false },
]);

const renderTag: SelectRenderTag = ({ option, handleClose }) => {
  return h(
    NTag,
    {
      closable: true,
      size: 'small',
      type: 'info',
      bordered: false,
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

const scrollSearch = (e: Event) => {
  const currentTarget = e.currentTarget as HTMLElement;
  if (currentTarget.scrollTop + currentTarget.clientHeight >= currentTarget.scrollHeight - 10) {
    // 到达底部，加载更多数据
    if (!searchLoading.value && hasMore.value && currentQuery.value) {
      currentPage.value += 1;
      searchUsers(currentQuery.value, true);
    }
  }
};

const debounceSearch = useDebounceFn((query: string) => searchUsers(query, false), 300);
const handleChangeExpired = createSingleSelectHandler(
  expiredOptions,
  'value',
  'checked',
  (value) => {
    shareLinkRequest.expiredTime = value;
  },
);
const handleChangeActionPerm = createSingleSelectHandler(
  actionsPermOptions,
  'value',
  'checked',
  (value) => {
    shareLinkRequest.actionPerm = value;
  },
);

/**
 * @description 创建会话分享链接
 */
const handleCreateLink = () => {
  if (!shareInfo.value.sessionId) {
    return message.error(t('FailedCreateConnection'));
  }
  console.log(shareInfo.value.sessionId);
  // createShareLink(shareLinkRequest);
  console.log('shareLinkRequest:', shareLinkRequest);

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
      shareInfo.value.shareURL = generateShareURL(res.id, res.verify_code);
    })
    .catch((error: any) => {
      console.error('Create share URL error:', error);
      message.error(t('CreateLinkFailed'));
    })
    .finally(() => {});
};

const generateShareURL = (shareId: string, shareCode: string) => {
  const encodedShareCode = encodeURIComponent(shareCode);
  return `${BASE_URL}/luna/share/${shareId}?type=lion&code=${encodedShareCode}`;
};

/**
 * @description 复制会话分享链接
 */
const handleCopyShareURL = () => {
  const url = shareInfo.value.shareURL;
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

/**
 * @description 返回到上一层
 */
const handleBack = () => {
  resetShareState();
};

const resetShareState = () => {
  showLinkResult.value = false;
  shareLinkRequest.expiredTime = 10;
  shareLinkRequest.actionPerm = 'writable';
  shareLinkRequest.users = [];
};
</script>

<template>
  <n-descriptions v-if="!showLinkResult" label-placement="top" :column="1">
    <n-descriptions-item>
      <template #label>
        <n-text class="text-xs-plus" depth="1">
          {{ t('ExpiredTime') }}
        </n-text>
      </template>

      <n-flex align="center" class="mt-2 cursor-pointer">
        <n-card
          v-for="item in expiredOptions"
          :key="item.value"
          bordered
          hoverable
          size="small"
          :content-style="{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
          }"
          :style="{
            border: item.checked ? `1px solid ${lighten(20)}` : '',
          }"
          style="width: 110px; height: 45px"
          @click="handleChangeExpired(item.value)"
        >
          <n-text depth="2" class="text-xs-plus">
            {{ item.label }}
          </n-text>
        </n-card>
      </n-flex>
    </n-descriptions-item>

    <n-descriptions-item>
      <n-divider dashed class="!my-1" />
    </n-descriptions-item>

    <n-descriptions-item>
      <template #label>
        <n-text class="text-xs-plus" depth="1">
          {{ t('ActionPerm') }}
        </n-text>
      </template>

      <n-flex align="center" :wrap="false" class="mt-2 cursor-pointer">
        <n-card
          v-for="item in actionsPermOptions"
          :key="item.value"
          bordered
          hoverable
          size="small"
          :content-style="{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
          }"
          :style="{
            border: item.checked ? `1px solid ${lighten(20)}` : '',
          }"
          style="width: 50%"
          @click="handleChangeActionPerm(item.value)"
        >
          <n-text depth="1" class="text-xs-plus">
            {{ item.label }}
          </n-text>
        </n-card>
      </n-flex>
    </n-descriptions-item>

    <n-descriptions-item>
      <n-divider dashed class="!my-1" />
    </n-descriptions-item>

    <n-descriptions-item>
      <template #label>
        <n-text class="text-xs-plus">
          {{ t('ShareUser') }}
        </n-text>
      </template>

      <n-flex vertical class="mt-2">
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
          @scroll="scrollSearch"
          @focus="() => debounceSearch('')"
        />
      </n-flex>
    </n-descriptions-item>

    <n-descriptions-item>
      <n-divider class="!my-1" />
    </n-descriptions-item>

    <n-descriptions-item>
      <n-button
        block
        secondary
        type="primary"
        class="mt-2 !text-xs-plus"
        :disabled="disabledCreateLink"
        @click="handleCreateLink"
      >
        {{ t('CreateLink') }}
      </n-button>
    </n-descriptions-item>
  </n-descriptions>

  <n-descriptions v-else label-placement="top" :column="1">
    <n-descriptions-item>
      <n-input placeholder="Link" round size="small" readonly :value="shareInfo.shareURL">
        <template #prefix>
          <Link :size="14" />
        </template>
      </n-input>

      <n-card
        size="small"
        :content-style="{
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          justifyContent: 'center',
        }"
        class="mt-4"
      >
        <n-text>
          {{ t('VerifyCode') }}
        </n-text>

        <n-text depth="2" class="text-2xl tracking-widest">
          {{ shareInfo.shareCode }}
        </n-text>
      </n-card>

      <n-flex align="center" :wrap="false" class="w-full mt-4">
        <n-button secondary type="success" class="!w-1/2" @click="handleCopyShareURL">
          <template #icon>
            <Copy />
          </template>

          {{ t('CopyLink') }}
        </n-button>
        <n-button secondary class="!w-1/2" @click="handleBack">
          <template #icon>
            <ArrowLeft />
          </template>

          {{ t('Back') }}
        </n-button>
      </n-flex>
    </n-descriptions-item>
  </n-descriptions>
</template>
