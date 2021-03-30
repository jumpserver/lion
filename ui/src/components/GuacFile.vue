<template>
  <div>{{ name }}-{{ type }}</div>
</template>

<script>
import {FileType} from '../utils/common'

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
      default: ()=>({}),
    }
  }
}
</script>

<style scoped>

</style>