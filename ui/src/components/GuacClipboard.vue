<template>
  <el-input type="textarea" class="clipboard" :rows="10" :value="value" @input="debounceInput" />
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
  -moz-border-radius: 0.25em;
  -webkit-border-radius: 0.25em;
  -khtml-border-radius: 0.25em;
  border-radius: 0.25em;
  white-space: pre;
  font-size: 1em;
  overflow: auto;
  padding-left: 10px;
  height: 100%;
  width: calc(100% - 20px);
}

.clipboard div {
  margin: 0;
}

</style>
