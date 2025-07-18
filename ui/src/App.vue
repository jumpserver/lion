<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { BASE_URL, LanguageCode, ThemeCode } from '@/utils/config';
import { useDebounceFn } from '@vueuse/core';
import { alovaInstance } from '@/api';
import type { GlobalThemeOverrides, NLocale } from 'naive-ui';
import { onMounted, ref, nextTick, provide, onUnmounted } from 'vue';
import { createThemeOverrides } from './overrides';
import { darkTheme, dateZhCN, enUS, esAR, jaJP, koKR, ptBR, ruRU, zhCN, zhTW } from 'naive-ui';
const { mergeLocaleMessage } = useI18n();
import { lunaCommunicator } from './utils/lunaBus';
import { LUNA_MESSAGE_TYPE, type LunaMessage } from '@/types/postmessage.type';
const langCodeMap = new Map(
  Object.entries({
    ko: koKR,
    ru: ruRU,
    ja: jaJP,
    es: esAR,
    en: enUS,
    'pt-br': ptBR,
    'zh-hant': zhTW,
    'zh-hans': zhCN,
    'zh-cn': zhCN,
  }),
);
const themeOverrides = ref<GlobalThemeOverrides | null>(null);
const langCode = langCodeMap.get(LanguageCode);
const loaded = ref(false);
const componentsLocale = ref<NLocale | null>(null);

provide('manual-set-theme', (theme: string) => {
  themeOverrides.value = createThemeOverrides(theme as 'default' | 'deepBlue' | 'darkGary');
});

const handleMainThemeChange = useDebounceFn((data: LunaMessage) => {
  console.log('Theme changed to:', data.data);
  const theme = data.data;
  themeOverrides.value = createThemeOverrides(theme as 'default' | 'deepBlue' | 'darkGary');
}, 300);

onMounted(async () => {
  loaded.value = false;
  componentsLocale.value = langCode || enUS;
  themeOverrides.value = createThemeOverrides(ThemeCode as 'default' | 'deepBlue' | 'darkGary');
  try {
    const translations = await alovaInstance
      .Get(`/api/v1/settings/i18n/lion/?lang=${LanguageCode}&flat=0`)
      .then((response) => (response as Response).json());

    for (const [key, value] of Object.entries(translations)) {
      mergeLocaleMessage(key, value);
    }
  } catch (e) {
    throw new Error(`${e}`);
  } finally {
    nextTick(() => {
      loaded.value = true;
    });
  }
  lunaCommunicator.onLuna(LUNA_MESSAGE_TYPE.CHANGE_MAIN_THEME, handleMainThemeChange);
});

onUnmounted(() => {
  lunaCommunicator.offLuna(LUNA_MESSAGE_TYPE.CHANGE_MAIN_THEME);
});
</script>

<template>
  <n-config-provider
    :locale="componentsLocale"
    :theme="darkTheme"
    :date-locale="dateZhCN"
    :theme-overrides="themeOverrides"
    class="flex items-center justify-center h-full w-full overflow-hidden"
  >
    <n-dialog-provider>
      <n-notification-provider placement="bottom-right">
        <n-message-provider>
          <RouterView v-if="loaded" />
        </n-message-provider>
      </n-notification-provider>
    </n-dialog-provider>
  </n-config-provider>
</template>

<style scoped></style>
