<script lang="ts" setup>
import { ref } from 'vue';
import { readClipboardText } from '@/utils/clipboard';
import { useDebounceFn } from '@vueuse/core';
import { NInput } from 'naive-ui';
import CardContainer from '@/components/CardContainer/index.vue';
const emit = defineEmits(['update:text']);
import { useMessage } from 'naive-ui';
import { useI18n } from 'vue-i18n';
const { t } = useI18n();
// 内部输入值
const inputValue = ref<string>('');
const isLoading = ref<boolean>(false);
const message = useMessage();
const props = defineProps<{
  remoteText?: string;
  disabled?: boolean;
}>();

const showRemoteText = ref<boolean>(false);

// 手动读取剪贴板内容
const loadClipboardText = async () => {
  try {
    isLoading.value = true;
    const text = await readClipboardText();
    inputValue.value = text;
    await handleInput(text);
  } catch (error) {
    console.log('Failed to read clipboard text:', error);
    // 可以添加用户友好的错误提示
  } finally {
    isLoading.value = false;
  }
};

// 处理输入事件
const handleInput = useDebounceFn((value: string) => {
  emit('update:text', value);
}, 300);

// 处理焦点事件，尝试自动读取剪贴板
const handleFocus = async () => {
  // 只有在输入框为空时才自动读取
  if (!inputValue.value.trim()) {
    try {
      await loadClipboardText();
    } catch (error) {
      // 静默处理错误，不影响用户体验
      console.debug('Auto-read clipboard failed, user can click button to read manually');
    }
  }
};

const noSideSpace = (value: string) => {
  return !value.startsWith(' ') && !value.endsWith(' ') && !value.startsWith('\n');
};

const size = {
  minRows: 4,
  maxRows: 6,
};

const maxlength = 1024 * 4;
</script>

<template>
  <CardContainer :title="t('Clipboard')">
    <n-form-item :label="t('ShowRemoteClip')" label-placement="left">
      <n-switch v-model:value="showRemoteText" :disabled="props.disabled" />
    </n-form-item>
    <n-input
      v-model:value="inputValue"
      @input="handleInput"
      @focus="handleFocus"
      type="textarea"
      :allow-input="noSideSpace"
      :autosize="size"
      :maxlength="maxlength"
      show-count
      clearable
      :placeholder="t('AutoPasteOnClick')"
      :disabled="props.disabled"
    >
    </n-input>
    <n-form-item v-if="showRemoteText">
      <n-input
        :value="props.remoteText"
        type="textarea"
        :autosize="size"
        readonly
        placeholder=""
        show-count
        :disabled="props.disabled"
      />
    </n-form-item>
  </CardContainer>
</template>

<!-- <template>
  <n-card class="w-full" :title="t('Clipboard')">
    <n-input
      v-model:value="inputValue"
      @input="handleInput"
      @focus="handleFocus"
      type="textarea"
      :allow-input="noSideSpace"
      :autosize="size"
      :maxlength="maxlength"
      show-count
      clearable
      :placeholder="t('AutoPasteOnClick')"
      :disabled="props.disabled"
    >
    </n-input>
  </n-card> -->
<!-- <n-space vertical> -->

<!-- <n-space> -->
<!-- <n-button
        @click="loadClipboardText"
        type="primary"
        size="small"
      >
       从剪贴板粘贴
      </n-button> -->
<!-- <n-button
        @click="loadRemoteClipboardText"
        type="primary"
        size="small"
        :disabled="props.disabled"
      >
        显示远程同步的剪贴板信息</n-button
      > -->
<!-- </n-space> -->
<!-- <n-input
      v-if="showRemoteText"
      :value="props.remoteText"
      type="textarea"
      :autosize="size"
      readonly
      placeholder="远程同步的剪贴板内容"
      :disabled="props.disabled"
    /> -->
<!-- </n-space> -->
<!-- </template> -->
