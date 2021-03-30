<template>
  <el-row>
    <textarea class="clipboard" @input="debounceInput" v-bind:value="value"/>
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
  data() {
    return {
      clipboardData: {
        type: 'text/plain',
        data: ''
      }
    }
  },

  methods: {
    getLocalClipboard() {
      if (navigator.clipboard && navigator.clipboard.readText) {
        navigator.clipboard.readText().then(function textRead(text) {
          this.clipboardData.data = text
        })
      }
    },
    debounceInput(event) {
      this.clipboardText = null
      clearTimeout(this.debounce)
      this.debounce = setTimeout(() => {
        this.clipboardText = event.target.value
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