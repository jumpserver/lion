<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
import { NTag, useMessage } from 'naive-ui';
import type { Composer } from 'vue-i18n';
import { removeShareUser } from '@/api';
import CardContainer from '@/components/CardContainer/index.vue';
import CreateLink from '@/components/SessionShare/widget/CreateLink.vue';

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
</script>

<template>
  <n-flex vertical align="center">
    <CardContainer>
      <template #custom-header>
        <n-text class="text-xs-plus"> {{ t('OnlineUser') }} </n-text>
        <NTag round :bordered="false" type="success" size="small" class="ml-2">
          {{ props.users?.length || 0 }} 人
        </NTag>
      </template>

      <n-flex v-if="props.users && props.users?.length > 0" class="w-full mb-4">
        <UserItem
          v-for="currentUser in props.users"
          :meta="currentUser"
          :key="currentUser.user_id"
          :username="currentUser.user"
          :primary="currentUser.primary"
          :writable="currentUser.writable"
          :user-id="currentUser.user_id"
          @remove-user="handleRemoveShareUser"
        />
      </n-flex>
    </CardContainer>
    <CardContainer :title="t('CreateLink')">
      <CreateLink :session="session" :disabled-create-link="props.disableCreate" />
    </CardContainer>
  </n-flex>
</template>
