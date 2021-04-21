<template>
  <el-row>
    <textarea class="clipboard" :value="value" @input="debounceInput" />
  </el-row>
</template>

<script>
export default {
  name: 'GuacClipboard',
  props: {
    value: {
      type: String,
      default: ''
    }
  },
  methods: {
    debounceInput(event) {
      clearTimeout(this.debounce)
      this.debounce = setTimeout(() => {
        this.$emit('ClipboardChange', event.target.value)
      }, 600)
    }
  }

}
</script>

<style scoped>

.clipboard {
  position: relative;
  border: 1px solid #AAA;
  -moz-border-radius: 0.25em;
  -webkit-border-radius: 0.25em;
  -khtml-border-radius: 0.25em;
  border-radius: 0.25em;
  width: 100%;
  height: 2in;
  white-space: pre;
  font-size: 1em;
  overflow: auto;
  padding: 0.25em;
}

.clipboard div {
  margin: 0;
}

</style>
