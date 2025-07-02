<script lang="ts" setup>
import { type UploadFileInfo } from 'naive-ui';
import { useMessage } from 'naive-ui';

const message = useMessage();
const props = defineProps<{
  fileList: Array<UploadFileInfo>;
}>();

const emit = defineEmits(['remove']);

const handleRemove = (options: any) => {
  console.log('File removed:', options);
  if (options.file.status === 'uploading') {
    console.warn('Cannot remove a file that is currently uploading.');
    message.warning('Cannot remove a file that is currently uploading.');
    return;
  }
  emit('remove', options);
};
</script>

<template>
  <n-upload
    abstract
    :show-preview-button="false"
    :file-list="props.fileList"
    @remove="handleRemove"
  >
    <n-upload-file-list />
  </n-upload>

  <!-- displayUploadingFiles -->
</template>
