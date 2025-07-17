<script lang="ts" setup>
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
const { t } = useI18n();

const props = defineProps<{
  isRemoteApp: boolean;
}>();

const emit = defineEmits(['combine-keys']);

const combinationKeys = [
  {
    keys: ['65307'],
    name: 'Esc',
  },
  {
    keys: ['65480'],
    name: 'F11',
  },
  {
    keys: ['65507', '65513', '65535'],
    name: 'Ctrl+Alt+Delete',
  },
  {
    keys: ['65507', '65513', '65288'],
    name: 'Ctrl+Alt+Backspace',
  },
  {
    keys: ['65515', '100'],
    name: 'Windows+D',
  },
  {
    keys: ['65515', '101'],
    name: 'Windows+E',
  },
  {
    keys: ['65515', '114'],
    name: 'Windows+R',
  },
  {
    keys: ['65515', '120'],
    name: 'Windows+X',
  },
  {
    keys: ['65515'],
    name: 'Windows',
  },
  {
    keys: ['65513', '65289'],
    name: 'Alt+Tab',
  },
];

const remoteAppCombinationKeys = [
  {
    keys: ['65513', '65289'],
    name: 'Alt+Tab',
  },
];

const keyboardList = computed(() => {
  const result: any[] = [];
  const keys = props.isRemoteApp ? remoteAppCombinationKeys : combinationKeys;
  keys.forEach((item) => {
    result.push({
      label: item.name,
      click: () => {
        // Handle key combination click
        emit('combine-keys', item.keys);
      },
    });
  });
  return result;
});
</script>

<template>
  <div>
    <n-divider title-placement="left" dashed class="!mb-3 !mt-0">
      <n-text depth="2" class="text-sm opacity-70">{{ t('AvailableShortcutKey') }} </n-text>
    </n-divider>
    <n-grid x-gap="8" y-gap="8" :cols="2">
      <n-gi v-for="item in keyboardList" :key="item.label">
        <n-card
          hoverable
          class="cursor-pointer transition-all duration-200 border-transparent hover:border-white/20"
          :content-style="{ padding: '12px' }"
          @click="item.click"
        >
          <template #default>
            <n-flex align="center" :size="12" class="px-2 py-1">
              <n-text class="text-sm text-white/90">
                {{ item.label }}
              </n-text>
            </n-flex>
          </template>
        </n-card>
      </n-gi>
    </n-grid>
  </div>
</template>
