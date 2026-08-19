import { createRouter, createWebHistory } from 'vue-router';
import { isAuthenticated } from '@/entities/user';
import { routes } from './routes';

export const router = createRouter({
  history: createWebHistory(),
  routes,
});

router.beforeEach((to) => {
  const authed = isAuthenticated();
  if (to.meta.requiresAuth && !authed) {
    return { name: 'login', query: { redirect: to.fullPath } };
  }
  if (to.meta.guest && authed) {
    return { path: '/projects' };
  }
  if (to.meta.requiresAdmin) {
    const user = JSON.parse(localStorage.getItem('gdh_user') || 'null');
    if (user?.role !== 'USER_ROLE_ADMIN') {
      return { path: '/projects' };
    }
  }
});

export default router;
