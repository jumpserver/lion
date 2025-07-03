<script lang="ts" setup>
import { defineComponent, ref, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { NButton, NIcon, NSelect } from 'naive-ui';

const { t } = useI18n();

const props = defineProps<{
  keyboard?: string;
  opened: boolean;
}>();

const generalOptions = ref([
  { label: t('German (Qwertz)'), value: 'de-de-qwertz' },
  { label: t('US English (Qwerty)'), value: 'en-us-qwerty' },
  { label: t('Spanish (Qwerty)'), value: 'es-es-qwerty' },
  { label: t('French (Azerty)'), value: 'fr-fr-azerty' },
  { label: t('Italian (Qwerty)'), value: 'it-it-qwerty' },
  { label: t('Dutch (QWERTY)'), value: 'nl-nl-qwerty' },
  { label: t('Russian (QWERTY)'), value: 'ru-ru-qwerty' },
]);

const emit = defineEmits(['update:keyboard', 'update:opened']);

const handleUpdateValue = (value: string) => {
  emit('update:keyboard', value);
};

const handleSwitchOpen = (value: boolean) => {
  emit('update:opened', value);
  console.log('handleSwitchOpen', value);
};
</script>

<template>
  <n-card title="虚拟键盘" class="w-full">
    <n-grid x-gap="12" :cols="4">
      <n-gi>
        <n-form-item label="开启" label-placement="left">
          <n-switch :value="props.opened" @update:value="handleSwitchOpen" />
        </n-form-item>
      </n-gi>
      <n-gi :span="3">
        <n-form-item label="键盘布局" label-placement="left">
          <n-select
            :value="props.keyboard"
            :options="generalOptions"
            @update:value="handleUpdateValue"
          />
        </n-form-item>
      </n-gi>
    </n-grid>
  </n-card>
</template>
