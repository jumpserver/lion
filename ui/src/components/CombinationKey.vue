<script lang="ts" setup>
import { ref, onMounted, computed, handleError } from 'vue';
import { useI18n } from 'vue-i18n';
import { NButton, NIcon } from 'naive-ui';
import { CirclePlus, ChevronLeft, ChevronDown, ChevronsDown } from 'lucide-vue-next';
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
const keyboardList = ref<{ label: string; icon?: any; click: () => void }[]>([]);

onMounted(() => {
  const keysList: any[] = [];
  const keys = props.isRemoteApp ? remoteAppCombinationKeys : combinationKeys;
  keys.forEach((item) => {
    keysList.push({
      label: item.name,
      icon: CirclePlus,
      click: () => {
        // Handle key combination click
        emit('combine-keys', item.keys);
      },
    });
  });
  keyboardList.value = keysList;
});

const collapsed = ref('');
const handleExpandedNames = (names: string[]) => {
  collapsed.value = names.join(', ');
};

const isCollapsed = computed(() => {
  return collapsed.value.length > 0;
});

const handleCollapsed = (e: any) => {
  collapsed.value = '1';
};
</script>

<template>
  <div>
    <n-divider title-placement="left" dashed class="!mb-3 !mt-0">
      <n-text depth="2" class="text-sm opacity-70"> {{ t('AvailableShortcutKey') }} </n-text>
    </n-divider>

    <n-collapse
      :expanded-names="collapsed"
      @update:expanded-names="handleExpandedNames"
      :accordion="true"
    >
      <n-collapse-item name="1" trigger-areas="extra">
        <template #arrow>
          <n-icon />
        </template>
        <template #header-extra>
          <ChevronLeft v-if="!isCollapsed" />
          <ChevronDown v-else />
        </template>
        <div>
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
                    <!-- <component v-if="item.icon" :is="item.icon" :size="20" class="text-white/90 flex-shrink-0" /> -->
                    <n-text class="text-sm text-white/90">
                      {{ item.label }}
                    </n-text>
                  </n-flex>
                </template>
              </n-card>
            </n-gi>
          </n-grid>
        </div>
      </n-collapse-item>
    </n-collapse>
    <div v-if="!isCollapsed" class="flex justify-center">
      <ChevronsDown @click="handleCollapsed" />
    </div>
  </div>
</template>
