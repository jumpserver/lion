<script lang="ts" setup>
    import {nextTick, ref, watch} from 'vue';
    import {readClipboardText} from '@/utils/clipboard';
    import {useDebounceFn} from '@vueuse/core';
    import {useMessage} from 'naive-ui';
    import {useI18n} from 'vue-i18n';

    const emit = defineEmits(['update:text']);

    const {t} = useI18n();
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
            handleInput(text);
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

    const isFocusHandling = ref(false);

    const handleFocus = () => {
        if (isFocusHandling.value) return;

        isFocusHandling.value = true;
        nextTick(() => {
            if (!inputValue.value.trim()) {
                loadClipboardText().catch(error => {
                    console.debug('Auto-read clipboard failed');
                }).finally(() => {
                    isFocusHandling.value = false;
                });
            } else {
                isFocusHandling.value = false;
            }
        });
    };

    const noSideSpace = (value: string) => {
        return !value.startsWith(' ') && !value.endsWith(' ') && !value.startsWith('\n');
    };

    const debouncedHiden = useDebounceFn(() => {
        showRemoteText.value = false;
    }, 1000 * 5);

    const loadRemoteClipboardText = async () => {
        if (!props.remoteText) {
            message.warning('远程剪贴板未返回内容');
            return;
        }
        showRemoteText.value = true;
        debouncedHiden();
    };

    const size = {
        minRows: 4,
        maxRows: 6,
    };

    const maxlength = 1024 * 4;

    // 监听远程剪贴板变化，自动更新textarea内容
    watch(
        () => props.remoteText,
        (newRemoteText) => {
            if (newRemoteText && newRemoteText !== inputValue.value) {
                inputValue.value = newRemoteText;
                handleInput(newRemoteText);
            }
        },
        {immediate: true}
    );
</script>

<template>
  <div>
    <n-divider title-placement="left" dashed class="!mb-3 !mt-0">
      <n-text depth="2" class="text-sm opacity-70"> {{ t('Clipboard') }}</n-text>
    </n-divider>
    <n-input
      v-model:value="inputValue"
      @input="handleInput"
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
  </div>
</template>

<!--<template>-->
<!--  <n-card class="w-full" :title="t('Clipboard')">-->
<!--    <n-input-->
<!--      v-model:value="inputValue"-->
<!--      @input="handleInput"-->
<!--      @focus="handleFocus"-->
<!--      type="textarea"-->
<!--      :allow-input="noSideSpace"-->
<!--      :autosize="size"-->
<!--      :maxlength="maxlength"-->
<!--      show-count-->
<!--      clearable-->
<!--      :placeholder="t('AutoPasteOnClick')"-->
<!--      :disabled="props.disabled"-->
<!--    >-->
<!--    </n-input>-->
<!--  </n-card>-->
<!-- <n-space vertical>-->

<!-- <n-space>-->
<!--<n-button-->
<!--        @click="loadClipboardText"-->
<!--        type="primary"-->
<!--        size="small"-->
<!--      >-->
<!--       从剪贴板粘贴-->
<!--      </n-button>-->
<!--<n-button-->
<!--        @click="loadRemoteClipboardText"-->
<!--        type="primary"-->
<!--        size="small"-->
<!--        :disabled="props.disabled"-->
<!--      >-->
<!--        显示远程同步的剪贴板信息</n-button-->
<!--      >-->
<!-- </n-space>-->
<!--<n-input-->
<!--      v-if="showRemoteText"-->
<!--      :value="props.remoteText"-->
<!--      type="textarea"-->
<!--      :autosize="size"-->
<!--      readonly-->
<!--      placeholder="远程同步的剪贴板内容"-->
<!--      :disabled="props.disabled"-->
<!--    />-->
<!-- </n-space>-->
<!-- </template>-->
