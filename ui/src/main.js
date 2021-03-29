import Vue from 'vue'
import ElementUI from 'element-ui'
import Guacamole from 'guacamole-common-js'
import draggable from 'vuedraggable'
import 'element-ui/lib/theme-chalk/index.css'
import App from './App.vue'
import i18n from './i18n'


Vue.use(ElementUI)
Vue.use(Guacamole)
Vue.use(draggable)
Vue.use(i18n)
Vue.config.productionTip = false

new Vue({
  i18n,
  render: h => h(App),
}).$mount('#app')
