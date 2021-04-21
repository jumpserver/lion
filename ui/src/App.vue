<template>
  <div id="app">
    <GuacamoleConnect />
  </div>
</template>

<script>
import GuacamoleConnect from './components/GuacamoleConnect'

export default {
  name: 'App',
  components: {
    GuacamoleConnect
  }
}
</script>

<style>
#app {
  font-family: Avenir, Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  color: #2c3e50;
  margin: 0;
  padding:0;
}
body {
  margin: 0;
}
</style>
