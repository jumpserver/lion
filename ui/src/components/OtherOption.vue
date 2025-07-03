<script lang="ts" setup>
import { ref, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { NButton, NIcon } from 'naive-ui';
import { CirclePlus, CircleMinus } from 'lucide-vue-next';
const { t } = useI18n();

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
const options = ref<{ label: string; key: string }[]>([]);

onMounted(() => {
  // Add remote app specific keys if needed
  let keys = combinationKeys;
  if (props.isRemoteApp) {
    keys = remoteAppCombinationKeys;
  }
  options.value = keys.map((item) => ({
    label: item.name,
    key: item.keys.join('+'),
  }));
});

const emit = defineEmits(['combine-keys', 'update:autoFit', 'update:fitPercentage', 'updateScale']);

const handleSelect = (value: string) => {
  emit('combine-keys', value.split('+'));
};

const props = defineProps<{
  isRemoteApp: boolean;
  autoFit: boolean;
  fitPercentage: number;
}>();

const handleAutoFitUpdate = (value: boolean) => {
  emit('update:autoFit', value);
};

const percentage = ref<number>(props.fitPercentage);

onMounted(() => {
  percentage.value = props.fitPercentage;
});
const handleCircleClick = (value: number) => {
  const newPercentage = percentage.value + value;
  if (newPercentage < 10) {
    console.warn('Fit percentage cannot be less than 10%');
    return;
  }
  percentage.value = newPercentage;
  emit('update:autoFit', false);
  emit('updateScale', newPercentage); // Simulate Ctrl+Alt+Backspace
};

const handleCircleMinusClick = (e: any) => {
  handleCircleClick(-5);
};
const handleCirclePlusClick = (e: any) => {
  handleCircleClick(5);
};
</script>

<template>
  <n-card :title="t('Other')" class="w-full">
    <n-grid x-gap="12" :cols="2">
      <n-gi>
        <n-dropdown trigger="hover" :options="options" @select="handleSelect">
          <n-button>{{ t('Shortcuts') }}</n-button>
        </n-dropdown>
      </n-gi>
      <n-gi>
        <n-form-item :label="t('AutoFit')" label-placement="left">
          <n-switch :value="props.autoFit" @update:value="handleAutoFitUpdate" />
          <CircleMinus @click="handleCircleMinusClick" /><span class="text-xs"
            >{{ props.fitPercentage }}%</span
          >
          <CirclePlus @click="handleCirclePlusClick" />
        </n-form-item>
      </n-gi>
    </n-grid>
  </n-card>
</template>
