import Vue from 'vue'
import VueRouter from 'vue-router'
import ElementUI from 'element-ui'
import Guacamole from 'guacamole-common-js'
import VueCookies from 'vue-cookies'
import 'element-ui/lib/theme-chalk/index.css'
import App from './App.vue'
import i18n from './i18n'
import plugins from './plugins'
import router from './router'

import '@/styles/index.scss'

Vue.use(VueRouter)
Vue.use(ElementUI)
Vue.use(Guacamole)
Vue.use(plugins)
Vue.use(VueCookies)
Vue.use(i18n)
Vue.config.productionTip = false

// logger
import VueLogger from 'vuejs-logger'
import loggerOptions from './utils/logger'
Vue.use(VueLogger, loggerOptions)

new Vue({
  i18n,
  router,
  render: h => h(App)
}).$mount('#app')
