<template>
  <el-row>
      <i v-if="type==='NORMAL'" class="el-icon-document"></i>
      <i v-else class="el-icon-folder"></i>
      <span> {{ name }}</span>
  </el-row>

</template>

<script>
import { FileType } from '../utils/common'

export default {
  name: 'GuacFile',
  props: {
    mimetype: {
      type: String,
      default: ''
    },
    streamName: {
      type: String,
      default: ''
    },
    type: {
      type: String,
      default: FileType.NORMAL
    },
    name: {
      type: String,
      default: ''
    },
    parent: {
      type: Object,
      default: null
    },
    files: {
      type: Object,
      default: () => ({}),
    }
  },
}
</script>

<style scoped>

</style>