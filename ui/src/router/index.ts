import { createRouter, createWebHistory } from 'vue-router'
import ConnectView from '../views/ConnectView.vue'


console.log('router init',import.meta.env)

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/connect/',
      name: 'connect',
      component: ConnectView,
    },
    {
      path: '/monitor/',
      name: 'monitor',
      // route level code-splitting
      // this generates a separate chunk (About.[hash].js) for this route
      // which is lazy-loaded when the route is visited.
      component: () => import('../views/MonitorView.vue'),
    },
    {
      path: '/share/:id/',
      name: 'share',
      component: () => import('../views/ShareView.vue'),
    }
  ],
})

export default router
