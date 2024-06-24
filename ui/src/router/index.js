import Vue from 'vue'
import Router from 'vue-router'

Vue.use(Router)

export const constantRoutes = []

/**
 * admin
 * the routes that need to be dynamically loaded based on admin roles
 */
export const allRoleRoutes = [
  {
    path: '/connect',
    component: () => import('../components/GuacamoleConnect')
  },
  {
    path: '/monitor',
    component: () => import('../components/GuacamoleMonitor')
  },
  {
    path: '/share/:id/',
    name: 'Share',
    component: () => import('../components/GuacamoleShare')
  }
]

const createRouter = () => new Router({
  mode: 'history', // require service support
  // scrollBehavior: () => ({y: 0}),
  base: '/lion/',
  routes: allRoleRoutes
})

const router = createRouter()

// Detail see: https://github.com/vuejs/vue-router/issues/1234#issuecomment-357941465
export function resetRouter() {
  const newRouter = createRouter()
  router.matcher = newRouter.matcher // reset router
}

export default router
