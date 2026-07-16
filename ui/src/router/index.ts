import { createRouter, createWebHashHistory } from 'vue-router';
import ConnectView from '../views/ConnectView.vue';
import { LION_BASE } from '@/utils/base';

const normalizeLegacyPath = () => {
  const basePath = LION_BASE.endsWith('/') ? LION_BASE.slice(0, -1) : LION_BASE;
  const { hash, pathname, search } = window.location;

  if (pathname === basePath) {
    window.history.replaceState(window.history.state, '', `${LION_BASE}${search}${hash}`);
    return;
  }

  if (hash || !basePath || !pathname.startsWith(basePath)) {
    return;
  }

  const routePath = pathname.slice(basePath.length) || '/';
  if (!routePath || routePath === '/') {
    return;
  }

  const normalizedRoutePath = routePath.startsWith('/') ? routePath : `/${routePath}`;
  window.history.replaceState(window.history.state, '', `${LION_BASE}#${normalizedRoutePath}${search}`);
};

normalizeLegacyPath();

const router = createRouter({
  history: createWebHashHistory(LION_BASE),
  routes: [
    {
      path: '/',
      redirect: '/connect/',
    },
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
    },
  ],
});

export default router;
