<script lang="ts" setup>
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { NSelect } from 'naive-ui';
import CardContainer from '@/components/CardContainer/index.vue';
const { t } = useI18n();

const props = defineProps<{
  keyboard?: string;
  opened: boolean;
}>();

const generalOptions = ref([
  { label: 'German (Qwertz)', value: 'de-de-qwertz' },
  { label: 'US English (Qwerty)', value: 'en-us-qwerty' },
  { label: 'Spanish (Qwerty)', value: 'es-es-qwerty' },
  { label: 'French (Azerty)', value: 'fr-fr-azerty' },
  { label: 'Italian (Qwerty)', value: 'it-it-qwerty' },
  { label: 'Dutch (QWERTY)', value: 'nl-nl-qwerty' },
  { label: 'Russian (QWERTY)', value: 'ru-ru-qwerty' },
]);

const emit = defineEmits(['update:keyboard', 'update:opened']);

const handleUpdateValue = (value: string) => {
  emit('update:keyboard', value);
};

const handleSwitchOpen = (value: boolean) => {
  emit('update:opened', value);
};
</script>

<template>
  <CardContainer :title="t('VirtualKeyboard')">
    <n-grid x-gap="12" :cols="4">
      <n-gi>
        <n-form-item :label="t('Enable')" label-placement="left">
          <n-switch :value="props.opened" @update:value="handleSwitchOpen" />
        </n-form-item>
      </n-gi>
      <n-gi :span="3">
        <n-form-item :label="t('KeyboardLayout')" label-placement="left">
          <n-select
            :value="props.keyboard"
            :options="generalOptions"
            @update:value="handleUpdateValue"
          />
        </n-form-item>
      </n-gi>
    </n-grid>
  </CardContainer>
</template>

<!-- <template>
  <n-card :title="t('VirtualKeyboard')" class="w-full">
    <n-grid x-gap="12" :cols="4">
      <n-gi>
        <n-form-item :label="t('Enable')" label-placement="left">
          <n-switch :value="props.opened" @update:value="handleSwitchOpen" />
        </n-form-item>
      </n-gi>
      <n-gi :span="3">
        <n-form-item :label="t('KeyboardLayout')" label-placement="left">
          <n-select
            :value="props.keyboard"
            :options="generalOptions"
            @update:value="handleUpdateValue"
          />
        </n-form-item>
      </n-gi>
    </n-grid>
  </n-card>
</template> -->
